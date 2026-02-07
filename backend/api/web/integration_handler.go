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

// ScanFile godoc
// @Summary      Сканировать файл через VirusTotal
// @Description  Загружает файл и отправляет его на проверку во внешние антивирусные сервисы
// @Tags         Integrations
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Бинарный файл для сканирования"
// @Success      200   {object}  models.ScanResult
// @Failure      400   {object}  map[string]string "Файл не предоставлен"
// @Failure      500   {object}  map[string]string "Ошибка сканирования"
// @Security     BearerAuth
// @Router       /integrations/scan [post]
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
