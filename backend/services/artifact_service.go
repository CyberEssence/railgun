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

type ArtifactService struct {
	db      *bun.DB
	elastic *elasticsearch.Client
}

func NewArtifactService(db *bun.DB, elasticURL string) *ArtifactService {
	cfg := elasticsearch.Config{
		Addresses: []string{elasticURL},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("Error creating Elasticsearch client: %s", err))
	}

	return &ArtifactService{
		db:      db,
		elastic: es,
	}
}

// GetArtifactsByHost получает артефакты для конкретного хоста
func (s *ArtifactService) GetArtifactsByHost(ctx context.Context, hostID string, artifactType string, from, to time.Time) ([]models.WindowsArtifact, error) {
	var artifacts []models.WindowsArtifact

	query := s.db.NewSelect().
		Model(&artifacts).
		Where("host_id = ?", hostID).
		Where("timestamp BETWEEN ? AND ?", from, to)

	if artifactType != "" {
		query = query.Where("type = ?", artifactType)
	}

	err := query.Order("timestamp DESC").Scan(ctx)

	return artifacts, err
}

// GetArtifactByID получает конкретный артефакт по ID
func (s *ArtifactService) GetArtifactByID(ctx context.Context, id int64) (*models.WindowsArtifact, error) {
	artifact := new(models.WindowsArtifact)

	err := s.db.NewSelect().
		Model(artifact).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return artifact, nil
}

// SaveArtifact сохраняет новый артефакт
func (s *ArtifactService) SaveArtifact(ctx context.Context, artifact models.WindowsArtifact) error {
	_, err := s.db.NewInsert().Model(&artifact).Exec(ctx)
	if err != nil {
		return err
	}

	// Также сохраняем в Elasticsearch
	artifactJSON, err := json.Marshal(artifact)
	if err != nil {
		return err
	}

	// Создаем io.Reader из []byte
	queryReader := bytes.NewReader(artifactJSON)

	_, err = s.elastic.Index(
		"windows-artifacts",
		queryReader,
		s.elastic.Index.WithContext(ctx),
		s.elastic.Index.WithDocumentID(fmt.Sprintf("%d", artifact.ID)),
	)

	return err
}

// SearchArtifacts ищет артефакты по заданным критериям
// Fixed SearchArtifacts method
// SearchArtifacts ищет артефакты по заданным критериям
func (s *ArtifactService) SearchArtifacts(ctx context.Context, query string) ([]models.WindowsArtifact, error) {
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"path", "value", "owner", "type"},
				"type":   "best_fields",
			},
		},
	}
	queryBytes, err := json.Marshal(esQuery)
	if err != nil {
		return nil, err
	}

	// Create io.Reader from []byte
	queryReader := bytes.NewReader(queryBytes)

	res, err := s.elastic.Search(
		s.elastic.Search.WithContext(ctx),
		s.elastic.Search.WithIndex("windows-artifacts"),
		s.elastic.Search.WithBody(queryReader),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("error parsing elasticsearch error: %s", err)
		}
		return nil, fmt.Errorf("elasticsearch error: %v", e)
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string                 `json:"_id"`
				Source models.WindowsArtifact `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	artifacts := make([]models.WindowsArtifact, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		artifacts[i] = hit.Source
	}
	return artifacts, nil
}
