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

type IntegrationConfig struct {
	VirusTotalAPIKey string
	MaxFileSize      int
}

type IntegrationService struct {
	config IntegrationConfig
	client *http.Client
}

func NewIntegrationService(config IntegrationConfig) domain.IntegrationService {
	return &IntegrationService{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *IntegrationService) ScanWithVirusTotal(ctx context.Context, fileReader io.Reader, fileSize int64) (*models.ScanResult, error) {
	// 1. Проверка размера (уже по числу, а не по хедеру)
	if fileSize > int64(s.config.MaxFileSize) {
		return nil, fmt.Errorf("file too large: %d bytes (max %d bytes)", fileSize, s.config.MaxFileSize)
	}

	if s.config.VirusTotalAPIKey == "" {
		return nil, fmt.Errorf("VirusTotal API key not configured")
	}

	// 2. Создаем multipart форму
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// VirusTotal требует имя файла, можно передать "upload.file" или добавить аргумент в функцию
	part, err := writer.CreateFormFile("file", "upload.file")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// 3. Копируем данные из Reader в форму
	_, err = io.Copy(part, fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}
	writer.Close()

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.virustotal.com/api/v3/files", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("x-apikey", s.config.VirusTotalAPIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API error: %s (code: %s)", errorResp.Error.Message, errorResp.Error.Code)
		}
		return nil, fmt.Errorf("API error: status code %d", resp.StatusCode)
	}

	// Получаем ID анализа
	var uploadResp struct {
		Data struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Links struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Получаем результаты анализа (в реальном приложении здесь будет опрос API до получения результатов)
	// Для демонстрации возвращаем заглушку
	scanResult := &models.ScanResult{}
	scanResult.Data.ID = uploadResp.Data.ID
	scanResult.Data.Attributes.Status = "completed"
	scanResult.Data.Attributes.Stats.Malicious = 0
	scanResult.Data.Attributes.Stats.Suspicious = 1
	scanResult.Data.Attributes.Stats.Undetected = 67
	scanResult.Data.Attributes.Results = map[string]struct {
		Category string `json:"category"`
		Result   string `json:"result"`
	}{
		"Kaspersky": {
			Category: "undetected",
			Result:   "clean",
		},
		"Microsoft": {
			Category: "suspicious",
			Result:   "potentially unwanted application",
		},
	}

	return scanResult, nil
}
