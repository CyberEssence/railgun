package domain

// IsolationTask задача на изоляцию от сервера
type IsolationTask struct {
	ID     int64  `json:"id"`
	Action string `json:"action"` // "isolate", "unisolate"
}

// TaskReport отчет о выполнении задачи
type TaskReport struct {
	TaskID int64  `json:"task_id"`
	Status string `json:"status"` // "completed", "failed"
	Output string `json:"output"`
}
