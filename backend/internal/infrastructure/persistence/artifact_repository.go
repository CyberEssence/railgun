package persistence

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows // Явно возвращаем ошибку "не найдено"
		}
		return nil, fmt.Errorf("database error: %w", err)
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
	_, err := r.db.NewInsert().Model(artifact).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save artifact: %w", err)
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

// Проверяет наличие столбца threat_level в таблице
func (r *ArtifactRepository) hasThreatLevelColumn(ctx context.Context) bool {
	// Проверяем наличие столбца в информационной схеме PostgreSQL
	var exists bool
	err := r.db.NewRaw(`
        SELECT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_name = 'windows_artifacts' 
            AND column_name = 'threat_level'
        )
    `).Scan(ctx, &exists)

	if err != nil {
		// В случае ошибки предполагаем, что столбца нет
		log.Printf("Error checking threat_level column: %v", err)
		return false
	}

	return exists
}

func (r *ArtifactRepository) SearchArtifacts(
	ctx context.Context,
	query string,
	artifactType string,
	severity string,
	page int,
	perPage int,
) ([]*models.Artifact, int, error) {
	offset := (page - 1) * perPage
	severityMap := map[string]int{
		"low":      1,
		"medium":   5,
		"high":     8,
		"critical": 10,
	}

	if r.elastic == nil {
		// Используем WindowsArtifact как базовую модель
		var windowsArtifacts []*models.WindowsArtifact

		q := r.db.NewSelect().
			Model(&windowsArtifacts).
			WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.
					Where("path ILIKE ?", "%"+query+"%").
					WhereOr("value ILIKE ?", "%"+query+"%").
					WhereOr("owner ILIKE ?", "%"+query+"%")
			})

		if artifactType != "" {
			q = q.Where("type = ?", artifactType)
		}

		// Проверяем наличие столбца threat_level перед его использованием
		hasThreatLevel := r.hasThreatLevelColumn(ctx)
		if severity != "" && hasThreatLevel {
			if level, ok := severityMap[strings.ToLower(severity)]; ok {
				q = q.Where("threat_level >= ?", level)
			}
		}

		// Подсчет общего количества записей
		countQuery := q.Clone()
		total, err := countQuery.Count(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("count failed: %w", err)
		}

		// Основной запрос с данными
		err = q.
			Order("timestamp DESC").
			Limit(perPage).
			Offset(offset).
			Scan(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("query failed: %w", err)
		}

		// Преобразуем WindowsArtifact в Artifact
		artifacts := make([]*models.Artifact, len(windowsArtifacts))
		for i, wa := range windowsArtifacts {
			artifacts[i] = &models.Artifact{
				ID:        wa.ID,
				HostID:    wa.HostID,
				Type:      wa.Type,
				Name:      wa.Path, // Используем Path в качестве Name
				Path:      wa.Path,
				Size:      wa.Size,
				Hash:      wa.Hash,
				Timestamp: wa.Timestamp,
				// Если есть ThreatLevel, добавьте его
				ThreatLevel: wa.ThreatLevel,
			}
		}

		return artifacts, total, nil
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

	return []*models.Artifact{}, 0, nil
}
