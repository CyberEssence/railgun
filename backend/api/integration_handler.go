package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"railgun-core/internal/domain"
)

type IntegrationHandler struct {
	integrationService domain.IntegrationService
}

func NewIntegrationHandler(integrationService domain.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
	}
}

// ScanFile сканирует файл с помощью внешних сервисов
func (h *IntegrationHandler) ScanFile(c *gin.Context) {
	// Получаем файл из запроса
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Сканируем файл с помощью VirusTotal
	result, err := h.integrationService.ScanWithVirusTotal(c, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
