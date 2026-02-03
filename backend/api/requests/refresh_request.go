package requests

// RefreshRequest структура для запроса на обновление токена
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
