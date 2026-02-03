package requests

// Verify2FARequest структура для запроса на проверку 2FA
type Verify2FARequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Token  string `json:"token" binding:"required"`
}
