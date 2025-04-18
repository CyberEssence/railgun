package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/uptrace/bun"

	"railgun-core/models"
)

type TrafficService interface {
	GetTrafficByHost(ctx context.Context, hostID string, from, to time.Time) ([]models.NetworkTraffic, error)
	GetTrafficStats(ctx context.Context, hostID string, from, to time.Time) (models.TrafficStats, error)
	SaveTraffic(ctx context.Context, traffic *models.NetworkTraffic) error
}

type TrafficServiceImpl struct {
	db      *bun.DB
	elastic *elasticsearch.Client
}

func NewTrafficService(db *bun.DB, elasticURL string) *TrafficServiceImpl {
	cfg := elasticsearch.Config{
		Addresses: []string{elasticURL},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("Error creating Elasticsearch client: %s", err))
	}

	_, err = es.Indices.Create("network-traffic")
	if err != nil {
		// Игнорируем ошибку, если индекс уже существует
		if !strings.Contains(err.Error(), "resource_already_exists_exception") {
			panic(fmt.Sprintf("Error creating Elasticsearch index: %s", err))
		}
	}

	return &TrafficServiceImpl{
		db:      db,
		elastic: es,
	}
}

func (s *TrafficServiceImpl) GetTrafficByHost(ctx context.Context, hostID string, from, to time.Time) ([]models.NetworkTraffic, error) {
	var traffic []models.NetworkTraffic
	query := s.db.NewSelect().
		Model(&traffic).
		Where("host_id = ?", hostID).
		Where("timestamp BETWEEN ? AND ?", from, to).
		Order("timestamp DESC")
	err := query.Scan(ctx)
	return traffic, err
}

func (s *TrafficServiceImpl) GetTrafficStats(ctx context.Context, hostID string, from, to time.Time) (models.TrafficStats, error) {
	stats := models.TrafficStats{}
	query := `SELECT 
        COALESCE(SUM(bytes_sent), 0) as total_bytes_sent,
        COALESCE(SUM(bytes_recv), 0) as total_bytes_recv,
        COALESCE(SUM(packets_sent), 0) as total_packets_sent,
        COALESCE(SUM(packets_recv), 0) as total_packets_recv,
        COALESCE(AVG(duration), 0) as average_duration
    FROM network_traffic
    WHERE host_id = ? AND timestamp BETWEEN ? AND ?`

	// Используем Bun для выполнения запроса
	err := s.db.NewRaw(query, hostID, from, to).Scan(ctx, &stats)

	if err != nil {
		log.Printf("Query failed: %v", err)
		return stats, fmt.Errorf("query failed: %w", err)
	}

	return stats, nil
}

// SaveTraffic сохраняет новую запись о трафике
func (s *TrafficServiceImpl) SaveTraffic(ctx context.Context, traffic models.NetworkTraffic) error {
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
