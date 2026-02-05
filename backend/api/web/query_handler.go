package api

import (
	"net/http"
	"time"

	"railgun-core/internal/domain"

	repository "railgun-core/internal/domain/repository"

	"github.com/gin-gonic/gin"
)

type QueryHandler struct {
	trafficRepo   repository.TrafficRepository
	analyticsRepo domain.AnalyticsRepository
}

func NewQueryHandler(tr repository.TrafficRepository, ar domain.AnalyticsRepository) *QueryHandler {
	return &QueryHandler{
		trafficRepo:   tr,
		analyticsRepo: ar,
	}
}

func (h *QueryHandler) GetTrafficByHost(c *gin.Context) {
	hostID := c.Param("hostId")
	from, to := parseTimeRange(c)

	traffic, err := h.trafficRepo.GetTrafficByHost(c.Request.Context(), hostID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, traffic)
}

func (h *QueryHandler) GetThreatHeatmap(c *gin.Context) {
	from, to := parseTimeRange(c)
	data, err := h.analyticsRepo.GetThreatHeatmap(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// Вспомогательная функция для парсинга времени
func parseTimeRange(c *gin.Context) (time.Time, time.Time) {
	from, _ := time.Parse(time.RFC3339, c.DefaultQuery("from", time.Now().Add(-24*time.Hour).Format(time.RFC3339)))
	to, _ := time.Parse(time.RFC3339, c.DefaultQuery("to", time.Now().Format(time.RFC3339)))
	return from, to
}
