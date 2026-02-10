package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"linux-agent/config"
)

type KafkaSender struct {
	writer *kafka.Writer
	config *config.KafkaConfig
	hostID string
}

func NewKafkaSender(cfg *config.KafkaConfig, hostID string) (*KafkaSender, error) {
	// Настройка механизма аутентификации
	var mechanism plain.Mechanism
	if cfg.SASL {
		mechanism = plain.Mechanism{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	// Создание writer с правильными параметрами
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
		writer: writer,
		config: cfg,
		hostID: hostID,
	}, nil
}

func (k *KafkaSender) Send(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// Определяем тип данных
	dataType := getDataType(data)

	message := kafka.Message{
		Key:   []byte(k.hostID),
		Value: jsonData,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "host_id", Value: []byte(k.hostID)},
			{Key: "type", Value: []byte(dataType)},
			{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = k.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %v", err)
	}

	return nil
}

func (k *KafkaSender) SendBatch(data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(data))

	for _, item := range data {
		jsonData, err := json.Marshal(item)
		if err != nil {
			log.Printf("Failed to marshal item: %v", err)
			continue
		}

		dataType := getDataType(item)

		messages = append(messages, kafka.Message{
			Key:   []byte(k.hostID),
			Value: jsonData,
			Time:  time.Now(),
			Headers: []kafka.Header{
				{Key: "host_id", Value: []byte(k.hostID)},
				{Key: "type", Value: []byte(dataType)},
			},
		})
	}

	if len(messages) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := k.writer.WriteMessages(ctx, messages...)
	if err != nil {
		return fmt.Errorf("failed to write batch to Kafka: %v", err)
	}

	return nil
}

func (k *KafkaSender) Close() error {
	if k.writer != nil {
		return k.writer.Close()
	}
	return nil
}

// Вспомогательная функция для определения типа данных
func getDataType(data interface{}) string {
	// Проверяем структуру через JSON маршалинг
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "unknown"
	}

	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		return "unknown"
	}

	// Ищем поле type
	if typeVal, ok := temp["type"].(string); ok {
		return typeVal
	}

	// Определяем по структуре
	switch v := data.(type) {
	case map[string]interface{}:
		if t, ok := v["type"].(string); ok {
			return t
		}
	}

	return "unknown"
}
