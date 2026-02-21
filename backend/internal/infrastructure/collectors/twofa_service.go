package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"railgun-core/internal/config"
	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

type TwoFAService struct {
	userRepo domain.UserRepository
	encKey   []byte
	config   *config.Config
}

func NewTwoFAService(userRepo domain.UserRepository, cfg *config.Config) *TwoFAService {
	// Получаем или генерируем ключ шифрования
	encKey := getEncryptionKey(cfg)

	log.Printf("TwoFA service initialized with key size: %d bytes", len(encKey))

	return &TwoFAService{
		userRepo: userRepo,
		encKey:   encKey,
		config:   cfg,
	}
}

func (s *TwoFAService) GenerateToken(ctx context.Context, userID string) (string, error) {
	// Проверка существования пользователя
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Генерируем случайный токен
	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Кодируем токен в base64
	token := base64.StdEncoding.EncodeToString(tokenBytes)

	// Вычисляем хеш токена для хранения в БД
	tokenHash := s.hashToken(token)

	// Создаем запись о токене
	twoFAToken := models.TwoFAToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	// Сохраняем токен в БД
	err = s.userRepo.SaveTwoFAToken(ctx, twoFAToken)
	if err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	return token, nil
}

// VerifySetup проверяет первый код для подтверждения настройки
func (s *TwoFAService) VerifySetup(ctx context.Context, token string, userID string) (bool, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return false, err
	}

	// Расшифровываем секрет
	decryptedSecret, err := s.decryptSecret(user.TOTPSecret)
	if err != nil {
		return false, err
	}

	// Проверяем код с расширенным окном (для настройки)
	valid, err := totp.ValidateCustom(
		token,
		decryptedSecret,
		time.Now().UTC(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      2, // Разрешаем ±2 интервала для настройки
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)

	if valid {
		log.Printf("2FA setup verified successfully for user %s", userID)
	}

	return valid, err
}

func (s *TwoFAService) Validate2FAToken(ctx context.Context, token string, userID string) (bool, error) {
	// Вычисляем хеш токена
	tokenHash := s.hashToken(token)

	// Получаем токен из БД
	twoFAToken, err := s.userRepo.GetTwoFAToken(ctx, tokenHash, userID)
	if err != nil {
		// Добавьте логирование для отладки
		log.Printf("Error getting token: %v", err)
		return false, fmt.Errorf("invalid token")
	}

	// Проверяем, не истек ли срок действия токена
	if time.Now().After(twoFAToken.ExpiresAt) {
		return false, fmt.Errorf("token expired")
	}

	// Проверяем, не был ли токен уже использован
	if twoFAToken.Used {
		return false, fmt.Errorf("token already used")
	}

	// Помечаем токен как использованный
	err = s.userRepo.MarkTokenAsUsed(ctx, twoFAToken.ID)
	if err != nil {
		return false, fmt.Errorf("failed to mark token as used: %w", err)
	}

	return true, nil
}

// Вспомогательный метод для хеширования токена
func (s *TwoFAService) hashToken(token string) string {
	h := hmac.New(sha256.New, []byte("twofa-secret-key"))
	h.Write([]byte(token))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type TwoFAConfig struct {
	Issuer        string
	EncryptionKey string
}

// Enable2FA включает 2FA для пользователя
func (s *TwoFAService) Enable2FA(ctx context.Context, userID string, username string) (map[string]interface{}, error) {
	log.Printf("Enabling 2FA for user %s (%s)", userID, username)

	// Используйте правильного issuer
	issuer := s.config.JWTConfig.Issuer
	if issuer == "" {
		issuer = "Railgun SIEM" // значение по умолчанию
	}

	// Генерация TOTP ключа
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
		Period:      30,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
		Secret:      nil,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %v", err)
	}

	secret := key.Secret()
	log.Printf("Generated TOTP secret for user %s", userID)

	// Шифрование секрета
	encryptedSecret, err := s.encryptSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %v", err)
	}

	// Генерация резервных кодов
	backupCodes := s.generateBackupCodes(10)

	// Сохранение в БД
	err = s.userRepo.Enable2FA(ctx, userID, encryptedSecret, backupCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to save 2FA settings: %v", err)
	}

	// Генерация QR-кода
	qrCode, err := s.generateQRCode(key)
	if err != nil {
		log.Printf("Warning: failed to generate QR code: %v", err)
	}

	return map[string]interface{}{
		"secret":       secret, // Показываем только один раз!
		"url":          key.URL(),
		"qr_code":      qrCode,
		"backup_codes": backupCodes,
		"message":      "Сканируйте QR-код в Google Authenticator или Яндекс Ключ. Сохраните резервные коды!",
		"warning":      "Секрет показывается только один раз. Сохраните его в надежном месте.",
	}, nil
}

