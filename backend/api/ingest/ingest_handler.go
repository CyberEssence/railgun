package api

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"
	repository "railgun-core/internal/domain/repository"

	requests "railgun-core/api/requests"

	"github.com/gin-gonic/gin"
)

type AgentMonitor struct {
	agents map[string]*AgentStatus
	mu     sync.RWMutex
}

type AgentStatus struct {
	HostID      string    `json:"host_id"`
	Hostname    string    `json:"hostname"`
	LastSeen    time.Time `json:"last_seen"`
	MetricsSent int64     `json:"metrics_sent"`
	LastError   string    `json:"last_error,omitempty"`
	Online      bool      `json:"online"`
	Version     string    `json:"version,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	AgentType   string    `json:"agent_type"`
}

func NewAgentMonitor() *AgentMonitor {
	return &AgentMonitor{
		agents: make(map[string]*AgentStatus),
	}
}

type IngestHandler struct {
	trafficRepo    repository.TrafficRepository
	networkLogRepo domain.NetworkLogRepository
	engine         domain.DetectionEngine
	agentMonitor   *AgentMonitor
}

func NewIngestHandler(
	tr repository.TrafficRepository,
	nl domain.NetworkLogRepository,
	de domain.DetectionEngine,
	monitor *AgentMonitor) *IngestHandler {

	return &IngestHandler{
		trafficRepo:    tr,
		networkLogRepo: nl,
		engine:         de,
		agentMonitor:   monitor,
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
		// Если тело запроса пустое (EOF), возвращаем понятную ошибку
		if errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request body is empty"})
			return
		}
		// Остальные ошибки парсинга JSON
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if traffic.SrcIP == "" || traffic.DstIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "src_ip and dst_ip are required"})
		return
	}

	if traffic.Timestamp.IsZero() {
		traffic.Timestamp = time.Now()
	}

	// Сохраняем в БД
	if err := h.trafficRepo.SaveTraffic(c.Request.Context(), traffic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go func(t models.NetworkTraffic) {
		h.engine.AddEvent(context.Background(), models.EventCorrelation{
			Type:      "network_flow",
			SourceIP:  t.SrcIP,
			HostID:    t.HostID,
			Timestamp: t.Timestamp,
			Success:   true,
		})
	}(traffic)

	c.JSON(http.StatusCreated, gin.H{"status": "accepted"})
}

type NetworkLogRequest struct {
	HostID  string `json:"host_id" binding:"required"`
	LogType string `json:"log_type" binding:"required"`
	LogData string `json:"log_data" binding:"required"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Вызываем метод репозитория
	err := h.trafficRepo.IsolateHost(c.Request.Context(), request.HostID, request.Reason, request.Duration)
	if err != nil {
		// Обрабатываем ошибку "хост не найден" отдельно
		if errors.Is(err, repository.ErrHostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to isolate host: " + err.Error()})
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

	stats, err := h.trafficRepo.GetTrafficStats(c.Request.Context(), hostID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *IngestHandler) ProcessAgentData(c *gin.Context) {
	var payload struct {
		HostID string        `json:"host_id"`
		Events []interface{} `json:"events"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		// Обрабатываем случай пустого тела запроса
		if errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request body is empty"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем и проверяем, что массив не пуст
	log.Printf("Received %d events from host %s", len(payload.Events), payload.HostID)

	if len(payload.Events) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "accepted", "message": "no events to process"})
		return
	}

	// Обработка разных типов событий
	for _, event := range payload.Events {
		eventMap, ok := event.(map[string]interface{})
		if !ok {
			continue
		}

		eventType, _ := eventMap["type"].(string)

		switch eventType {
		case "system_info":
			h.processSystemInfo(eventMap, payload.HostID)
		case "network_info":
			h.processNetworkInfo(eventMap, payload.HostID)
		case "security_info":
			h.processSecurityInfo(eventMap, payload.HostID)
		case "process_info":
			h.processProcessInfo(eventMap, payload.HostID)
		}
	}

	// Анализ в реальном времени (безопасный запуск)
	// Создаем копию слайса, чтобы избежать гонки данных, если слайс изменится извне
	eventsCopy := make([]interface{}, len(payload.Events))
	copy(eventsCopy, payload.Events)

	go h.analyzeRealtime(payload.HostID, eventsCopy)

	c.JSON(http.StatusOK, gin.H{"status": "processed", "count": len(payload.Events)})
}

func (h *IngestHandler) processSystemInfo(data map[string]interface{}, hostID string) {
	// Сохранение в Elasticsearch
	//indexName := "system-metrics-" + time.Now().Format("2006.01.02")

	doc := map[string]interface{}{
		"@timestamp": data["timestamp"],
		"host": map[string]interface{}{
			"id":   hostID,
			"name": data["hostname"],
		},
		"system": data,
		"event": map[string]interface{}{
			"type":   "metrics",
			"module": "system",
		},
	}

	// Сохранение в БД SIEM
	h.saveToDatabase("system_metrics", doc)

	// Проверка на аномалии
	h.checkForAnomalies(doc)
}

func (m *AgentMonitor) UpdateStatus(hostID, hostname string, metricsSent int, agentInfo ...map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.agents[hostID]
	if !exists {
		status = &AgentStatus{
			HostID:    hostID,
			Hostname:  hostname,
			LastSeen:  time.Now(),
			Online:    true,
			AgentType: "linux",
		}
		m.agents[hostID] = status
	}

	status.LastSeen = time.Now()
	status.MetricsSent += int64(metricsSent)
	status.Online = true

	if len(agentInfo) > 0 {
		if ip, ok := agentInfo[0]["ip"]; ok {
			status.IPAddress = ip
		}
		if version, ok := agentInfo[0]["version"]; ok {
			status.Version = version
		}
	}
}

func (m *AgentMonitor) SetError(hostID, errorMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if status, exists := m.agents[hostID]; exists {
		status.LastError = errorMsg
	}
}

func (m *AgentMonitor) SetOffline(hostID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if status, exists := m.agents[hostID]; exists {
		status.Online = false
	}
}

func (m *AgentMonitor) GetAllStatus() []*AgentStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*AgentStatus, 0, len(m.agents))
	for _, status := range m.agents {
		// Проверяем, если агент не активен более 5 минут
		if time.Since(status.LastSeen) > 5*time.Minute {
			status.Online = false
		}
		result = append(result, status)
	}

	return result
}

func (m *AgentMonitor) GetAgentStatus(hostID string) (*AgentStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.agents[hostID]
	return status, exists
}

func (h *IngestHandler) GetAgentsStatus(c *gin.Context) {
	status := h.agentMonitor.GetAllStatus()
	c.JSON(http.StatusOK, gin.H{
		"agents": status,
		"total":  len(status),
		"online": countOnline(status),
	})
}

func countOnline(agents []*AgentStatus) int {
	count := 0
	for _, agent := range agents {
		if agent.Online {
			count += 1
		}
	}
	return count
}

func (h *IngestHandler) saveToDatabase(indexName string, doc map[string]interface{}) {
	// Реализация сохранения в базу данных (Elasticsearch, PostgreSQL и т.д.)
	log.Printf("Saving to %s: %v", indexName, doc)
}

func (h *IngestHandler) checkForAnomalies(doc map[string]interface{}) {
	// Проверка на аномалии в данных
	// Например, проверка значений CPU, памяти, диска
	// Можно интегрировать с h.engine
}

func (h *IngestHandler) processNetworkInfo(data map[string]interface{}, hostID string) {
	// Обработка сетевой информации
	doc := map[string]interface{}{
		"@timestamp": data["timestamp"],
		"host": map[string]interface{}{
			"id": hostID,
		},
		"network": data,
		"event": map[string]interface{}{
			"type":   "metrics",
			"module": "network",
		},
	}
	h.saveToDatabase("network_metrics", doc)
}

func (h *IngestHandler) processSecurityInfo(data map[string]interface{}, hostID string) {
	// Обработка информации о безопасности
	doc := map[string]interface{}{
		"@timestamp": data["timestamp"],
		"host": map[string]interface{}{
			"id": hostID,
		},
		"security": data,
		"event": map[string]interface{}{
			"type":   "security",
			"module": "threat_intel",
		},
	}
	h.saveToDatabase("security_events", doc)
}

func (h *IngestHandler) processProcessInfo(data map[string]interface{}, hostID string) {
	// Обработка информации о процессах
	doc := map[string]interface{}{
		"@timestamp": data["timestamp"],
		"host": map[string]interface{}{
			"id": hostID,
		},
		"process": data,
		"event": map[string]interface{}{
			"type":   "process",
			"module": "system",
		},
	}
	h.saveToDatabase("process_events", doc)
}

func (h *IngestHandler) analyzeRealtime(hostID string, events []interface{}) {
	// Желательно использовать контекст с таймаутом для фоновых задач
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, event := range events {
		eventMap, ok := event.(map[string]interface{})
		if !ok {
			continue
		}

		// Извлекаем timestamp если есть, иначе ставим текущий
		timestamp := time.Now()
		if ts, ok := eventMap["timestamp"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				timestamp = t
			}
		}

		h.engine.AddEvent(ctx, models.EventCorrelation{
			Type:      "agent_data",
			HostID:    hostID,
			Timestamp: timestamp,
			Data:      eventMap,
		})
	}
}
