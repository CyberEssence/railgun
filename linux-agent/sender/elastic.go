package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"linux-agent/config"
	"linux-agent/pkg/models"
)

type ElasticSender struct {
	client   *http.Client
	config   *config.ElasticConfig
	hostID   string
	hostname string
	baseURL  string
}

func NewElasticSender(cfg *config.ElasticConfig, hostID, hostname string) (*ElasticSender, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 90 * time.Second,
		},
	}

	baseURL := strings.TrimSuffix(cfg.URL, "/")

	sender := &ElasticSender{
		client:   client,
		config:   cfg,
		hostID:   hostID,
		hostname: hostname,
		baseURL:  baseURL,
	}

	// Проверяем подключение
	if err := sender.checkConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %v", err)
	}

	return sender, nil
}

func (e *ElasticSender) SendBatch(batch []*models.MetricBatch) error {
	if len(batch) == 0 {
		return nil
	}

	for _, metrics := range batch {
		// Формируем индекс
		indexName := e.getIndexName()

		// Создаем документ
		doc := map[string]interface{}{
			"@timestamp": time.Now().UTC(),
			"host": map[string]interface{}{
				"id":   e.hostID,
				"name": e.hostname,
			},
			"metrics": metrics,
		}

		docJSON, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/%s/_doc", e.baseURL, indexName)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(docJSON))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		if e.config.Username != "" && e.config.Password != "" {
			req.SetBasicAuth(e.config.Username, e.config.Password)
		}

		resp, err := e.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("Elasticsearch error: %d", resp.StatusCode)
		}
	}

	return nil
}

func (e *ElasticSender) Close() error {
	return nil
}

func (e *ElasticSender) checkConnection() error {
	req, err := http.NewRequest("GET", e.baseURL, nil)
	if err != nil {
		return err
	}

	if e.config.Username != "" && e.config.Password != "" {
		req.SetBasicAuth(e.config.Username, e.config.Password)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("connection failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (e *ElasticSender) getIndexName() string {
	if e.config.Index != "" {
		indexTemplate := e.config.Index
		if strings.Contains(indexTemplate, "%{") {
			now := time.Now()
			indexTemplate = strings.ReplaceAll(indexTemplate, "%{+yyyy.MM.dd}", now.Format("2006.01.02"))
			indexTemplate = strings.ReplaceAll(indexTemplate, "%{+yyyy.MM}", now.Format("2006.01"))
			indexTemplate = strings.ReplaceAll(indexTemplate, "%{+yyyy}", now.Format("2006"))
		}
		return indexTemplate
	}
	return fmt.Sprintf("siem-logs-%s", time.Now().Format("2006.01.02"))
}
