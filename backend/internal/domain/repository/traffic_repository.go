package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/uptrace/bun"

	"railgun-core/internal/domain/models"
)

type TrafficRepository struct {
	db      *bun.DB
	elastic *elasticsearch.Client
}

func NewTrafficRepository(db *bun.DB, elasticURL string) *TrafficRepository {
	cfg := elasticsearch.Config{
		Addresses: []string{elasticURL},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Printf("Warning: Elasticsearch client creation failed: %v", err)
	} else {
		// Создаем индексы, если они не существуют
		_, err = es.Indices.Create("network-traffic")
		if err != nil && !strings.Contains(err.Error(), "resource_already_exists_exception") {
			log.Printf("Warning: Failed to create Elasticsearch index: %v", err)
		}

		_, err = es.Indices.Create("network-logs")
		if err != nil && !strings.Contains(err.Error(), "resource_already_exists_exception") {
			log.Printf("Warning: Failed to create Elasticsearch index: %v", err)
		}
	}

	return &TrafficRepository{
		db:      db,
		elastic: es,
	}
}

func (r *TrafficRepository) GetTrafficByHost(ctx context.Context, hostID string, from, to time.Time) ([]models.NetworkTraffic, error) {
	var traffic []models.NetworkTraffic
	err := r.db.NewSelect().
		Model(&traffic).
		Where("host_id = ?", hostID).
		Where("timestamp BETWEEN ? AND ?", from, to).
		Order("timestamp DESC").
		Scan(ctx)

	return traffic, err
}

