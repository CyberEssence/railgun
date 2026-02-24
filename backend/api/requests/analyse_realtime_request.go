package requests

type AnalyzeRealtimeRequest struct {
	Data     []string `json:"data" binding:"required"`
	DataType string   `json:"dataType" binding:"required"`
	HostID   string   `json:"hostId" binding:"required"`
}
