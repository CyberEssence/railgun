package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"linux-agent/internal/adapters/config"
	"linux-agent/internal/app"
)

func main() {
	// Парсим аргументы командной строки
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "Path to config file")
	flag.StringVar(&configPath, "c", "config.yaml", "Path to config file (shorthand)")
	flag.Parse()

	// Загрузка конфигурации
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Создаем приложение
	application, err := app.New(cfg)
	if err != nil {
		log.Fatal("Failed to create application:", err)
	}

	// Запускаем приложение
	if err := application.Start(); err != nil {
		log.Fatal("Failed to start application:", err)
	}

	// Обработка сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала завершения
	sig := <-sigChan
	log.Printf("Received signal: %v", sig)

	// Останавливаем приложение
	application.Stop()
	log.Println("Application stopped")
}

func loadConfig(configPath string) (*config.Config, error) {
	// Пробуем загрузить из указанного файла
	if _, err := os.Stat(configPath); err == nil {
		log.Printf("Loading config from: %s", configPath)
		return config.Load(configPath)
	}

	// Если файл не найден, пробуем загрузить из env
	log.Println("Config file not found, loading from environment variables")
	return config.LoadFromEnv()
}
