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
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Открываем файл, чтобы получить io.Reader
	fileStream, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileStream.Close()

	// Передаем в сервис: контекст, поток (Reader) и размер (int64)
	result, err := h.integrationService.ScanWithVirusTotal(c.Request.Context(), fileStream, fileHeader.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
