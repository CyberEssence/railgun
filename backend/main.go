package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"railgun-core/api"
	"railgun-core/models"
	"railgun-core/services"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func main() {
	// Подключение к PostgreSQL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:(2a+3b=0c+1d)@localhost:5432/siem?sslmode=disable"
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	// Миграции
	ctx := context.Background()
	if err := models.RunMigrations(ctx, db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Настройка сервисов
	elasticURL := os.Getenv("ELASTIC_URL")
	if elasticURL == "" {
		elasticURL = "http://127.0.0.1:9200"
	}
	trafficSvc := services.NewTrafficService(db, elasticURL)
	artifactSvc := services.NewArtifactService(db, elasticURL)

	// Настройка Gin
	r := gin.Default()

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
	api.RegisterRoutes(r, db, trafficSvc, artifactSvc)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
