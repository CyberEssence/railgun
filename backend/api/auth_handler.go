package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"railgun-core/internal/domain"
	"railgun-core/internal/infrastructure/persistence"
	"railgun-core/internal/models"
)

type AuthHandler struct {
	config       *domain.Config
	twoFAService domain.TwoFAService
	userRepo     *persistence.UserRepository
}

func NewAuthHandler(config *domain.Config, twoFAService domain.TwoFAService, userRepo *persistence.UserRepository) *AuthHandler {
	return &AuthHandler{
		config:       config,
		twoFAService: twoFAService,
		userRepo:     userRepo,
	}
}

// LoginRequest структура для запроса на вход
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	TwoFAToken string `json:"two_fa_token,omitempty"`
}

// LoginResponse структура для ответа на вход
type LoginResponse struct {
	RequiresTwoFA bool   `json:"requires_2fa"`
	UserID        int64  `json:"user_id,omitempty"`
	Message       string `json:"message"`
}

// Verify2FARequest структура для запроса на проверку 2FA
type Verify2FARequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

// TokenResponse структура для ответа с токеном
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RegisterRequest структура для запроса на регистрацию
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RefreshRequest структура для запроса на обновление токена
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// api/auth_handler.go
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := h.userRepo.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Если включена 2FA
	if h.twoFAService != nil {
		token, err := h.twoFAService.GenerateToken(c.Request.Context(), user.ID)
		if err != nil {
			// Логируем ошибку для отладки
			log.Printf("2FA token generation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate 2FA token"})
			return
		}

		c.JSON(http.StatusOK, models.LoginResponse{
			RequiresTwoFA: true,
			UserID:        user.ID,
			Message:       "2FA token required",
			TwoFAToken:    token,
		})
		return
	}

	// Генерация JWT токенов если 2FA не используется
	accessToken, refreshToken, expiresIn, err := h.generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	})
}

// Verify2FA проверяет 2FA токен и выдает JWT токен
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	// Логируем входящий запрос
	body, _ := c.GetRawData()
	log.Printf("Incoming 2FA verification request: %s", string(body))
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // Возвращаем body для повторного чтения

	var req struct {
		UserID int64  `json:"user_id" binding:"required"`
		Token  string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": gin.H{
				"expected_fields": []string{"user_id (number)", "token (string)"},
				"error":           err.Error(),
			},
		})
		return
	}

	log.Printf("Verifying 2FA token for user %d", req.UserID)

	valid, err := h.twoFAService.Validate2FAToken(c, req.Token, req.UserID)
	if err != nil || !valid {
		log.Printf("Token validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid 2FA token"})
		return
	}

	accessToken, refreshToken, expiresIn, err := h.generateTokens(req.UserID)
	if err != nil {
		log.Printf("Token generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	})
}

// Register обрабатывает запрос на регистрацию
// Register обрабатывает запрос на регистрацию
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка, что пользователь с таким именем не существует
	existingUser, err := h.userRepo.GetUserByUsername(c.Request.Context(), req.Username)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Проверка сложности пароля
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		return
	}

	// Создание нового пользователя
	newUser := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password, // UserRepository выполнит хеширование
		CreatedAt:    time.Now(),
	}

	// Сохранение пользователя в базе данных
	err = h.userRepo.CreateUser(c.Request.Context(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	// Опционально: создание и отправка токена подтверждения email
	// h.sendVerificationEmail(newUser.Email, verificationToken)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"username": newUser.Username,
			"email":    newUser.Email,
		},
	})
}

// RefreshToken обновляет JWT токен
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.Auth.Secret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Получаем ID пользователя из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token claims"})
		return
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	// Генерируем новые токены
	accessToken, refreshToken, expiresIn, err := h.generateTokens(int64(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	})
}

// AuthMiddleware middleware для проверки JWT токена
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Извлекаем токен из заголовка
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		// Проверяем токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.config.Auth.Secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Получаем ID пользователя из токена
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token claims"})
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
			return
		}

		// Сохраняем ID пользователя в контексте
		c.Set("user_id", int64(userID))
		c.Next()
	}
}

// generateTokens генерирует JWT токены
func (h *AuthHandler) generateTokens(userID int64) (string, string, int, error) {
	// Создаем access token
	accessTokenTTL := time.Duration(h.config.Auth.TokenTTL) * time.Second
	accessTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(accessTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
		"iss":     h.config.Auth.IssuerURL,
		"type":    "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(h.config.Auth.Secret))
	if err != nil {
		return "", "", 0, err
	}

	// Создаем refresh token (с более длительным сроком действия)
	refreshTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 дней
		"iat":     time.Now().Unix(),
		"iss":     h.config.Auth.IssuerURL,
		"type":    "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(h.config.Auth.Secret))
	if err != nil {
		return "", "", 0, err
	}

	return accessTokenString, refreshTokenString, int(accessTokenTTL.Seconds()), nil
}
