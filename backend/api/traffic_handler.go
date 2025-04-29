package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"railgun-core/internal/domain"
	"railgun-core/internal/models"
)

type TrafficHandler struct {
	trafficRepo domain.TrafficRepository
}

func NewTrafficHandler(trafficRepo domain.TrafficRepository) *TrafficHandler {
	return &TrafficHandler{
		trafficRepo: trafficRepo,
	}
}

// GetTrafficByHost возвращает трафик для указанного хоста
func (h *TrafficHandler) GetTrafficByHost(c *gin.Context) {
	hostID := c.Param("hostId")
	if hostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host ID is required"})
		return
	}

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

	// Получаем трафик
	traffic, err := h.trafficRepo.GetTrafficByHost(c, hostID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, traffic)
}

// GetTrafficStats возвращает статистику трафика для указанного хоста
func (h *TrafficHandler) GetTrafficStats(c *gin.Context) {
	hostID := c.Param("hostId")
	if hostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host ID is required"})
		return
	}

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

	// Получаем статистику
	stats, err := h.trafficRepo.GetTrafficStats(c, hostID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SaveTraffic сохраняет запись о трафике
func (h *TrafficHandler) SaveTraffic(c *gin.Context) {
	var traffic models.NetworkTraffic
	if err := c.ShouldBindJSON(&traffic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем время, если не указано
	if traffic.Timestamp.IsZero() {
		traffic.Timestamp = time.Now()
	}

	// Сохраняем трафик
	err := h.trafficRepo.SaveTraffic(c, traffic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Traffic saved successfully", "id": traffic.ID})
}

// ProcessNetworkLog обрабатывает лог сетевого трафика
func (h *TrafficHandler) ProcessNetworkLog(c *gin.Context) {
	var logRequest struct {
		HostID  string `json:"host_id" binding:"required"`
		LogData string `json:"log_data" binding:"required"`
		LogType string `json:"log_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&logRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обрабатываем лог
	traffic, err := h.trafficRepo.ProcessNetworkLog(c, logRequest.HostID, logRequest.LogData, logRequest.LogType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Log processed successfully",
		"traffic_entries": len(traffic),
		"data":            traffic,
	})
}

// IsolateHost изолирует хост от сети
func (h *TrafficHandler) IsolateHost(c *gin.Context) {
	var request struct {
		HostID   string `json:"host_id" binding:"required"`
		Reason   string `json:"reason" binding:"required"`
		Duration int    `json:"duration"` // в минутах, 0 = бессрочно
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Изолируем хост
	err := h.trafficRepo.IsolateHost(c, request.HostID, request.Reason, request.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Host isolated successfully",
		"host_id":  request.HostID,
		"duration": request.Duration,
	})
}

// GetThreatHeatmap возвращает тепловую карту угроз
func (h *TrafficHandler) GetThreatHeatmap(c *gin.Context) {
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

	// Получаем данные для тепловой карты
	heatmapData, err := h.trafficRepo.GetThreatHeatmap(c, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, heatmapData)
}
