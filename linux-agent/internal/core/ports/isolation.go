package ports

import (
	"linux-agent/internal/core/domain"
)

// TaskFetcher порт для получения задач от SIEM сервера
type TaskFetcher interface {
	FetchTask(hostID string) (*domain.IsolationTask, error)
	ReportResult(report *domain.TaskReport) error
}

// IsolationExecutor порт для воздействия на ОС (iptables/firewall)
type IsolationExecutor interface {
	Isolate(serverIP string) error
	Unisolate() error
}
