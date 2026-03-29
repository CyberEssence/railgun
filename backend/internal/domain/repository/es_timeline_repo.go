package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"railgun-core/internal/domain/models"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type ESTimelineRepository struct {
	client *elasticsearch.Client
	index  string
}

func NewESTimelineRepository(client *elasticsearch.Client, index string) *ESTimelineRepository {
	return &ESTimelineRepository{client: client, index: index}
}

func (r *ESTimelineRepository) GetHostTimeline(ctx context.Context, hostID string, start, end time.Time) ([]models.APTEpoch, error) {
	// Формируем запрос в виде map (который сериализуется в JSON для ES)
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{"term": map[string]interface{}{"host_id.keyword": hostID}},
					map[string]interface{}{"range": map[string]interface{}{"@timestamp": map[string]interface{}{"gte": start, "lte": end}}},
					map[string]interface{}{"range": map[string]interface{}{"malicious_probability": map[string]interface{}{"gt": 0.7}}},
				},
			},
		},
		"sort": []interface{}{
			map[string]interface{}{"@timestamp": map[string]interface{}{"order": "asc"}},
		},
		"size": 50,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	// Выполняем запрос (синтаксис v8)
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.index),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("elastic error [%s]: %s", res.Status(), e)
	}

	// Парсим ответ от ES
	var esResponse struct {
		Hits struct {
			Hits []struct {
				Source json.RawMessage `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		return nil, err
	}

	// Маппим в нашу структуру
	epochs := make([]models.APTEpoch, 0, len(esResponse.Hits.Hits))

	for _, hit := range esResponse.Hits.Hits {
		var event struct {
			Timestamp      time.Time `json:"@timestamp"`
			MitreStage     string    `json:"mitre_stage"`
			MainIndicator  string    `json:"indicator"`
			MaliciousScore float64   `json:"malicious_probability"`
		}

		if err := json.Unmarshal(hit.Source, &event); err != nil {
			continue // Пропускаем кривые документы
		}

		epochs = append(epochs, models.APTEpoch{
			Timestamp: event.Timestamp.Unix(),
			Stage:     event.MitreStage,
			Indicator: event.MainIndicator,
			RiskScore: event.MaliciousScore,
		})
	}

	return epochs, nil
}
