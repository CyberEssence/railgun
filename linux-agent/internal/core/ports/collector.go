package ports

import (
	"context"
	"linux-agent/internal/core/domain"
)

// Collector порт для сбора данных
type Collector interface {
	Name() string
	Enabled() bool
	Collect(ctx context.Context) (interface{}, error)
}

// SystemCollector порт для системного коллектора
type SystemCollector interface {
	Collector
	CollectSystem(ctx context.Context) (*domain.SystemInfo, error)
}

// NetworkCollector порт для сетевого коллектора
type NetworkCollector interface {
	Collector
	CollectNetwork(ctx context.Context) (*domain.NetworkInfo, error)
}

// ProcessCollector порт для коллектора процессов
type ProcessCollector interface {
	Collector
	CollectProcesses(ctx context.Context) (*domain.ProcessInfo, error)
}

// SecurityCollector порт для коллектора безопасности
type SecurityCollector interface {
	Collector
	CollectSecurity(ctx context.Context) (*domain.SecurityInfo, error)
}
