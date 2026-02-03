package api

import (
	"github.com/gin-gonic/gin"

	"railgun-core/internal/config"
	"railgun-core/internal/domain"
	"railgun-core/internal/infrastructure/persistence"
)

func RegisterRoutes(
	r *gin.Engine,
	cfg *config.Config,
	trafficRepo domain.TrafficRepository,
	artifactRepo domain.ArtifactRepository,
	aiService domain.AIService,
	integrationService domain.IntegrationService,
	twoFAService domain.TwoFAService,
	userRepo *persistence.UserRepository,
) {
	// Инициализация обработчиков
	authHandler := NewAuthHandler(cfg, twoFAService, userRepo)
	aiHandler := NewAIHandler(aiService)
	artifactHandler := NewArtifactHandler(artifactRepo)
	integrationHandler := NewIntegrationHandler(integrationService)
	dashboardHandler := NewDashboardHandler(trafficRepo, aiService)
	trafficHandler := NewTrafficHandler(trafficRepo)

	// Группа аутентификации
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/verify-2fa", authHandler.Verify2FA)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}

	// Группа API с аутентификацией
	apiGroup := r.Group("/api")
	apiGroup.Use(authHandler.AuthMiddleware())
	{
		// Трафик
		apiGroup.GET("/traffic/:hostId", trafficHandler.GetTrafficByHost)
		apiGroup.GET("/traffic/stats/:hostId", trafficHandler.GetTrafficStats)
		apiGroup.POST("/traffic", trafficHandler.SaveTraffic)
		apiGroup.POST("/traffic/logs", trafficHandler.ProcessNetworkLog)
		apiGroup.POST("/traffic/isolate", trafficHandler.IsolateHost)
		apiGroup.GET("/traffic/heatmap", trafficHandler.GetThreatHeatmap)

		// Артефакты
		apiGroup.GET("/artifacts/:hostId", artifactHandler.GetArtifactsByHost)
		apiGroup.GET("/artifacts/id/:id", artifactHandler.GetArtifactByID)
		apiGroup.POST("/artifacts", artifactHandler.SaveArtifact)
		apiGroup.GET("/artifacts/search", artifactHandler.SearchArtifacts)

		// AI и анализ
		apiGroup.POST("/ai/analyze", aiHandler.AnalyzeRealtime)
		apiGroup.GET("/ai/patterns", aiHandler.GetAttackPatterns)
		apiGroup.POST("/ai/counter-attack", aiHandler.ExecuteCounterAttack)
		apiGroup.GET("/ai/apt-timeline", aiHandler.GetAPTTimeline)
		apiGroup.POST("/ai/models/update", aiHandler.UpdateModels)
		apiGroup.POST("/ai/models/train", aiHandler.TrainModel)
		apiGroup.GET("/ai/models", aiHandler.ListModels)

		// Интеграции
		apiGroup.POST("/integration/scan", integrationHandler.ScanFile)

		// Дашборд
		apiGroup.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
	}

	// Публичные эндпоинты
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