// GetTrafficStats получает статистику трафика из Elasticsearch
func (r *TrafficRepository) GetTrafficStats(ctx context.Context, hostID string, from, to time.Time) (*models.TrafficStats, error) {
	var buf bytes.Buffer

	query := map[string]interface{}{
		"size": 0, // Не возвращаем сами документы, только агрегации
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					// Фильтр по Host ID
					{"term": map[string]interface{}{"host.id": hostID}},
					// Фильтр по типу документа (чтобы не считать file/log)
					{"exists": map[string]interface{}{"field": "network.bandwidth"}},
					// Фильтр по времени
					{"range": map[string]interface{}{
						"@timestamp": map[string]interface{}{
							"gte": from.Format(time.RFC3339),
							"lte": to.Format(time.RFC3339),
						},
					}},
				},
			},
		},
		"aggs": map[string]interface{}{
			"total_sent": map[string]interface{}{
				"sum": map[string]interface{}{"field": "network.bandwidth.bytes_sent"},
			},
			"total_recv": map[string]interface{}{
				"sum": map[string]interface{}{"field": "network.bandwidth.bytes_recv"},
			},
			"total_pkts_sent": map[string]interface{}{
				"sum": map[string]interface{}{"field": "network.bandwidth.packets_sent"},
			},
			"total_pkts_recv": map[string]interface{}{
				"sum": map[string]interface{}{"field": "network.bandwidth.packets_recv"},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	// Выполняем запрос к ES
	// Используем индекс siem-logs-* для охвата всех дат
	req := esapi.SearchRequest{
		Index: []string{"siem-logs-*"},
		Body:  &buf,
	}

	res, err := req.Do(ctx, r.elastic)
	if err != nil {
		return nil, fmt.Errorf("failed to search ES: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("ES error response parsing failed: %s", res.String())
		}
		return nil, fmt.Errorf("ES error: %v", e)
	}

	// Парсим ответ
	var esResp struct {
		Aggregations struct {
			TotalSent struct {
				Value float64 `json:"value"`
			} `json:"total_sent"`
			TotalRecv struct {
				Value float64 `json:"value"`
			} `json:"total_recv"`
			TotalPktsSent struct {
				Value float64 `json:"value"`
			} `json:"total_pkts_sent"`
			TotalPktsRecv struct {
				Value float64 `json:"value"`
			} `json:"total_pkts_recv"`
		} `json:"aggregations"`
		// Hits тоже парсим, чтобы проверить, есть ли вообще данные
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("failed to decode ES response: %w", err)
	}

	// Формируем результат
	stats := &models.TrafficStats{
		TotalBytesSent:   int64(esResp.Aggregations.TotalSent.Value),
		TotalBytesRecv:   int64(esResp.Aggregations.TotalRecv.Value),
		TotalPacketsSent: int64(esResp.Aggregations.TotalPktsSent.Value),
		TotalPacketsRecv: int64(esResp.Aggregations.TotalPktsRecv.Value),
	}

	return stats, nil
}

func (r *TrafficRepository) SaveTraffic(ctx context.Context, traffic models.NetworkTraffic) error {
	// Сохраняем в PostgreSQL
	_, err := r.db.NewInsert().Model(&traffic).Exec(ctx)
	if err != nil {
		return err
	}

	// Если Elasticsearch доступен, индексируем и там
	if r.elastic != nil {
		trafficJSON, err := json.Marshal(traffic)
		if err != nil {
			return err
		}

		_, err = r.elastic.Index(
			"network-traffic",
			bytes.NewReader(trafficJSON),
			r.elastic.Index.WithContext(ctx),
			r.elastic.Index.WithDocumentID(fmt.Sprintf("%s", traffic.ID)),
		)
		if err != nil {
			log.Printf("Warning: Failed to index traffic in Elasticsearch: %v", err)
		}
	}

	return nil
}

func (r *TrafficRepository) ProcessNetworkLog(ctx context.Context, hostID, logData, logType string) ([]models.NetworkLog, error) {
	// Парсим входные данные в структуру NetworkLog
	parsedLog := models.NetworkLog{
		SourceIP:  "",
		LogType:   logType,
		RawData:   logData,
		Timestamp: time.Now(),
		Severity:  "info",
	}

	// Предполагаем, что logData содержит строки вида "src_ip=X.X.X.X dst_ip=Y.Y.Y.Y"
	if strings.Contains(logData, "src_ip=") {
		parts := strings.Split(logData, " ")
		for _, part := range parts {
			if strings.HasPrefix(part, "src_ip=") {
				parsedLog.SourceIP = strings.TrimPrefix(part, "src_ip=")
			}
			if strings.HasPrefix(part, "dst_ip=") {
				parsedLog.DestinationIP = strings.TrimPrefix(part, "dst_ip=")
			}
		}
	}

	// Сохраняем в PostgreSQL
	err := r.db.NewInsert().Model(&parsedLog).Scan(ctx, &parsedLog.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to save log to database: %w", err)
	}

	// Если Elasticsearch доступен, индексируем и там
	if r.elastic != nil {
		logJSON, _ := json.Marshal(parsedLog)
		req := esapi.IndexRequest{
			Index:      "network-logs",
			Body:       bytes.NewReader(logJSON),
			DocumentID: fmt.Sprintf("%s", parsedLog.ID),
			Refresh:    "true",
		}
		res, err := req.Do(ctx, r.elastic)
		if err != nil {
			log.Printf("Warning: Failed to index log in Elasticsearch: %v", err)
		}
		defer res.Body.Close()
	}

	// Возвращаем связанные записи трафика
	var recentLogs []models.NetworkLog
	err = r.db.NewSelect().
		Model(&recentLogs).
		Where("host_id = ?", hostID).
		Where("timestamp > ?", time.Now().Add(-1*time.Hour)).
		Order("timestamp DESC").
		Limit(100).
		Scan(ctx)

	if err != nil {
		return []models.NetworkLog{}, nil
	}

	return []models.NetworkLog{parsedLog}, nil
}

// Вспомогательная функция для парсинга лога
func parseNetworkLog(sourceIP, destinationIP, logData string) (models.NetworkLog, error) {
	return models.NetworkLog{
		SourceIP:      sourceIP,
		DestinationIP: destinationIP,
		RawData:       logData,
		Timestamp:     time.Now(),
		// Дополнительные поля можно заполнить из logData
	}, nil
}

// Вспомогательная функция для получения связанного трафика
func (r *TrafficRepository) getRelatedTraffic(ctx context.Context, sourceIP, destinationIP string) ([]models.NetworkTraffic, error) {
	var traffic []models.NetworkTraffic
	err := r.db.NewSelect().
		Model(&traffic).
		Where("source_ip = ? OR destination_ip = ?", sourceIP, destinationIP).
		Where("timestamp > ?", time.Now().Add(-24*time.Hour)).
		Order("timestamp DESC").
		Limit(100).
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return traffic, nil
}

func (r *TrafficRepository) IsolateHost(ctx context.Context, hostID string, reason string, duration int) error {
	// Валидация входных данных
	if hostID == "" {
		return fmt.Errorf("invalid host ID")
	}

	// Проверяем существование хоста
	var host models.Host
	err := r.db.NewSelect().
		Model(&host).
		Where("id = ?", hostID).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			// Создаем хост, если он не существует
			host = models.Host{
				ID:          hostID,
				Hostname:    "auto-created",
				IPAddress:   "",
				LastSeen:    time.Now(),
				OSVersion:   "unknown",
				Status:      "isolated", // Сразу устанавливаем статус "isolated"
				Description: reason,
			}
			_, err = r.db.NewInsert().Model(&host).Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to create host: %w", err)
			}
			return nil // Хост создан и уже изолирован
		}
		return fmt.Errorf("host not found: %w", err)
	}

	// Обновляем статус хоста
	_, err = r.db.NewUpdate().
		Model(&host).
		Set("status = ?", "isolated").
		Where("id = ?", hostID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update host status: %w", err)
	}

	return nil
}

