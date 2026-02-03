package requests

// LoginRequest структура для запроса на вход
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	TwoFAToken string `json:"two_fa_token,omitempty"`
}
