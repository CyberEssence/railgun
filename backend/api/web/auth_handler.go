package api

import (
	"fmt"
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

	// Если у пользователя включена 2FA, возвращаем специальный ответ
	if user.TOTPEnabled {
		c.JSON(http.StatusOK, gin.H{
			"requires_2fa": true,
			"user_id":      user.ID,
			"message":      "Please enter your 2FA code",
		})
		return
	}

	// Если 2FA не включена - сразу выдаем токены
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
	var req requests.Verify2FARequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Verifying TOTP code for user %s", req.UserID)

	valid, err := h.twoFAService.ValidateTOTPToken(c.Request.Context(), req.Token, req.UserID)
	if err != nil || !valid {
		log.Printf("TOTP validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid 2FA code",
			"details": err.Error(), // Добавляем детали для отладки
		})
		return
	}

	log.Printf("TOTP validation successful for user %d", req.UserID)

	// Генерация JWT токенов
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

	// Проверка сложности пароля
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		return
	}

	// Создание нового пользователя
	newUser := &models.User{
		Username:        req.Username,
		Email:           req.Email,
		PasswordHash:    req.Password, // Будет захеширован в репозитории
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		TOTPEnabled:     false,
		TOTPSecret:      "",
		TOTPBackupCodes: []string{},
	}

	// Сохранение пользователя в базе данных
	err := h.userRepo.CreateUser(c.Request.Context(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	// Получаем созданного пользователя для ответа
	createdUser, _ := h.userRepo.GetUserByUsername(c.Request.Context(), newUser.Username)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":         createdUser.ID,
			"username":   createdUser.Username,
			"email":      createdUser.Email,
			"created_at": createdUser.CreatedAt,
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

	userID, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	// Генерируем новые токены
	accessToken, refreshToken, expiresIn, err := h.generateTokens(userID)
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
func (h *AuthHandler) generateTokens(userID string) (string, string, int, error) {
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
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
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

// Enable2FA godoc
// @Summary      Включение 2FA
// @Description  Генерирует секрет и QR-код для настройки Google Authenticator/Yandex Key
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.Enable2FARequest true "ID пользователя"
// @Success      200  {object}  map[string]interface{} "Данные для настройки 2FA"
// @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// @Router       /api/auth/2fa/enable [post]
func (h *AuthHandler) Enable2FA(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	result, err := h.twoFAService.Enable2FA(c.Request.Context(), userID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to enable 2FA",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Verify2FASetup godoc
// @Summary      Подтверждение настройки 2FA
// @Description  Проверяет первый код из приложения для подтверждения корректной настройки
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.Verify2FASetupRequest true "ID пользователя и код"
// @Success      200  {object}  map[string]string "Статус подтверждения"
// @Failure      400  {object}  map[string]string "Неверный код"
// @Router       /api/auth/2fa/verify-setup [post]
func (h *AuthHandler) Verify2FASetup(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token required"})
		return
	}

	userID := c.GetString("user_id")
	valid, err := h.twoFAService.VerifySetup(c.Request.Context(), req.Token, userID)

	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid verification code",
			"message": "Please check the code from your authenticator app",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "2FA has been successfully enabled",
	})
}

// Get2FAStatus godoc
// @Summary      Получение статуса 2FA
// @Description  Проверяет, включена ли 2FA для пользователя
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{} "Статус 2FA"
// @Router       /api/auth/2fa/status [get]
func (h *AuthHandler) Get2FAStatus(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID := userIDValue.(string)
	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	hasBackupCodes := len(user.TOTPBackupCodes) > 0

	c.JSON(http.StatusOK, gin.H{
		"enabled":            user.TOTPEnabled,
		"has_backup_codes":   hasBackupCodes,
		"backup_codes_count": len(user.TOTPBackupCodes),
	})
}

// Disable2FA godoc
// @Summary      Отключение 2FA
// @Description  Отключает двухфакторную аутентификацию для пользователя
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.Disable2FARequest true "ID пользователя и подтверждение"
// @Success      200  {object}  map[string]string "Статус отключения"
// @Failure      400  {object}  map[string]string "Неверный запрос"
// @Router       /api/auth/2fa/disable [post]
func (h *AuthHandler) Disable2FA(c *gin.Context) {
	// Берем user_id из JWT
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID := userIDValue.(string)

	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Проверяем пароль
	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	err = h.twoFAService.Disable2FA(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Failed to disable 2FA: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "2FA has been disabled",
	})
}

func (h *AuthHandler) GenerateNewBackupCodes(c *gin.Context) {
	userID := c.GetString("user_id")

	backupCodes, err := h.twoFAService.GenerateNewBackupCodes(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate backup codes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"backup_codes": backupCodes,
		"message":      "Save these codes in a secure place!",
	})
}

// Аутентификация агентов в вашем SIEM
func (h *AuthHandler) AgentAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверка токена агента
		token := c.GetHeader("X-Agent-Token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Agent token required"})
			return
		}

		// Валидация токена (можно использовать JWT или статические токены)
		agentID, err := validateAgentToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid agent token"})
			return
		}

		c.Set("agent_id", agentID)
		c.Next()
	}
}

// Генерация токенов для агентов
func GenerateAgentToken(hostID string) (string, error) {
	claims := jwt.MapClaims{
		"host_id": hostID,
		"type":    "agent",
		"exp":     time.Now().Add(24 * 365 * time.Hour).Unix(), // 1 год
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-agent-secret"))
}

// validateAgentToken проверяет токен агента
func validateAgentToken(token string) (string, error) {
	// Для начала можно использовать простую проверку JWT
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// В реальном приложении используйте секрет из конфигурации
		return []byte("your-agent-secret"), nil
	})

	if err != nil || !parsedToken.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	hostID, ok := claims["host_id"].(string)
	if !ok {
		return "", fmt.Errorf("host_id not found in token")
	}

	return hostID, nil
}
