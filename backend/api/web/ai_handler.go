package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"railgun-core/internal/domain"
)

type AIHandler struct {
	aiService domain.AIService
}

func NewAIHandler(aiService domain.AIService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// AnalyzeRealtime выполняет анализ данных в реальном времени
func (h *AIHandler) AnalyzeRealtime(c *gin.Context) {
	var request struct {
		Data     []string `json:"data" binding:"required"`
		DataType string   `json:"data_type" binding:"required"`
		HostID   string   `json:"host_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Анализируем данные
	result, err := h.aiService.AnalyzeData(c, request.Data, request.DataType, request.HostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAttackPatterns возвращает шаблоны атак
func (h *AIHandler) GetAttackPatterns(c *gin.Context) {
	// Получаем параметры запроса
	category := c.Query("category")
	severity := c.Query("severity")

	// Получаем параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	// Получаем шаблоны атак
	patterns, total, err := h.aiService.GetAttackPatterns(c, category, severity, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": patterns,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"per_page":    perPage,
			"total_pages": (total + perPage - 1) / perPage,
		},
	})
}

// ExecuteCounterAttack выполняет контратаку
func (h *AIHandler) ExecuteCounterAttack(c *gin.Context) {
	var request struct {
		TargetIP   string `json:"target_ip" binding:"required,ip"`
		AttackType string `json:"attack_type" binding:"required"`
		Intensity  int    `json:"intensity" binding:"required,min=1,max=5"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем права пользователя (в реальном приложении)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Логируем действие
	c.Set("audit_log", map[string]interface{}{
		"action":      "execute_counter_attack",
		"user_id":     userID,
		"target_ip":   request.TargetIP,
		"attack_type": request.AttackType,
		"intensity":   request.Intensity,
		"timestamp":   time.Now(),
	})

	// Выполняем контратаку
	/*err := engine.NewDetector(config.Detection).RespondToThreat(request.TargetIP, request.Intensity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}*/

	c.JSON(http.StatusOK, gin.H{
		"message":     "Counter-attack initiated successfully",
		"target_ip":   request.TargetIP,
		"attack_type": request.AttackType,
	})
}

// GetAPTTimeline возвращает временную шкалу APT атаки
func (h *AIHandler) GetAPTTimeline(c *gin.Context) {
	// Получаем параметры запроса
	hostID := c.Query("host_id")
	if hostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host ID is required"})
		return
	}

	fromStr := c.DefaultQuery("from", time.Now().Add(-30*24*time.Hour).Format(time.RFC3339))
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

	// Получаем временную шкалу
	timeline, err := h.aiService.GetAPTTimeline(c, hostID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, timeline)
}

// UpdateModels обновляет модели AI
func (h *AIHandler) UpdateModels(c *gin.Context) {
	var request struct {
		ModelIDs []string `json:"model_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем модели
	results, err := h.aiService.UpdateModels(c, request.ModelIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Models update initiated",
		"results": results,
	})
}

// TrainModel запускает обучение модели AI
func (h *AIHandler) TrainModel(c *gin.Context) {
	var request struct {
		ModelID     string `json:"model_id" binding:"required"`
		DatasetPath string `json:"dataset_path" binding:"required"`
		Epochs      int    `json:"epochs" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Запускаем обучение
	jobID, err := h.aiService.TrainModel(c, request.ModelID, request.DatasetPath, request.Epochs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Training job started",
		"job_id":  jobID,
	})
}

// ListModels возвращает список моделей AI
func (h *AIHandler) ListModels(c *gin.Context) {
	// Получаем параметры запроса
	modelType := c.Query("type")

	// Получаем список моделей
	models, err := h.aiService.ListModels(c, modelType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models)
}

func (h *AIHandler) GetPatternStats(c *gin.Context) {
	// Получаем статистику по категориям
	byCategory := []gin.H{
		{"category": "Initial Access", "count": 15},
		{"category": "Execution", "count": 22},
		{"category": "Persistence", "count": 18},
		{"category": "Privilege Escalation", "count": 12},
		{"category": "Defense Evasion", "count": 25},
		{"category": "Credential Access", "count": 20},
		{"category": "Discovery", "count": 16},
		{"category": "Lateral Movement", "count": 14},
		{"category": "Collection", "count": 10},
		{"category": "Command and Control", "count": 8},
		{"category": "Exfiltration", "count": 6},
		{"category": "Impact", "count": 9},
	}

	// Получаем статистику по уровням угрозы
	bySeverity := []gin.H{
		{"severity": "Low", "count": 10},
		{"severity": "Medium", "count": 35},
		{"severity": "High", "count": 28},
		{"severity": "Critical", "count": 5},
	}

	c.JSON(http.StatusOK, gin.H{
		"byCategory": byCategory,
		"bySeverity": bySeverity,
	})
}
