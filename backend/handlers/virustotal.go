package handlers

import (
	"net/http"                     // Добавьте импорт
	"railgun-core/internal/domain" // Импортируем домен для интерфейса

	"github.com/gin-gonic/gin"
)

//дубль api/integration_handler.go, потом удалить

type VirusTotalHandler struct {
	// Используем интерфейс из domain, а не конкретный сервис из infrastructure
	integrationSvc domain.IntegrationService
}

func NewVirusTotalHandler(integrationSvc domain.IntegrationService) *VirusTotalHandler {
	return &VirusTotalHandler{
		integrationSvc: integrationSvc,
	}
}

func (h *VirusTotalHandler) HandleFileScan(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Открываем файл для получения io.Reader
	fileStream, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer fileStream.Close()

	// 2. Передаем контекст, поток данных и размер (3 аргумента)
	result, err := h.integrationSvc.ScanWithVirusTotal(c.Request.Context(), fileStream, fileHeader.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "scan failed",
			"detail": err.Error(),
		})
		return
	}

	response := gin.H{
		"scan_id":   result.Data.ID,
		"malicious": result.Data.Attributes.Stats.Malicious,
		"status":    result.Data.Attributes.Status,
	}

	if result.Data.Attributes.Status == "completed" {
		response["results"] = result.Data.Attributes.Results
	}

	c.JSON(http.StatusOK, response)
}
