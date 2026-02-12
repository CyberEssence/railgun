package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"linux-agent/config"
	"linux-agent/internal/agent"
)

func main() {
	// Загрузка конфигурации
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Создаем агента
	ag, err := agent.New(cfg)
	if err != nil {
		log.Fatal("Failed to create agent:", err)
	}

	// Запускаем агента
	if err := ag.Start(); err != nil {
		log.Fatal("Failed to start agent:", err)
	}

	// Обработка сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала завершения
	sig := <-sigChan
	log.Printf("Received signal: %v", sig)

	// Останавливаем агента
	ag.Stop()
}

func loadConfig() (*config.Config, error) {
	// Пробуем загрузить из файла
	if _, err := os.Stat("config.yaml"); err == nil {
		return config.Load("config.yaml")
	}
	// Если нет файла, загружаем из env
	return config.LoadFromEnv()
}
