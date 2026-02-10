package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"linux-agent/config"
)

type ElasticSender struct {
	client  *http.Client
	config  *config.ElasticConfig
	hostID  string
	baseURL string
}

func NewElasticSender(cfg *config.ElasticConfig, hostID string) (*ElasticSender, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Убираем trailing slash
	baseURL := strings.TrimSuffix(cfg.URL, "/")

	sender := &ElasticSender{
		client:  client,
		config:  cfg,
		hostID:  hostID,
		baseURL: baseURL,
	}

	// Проверяем подключение
	if err := sender.checkConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %v", err)
	}

	return sender, nil
}

func (e *ElasticSender) Send(data interface{}) error {
	// Определяем индекс
	indexName := e.getIndexName(data)

	// Маршалим данные
	/*jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}*/

	// Создаем документ с метаданными
	doc := map[string]interface{}{
		"@timestamp": time.Now().UTC(),
		"host": map[string]interface{}{
			"id":   e.hostID,
			"name": getHostname(data),
		},
		"event": map[string]interface{}{
			"type":   getDataType(data),
			"module": "linux_agent",
		},
		"agent": map[string]interface{}{
			"id":   e.hostID,
			"type": "linux",
		},
		"data": data,
	}

	docJSON, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	// URL для индексации
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
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("Elasticsearch error: %v", errorResp)
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

func (e *ElasticSender) getIndexName(data interface{}) string {
	// Определяем тип данных для индекса
	dataType := getDataType(data)

	// Используем шаблон из конфига или создаем свой
	if e.config.Index != "" {
		indexTemplate := e.config.Index
		// Подставляем дату если есть шаблон
		if strings.Contains(indexTemplate, "%{") {
			now := time.Now()
			indexTemplate = strings.ReplaceAll(indexTemplate, "%{+yyyy.MM.dd}", now.Format("2006.01.02"))
			indexTemplate = strings.ReplaceAll(indexTemplate, "%{+yyyy.MM}", now.Format("2006.01"))
			indexTemplate = strings.ReplaceAll(indexTemplate, "%{+yyyy}", now.Format("2006"))
		}
		return indexTemplate
	}

	// Дефолтный индекс
	return fmt.Sprintf("siem-%s-%s", dataType, time.Now().Format("2006.01.02"))
}

func getHostname(data interface{}) string {
	// Извлекаем hostname из данных
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "unknown"
	}

	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		return "unknown"
	}

	if hostname, ok := temp["hostname"].(string); ok {
		return hostname
	}

	return "unknown"
}
