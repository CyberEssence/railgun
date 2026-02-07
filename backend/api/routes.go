package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	ingest "railgun-core/api/ingest"
	web "railgun-core/api/web"
	_ "railgun-core/docs"
	"railgun-core/internal/config"
	"railgun-core/internal/domain"
	repository "railgun-core/internal/domain/repository"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(
	r *gin.Engine,
	cfg *config.Config,
	trafficRepo repository.TrafficRepository,
	artifactRepo domain.ArtifactRepository,
	aiService domain.AIService,
	integrationService domain.IntegrationService,
	twoFAService domain.TwoFAService,
	userRepo *repository.UserRepository,
	incidentRepo domain.IncidentRepository,
	networkLogRepo domain.NetworkLogRepository,
	detectionRepo domain.DetectionEngine,
	analyticsRepo domain.AnalyticsRepository,
) {
	// Инициализация обработчиков
	authHandler := web.NewAuthHandler(cfg, twoFAService, userRepo)
	aiHandler := web.NewAIHandler(aiService)
	artifactHandler := web.NewArtifactHandler(artifactRepo)
	integrationHandler := web.NewIntegrationHandler(integrationService)
	dashboardHandler := web.NewDashboardHandler(trafficRepo, aiService)
	//trafficHandler := NewTrafficHandler(trafficRepo)
	incidentHandler := web.NewIncidentHandler(incidentRepo)
	ingestHandler := ingest.NewIngestHandler(trafficRepo, networkLogRepo, detectionRepo)
	queryHandler := web.NewQueryHandler(trafficRepo, analyticsRepo)

	// Тестовый маршрут для проверки middleware
	r.GET("/api/test-auth", authHandler.AuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{
			"message": "Authenticated!",
			"user_id": userID,
		})
	})

	// Группа аутентификации
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/verify-2fa", authHandler.Verify2FA)
		authGroup.POST("/refresh", authHandler.RefreshToken)

		protected := authGroup.Group("")
		protected.Use(authHandler.AuthMiddleware())
		{
			protected.POST("/2fa/enable", authHandler.Enable2FA)
			protected.POST("/2fa/verify-setup", authHandler.Verify2FASetup)
			protected.POST("/2fa/disable", authHandler.Disable2FA)
			protected.POST("/2fa/new-backup-codes", authHandler.GenerateNewBackupCodes)
			protected.GET("/2fa/status", authHandler.Get2FAStatus)
		}
	}

	// Группа API с аутентификацией
	apiGroup := r.Group("/api")
	apiGroup.Use(authHandler.AuthMiddleware())
	{
		// Трафик
		apiGroup.GET("/traffic/stats/:hostId", ingestHandler.GetTrafficStats)
		apiGroup.POST("/traffic/isolate", ingestHandler.IsolateHost)

		// Артефакты
		apiGroup.GET("/artifacts/host/:hostId", artifactHandler.GetArtifactsByHost)
		apiGroup.GET("/artifacts/id/:id", artifactHandler.GetArtifactByID)
		apiGroup.POST("/artifacts", artifactHandler.SaveArtifact)
		apiGroup.GET("/artifacts/search", artifactHandler.SearchArtifacts)

		// AI и анализ
		apiGroup.POST("/ai/analyze", aiHandler.AnalyzeRealtime)
		apiGroup.GET("/ai/patterns", aiHandler.GetAttackPatterns)
		apiGroup.GET("/ai/patterns/stats", aiHandler.GetPatternStats)
		apiGroup.POST("/ai/counter-attack", aiHandler.ExecuteCounterAttack)
		apiGroup.GET("/ai/apt-timeline", aiHandler.GetAPTTimeline)
		apiGroup.POST("/ai/models/update", aiHandler.UpdateModels)
		apiGroup.POST("/ai/models/train", aiHandler.TrainModel)
		apiGroup.GET("/ai/models", aiHandler.ListModels)

		// Интеграции
		apiGroup.POST("/integration/scan", integrationHandler.ScanFile)

		// Дашборд
		apiGroup.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)

		// Инциденты
		apiGroup.GET("/incidents", incidentHandler.GetIncidents)

		apiGroup.POST("/traffic", ingestHandler.SaveTraffic)
		apiGroup.POST("traffic/logs", ingestHandler.ProcessNetworkLog)

		apiGroup.GET("/traffic/:hostId", queryHandler.GetTrafficByHost)
		apiGroup.GET("/traffic/heatmap", queryHandler.GetThreatHeatmap)
	}

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Публичные эндпоинты
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
