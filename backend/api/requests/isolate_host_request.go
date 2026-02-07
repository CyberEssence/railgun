package requests

type IsolateHostRequest struct {
	HostID   string `json:"host_id" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
	Duration int    `json:"duration"` // в минутах, 0 = бессрочно
}
