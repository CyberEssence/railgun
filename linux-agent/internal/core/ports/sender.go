package ports

import (
	"context"
	"linux-agent/internal/core/domain"
)

// Sender порт для отправки данных
type Sender interface {
	Send(ctx context.Context, batch *domain.MetricBatch) error
	Close() error
}

// BatchSender порт для отправки батчей
type BatchSender interface {
	Sender
	SendBatch(ctx context.Context, batches []*domain.MetricBatch) error
}