// Validate2FAToken проверяет TOTP код
func (s *TwoFAService) ValidateTOTPToken(ctx context.Context, token string, userID string) (bool, error) {
	// Получаем пользователя
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("user not found: %v", err)
	}

	// Проверяем, включена ли 2FA
	if !user.TOTPEnabled || user.TOTPSecret == "" {
		return false, fmt.Errorf("2FA is not enabled for this user")
	}

	// Расшифровываем секрет
	decryptedSecret, err := s.decryptSecret(user.TOTPSecret)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt secret: %v", err)
	}

	// Проверяем TOTP код
	valid := totp.Validate(token, decryptedSecret)

	// Если TOTP не прошел, проверяем резервные коды
	if !valid && user.TOTPBackupCodes != "" {
		if s.validateBackupCode(user.TOTPBackupCodes, token) {
			valid = true
			// Удаляем использованный резервный код
			go s.removeUsedBackupCode(ctx, userID, token)
		}
	}

	return valid, nil
}

// Disable2FA отключает 2FA для пользователя
func (s *TwoFAService) Disable2FA(ctx context.Context, userID string) error {
	return s.userRepo.Disable2FA(ctx, userID)
}

// GenerateNewBackupCodes генерирует новые резервные коды
func (s *TwoFAService) GenerateNewBackupCodes(ctx context.Context, userID string) ([]string, error) {
	backupCodes := s.generateBackupCodes(10)

	// Преобразуем в JSON
	backupCodesJSON, err := json.Marshal(backupCodes)
	if err != nil {
		return nil, err
	}

	// Обновляем в БД
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.TOTPBackupCodes = string(backupCodesJSON)
	err = s.userRepo.UpdateUser(ctx, user)

	return backupCodes, err
}

// generateBackupCodes генерирует резервные коды
func (s *TwoFAService) generateBackupCodes(count int) []string {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		// Генерация 10 символов
		b := make([]byte, 8)
		rand.Read(b)
		code := base64.StdEncoding.EncodeToString(b)

		// Убираем спецсимволы, делаем uppercase
		code = strings.ToUpper(strings.NewReplacer(
			"+", "",
			"/", "",
			"=", "",
		).Replace(code))

		// Форматируем: XXXX-XXXX-XXXX
		if len(code) >= 12 {
			formatted := fmt.Sprintf("%s-%s-%s",
				code[0:4], code[4:8], code[8:12])
			codes[i] = formatted
		} else {
			codes[i] = code
		}
	}

	return codes
}

// generateQRCode создает QR-код в base64
func (s *TwoFAService) generateQRCode(key *otp.Key) (string, error) {
	qrCode, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(qrCode), nil
}

// validateBackupCode проверяет резервный код
func (s *TwoFAService) validateBackupCode(backupCodesJSON, code string) bool {
	var backupCodes []string
	err := json.Unmarshal([]byte(backupCodesJSON), &backupCodes)
	if err != nil {
		return false
	}

	for _, backupCode := range backupCodes {
		if backupCode == code {
			return true
		}
	}

	return false
}

// removeUsedBackupCode удаляет использованный резервный код
func (s *TwoFAService) removeUsedBackupCode(ctx context.Context, userID string, usedCode string) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user for backup code removal: %v", err)
		return
	}

	var backupCodes []string
	err = json.Unmarshal([]byte(user.TOTPBackupCodes), &backupCodes)
	if err != nil {
		log.Printf("Failed to unmarshal backup codes: %v", err)
		return
	}

	// Удаляем использованный код
	newBackupCodes := []string{}
	for _, code := range backupCodes {
		if code != usedCode {
			newBackupCodes = append(newBackupCodes, code)
		}
	}

	// Сохраняем обновленный список
	newBackupCodesJSON, err := json.Marshal(newBackupCodes)
	if err != nil {
		log.Printf("Failed to marshal new backup codes: %v", err)
		return
	}

	user.TOTPBackupCodes = string(newBackupCodesJSON)
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		log.Printf("Failed to update user backup codes: %v", err)
	}
}

// encryptSecret шифрует TOTP секрет
func (s *TwoFAService) encryptSecret(secret string) (string, error) {
	if len(s.encKey) == 0 {
		return secret, nil // Не шифруем если нет ключа
	}

	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptSecret расшифровывает TOTP секрет
func (s *TwoFAService) decryptSecret(encrypted string) (string, error) {
	if len(s.encKey) == 0 {
		return encrypted, nil // Не расшифровываем если нет ключа
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// getEncryptionKey возвращает ключ шифрования
func getEncryptionKey(cfg *config.Config) []byte {
	// 1. Из конфигурации
	if cfg.JWTConfig.EncryptionKey != "" {
		key := []byte(cfg.JWTConfig.EncryptionKey)
		if isValidKeySize(len(key)) {
			return key
		}
	}

	return nil
}

func isValidKeySize(size int) bool {
	return size == 16 || size == 24 || size == 32
}
