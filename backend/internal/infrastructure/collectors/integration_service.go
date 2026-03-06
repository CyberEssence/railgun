package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"
)

// IntegrationService реализует интерфейс domain.IntegrationService
type IntegrationService struct {
	client *http.Client
	config IntegrationConfig
}

// IntegrationConfig определяет параметры для внешних API
type IntegrationConfig struct {
	VirusTotalAPIKey string
	MaxFileSize      int64
	PollTimeout      time.Duration
	PollInterval     time.Duration
}

func NewIntegrationService(config IntegrationConfig) domain.IntegrationService {
	// Устанавливаем значения по умолчанию, если они не заданы
	if config.PollTimeout == 0 {
		config.PollTimeout = 2 * time.Minute
	}
	if config.PollInterval == 0 {
		config.PollInterval = 5 * time.Second
	}

	return &IntegrationService{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *IntegrationService) ScanWithVirusTotal(ctx context.Context, fileReader io.Reader, fileSize int64) (*models.ScanResult, error) {
	// Проверка конфигурации
	if s.config.VirusTotalAPIKey == "" {
		return nil, fmt.Errorf("VirusTotal API key not configured")
	}

	// Сравниваем с размером в байтах (предполагаем, что MaxFileSize уже в байтах)
	if fileSize > s.config.MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d bytes)", fileSize, s.config.MaxFileSize)
	}

	// Загрузка файла (Upload)
	uploadURL := "https://www.virustotal.com/api/v3/files"
	analysisID, err := s.uploadFileToVT(ctx, uploadURL, fileReader)
	if err != nil {
		return nil, fmt.Errorf("file upload failed: %w", err)
	}

	// Ожидание и получение отчета (Polling)
	reportURL := fmt.Sprintf("https://www.virustotal.com/api/v3/analyses/%s", analysisID)

	result, err := s.pollForReport(ctx, reportURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis report: %w", err)
	}

	return result, nil
}

func (s *IntegrationService) uploadFileToVT(ctx context.Context, url string, reader io.Reader) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "upload.file")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, reader); err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-apikey", s.config.VirusTotalAPIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var uploadResp struct {
		Data struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", fmt.Errorf("failed to decode upload response: %w", err)
	}

	if uploadResp.Data.ID == "" {
		return "", fmt.Errorf("empty analysis ID received")
	}

	return uploadResp.Data.ID, nil
}

// pollForReport опрашивает API
func (s *IntegrationService) pollForReport(ctx context.Context, url string) (*models.ScanResult, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.PollTimeout)
	defer cancel()

	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for scan result")
		case <-ticker.C:
			req, err := http.NewRequestWithContext(timeoutCtx, "GET", url, nil)
			if err != nil {
				return nil, err
			}

			req.Header.Set("x-apikey", s.config.VirusTotalAPIKey)
			req.Header.Set("Accept", "application/json")

			resp, err := s.client.Do(req)
			if err != nil {
				continue // Повторная попытка при сетевой ошибке
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				if resp.StatusCode == http.StatusNotFound {
					continue
				}
				return nil, fmt.Errorf("unexpected status code %d fetching report", resp.StatusCode)
			}

			var result models.ScanResult
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("failed to decode report: %w", err)
			}
			resp.Body.Close()

			if result.Data.Attributes.Status == "completed" {
				return &result, nil
			}
		}
	}
}
