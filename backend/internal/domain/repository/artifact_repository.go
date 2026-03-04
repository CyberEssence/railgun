package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"railgun-core/internal/domain/models"

	"github.com/elastic/go-elasticsearch/v8"
)

// ArtifactRepository реализует интерфейс для работы с артефактами через Elasticsearch
type ArtifactRepository struct {
	client      *elasticsearch.Client
	searchIndex string // для поиска используем паттерн
	writeIndex  string // для записи используем конкретный индекс
}

// NewArtifactRepository создает новый репозиторий
func NewArtifactRepository(addresses []string, username, password, indexPattern string) (*ArtifactRepository, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// Проверяем подключение
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	// Для поиска используем паттерн
	searchIndex := indexPattern
	if searchIndex == "" || strings.Contains(searchIndex, "*") {
		searchIndex = "siem-logs-*"
	}

	return &ArtifactRepository{
		client:      client,
		searchIndex: searchIndex,
		writeIndex:  "siem-logs-" + time.Now().Format("2006.01.02"),
	}, nil
}

// GetArtifactsByHost получает артефакты по хосту
func (r *ArtifactRepository) GetArtifactsByHost(ctx context.Context, hostID string, page, perPage int) ([]*models.Artifact, int, error) {
	from := (page - 1) * perPage

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"host_id": hostID,
			},
		},
		"sort": []map[string]interface{}{
			{
				"@timestamp": map[string]interface{}{ // Агент отправляет @timestamp
					"order": "desc",
				},
			},
		},
		"from": from,
		"size": perPage,
	}

	return r.search(ctx, query)
}

// GetArtifactByID получает артефакт по ID
func (r *ArtifactRepository) GetArtifactByID(ctx context.Context, id int64) (*models.Artifact, error) {
	idStr := fmt.Sprintf("%d", id)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"id": idStr,
						},
					},
					{
						"term": map[string]interface{}{
							"_id": idStr,
						},
					},
				},
				"minimum_should_match": 1,
			},
		},
		"size": 1,
	}

	artifacts, _, err := r.search(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(artifacts) == 0 {
		return nil, fmt.Errorf("artifact with ID %d not found", id)
	}

	return artifacts[0], nil
}

// SearchArtifacts ищет артефакты
func (r *ArtifactRepository) SearchArtifacts(ctx context.Context, query, artifactType, severity string, page, perPage int) ([]*models.Artifact, int, error) {
	from := (page - 1) * perPage

	// Базовый запрос
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{
			{
				"@timestamp": map[string]interface{}{ // Сортируем по @timestamp
					"order": "desc",
				},
			},
		},
		"from": from,
		"size": perPage,
	}

	// Добавляем полнотекстовый поиск
	if query != "" && query != "system" {
		must := esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"type", "path", "value", "host_id", "system.*", "network.*", "security.*"},
			},
		})
		esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = must
	}

	// Добавляем фильтры
	filters := []map[string]interface{}{}

	// Фильтр по типу (ищем в разных местах)
	if artifactType != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"type": artifactType,
			},
		})
	}

	// Фильтр по severity
	if severity != "" {
		severityMap := map[string]int{
			"low":      1,
			"medium":   5,
			"high":     8,
			"critical": 10,
		}
		if level, ok := severityMap[strings.ToLower(severity)]; ok {
			// Ищем threat_level в разных местах
			filters = append(filters, map[string]interface{}{
				"range": map[string]interface{}{
					"threat_level": map[string]interface{}{
						"gte": level,
					},
				},
			})
		}
	}

	if len(filters) > 0 {
		esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = filters
	}

	return r.search(ctx, esQuery)
}

// SaveArtifact сохраняет артефакт
func (r *ArtifactRepository) SaveArtifact(ctx context.Context, artifact *models.WindowsArtifact) error {
	// Обновляем writeIndex если день изменился
	currentIndex := "siem-logs-" + time.Now().Format("2006.01.02")
	if currentIndex != r.writeIndex {
		r.writeIndex = currentIndex
	}

	doc := map[string]interface{}{
		"@timestamp":   time.Now().UTC(), // Добавляем @timestamp для сортировки
		"uuid":         artifact.UUID,
		"host_id":      artifact.HostID,
		"type":         artifact.Type,
		"path":         artifact.Path,
		"size":         artifact.Size,
		"hash":         artifact.Hash,
		"value":        artifact.Value,
		"owner":        artifact.Owner,
		"permissions":  artifact.Permissions,
		"timestamp":    artifact.Timestamp,
		"threat_level": artifact.ThreatLevel,
		"created_at":   time.Now().UTC(),
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal artifact: %w", err)
	}

	res, err := r.client.Index(
		r.writeIndex,
		bytes.NewReader(data),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithDocumentID(artifact.UUID),
		r.client.Index.WithRefresh("wait_for"),
	)
	if err != nil {
		return fmt.Errorf("failed to index artifact: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch error: %s", res.String())
	}

	return nil
}

