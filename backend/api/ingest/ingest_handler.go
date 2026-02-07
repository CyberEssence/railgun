package api

import (
	"context"
	"net/http"
	"time"

	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"
	repository "railgun-core/internal/domain/repository"

	"github.com/gin-gonic/gin"
)

type IngestHandler struct {
	trafficRepo    repository.TrafficRepository
	networkLogRepo domain.NetworkLogRepository
	engine         domain.DetectionEngine // Добавляем движок для анализа на лету
}

func NewIngestHandler(tr repository.TrafficRepository, nl domain.NetworkLogRepository, de domain.DetectionEngine) *IngestHandler {
	return &IngestHandler{
		trafficRepo:    tr,
		networkLogRepo: nl,
		engine:         de,
	}
}

// SaveTraffic — прием структурированных данных от агентов
func (h *IngestHandler) SaveTraffic(c *gin.Context) {
	var traffic models.NetworkTraffic
	if err := c.ShouldBindJSON(&traffic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if traffic.Timestamp.IsZero() {
		traffic.Timestamp = time.Now()
	}

	// 1. Сохраняем в БД
	if err := h.trafficRepo.SaveTraffic(c.Request.Context(), traffic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. СРАЗУ отправляем в Engine для корреляции (асинхронно)
	go h.engine.AddEvent(context.Background(), models.EventCorrelation{
		Type:      "network_flow",
		SourceIP:  traffic.SrcIP,
		HostID:    traffic.HostID,
		Timestamp: traffic.Timestamp,
		Success:   true,
	})

	c.JSON(http.StatusCreated, gin.H{"status": "accepted"})
}

// ProcessNetworkLog — прием сырых логов (syslog, и т.д.)
func (h *IngestHandler) ProcessNetworkLog(c *gin.Context) {
	var logReq struct {
		HostID  string `json:"host_id" binding:"required"`
		LogData string `json:"log_data" binding:"required"`
		LogType string `json:"log_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&logReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	traffic, err := h.networkLogRepo.ProcessNetworkLog(c.Request.Context(), logReq.HostID, logReq.LogData, logReq.LogType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entries_processed": len(traffic)})
}

// IsolateHost изолирует хост от сети
func (h *IngestHandler) IsolateHost(c *gin.Context) {
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

// GetTrafficStats возвращает статистику трафика для указанного хоста
func (h *IngestHandler) GetTrafficStats(c *gin.Context) {
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
