package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/uptrace/bun"

	"railgun-core/models"
)

type TrafficService struct {
	db      *bun.DB
	elastic *elasticsearch.Client
}

func NewTrafficService(db *bun.DB, elasticURL string) *TrafficService {
	cfg := elasticsearch.Config{
		Addresses: []string{elasticURL},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("Error creating Elasticsearch client: %s", err))
	}

	return &TrafficService{
		db:      db,
		elastic: es,
	}
}

// GetTrafficByHost получает трафик для конкретного хоста
func (s *TrafficService) GetTrafficByHost(ctx context.Context, hostID string, from, to time.Time) ([]models.NetworkTraffic, error) {
	var traffic []models.NetworkTraffic

	err := s.db.NewSelect().
		Model(&traffic).
		Where("host_id = ?", hostID).
		Where("timestamp BETWEEN ? AND ?", from, to).
		Order("timestamp DESC").
		Scan(ctx)

	return traffic, err
}

// GetTrafficStats получает статистику трафика для хоста
func (s *TrafficService) GetTrafficStats(ctx context.Context, hostID string, from, to time.Time) (map[string]interface{}, error) {
	var result map[string]interface{}

	// Формируем запрос к Elasticsearch для агрегации данных
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"host_id": hostID,
						},
					},
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": from.Format(time.RFC3339),
								"lte": to.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"total_bytes_sent": map[string]interface{}{
				"sum": map[string]interface{}{
					"field": "bytes_sent",
				},
			},
			"total_bytes_recv": map[string]interface{}{
				"sum": map[string]interface{}{
					"field": "bytes_recv",
				},
			},
			"by_protocol": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "protocol",
					"size":  10,
				},
			},
			"traffic_over_time": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":    "timestamp",
					"interval": "1h",
				},
				"aggs": map[string]interface{}{
					"bytes_sent": map[string]interface{}{
						"sum": map[string]interface{}{
							"field": "bytes_sent",
						},
					},
					"bytes_recv": map[string]interface{}{
						"sum": map[string]interface{}{
							"field": "bytes_recv",
						},
					},
				},
			},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	queryReader := bytes.NewReader(queryBytes)
	// Выполняем запрос к Elasticsearch
	res, err := s.elastic.Search(
		s.elastic.Search.WithContext(ctx),
		s.elastic.Search.WithIndex("network-traffic"),
		s.elastic.Search.WithBody(queryReader),
	)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	result = r["aggregations"].(map[string]interface{})
	return result, nil
}

// SaveTraffic сохраняет новую запись о трафике
func (s *TrafficService) SaveTraffic(ctx context.Context, traffic models.NetworkTraffic) error {
	_, err := s.db.NewInsert().Model(&traffic).Exec(ctx)
	if err != nil {
		return err
	}

	// Также сохраняем в Elasticsearch
	trafficJSON, err := json.Marshal(traffic)
	if err != nil {
		return err
	}

	// Создаем io.Reader из []byte
	queryReader := bytes.NewReader(trafficJSON)

	_, err = s.elastic.Index(
		"network-traffic",
		queryReader,
		s.elastic.Index.WithContext(ctx),
		s.elastic.Index.WithDocumentID(fmt.Sprintf("%d", traffic.ID)),
	)

	return err
}
