package sender

import (
	"fmt"

	"linux-agent/config"
	"linux-agent/pkg/models"
)

type Sender interface {
	SendBatch(batch []*models.MetricBatch) error
	Close() error
}

func NewSender(cfg *config.Config, hostID, hostname string) (Sender, error) {
	// Приоритет: SIEM > Kafka > Elastic
	if cfg.SIEM.Enabled && cfg.SIEM.URL != "" {
		return NewSIEMSender(&cfg.SIEM, hostID, hostname)
	}

	if cfg.Kafka.Enabled && len(cfg.Kafka.Brokers) > 0 {
		return NewKafkaSender(&cfg.Kafka, hostID, hostname)
	}

	if cfg.Elastic.Enabled && cfg.Elastic.URL != "" {
		return NewElasticSender(&cfg.Elastic, hostID, hostname)
	}

	return nil, fmt.Errorf("no senders configured")
}
