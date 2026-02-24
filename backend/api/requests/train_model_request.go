package requests

type TrainModelRequest struct {
	ModelID     string `json:"modelId" binding:"required"`
	DatasetPath string `json:"datasetPath" binding:"required"`
	Epochs      int    `json:"epochs" binding:"required,min=1"`
}
