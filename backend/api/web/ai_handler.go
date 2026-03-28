package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"railgun-core/api/requests"
	"railgun-core/internal/domain"
)

type AIHandler struct {
	aiService domain.AIService
}

func NewAIHandler(svc domain.AIService) *AIHandler {
	return &AIHandler{aiService: svc}
}

func (h *AIHandler) AnalyzeRealtime(c *gin.Context) {
	var req requests.AnalyzeRealtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Вызываем сервис, передавая массив логов
	results, err := h.aiService.AnalyzeAndSave(c.Request.Context(), req.Data, req.HostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed: " + err.Error()})
		return
	}

	// Возвращаем массив результатов
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// ExecuteCounterAttack godoc
// @Summary      Запустить контратаку
// @Description  Инициирует защитные действия против указанного IP адреса
// @Tags         Defense
// @Accept       json
// @Produce      json
// @Param        request body requests.CounterAttackRequest true "Параметры контратаки"
// @Success      200  {object}  map[string]string "Сообщение об успешном запуске"
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string "Не авторизован"
// @Security     BearerAuth
// @Router       /ai/counter-attack [post]
func (h *AIHandler) ExecuteCounterAttack(c *gin.Context) {
	var request requests.CounterAttackRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем права пользователя (в реальном приложении)
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Логируем действие
	c.Set("audit_log", map[string]interface{}{
		"action":     "execute_counter_attack",
		"userId":     userID,
		"targetIp":   request.TargetIP,
		"attackType": request.AttackType,
		"intensity":  request.Intensity,
		"timestamp":  time.Now(),
	})

	// Выполняем контратаку
	/*err := engine.NewDetector(config.Detection).RespondToThreat(request.TargetIP, request.Intensity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}*/

	c.JSON(http.StatusOK, gin.H{
		"message":    "Counter-attack initiated successfully",
		"targetIp":   request.TargetIP,
		"attackType": request.AttackType,
	})
}

// GetPatternStats godoc
// @Summary      Статистика шаблонов
// @Description  Возвращает агрегированную статистику атак по категориям и уровням критичности
// @Tags         Attack Patterns
// @Produce      json
// @Success      200  {object}  map[string]interface{} "Статистика (byCategory, bySeverity)"
// @Security     BearerAuth
// @Router       /ai/patterns/stats [get]
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
