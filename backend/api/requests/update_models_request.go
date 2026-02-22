package requests

type UpdateModelsRequest struct {
	ModelIDs []string `json:"modelIds" binding:"required"`
}
