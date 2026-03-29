package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"railgun-core/api"
	"railgun-core/internal/config"
	repository "railgun-core/internal/domain/repository"
	engine "railgun-core/internal/engine/detection"
	services "railgun-core/internal/infrastructure/collectors"
	persistence "railgun-core/internal/infrastructure/persistence"
	"railgun-core/internal/usecase"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer <your_token>

// @title           Railgun API
// @version         1.0
// @description     Railgun AI-driven SIEM API
// @host            localhost:8080
// @BasePath        /api
func main() {
	// Зарузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.Database.DSN)))

	// Create bun.DB
	db := bun.NewDB(sqldb, pgdialect.New())

	// Настройка контекста с обработкой сигналов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов остановки
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		log.Println("Shutting down...")
		cancel()
	}()

	// Миграции
	if err := persistence.RunMigrations(ctx, db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Инициализация репозиториев
	trafficRepo := repository.NewTrafficRepository(db, cfg.Elastic.URL)
	// Подключаемся к Elasticsearch если настроен
	var elasticAddrs []string
	if cfg.Elastic.URL != "" {
		elasticAddrs = []string{cfg.Elastic.URL}
	}

	// Создаем репозиторий артефактов с Elasticsearch
	artifactRepo, err := repository.NewArtifactRepository(
		db,
		elasticAddrs,
		cfg.Elastic.Username,
		cfg.Elastic.Password,
		"siem-logs-*", // Имя индекса
	)
	if err != nil {
		log.Fatal("Failed to create artifact repository:", err)
	}
	userRepo := repository.NewUserRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)
	agentRepo := repository.NewAgentRepository(db)
	analysisRepo := repository.NewAnalysisRepository(db)

	detEngine := engine.NewDetector(cfg.Detection, incidentRepo)

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.Elastic.URL},
	})
	if err != nil {
		log.Fatal("Failed to create ES client for timeline: ", err)
	}
	timelineRepo := repository.NewESTimelineRepository(esClient, "railgun-logs-*")

	// Инициализация сервисов
	aiService := services.NewAIService(analysisRepo, timelineRepo)
	integrationService := services.NewIntegrationService(services.IntegrationConfig{
		VirusTotalAPIKey: cfg.VirusTotal.VirusTotalAPIKey,
		MaxFileSize:      cfg.VirusTotal.MaxFileSizeMB,
	})
	twoFAService := services.NewTwoFAService(userRepo, cfg)
	agentService := usecase.NewAgentService(agentRepo)

	// Настройка Gin
	r := gin.Default()
	r.MaxMultipartMemory = 32 << 20

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Регистрация маршрутов API
	api.RegisterRoutes(
		r,
		cfg,
		*trafficRepo,
		artifactRepo,
		aiService,
		integrationService,
		twoFAService,
		userRepo,
		incidentRepo,
		trafficRepo,
		detEngine,
		trafficRepo,
		*agentService,
	)

	// Запуск сервера
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Fatalf("Database connection close failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}
