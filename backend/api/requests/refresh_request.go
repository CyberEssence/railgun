package requests

// RefreshRequest структура для запроса на обновление токена
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
