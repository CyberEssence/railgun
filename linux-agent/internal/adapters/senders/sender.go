package senders

import (
	"fmt"

	"linux-agent/internal/core/ports"
)

// NewSender создает отправитель на основе конфигурации
func NewSender(config interface{}, hostID, hostname string) (ports.Sender, error) {
	switch cfg := config.(type) {
	case ElasticsearchConfig:
		return NewElasticsearchSender(cfg, hostID, hostname)
	default:
		return nil, fmt.Errorf("unsupported sender type")
	}
}
