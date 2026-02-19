package requests

// Verify2FARequest структура для запроса на проверку 2FA
type Verify2FARequest struct {
	UserID string `json:"user_id" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

// LoginRequest структура для запроса на вход
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	TwoFAToken string `json:"two_fa_token,omitempty"`
}

// RegisterRequest структура для запроса на регистрацию
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Verify2FASetupRequest
type Verify2FASetupRequest struct {
	Token string `json:"token" binding:"required"` // user_id из JWT
}

// Disable2FARequest
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"` // user_id из JWT
}