// GetArtifactByUUID получает артефакт по UUID
func (r *ArtifactRepository) GetArtifactByUUID(ctx context.Context, uuid string) (*models.Artifact, error) {

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": uuid,
			},
		},
		"size": 1,
	}

	artifacts, _, err := r.search(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(artifacts) == 0 {
		return nil, fmt.Errorf("artifact with UUID %s not found", uuid)
	}

	return artifacts[0], nil
}

// search выполняет поиск и возвращает результаты
func (r *ArtifactRepository) search(ctx context.Context, query map[string]interface{}) ([]*models.Artifact, int, error) {
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.searchIndex),
		r.client.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	artifacts := make([]*models.Artifact, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		artifact := r.convertToArtifact(hit.Source)
		if artifact != nil {
			artifacts = append(artifacts, artifact)
		}
	}

	return artifacts, result.Hits.Total.Value, nil
}

// convertToArtifact конвертирует Elasticsearch документ в Artifact
func (r *ArtifactRepository) convertToArtifact(source map[string]interface{}) *models.Artifact {
	artifact := &models.Artifact{
		UUID:      getString(source, "uuid"), // Берем UUID из поля uuid
		HostID:    getString(source, "host_id"),
		Type:      getString(source, "type"),
		Name:      getString(source, "path"),
		Path:      getString(source, "path"),
		Size:      getInt64(source, "size"),
		Hash:      getString(source, "hash"),
		Timestamp: getTime(source, "timestamp"), // Используем поле timestamp
	}

	// Если есть threat_level
	if threat, ok := source["threat_level"].(float64); ok {
		artifact.ThreatLevel = int(threat)
	}

	// Если это данные от Linux агента (system, network, security)
	if system, ok := source["system"].(map[string]interface{}); ok {
		artifact.Name = "system"
		artifact.Type = "system"
		artifact.ThreatLevel = calculateThreatLevel(system)

		// Если host_id нет в корне, берем из system
		if artifact.HostID == "" {
			if hostID, ok := system["host_id"].(string); ok {
				artifact.HostID = hostID
			}
		}

		// Если timestamp нет в корне, берем из system
		if artifact.Timestamp.IsZero() {
			if ts, ok := system["timestamp"].(string); ok {
				if t, err := time.Parse(time.RFC3339, ts); err == nil {
					artifact.Timestamp = t
				}
			}
		}
	}

	if network, ok := source["network"].(map[string]interface{}); ok {
		artifact.Name = "network"
		artifact.Type = "network"
		artifact.ThreatLevel = calculateNetworkThreatLevel(network)

		if artifact.HostID == "" {
			if hostID, ok := network["host_id"].(string); ok {
				artifact.HostID = hostID
			}
		}
	}

	if security, ok := source["security"].(map[string]interface{}); ok {
		artifact.Name = "security"
		artifact.Type = "security"
		artifact.ThreatLevel = calculateThreatLevel(security)

		if artifact.HostID == "" {
			if hostID, ok := security["host_id"].(string); ok {
				artifact.HostID = hostID
			}
		}
	}

	return artifact
}

// Вспомогательная функция для getTime
func getTime(m map[string]interface{}, key string) time.Time {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t
			}
		case time.Time:
			return v
		}
	}
	return time.Now()
}

// Вспомогательные функции
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		}
	}
	return 0
}

func calculateThreatLevel(system map[string]interface{}) int {
	level := 0

	if cpu, ok := system["cpu_usage"].(float64); ok && cpu > 90 {
		level += 3
	}

	if mem, ok := system["memory_percent"].(float64); ok && mem > 90 {
		level += 3
	}

	return level
}

func calculateNetworkThreatLevel(network map[string]interface{}) int {
	level := 0

	if conns, ok := network["connections"].([]interface{}); ok {
		for _, conn := range conns {
			if c, ok := conn.(map[string]interface{}); ok {
				if status, ok := c["status"].(string); ok && status != "ESTABLISHED" {
					level += 1
				}
			}
		}
	}

	return level
}
