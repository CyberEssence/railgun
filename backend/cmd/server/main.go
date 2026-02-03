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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"railgun-core/api"
	"railgun-core/internal/config"
	"railgun-core/internal/infrastructure/persistence"
	"railgun-core/internal/infrastructure/services"
)

func main() {
	// Загрузка конфигурации
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
	trafficRepo := persistence.NewTrafficRepository(db, cfg.Elastic.URL)
	artifactRepo := persistence.NewArtifactRepository(db, cfg.Elastic.URL)
	userRepo := persistence.NewUserRepository(db)

	// Инициализация сервисов
	aiService := services.NewAIService(db)
	integrationService := services.NewIntegrationService(services.IntegrationConfig{
		VirusTotalAPIKey: cfg.Integration.VirusTotalAPIKey,
		MaxFileSize:      cfg.Integration.MaxFileSizeMB,
	})
	twoFAService := services.NewTwoFAService(userRepo)

	// Настройка Gin
	r := gin.Default()
	r.MaxMultipartMemory = 32 << 20

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.CORSAllowOrigins,
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
		trafficRepo,
		artifactRepo,
		aiService,
		integrationService,
		twoFAService,
		userRepo,
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
