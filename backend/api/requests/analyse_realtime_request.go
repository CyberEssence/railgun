package requests

type AnalyzeRealtimeRequest struct {
	Data     []string `json:"data"`
	DataType string   `json:"dataType"`
	HostID   string   `json:"hostId"`
}