func (r *TrafficRepository) GetThreatHeatmap(ctx context.Context, from, to time.Time) ([]models.HeatmapPoint, error) {
	// Валидация временного диапазона
	if from.After(to) {
		return nil, fmt.Errorf("invalid time range: from cannot be after to")
	}

	// Получаем данные о трафике за указанный период
	var logs []models.NetworkLog
	err := r.db.NewSelect().
		Model(&logs).
		Where("timestamp BETWEEN ? AND ?", from, to).
		Where("severity IN ('warning', 'critical')").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}

	// Создаем карту для агрегации данных по IP
	ipMap := make(map[string]*models.HeatmapPoint)

	for _, log := range logs {
		// Получаем геоданные для IP
		geoInfo := getGeoInfoFromIP(log.SourceIP)
		if geoInfo == nil {
			continue
		}

		// Если точка для этого IP уже существует, увеличиваем вес
		if point, exists := ipMap[log.SourceIP]; exists {
			point.Weight++
			continue
		}

		// Создаем новую точку тепловой карты
		ipMap[log.SourceIP] = &models.HeatmapPoint{
			Latitude:  geoInfo.Latitude,
			Longitude: geoInfo.Longitude,
			Weight:    1,
			IP:        log.SourceIP,
			Country:   geoInfo.Country,
		}
	}

	// Преобразуем карту в слайс
	result := make([]models.HeatmapPoint, 0, len(ipMap))
	for _, point := range ipMap {
		result = append(result, *point)
	}

	return result, nil
}

func getGeoInfoFromIP(ipStr string) *models.GeoInfo {
	// Простая реализация для демонстрации
	// В реальном проекте используйте базу GeoIP или внешний API

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil
	}

	// Пример: преобразуем последний октет IP в "псевдо-координаты"
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		return nil
	}

	lastOctet, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil
	}

	// Генерируем "фейковые" координаты на основе IP
	// В реальном проекте это должно быть заменено реальными данными
	return &models.GeoInfo{
		Latitude:  30 + float64(lastOctet%20)/10,   // 30.0 - 31.9
		Longitude: -100 + float64(lastOctet%30)/10, // -100.0 - -97.1
		Country:   "US",                            // Пример - все IP считаются из США
	}
}

func (r *TrafficRepository) GetDashboardStats(ctx context.Context, from, to time.Time) (map[string]interface{}, error) {
	// Валидация временного диапазона
	if from.After(to) {
		return nil, fmt.Errorf("invalid time range: from cannot be after to")
	}

	stats := make(map[string]interface{})

	// Получаем общее количество событий за период
	var eventCount int
	err := r.db.NewRaw("SELECT COUNT(*) FROM events WHERE timestamp BETWEEN ? AND ?", from, to).Scan(ctx, &eventCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}
	stats["total_events"] = eventCount

	// Получаем количество активных соединений (последние 15 минут от конечной даты)
	var connectionCount int
	activeSince := to.Add(-15 * time.Minute)
	err = r.db.NewRaw("SELECT COUNT(*) FROM network_traffic WHERE timestamp BETWEEN ? AND ?",
		activeSince, to).Scan(ctx, &connectionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count connections: %w", err)
	}
	stats["active_connections"] = connectionCount

	// Получаем количество подозрительных активностей за период
	var suspiciousCount int
	err = r.db.NewRaw(`
        SELECT COUNT(*) FROM events 
        WHERE severity IN ('warning', 'critical') 
        AND timestamp BETWEEN ? AND ?`,
		from, to).Scan(ctx, &suspiciousCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count suspicious activities: %w", err)
	}
	stats["suspicious_activity"] = suspiciousCount

	// Определяем состояние системы
	var systemHealth string
	if suspiciousCount > 10 {
		systemHealth = "critical"
	} else if suspiciousCount > 5 {
		systemHealth = "warning"
	} else {
		systemHealth = "healthy"
	}
	stats["system_health"] = systemHealth

	// Добавляем метаданные о временном диапазоне
	stats["time_range"] = map[string]interface{}{
		"from": from.Format(time.RFC3339),
		"to":   to.Format(time.RFC3339),
	}

	return stats, nil
}

// Вспомогательная функция для определения региона по IP-адресу
func getRegionFromIP(ip string) string {
	// Упрощенная реализация - в реальном приложении здесь будет
	// использоваться геолокационная база данных
	if strings.HasPrefix(ip, "192.168.") {
		return "local"
	} else if strings.HasPrefix(ip, "10.") {
		return "internal"
	} else if strings.HasPrefix(ip, "172.16.") {
		return "dmz"
	} else {
		return "external"
	}
}
