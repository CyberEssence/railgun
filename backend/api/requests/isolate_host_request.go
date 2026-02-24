package requests

type IsolateHostRequest struct {
	HostID   string `json:"hostId" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
	Duration int    `json:"duration"`
}
