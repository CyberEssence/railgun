package sender

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"linux-agent/config"
	"linux-agent/pkg/models"
)

type KafkaSender struct {
	writer   *kafka.Writer
	config   *config.KafkaConfig
	hostID   string
	hostname string
}

func NewKafkaSender(cfg *config.KafkaConfig, hostID, hostname string) (*KafkaSender, error) {
	var mechanism plain.Mechanism
	if cfg.SASL {
		mechanism = plain.Mechanism{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafka.LeastBytes{},
		Async:                  true,
		Compression:            kafka.Snappy,
		BatchSize:              100,
		BatchTimeout:           1 * time.Second,
		AllowAutoTopicCreation: true,
		Transport: &kafka.Transport{
			SASL: mechanism,
		},
	}

	return &KafkaSender{
		writer:   writer,
		config:   cfg,
		hostID:   hostID,
		hostname: hostname,
	}, nil
}

func (k *KafkaSender) SendBatch(batch []*models.MetricBatch) error {
	if len(batch) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(batch))

	for _, metrics := range batch {
		jsonData, err := json.Marshal(metrics)
		if err != nil {
			continue
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(k.hostID),
			Value: jsonData,
			Time:  time.Now(),
			Headers: []kafka.Header{
				{Key: "host_id", Value: []byte(k.hostID)},
				{Key: "hostname", Value: []byte(k.hostname)},
			},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return k.writer.WriteMessages(ctx, messages...)
}

func (k *KafkaSender) Close() error {
	if k.writer != nil {
		return k.writer.Close()
	}
	return nil
}
