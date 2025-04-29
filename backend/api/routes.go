package api

import (
	"github.com/gin-gonic/gin"

	"railgun-core/internal/domain"
	"railgun-core/internal/infrastructure/persistence"
)

// RegisterRoutes регистрирует все маршруты API
func RegisterRoutes(
	r *gin.Engine,
	cfg *domain.Config,
	trafficRepo domain.TrafficRepository,
	artifactRepo domain.ArtifactRepository,
	aiService domain.AIService,
	integrationService domain.IntegrationService,
	twoFAService domain.TwoFAService,
	userRepo *persistence.UserRepository,
) {
	// Создаем обработчики
	authHandler := NewAuthHandler(cfg, twoFAService, userRepo)
	trafficHandler := NewTrafficHandler(trafficRepo)
	artifactHandler := NewArtifactHandler(artifactRepo)
	aiHandler := NewAIHandler(aiService)
	integrationHandler := NewIntegrationHandler(integrationService)
	dashboardHandler := NewDashboardHandler(trafficRepo, aiService)

	// Группа для аутентификации
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/verify-2fa", authHandler.Verify2FA)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Группа для API с аутентификацией
	api := r.Group("/api")
	api.Use(authHandler.AuthMiddleware())
	{
		// Трафик
		api.GET("/traffic/:hostId", trafficHandler.GetTrafficByHost)
		api.GET("/traffic/stats/:hostId", trafficHandler.GetTrafficStats)
		api.POST("/traffic", trafficHandler.SaveTraffic)
		api.POST("/traffic/logs", trafficHandler.ProcessNetworkLog)
		api.POST("/traffic/isolate", trafficHandler.IsolateHost)
		api.GET("/traffic/heatmap", trafficHandler.GetThreatHeatmap)

		// Артефакты
		api.GET("/artifacts/:hostId", artifactHandler.GetArtifactsByHost)
		api.GET("/artifacts/id/:id", artifactHandler.GetArtifactByID)
		api.POST("/artifacts", artifactHandler.SaveArtifact)
		api.GET("/artifacts/search", artifactHandler.SearchArtifacts)

		// AI и анализ
		api.POST("/ai/analyze", aiHandler.AnalyzeRealtime)
		api.GET("/ai/patterns", aiHandler.GetAttackPatterns)
		api.POST("/ai/counter-attack", aiHandler.ExecuteCounterAttack)
		api.GET("/ai/apt-timeline", aiHandler.GetAPTTimeline)
		api.POST("/ai/models/update", aiHandler.UpdateModels)
		api.POST("/ai/models/train", aiHandler.TrainModel)
		api.GET("/ai/models", aiHandler.ListModels)

		// Интеграции
		api.POST("/integration/scan", integrationHandler.ScanFile)

		// Дашборд
		api.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
	}

	// Публичные эндпоинты
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
