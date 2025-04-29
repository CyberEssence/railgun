package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"railgun-core/internal/domain"
	"railgun-core/internal/models"
)

type TwoFAService struct {
	userRepo domain.UserRepository
}

func NewTwoFAService(userRepo domain.UserRepository) domain.TwoFAService {
	return &TwoFAService{
		userRepo: userRepo,
	}
}

func (s *TwoFAService) GenerateToken(ctx context.Context, userID int64) (string, error) {
	// Генерируем случайный токен
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
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

func (s *TwoFAService) Validate2FAToken(ctx context.Context, token string, userID int64) (bool, error) {
	// Вычисляем хеш токена
	tokenHash := s.hashToken(token)

	// Получаем токен из БД
	twoFAToken, err := s.userRepo.GetTwoFAToken(ctx, tokenHash, userID)
	if err != nil {
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
