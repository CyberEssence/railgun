package handlers

import (
	"net/http"
	"railgun-core/internal/infrastructure/services"

	"github.com/gin-gonic/gin"
)

type VirusTotalHandler struct {
	integrationSvc *services.IntegrationService
}

func NewVirusTotalHandler(integrationSvc *services.IntegrationService) *VirusTotalHandler {
	return &VirusTotalHandler{
		integrationSvc: integrationSvc,
	}
}

func (h *VirusTotalHandler) HandleFileScan(c *gin.Context) {
	// Проверка размера файла
	if c.Request.ContentLength > 32<<20 { // 32MB
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.integrationSvc.ScanWithVirusTotal(c.Request.Context(), fileHeader)
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
