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

type SIEMSender struct {
	client   *http.Client
	config   *config.SIEMConfig
	hostID   string
	hostname string
	baseURL  string
}

func NewSIEMSender(cfg *config.SIEMConfig, hostID, hostname string) (*SIEMSender, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 60 * time.Second,
		},
	}

	return &SIEMSender{
		client:   client,
		config:   cfg,
		hostID:   hostID,
		hostname: hostname,
		baseURL:  strings.TrimSuffix(cfg.URL, "/"),
	}, nil
}

func (s *SIEMSender) SendBatch(batch []*models.MetricBatch) error {
	if len(batch) == 0 {
		return nil
	}

	// Подготавливаем данные для отправки
	for _, metrics := range batch {
		payload := map[string]interface{}{
			"host_id":   metrics.HostID,
			"hostname":  metrics.Hostname,
			"timestamp": metrics.Timestamp,
			"metrics":   metrics.Metrics,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal metrics: %v", err)
		}

		// Отправляем на эндпоинт /api/traffic/logs
		req, err := http.NewRequest("POST",
			fmt.Sprintf("%s/api/traffic/logs", s.baseURL),
			bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		if s.config.Token != "" {
			req.Header.Set("Authorization", "Bearer "+s.config.Token)
		}
		req.Header.Set("X-Host-ID", s.hostID)

		resp, err := s.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("SIEM API error: %d", resp.StatusCode)
		}
	}

	return nil
}

func (s *SIEMSender) Close() error {
	return nil
}
