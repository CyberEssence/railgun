package senders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"linux-agent/internal/core/domain"
	"linux-agent/internal/core/ports"

	"github.com/elastic/go-elasticsearch/v8"
)

// ElasticsearchSender отправляет данные в Elasticsearch
type ElasticsearchSender struct {
	client   *elasticsearch.Client
	index    string
	hostID   string
	hostname string
}

// ElasticsearchConfig конфигурация Elasticsearch
type ElasticsearchConfig struct {
	URLs     []string
	Username string
	Password string
	Index    string
}

// NewElasticsearchSender создает новый отправитель в Elasticsearch
func NewElasticsearchSender(cfg ElasticsearchConfig, hostID, hostname string) (ports.Sender, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.URLs,
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
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

	// Создаем шаблон индекса если его нет
	if err := createIndexTemplate(client); err != nil {
		log.Printf("Warning: failed to create index template: %v", err)
	}

	return &ElasticsearchSender{
		client:   client,
		index:    cfg.Index,
		hostID:   hostID,
		hostname: hostname,
	}, nil
}

// Send отправляет батч метрик в Elasticsearch
func (e *ElasticsearchSender) Send(ctx context.Context, batch *domain.MetricBatch) error {
	indexName := e.getIndexName()

	// Подготавливаем документ для Elasticsearch
	doc := map[string]interface{}{
		"@timestamp": time.Now().UTC(),
		"host": map[string]interface{}{
			"id":   batch.HostID,
			"name": batch.Hostname,
		},
		"system":    batch.System,
		"network":   batch.Network,
		"processes": batch.Processes,
		"security":  batch.Security,
	}

	// Удаляем nil значения
	e.cleanNil(doc)

	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// Отправляем в Elasticsearch
	res, err := e.client.Index(
		indexName,
		bytes.NewReader(data),
		e.client.Index.WithContext(ctx),
		e.client.Index.WithRefresh("wait_for"),
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch indexing error: %s", res.String())
	}

	log.Printf("Document indexed successfully to %s", indexName)
	return nil
}

// Close закрывает соединение
func (e *ElasticsearchSender) Close() error {
	return nil
}

func (e *ElasticsearchSender) getIndexName() string {
	if e.index != "" {
		indexTemplate := e.index
		now := time.Now()
		indexTemplate = strings.ReplaceAll(indexTemplate, "%{+yyyy.MM.dd}", now.Format("2006.01.02"))
		indexTemplate = strings.ReplaceAll(indexTemplate, "%{+yyyy.MM}", now.Format("2006.01"))
		indexTemplate = strings.ReplaceAll(indexTemplate, "%{+yyyy}", now.Format("2006"))
		return indexTemplate
	}
	return fmt.Sprintf("siem-logs-%s", time.Now().Format("2006.01.02"))
}

func (e *ElasticsearchSender) cleanNil(m map[string]interface{}) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
		}
		if vm, ok := v.(map[string]interface{}); ok {
			e.cleanNil(vm)
		}
	}
}

func createIndexTemplate(client *elasticsearch.Client) error {
	template := map[string]interface{}{
		"index_patterns": []string{"siem-logs-*"},
		"template": map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 1,
				"refresh_interval":   "5s",
			},
			"mappings": map[string]interface{}{
				"dynamic_templates": []map[string]interface{}{
					{
						"strings_as_keyword": map[string]interface{}{
							"match_mapping_type": "string",
							"mapping": map[string]interface{}{
								"type": "keyword",
							},
						},
					},
				},
				"properties": map[string]interface{}{
					"@timestamp": map[string]interface{}{
						"type": "date",
					},
					"host": map[string]interface{}{
						"properties": map[string]interface{}{
							"id":   map[string]interface{}{"type": "keyword"},
							"name": map[string]interface{}{"type": "keyword"},
						},
					},
					"system": map[string]interface{}{
						"properties": map[string]interface{}{
							"cpu_usage":      map[string]interface{}{"type": "float"},
							"memory_percent": map[string]interface{}{"type": "float"},
							"load_average":   map[string]interface{}{"type": "float"},
						},
					},
					"network": map[string]interface{}{
						"properties": map[string]interface{}{
							"connections": map[string]interface{}{
								"properties": map[string]interface{}{
									"remote_addr": map[string]interface{}{"type": "ip"},
									"local_port":  map[string]interface{}{"type": "integer"},
									"status":      map[string]interface{}{"type": "keyword"},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(template)
	if err != nil {
		return err
	}

	res, err := client.Indices.PutTemplate("siem-template", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
