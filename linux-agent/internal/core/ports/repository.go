package ports

import (
	"context"
)

// MetricRepository порт для хранения метрик
type MetricRepository interface {
	Save(ctx context.Context, index string, data interface{}) error
	SaveBatch(ctx context.Context, index string, data []interface{}) error
	Search(ctx context.Context, index string, query map[string]interface{}) ([]interface{}, error)
}

// HealthChecker порт для проверки здоровья
type HealthChecker interface {
	Ping(ctx context.Context) error
}
