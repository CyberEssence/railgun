package persistence

import (
	"context"
	"database/sql"
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

// GetUserByID получает пользователя по ID (UUID)
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByUsername получает пользователя по username
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.NewSelect().
		Model(&user).
		Where("username = ?", username).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetUserByEmail получает пользователя по email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.NewSelect().
		Model(&user).
		Where("email = ?", email).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// CreateUser создает нового пользователя
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	// Проверяем существование пользователя с таким email
	existingUser, err := r.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("user with this email already exists")
	}

	// Проверяем существование пользователя с таким username
	existingUser, err = r.GetUserByUsername(ctx, user.Username)
	if err != nil {
		return fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("user with this username already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	// Устанавливаем значения по умолчанию
	if user.TOTPBackupCodes == nil {
		user.TOTPBackupCodes = []string{} // Пустой слайс, а не строка "[]"
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	// Вставляем пользователя
	_, err = r.db.NewInsert().
		Model(user).
		Exec(ctx)

	return err
}

// UpdateUser обновляет данные пользователя
func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	// Если пароль изменился, хешируем его
	if len(user.PasswordHash) > 0 && !isHashedPassword(user.PasswordHash) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	user.UpdatedAt = time.Now()

	_, err := r.db.NewUpdate().
		Model(user).
		Where("id = ?", user.ID).
		Exec(ctx)

	return err
}

// ValidateCredentials проверяет учетные данные пользователя
func (r *UserRepository) ValidateCredentials(ctx context.Context, username, password string) (*models.User, error) {
	user, err := r.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Обновляем время последнего входа
	now := time.Now()
	user.LastLogin = now
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

// SaveTwoFAToken сохраняет 2FA токен
func (r *UserRepository) SaveTwoFAToken(ctx context.Context, token *models.TwoFAToken) error {
	// Проверяем существование таблицы
	_, err := r.db.NewCreateTable().
		Model((*models.TwoFAToken)(nil)).
		IfNotExists().
		ForeignKey(`(user_id) REFERENCES users(id) ON DELETE CASCADE`).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure table exists: %w", err)
	}

	// Сохраняем токен
	_, err = r.db.NewInsert().Model(token).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert token: %w", err)
	}
	return nil
}

// GetTwoFAToken получает 2FA токен
func (r *UserRepository) GetTwoFAToken(ctx context.Context, tokenHash string, userID string) (*models.TwoFAToken, error) {
	var token models.TwoFAToken
	err := r.db.NewSelect().
		Model(&token).
		Where("token_hash = ?", tokenHash).
		Where("user_id = ?", userID).
		Where("used = ?", false).
		Where("expires_at > ?", time.Now()).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Printf("Error querying token: %v", err)
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return &token, nil
}

// MarkTokenAsUsed отмечает токен как использованный
func (r *UserRepository) MarkTokenAsUsed(ctx context.Context, tokenID int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.TwoFAToken)(nil)).
		Set("used = ?", true).
		Where("id = ?", tokenID).
		Exec(ctx)

	return err
}

// Enable2FA включает 2FA для пользователя
func (r *UserRepository) Enable2FA(ctx context.Context, userID string, secret string, backupCodes []string) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("totp_secret = ?", secret).
		Set("totp_enabled = ?", true).
		Set("totp_backup_codes = ?", backupCodes). // Bun автоматически сериализует в JSON
		Where("id = ?", userID).
		Exec(ctx)

	return err
}

// Disable2FA отключает 2FA для пользователя
func (r *UserRepository) Disable2FA(ctx context.Context, userID string) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("totp_secret = ?", "").
		Set("totp_enabled = ?", false).
		Set("totp_backup_codes = ?", []string{}). // Пустой массив
		Where("id = ?", userID).
		Exec(ctx)

	return err
}

// GetTOTPSecret получает TOTP секрет пользователя
func (r *UserRepository) GetTOTPSecret(ctx context.Context, userID string) (string, error) {
	var user models.User
	err := r.db.NewSelect().
		Model(&user).
		Column("totp_secret").
		Where("id = ?", userID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get TOTP secret: %w", err)
	}

	return user.TOTPSecret, nil
}

// Вспомогательная функция для проверки, является ли строка хешированным паролем
func isHashedPassword(password string) bool {
	// Простая эвристика: bcrypt хеши начинаются с $2a$, $2b$ или $2y$
	return len(password) > 4 && (password[:4] == "$2a$" || password[:4] == "$2b$" || password[:4] == "$2y$")
}
