package api

import (
	"github.com/gin-gonic/gin"

	ingest "railgun-core/api/ingest"
	web "railgun-core/api/web"
	"railgun-core/internal/config"
	"railgun-core/internal/domain"
	repository "railgun-core/internal/domain/repository"
)

func RegisterRoutes(
	r *gin.Engine,
	cfg *config.Config,
	trafficRepo domain.TrafficRepository,
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
		/*apiGroup.GET("/traffic/:hostId", trafficHandler.GetTrafficByHost)
		apiGroup.GET("/traffic/stats/:hostId", trafficHandler.GetTrafficStats)
		apiGroup.POST("/traffic", trafficHandler.SaveTraffic)
		apiGroup.POST("/traffic/logs", trafficHandler.ProcessNetworkLog)
		apiGroup.POST("/traffic/isolate", trafficHandler.IsolateHost)
		apiGroup.GET("/traffic/heatmap", trafficHandler.GetThreatHeatmap)*/

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

		// Инциденты
		apiGroup.GET("/incidents", incidentHandler.GetIncidents)

		ingest := r.Group("/ingest")
		{
			ingest.POST("/traffic", ingestHandler.SaveTraffic)
			ingest.POST("/logs", ingestHandler.ProcessNetworkLog)
		}

		apiGroup.GET("/traffic/:hostId", queryHandler.GetTrafficByHost)
		apiGroup.GET("/heatmap", queryHandler.GetThreatHeatmap)
	}

	// Публичные эндпоинты
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
