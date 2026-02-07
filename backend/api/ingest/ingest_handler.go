package api

import (
	"context"
	"net/http"
	"time"

	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"
	repository "railgun-core/internal/domain/repository"

	requests "railgun-core/api/requests"

	"github.com/gin-gonic/gin"
)

type IngestHandler struct {
	trafficRepo    repository.TrafficRepository
	networkLogRepo domain.NetworkLogRepository
	engine         domain.DetectionEngine
}

func NewIngestHandler(tr repository.TrafficRepository, nl domain.NetworkLogRepository, de domain.DetectionEngine) *IngestHandler {
	return &IngestHandler{
		trafficRepo:    tr,
		networkLogRepo: nl,
		engine:         de,
	}
}

// SaveTraffic godoc
// @Summary      Прием структурированного трафика
// @Description  Принимает JSON с сетевой активностью от агентов, сохраняет в БД и отправляет в корреляционный движок
// @Tags         Ingestion
// @Accept       json
// @Produce      json
// @Param        traffic  body      models.NetworkTraffic  true  "Данные сетевого трафика"
// @Success      201      {object}  map[string]string      "status: accepted"
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /ingest/traffic [post]
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

type NetworkLogRequest struct {
	HostID  string `json:"host_id" binding:"required"`
	LogData string `json:"log_data" binding:"required"`
	LogType string `json:"log_type" binding:"required"`
}

// ProcessNetworkLog godoc
// @Summary      Прием сырых логов
// @Description  Принимает неструктурированные логи (syslog и др.), парсит их и преобразует в записи трафика
// @Tags         Ingestion
// @Accept       json
// @Produce      json
// @Param        log_request body NetworkLogRequest true "Сырые логи и метаданные"
// @Success      200      {object}  map[string]int     "Количество обработанных записей (entries_processed)"
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /ingest/logs [post]
func (h *IngestHandler) ProcessNetworkLog(c *gin.Context) {
	var logReq NetworkLogRequest

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

// IsolateHost godoc
// @Summary      Изоляция хоста
// @Description  Блокирует сетевую активность хоста по его ID на указанное время
// @Tags         Response
// @Accept       json
// @Produce      json
// @Param        request body requests.IsolateHostRequest true "Параметры изоляции (duration в минутах)"
// @Success      200      {object}  map[string]interface{} "Подтверждение изоляции"
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /ingest/isolate [post]
func (h *IngestHandler) IsolateHost(c *gin.Context) {
	var request requests.IsolateHostRequest

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

// GetTrafficStats godoc
// @Summary      Статистика трафика хоста
// @Description  Возвращает агрегированную статистику (объемы, протоколы) для конкретного хоста за период
// @Tags         Analytics
// @Produce      json
// @Param        hostId   path      string  true   "ID хоста"
// @Param        from     query     string  false  "Начало (RFC3339)"
// @Param        to       query     string  false  "Конец (RFC3339)"
// @Success      200      {object}  models.TrafficStats
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /ingest/stats/{hostId} [get]
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
