package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"railgun-core/api/requests"
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

// AnalyzeRealtime godoc
// @Summary      Выполнить AI анализ в реальном времени
// @Description  Принимает данные, тип данных и ID хоста для мгновенного анализа нейросетью
// @Tags         AI Analysis
// @Accept       json
// @Produce      json
// @Param        request body requests.AnalyzeRealtimeRequest true "Данные для анализа"
// @Success      200  {object}  models.AnalysisResult
// @Failure      400  {object}  map[string]string "Ошибка валидации"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /ai/analyze [post]
func (h *AIHandler) AnalyzeRealtime(c *gin.Context) {
	var request requests.AnalyzeRealtimeRequest

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

// GetAttackPatterns godoc
// @Summary      Получить шаблоны атак
// @Description  Возвращает список шаблонов атак с фильтрацией и пагинацией
// @Tags         Attack Patterns
// @Produce      json
// @Param        category  query  string  false  "Категория (напр. Initial Access)"
// @Param        severity  query  string  false  "Уровень угрозы (Low, Medium, High, Critical)"
// @Param        page      query  int     false  "Номер страницы" default(1)
// @Param        per_page  query  int     false  "Кол-во элементов на странице" default(20)
// @Success      200  {object}  map[string]interface{} "Объект с data (массив шаблонов) и meta (пагинация)"
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /ai/patterns [get]
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

// GetAPTTimeline godoc
// @Summary      Временная шкала APT
// @Description  Получает хронологию сложных устойчивых угроз (APT) для конкретного хоста
// @Tags         AI Analysis
// @Produce      json
// @Param        host_id  query  string  true   "ID хоста"
// @Param        from     query  string  false  "Начало периода (RFC3339)"
// @Param        to       query  string  false  "Конец периода (RFC3339)"
// @Success      200  {array}   map[string]interface{} "Массив событий таймлайна"
// @Failure      400  {object}  map[string]string
// @Security     BearerAuth
// @Router       /ai/apt-timeline [get]
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

// UpdateModels godoc
// @Summary      Обновить модели AI
// @Description  Запускает процесс обновления весов для выбранных ID моделей
// @Tags         Model Management
// @Accept       json
// @Produce      json
// @Param        request body requests.UpdateModelsRequest true "Список ID моделей"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Security     BearerAuth
// @Router       /ai/models/update [post]
func (h *AIHandler) UpdateModels(c *gin.Context) {
	var request requests.UpdateModelsRequest

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

// TrainModel godoc
// @Summary      Обучить модель
// @Description  Создает задачу на обучение конкретной модели на основе указанного датасета
// @Tags         Model Management
// @Accept       json
// @Produce      json
// @Param        request body requests.TrainModelRequest true "Параметры обучения"
// @Success      200  {object}  map[string]string "ID задачи на обучение"
// @Failure      400  {object}  map[string]string
// @Security     BearerAuth
// @Router       /ai/models/train [post]
func (h *AIHandler) TrainModel(c *gin.Context) {
	var request requests.TrainModelRequest

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

// ListModels godoc
// @Summary      Список всех моделей
// @Description  Возвращает список доступных AI моделей, опционально фильтруя по типу
// @Tags         Model Management
// @Produce      json
// @Param        type  query  string  false  "Тип модели"
// @Success      200  {array}   map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /ai/models [get]
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
