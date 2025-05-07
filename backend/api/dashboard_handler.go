package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"railgun-core/internal/domain"
)

type DashboardHandler struct {
	trafficRepo domain.TrafficRepository
	aiService   domain.AIService
}

func NewDashboardHandler(trafficRepo domain.TrafficRepository, aiService domain.AIService) *DashboardHandler {
	return &DashboardHandler{
		trafficRepo: trafficRepo,
		aiService:   aiService,
	}
}

// GetDashboardStats возвращает статистику для дашборда
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

	// Получаем статистику трафика
	trafficStats, err := h.trafficRepo.GetDashboardStats(c, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get traffic stats: " + err.Error()})
		return
	}

	// Получаем статистику угроз
	threatStats, err := h.aiService.GetThreatStats(c, from, to)
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
