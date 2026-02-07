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

	"railgun-core/internal/config"
	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"

	"railgun-core/api/requests"
	"railgun-core/api/responses"

	repository "railgun-core/internal/domain/repository"
)

type AuthHandler struct {
	config       *config.Config
	twoFAService domain.TwoFAService
	userRepo     *repository.UserRepository
}

func NewAuthHandler(config *config.Config, twoFAService domain.TwoFAService, userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		config:       config,
		twoFAService: twoFAService,
		userRepo:     userRepo,
	}
}

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентификация пользователя по имени и паролю. Может вернуть токены сразу или потребовать 2FA.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "Данные пользователя"
// @Success      200  {object}  models.TokenResponse "Если 2FA отключена"
// @Success      200  {object}  models.LoginResponse "Если требуется 2FA"
// @Failure      401  {object}  map[string]string "Неверные учетные данные"
// @Failure      500  {object}  map[string]string "Ошибка генерации токенов"
// @Router       /auth/login [post]
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

// Verify2FA godoc
// @Summary      Проверка 2FA кода
// @Description  Проверяет временный код подтверждения и выдает финальные JWT токены
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body requests.Verify2FARequest true "ID пользователя и код из приложения"
// @Success      200  {object}  models.TokenResponse "JWT токены"
// @Failure      400  {object}  map[string]interface{} "Неверный формат запроса"
// @Failure      401  {object}  map[string]string "Неверный 2FA код"
// @Router       /auth/verify-2fa [post]
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	// Логируем входящий запрос
	body, _ := c.GetRawData()
	log.Printf("Incoming 2FA verification request: %s", string(body))
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // Возвращаем body для повторного чтения

	var req requests.Verify2FARequest

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

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Создает новый аккаунт. Пароль должен быть не менее 8 символов.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body requests.RegisterRequest true "Данные для регистрации"
// @Success      201  {object}  map[string]interface{} "Пользователь успешно создан"
// @Failure      400  {object}  map[string]string "Ошибка валидации или слабый пароль"
// @Failure      409  {object}  map[string]string "Имя пользователя уже занято"
// @Failure      500  {object}  map[string]string
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req requests.RegisterRequest
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

// RefreshToken godoc
// @Summary      Обновить JWT токены
// @Description  Принимает Refresh Token и выдает новую пару Access/Refresh токенов
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body requests.RefreshRequest true "Refresh Token"
// @Success      200  {object}  responses.TokenResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string "Невалидный Refresh Token"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req requests.RefreshRequest
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

	c.JSON(http.StatusOK, responses.TokenResponse{
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
