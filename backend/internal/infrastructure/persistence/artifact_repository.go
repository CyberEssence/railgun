package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/uptrace/bun"

	"railgun-core/internal/models"
)

type ArtifactRepository struct {
	db      *bun.DB
	elastic *elasticsearch.Client
}

func NewArtifactRepository(db *bun.DB, elasticURL string) *ArtifactRepository {
	cfg := elasticsearch.Config{
		Addresses: []string{elasticURL},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Printf("Error creating Elasticsearch client: %v", err)
		return &ArtifactRepository{
			db:      db,
			elastic: nil,
		}
	}

	// Проверка соединения с Elasticsearch
	res, err := es.Ping()
	if err != nil {
		log.Printf("Elasticsearch ping failed: %v", err)
		return &ArtifactRepository{
			db:      db,
			elastic: nil,
		}
	}
	defer res.Body.Close()

	// Создание индекса с обработкой ошибок
	res, err = es.Indices.Create("windows-artifacts")
	if err != nil && !strings.Contains(err.Error(), "resource_already_exists_exception") {
		log.Printf("Error creating index: %v", err)
	}

	return &ArtifactRepository{
		db:      db,
		elastic: es,
	}
}

func (r *ArtifactRepository) GetArtifactsByHost(ctx context.Context, hostID string, page, perPage int) ([]*models.Artifact, int, error) {
	var windowsArtifacts []models.WindowsArtifact
	count, err := r.db.NewSelect().
		Model(&windowsArtifacts).
		Where("host_id = ?", hostID).
		Limit(perPage).
		Offset((page - 1) * perPage).
		ScanAndCount(ctx)

	// Конвертация в общий тип Artifact
	artifacts := make([]*models.Artifact, len(windowsArtifacts))
	for i, wa := range windowsArtifacts {
		artifacts[i] = &models.Artifact{
			ID:        wa.ID,
			HostID:    wa.HostID,
			Type:      wa.Type,
			Name:      wa.Path,
			Path:      wa.Path,
			Size:      wa.Size,
			Hash:      wa.Hash,
			Timestamp: wa.Timestamp,
		}
	}

	return artifacts, count, err
}

func (r *ArtifactRepository) GetArtifactByID(ctx context.Context, id int64) (*models.Artifact, error) {
	windowsArtifact := new(models.WindowsArtifact)
	err := r.db.NewSelect().
		Model(windowsArtifact).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	// Конвертируем WindowsArtifact в общий Artifact
	artifact := &models.Artifact{
		ID:          windowsArtifact.ID,
		HostID:      windowsArtifact.HostID,
		Type:        windowsArtifact.Type,
		Name:        windowsArtifact.Path, // Пример маппинга полей
		Path:        windowsArtifact.Path,
		Size:        windowsArtifact.Size,
		Hash:        windowsArtifact.Hash,
		Timestamp:   windowsArtifact.Timestamp,
		ThreatLevel: 0, // Задайте соответствующее значение
	}

	return artifact, nil
}

func (r *ArtifactRepository) SaveArtifact(ctx context.Context, artifact *models.WindowsArtifact) error {
	// Сохраняем в PostgreSQL
	_, err := r.db.NewInsert().Model(&artifact).Exec(ctx)
	if err != nil {
		return err
	}

	// Если Elasticsearch доступен, индексируем и там
	if r.elastic != nil {
		artifactJSON, err := json.Marshal(artifact)
		if err != nil {
			return err
		}

		_, err = r.elastic.Index(
			"windows-artifacts",
			bytes.NewReader(artifactJSON),
			r.elastic.Index.WithContext(ctx),
			r.elastic.Index.WithDocumentID(fmt.Sprintf("%d", artifact.ID)),
		)
		if err != nil {
			log.Printf("Warning: Failed to index artifact in Elasticsearch: %v", err)
		}
	}

	return nil
}

func (r *ArtifactRepository) SearchArtifacts(
	ctx context.Context,
	query string,
	artifactType string,
	severity string,
	page int,
	perPage int,
) ([]*models.Artifact, int, error) {

	// Рассчитываем offset для пагинации
	offset := (page - 1) * perPage

	if r.elastic == nil {
		// Поиск через PostgreSQL
		var artifacts []*models.Artifact
		q := r.db.NewSelect().
			Model(&artifacts).
			Where("(path ILIKE ? OR value ILIKE ? OR owner ILIKE ?)",
				"%"+query+"%", "%"+query+"%", "%"+query+"%")

		// Добавляем фильтры по типу и severity если они указаны
		if artifactType != "" {
			q = q.Where("type = ?", artifactType)
		}
		if severity != "" {
			q = q.Where("severity = ?", severity)
		}

		// Получаем общее количество записей для пагинации
		total, err := q.Count(ctx)
		if err != nil {
			return nil, 0, err
		}

		// Применяем сортировку и пагинацию
		err = q.
			Order("timestamp DESC").
			Limit(perPage).
			Offset(offset).
			Scan(ctx)

		return artifacts, total, err
	}

	// Поиск через Elasticsearch
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":  query,
							"fields": []string{"path", "value", "owner", "type"},
							"type":   "best_fields",
						},
					},
				},
			},
		},
	}

	// Добавляем фильтры в ES запрос
	filters := []map[string]interface{}{}
	if artifactType != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"type.keyword": artifactType,
			},
		})
	}
	if severity != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"severity.keyword": severity,
			},
		})
	}

	if len(filters) > 0 {
		esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = filters
	}

	queryBytes, err := json.Marshal(esQuery)
	if err != nil {
		return nil, 0, err
	}

	res, err := r.elastic.Search(
		r.elastic.Search.WithContext(ctx),
		r.elastic.Search.WithIndex("windows-artifacts"),
		r.elastic.Search.WithBody(bytes.NewReader(queryBytes)),
		r.elastic.Search.WithFrom(offset),
		r.elastic.Search.WithSize(perPage),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, 0, fmt.Errorf("error parsing elasticsearch error: %s", err)
		}
		return nil, 0, fmt.Errorf("elasticsearch error: %v", e)
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source *models.Artifact `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	artifacts := make([]*models.Artifact, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		artifacts[i] = hit.Source
	}

	return artifacts, result.Hits.Total.Value, nil
}
