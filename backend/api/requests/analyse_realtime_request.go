package requests

type AnalyzeRealtimeRequest struct {
	Data     []string `json:"data" binding:"required"`
	DataType string   `json:"data_type" binding:"required"`
	HostID   string   `json:"host_id" binding:"required"`
}
