package task_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"linux-agent/internal/core/domain"
	"linux-agent/internal/core/ports"
)

type HTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string) ports.TaskFetcher {
	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *HTTPClient) FetchTask(hostID string) (*domain.IsolationTask, error) {
	url := fmt.Sprintf("%s/agent/task", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Agent-ID", hostID)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		if resp.ContentLength == 0 {
			return nil, nil
		}

		var task domain.IsolationTask
		if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
			return nil, nil
		}

		if task.ID != 0 {
			return &task, nil
		}
	}

	return nil, nil
}

func (c *HTTPClient) ReportResult(report *domain.TaskReport) error {
	url := fmt.Sprintf("%s/agent/report", c.baseURL)
	body, err := json.Marshal(report)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	// В реальности стоит добавить подпись или API Key

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to report result, status: %d", resp.StatusCode)
	}

	log.Printf("Task %d result reported: %s", report.TaskID, report.Status)
	return nil
}
