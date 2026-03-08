package app

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"linux-agent/internal/adapters/collectors"
	"linux-agent/internal/adapters/config"
	"linux-agent/internal/adapters/executor"
	"linux-agent/internal/adapters/senders"
	"linux-agent/internal/adapters/task_client"
	"linux-agent/internal/core/services"
)

type Application struct {
	Config       *config.Config
	AgentSvc     *services.AgentService
	IsolationSvc *services.IsolationService
}

func New(cfg *config.Config) (*Application, error) {
	collectorSvc := services.NewCollectorService()
	var isolationSvc *services.IsolationService
	if cfg.Collector.System {
		collectorSvc.RegisterCollector(
			collectors.NewSystemCollector(cfg.HostID, cfg.Hostname, true),
		)
	}

	if cfg.Collector.Network {
		collectorSvc.RegisterCollector(
			collectors.NewNetworkCollector(cfg.HostID, cfg.Hostname, true),
		)
	}

	if cfg.Collector.Processes {
		collectorSvc.RegisterCollector(
			collectors.NewProcessCollector(cfg.HostID, cfg.Hostname, true),
		)
	}

	if cfg.Collector.Security {
		collectorSvc.RegisterCollector(
			collectors.NewSecurityCollector(cfg.HostID, cfg.Hostname, true),
		)
	}

	// Проверяем включен ли Elasticsearch
	if !cfg.Elastic.Enabled {
		return nil, logError("Elasticsearch is not enabled in config")
	}

	// Создаем отправитель для Elasticsearch
	esSender, err := senders.NewElasticsearchSender(
		cfg.ToElasticConfig(),
		cfg.HostID,
		cfg.Hostname,
	)
	if err != nil {
		return nil, logError("Failed to create Elasticsearch sender: %v", err)
	}

	if cfg.Isolation.Enabled {
		if cfg.Isolation.ServerURL == "" {
			return nil, fmt.Errorf("isolation is enabled but server_url is not set")
		}

		// Парсим URL чтобы получить чистый IP/Host для iptables
		serverIP, err := extractHostFromURL(cfg.Isolation.ServerURL)
		if err != nil {
			log.Printf("Warning: could not parse server URL for isolation whitelist: %v", err)
		}

		fetcher := task_client.NewHTTPClient(cfg.Isolation.ServerURL)
		exec := executor.NewFirewallExecutor()

		isolationSvc = services.NewIsolationService(
			fetcher,
			exec,
			cfg.HostID,
			serverIP,
			cfg.Isolation.PollInterval,
		)
	} else {
		log.Println("Isolation service is disabled.")
	}

	// Создаем сервис агента
	agentSvc := services.NewAgentService(
		collectorSvc,
		esSender,
		services.AgentConfig{
			HostID:        cfg.HostID,
			Hostname:      cfg.Hostname,
			Interval:      cfg.Collector.Interval,
			BatchSize:     cfg.Collector.BatchSize,
			FlushInterval: 10 * time.Second,
		},
	)

	return &Application{
		Config:       cfg,
		AgentSvc:     agentSvc,
		IsolationSvc: isolationSvc,
	}, nil
}

func extractHostFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	host := strings.Split(u.Host, ":")[0]
	return host, nil
}

// Start запускает приложение
func (a *Application) Start() error {
	log.Printf("Starting Linux Agent with config:")
	log.Printf("  Host ID: %s", a.Config.HostID)
	log.Printf("  Hostname: %s", a.Config.Hostname)
	log.Printf("  Collectors: system=%v, network=%v, processes=%v, security=%v",
		a.Config.Collector.System,
		a.Config.Collector.Network,
		a.Config.Collector.Processes,
		a.Config.Collector.Security,
	)
	log.Printf("  Elasticsearch: %v (enabled: %v)",
		a.Config.Elastic.URLs,
		a.Config.Elastic.Enabled,
	)

	if a.IsolationSvc != nil {
		a.IsolationSvc.Start()
	}

	return a.AgentSvc.Start()
}

// Stop останавливает приложение
func (a *Application) Stop() {
	if a.IsolationSvc != nil {
		a.IsolationSvc.Stop()
	}

	a.AgentSvc.Stop()
}

// Вспомогательная функция для логирования ошибок
func logError(format string, args ...interface{}) error {
	log.Printf("ERROR: "+format, args...)
	return fmt.Errorf(format, args...)
}
