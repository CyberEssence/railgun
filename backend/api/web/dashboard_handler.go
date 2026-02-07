package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"
	repository "railgun-core/internal/domain/repository"
)

type DashboardHandler struct {
	trafficRepo   repository.TrafficRepository
	analyticsRepo domain.AnalyticsRepository
	aiRepo        domain.AIService
}

func NewDashboardHandler(trafficRepo repository.TrafficRepository, aiRepo domain.AIService) *DashboardHandler {
	return &DashboardHandler{
		trafficRepo: trafficRepo,
		aiRepo:      aiRepo,
	}
}

// GetDashboardStats godoc
// @Summary      Получить статистику дашборда
// @Description  Возвращает агрегированные данные по трафику и угрозам за указанный период времени
// @Tags         Dashboard
// @Produce      json
// @Param        from  query     string  false  "Начало периода (RFC3339, например: 2023-01-01T00:00:00Z)"
// @Param        to    query     string  false  "Конец периода (RFC3339)"
// @Success      200   {object}  map[string]interface{} "Статистика трафика, угроз и подтвержденный временной диапазон"
// @Failure      400   {object}  map[string]string      "Неверный формат даты"
// @Failure      500   {object}  map[string]string      "Ошибка при получении данных из репозиториев"
// @Security     BearerAuth
// @Router       /dashboard/stats [get]
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	// Получаем параметры запроса
	fromStr := c.DefaultQuery("from", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	toStr := c.DefaultQuery("to", time.Now().Format(time.RFC3339))

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' parameter"})
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' parameter"})
		return
	}

	threatStats, err := h.aiRepo.GetThreatStats(c, from, to)
	if err != nil {
		// Если таблицы нет, возвращаем нулевую статистику вместо ошибки
		if strings.Contains(err.Error(), "отношение \"threats\" не существует") {
			threatStats = &models.ThreatStats{}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get threat stats",
				"details": err.Error(),
			})
			return
		}
	}

	// Получаем статистику трафика
	trafficStats, err := h.analyticsRepo.GetDashboardStats(c, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get traffic stats: " + err.Error()})
		return
	}

	// Получаем статистику угроз
	threatStats, err = h.aiRepo.GetThreatStats(c, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get threat stats: " + err.Error()})
		return
	}

	// Объединяем статистику
	stats := map[string]interface{}{
		"traffic": trafficStats,
		"threats": threatStats,
		"time_range": map[string]string{
			"from": from.Format(time.RFC3339),
			"to":   to.Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, stats)
}
