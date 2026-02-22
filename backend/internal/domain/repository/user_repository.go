package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"

	"railgun-core/internal/domain/models"
)

type UserRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Where("username = ?", username).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	// Check if user with this email already exists
	var existingUser models.User
	err := r.db.NewSelect().Model(&existingUser).
		Where("email = ?", user.Email).
		Limit(1).
		Scan(ctx)

	if err == nil {
		return fmt.Errorf("user with this email already exists")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check for existing user: %w", err)
	}

	// Check if user with this username already exists
	err = r.db.NewSelect().Model(&existingUser).
		Where("username = ?", user.Username).
		Limit(1).
		Scan(ctx)

	if err == nil {
		return fmt.Errorf("user with this username already exists")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check for existing user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	// Устанавливаем значения по умолчанию для 2FA полей
	if user.TOTPSecret == "" {
		user.TOTPSecret = ""
	}
	if !user.TOTPEnabled {
		user.TOTPEnabled = false
	}
	if len(user.TOTPBackupCodes) == 0 || string(user.TOTPBackupCodes) == "null" {
		user.TOTPBackupCodes = json.RawMessage("[]")
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	_, err = r.db.NewInsert().
		Model(user).
		Returning("id").
		Exec(ctx)

	return err
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	// Если пароль изменился, хешируем его
	if len(user.PasswordHash) > 0 && !isHashedPassword(user.PasswordHash) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	_, err := r.db.NewUpdate().
		Model(user).
		Where("id = ?", user.ID).
		Exec(ctx)

	return err
}

func (r *UserRepository) ValidateCredentials(ctx context.Context, username, password string) (*models.User, error) {
	user, err := r.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Обновляем время последнего входа
	user.LastLogin = time.Now()
	_, err = r.db.NewUpdate().
		Model(user).
		Set("last_login = ?", user.LastLogin).
		Where("id = ?", user.ID).
		Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return user, nil
}

func (r *UserRepository) SaveTwoFAToken(ctx context.Context, token models.TwoFAToken) error {
	// Проверяем существование таблицы
	_, err := r.db.NewCreateTable().
		Model((*models.TwoFAToken)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure table exists: %w", err)
	}

	// Сохраняем токен
	_, err = r.db.NewInsert().Model(&token).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert token: %w", err)
	}
	return nil
}

func (r *UserRepository) SaveTOTPSecret(ctx context.Context, userID string, secret string) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("totp_secret = ?", secret).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update totp_secret: %v", err)
	}

	log.Printf("TOTP secret saved for user %s", userID)
	return err
}

func (r *UserRepository) GetTwoFAToken(ctx context.Context, tokenHash string, userID string) (*models.TwoFAToken, error) {
	token := new(models.TwoFAToken)
	err := r.db.NewSelect().
		Model(token).
		Where("token_hash = ?", tokenHash).
		Where("user_id = ?", userID).
		Where("used = false").
		Where("expires_at > ?", time.Now()).
		Scan(ctx)

	if err != nil {
		// Добавьте логирование для отладки
		log.Printf("Error querying token: %v", err)
		return nil, err
	}

	return token, nil
}

func (r *UserRepository) MarkTokenAsUsed(ctx context.Context, tokenID int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.TwoFAToken)(nil)). // Используем правильную модель
		Set("used = true").
		Where("id = ?", tokenID).
		Exec(ctx)

	return err
}

func (r *UserRepository) Enable2FA(ctx context.Context, userID string, encryptedSecret string, backupCodes []string) error {
	log.Printf("Repository: Enabling 2FA for user %s", userID)

	// Конвертируем backupCodes в JSON для сохранения
	backupCodesJSON, err := json.Marshal(backupCodes)
	if err != nil {
		return fmt.Errorf("failed to marshal backup codes: %v", err)
	}

	// Используем bun Update
	_, err = r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("totp_secret = ?", encryptedSecret).
		Set("totp_backup_codes = ?", string(backupCodesJSON)).
		Set("totp_enabled = ?", true).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		log.Printf("Error updating user 2FA data: %v", err)
		return fmt.Errorf("failed to update user 2FA data: %v", err)
	}

	log.Printf("Repository: 2FA data saved successfully for user %s", userID)
	return nil
}

func (r *UserRepository) Disable2FA(ctx context.Context, userID string) error {
	_, err := r.db.NewUpdate().
		Model(&models.User{}).
		Set("totp_secret = ?", "").
		Set("totp_enabled = ?", false).
		Set("totp_backup_codes = ?", "[]").
		Where("id = ?", userID).
		Exec(ctx)

	return err
}

func (r *UserRepository) GetTOTPSecret(ctx context.Context, userID string) (string, error) {
	var user models.User
	err := r.db.NewSelect().
		Model(&user).
		Column("totp_secret").
		Where("id = ?", userID).
		Scan(ctx)

	if err != nil {
		return "", err
	}

	return user.TOTPSecret, nil
}

func (r *UserRepository) EnableTOTP(ctx context.Context, userID string) error {
	log.Printf("Enabling TOTP for user %s", userID)

	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("totp_enabled = ?", true).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	return err
}

// Вспомогательная функция для проверки, является ли строка хешированным паролем
func isHashedPassword(password string) bool {
	// Простая эвристика: bcrypt хеши начинаются с $2a$, $2b$ или $2y$
	return len(password) > 4 && (password[:4] == "$2a$" || password[:4] == "$2b$" || password[:4] == "$2y$")
}
