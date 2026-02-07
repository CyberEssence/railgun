package requests

type UpdateModelsRequest struct {
	ModelIDs []string `json:"model_ids" binding:"required"`
}
