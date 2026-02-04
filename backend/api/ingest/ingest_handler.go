package api

import (
	"context"
	"net/http"
	"time"

	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"

	"github.com/gin-gonic/gin"
)

type IngestHandler struct {
	trafficRepo    domain.TrafficRepository
	networkLogRepo domain.NetworkLogRepository
	engine         domain.DetectionEngine // Добавляем движок для анализа на лету
}

func NewIngestHandler(tr domain.TrafficRepository, nl domain.NetworkLogRepository, de domain.DetectionEngine) *IngestHandler {
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
