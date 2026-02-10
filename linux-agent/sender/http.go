package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"linux-agent/config"
)

type HTTPSender struct {
	client *http.Client
	config *config.HTTPConfig
	hostID string
	token  string
	batch  []interface{}
}

func NewHTTPSender(cfg *config.HTTPConfig, hostID string) *HTTPSender {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &HTTPSender{
		client: client,
		config: cfg,
		hostID: hostID,
		token:  cfg.Token,
		batch:  make([]interface{}, 0, cfg.BatchSize),
	}
}

func (h *HTTPSender) Send(data interface{}) error {
	// Добавляем в батч
	h.batch = append(h.batch, data)

	// Если батч заполнен, отправляем
	if len(h.batch) >= h.config.BatchSize {
		return h.flush()
	}

	return nil
}

func (h *HTTPSender) flush() error {
	if len(h.batch) == 0 {
		return nil
	}

	payload := map[string]interface{}{
		"host_id": h.hostID,
		"events":  h.batch,
		"count":   len(h.batch),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", h.config.URL+"/api/traffic",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.token)
	req.Header.Set("X-Host-ID", h.hostID)

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Сбрасываем батч
	h.batch = h.batch[:0]

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return nil
}

func (h *HTTPSender) Close() error {
	// Отправляем оставшиеся данные
	return h.flush()
}
