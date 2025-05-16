package persistence

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
	"github.com/uptrace/bun"

	"railgun-core/internal/models"
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

func (r *TrafficRepository) GetTrafficStats(ctx context.Context, hostID string, from, to time.Time) (*models.TrafficStats, error) {
	var stats struct {
		TotalBytesSent   int64   `bun:"total_bytes_sent"`
		TotalBytesRecv   int64   `bun:"total_bytes_recv"`
		TotalPacketsSent int64   `bun:"total_packets_sent"`
		TotalPacketsRecv int64   `bun:"total_packets_recv"`
		AverageDuration  float64 `bun:"average_duration"`
	}

	query := `
        SELECT 
            COALESCE(SUM(bytes_sent), 0) as total_bytes_sent,
            COALESCE(SUM(bytes_recv), 0) as total_bytes_recv,
            COALESCE(SUM(packets_sent), 0) as total_packets_sent,
            COALESCE(SUM(packets_recv), 0) as total_packets_recv,
            COALESCE(AVG(duration), 0) as average_duration
        FROM network_traffic
        WHERE host_id = ? AND timestamp BETWEEN ? AND ?
    `

	err := r.db.NewRaw(query, hostID, from, to).Scan(ctx, &stats)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Копируем значения в результат
	return &models.TrafficStats{
		TotalBytesSent:   stats.TotalBytesSent,
		TotalBytesRecv:   stats.TotalBytesRecv,
		TotalPacketsSent: stats.TotalPacketsSent,
		TotalPacketsRecv: stats.TotalPacketsRecv,
		AverageDuration:  stats.AverageDuration,
	}, nil
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
			r.elastic.Index.WithDocumentID(fmt.Sprintf("%d", traffic.ID)),
		)
		if err != nil {
			log.Printf("Warning: Failed to index traffic in Elasticsearch: %v", err)
		}
	}

	return nil
}

func (r *TrafficRepository) ProcessNetworkLog(ctx context.Context, hostID, logData, logType string) ([]models.NetworkTraffic, error) {
	// Парсим входные данные в структуру NetworkLog
	parsedLog := models.NetworkLog{
		SourceIP:  "", // Будет заполнено из logData
		LogType:   logType,
		RawData:   logData,
		Timestamp: time.Now(),
		Severity:  "info", // По умолчанию
	}

	// Простая логика парсинга (в реальном приложении будет сложнее)
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
	_, err := r.db.NewInsert().Model(&parsedLog).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to save log to database: %w", err)
	}

	// Если Elasticsearch доступен, индексируем и там
	if r.elastic != nil {
		logJSON, err := json.Marshal(parsedLog)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal log for Elasticsearch: %w", err)
		}

		_, err = r.elastic.Index(
			"network-logs",
			bytes.NewReader(logJSON),
			r.elastic.Index.WithContext(ctx),
			r.elastic.Index.WithDocumentID(fmt.Sprintf("%d", parsedLog.ID)),
		)
		if err != nil {
			log.Printf("Warning: Failed to index log in Elasticsearch: %v", err)
		}
	}

	// Возвращаем связанные записи трафика
	var relatedTraffic []models.NetworkTraffic
	err = r.db.NewSelect().
		Model(&relatedTraffic).
		Where("host_id = ?", hostID).
		Where("timestamp > ?", time.Now().Add(-1*time.Hour)).
		Order("timestamp DESC").
		Limit(100).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get related traffic: %w", err)
	}

	return relatedTraffic, nil
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
