package requests

type TrainModelRequest struct {
	ModelID     string `json:"model_id" binding:"required"`
	DatasetPath string `json:"dataset_path" binding:"required"`
	Epochs      int    `json:"epochs" binding:"required,min=1"`
}
