package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"linux-agent/config"
	"linux-agent/internal/collector"
	"linux-agent/sender"
)

type Collector interface {
	Name() string
	Collect() (interface{}, error)
	Enabled() bool
}

type Sender interface {
	Send(data interface{}) error
	Close() error
}

type Agent struct {
	config     *config.Config
	collectors []Collector
	senders    []Sender
	hostID     string
	hostname   string
	done       chan bool
}

func main() {
	// Загрузка конфигурации
	var cfg *config.Config

	// Пробуем загрузить из файла, если нет - из env
	if _, err := os.Stat("config.yaml"); err == nil {
		cfg, err = config.Load("config.yaml")
		if err != nil {
			log.Fatal("Failed to load config from file:", err)
		}
	} else {
		cfg, err = config.LoadFromEnv()
		if err != nil {
			log.Fatal("Failed to load config from env:", err)
		}
	}

	// Определяем hostname
	hostname := cfg.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	hostID := cfg.HostID
	if hostID == "" {
		hostID = generateHostID(hostname)
	}

	log.Printf("Starting Linux Agent v1.0.0 on %s (ID: %s)", hostname, hostID)
	log.Printf("Log level: %s", cfg.LogLevel)

	// Создаем агента
	agent := NewAgent(cfg, hostID, hostname)

	// Обработка сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Запуск сбора данных
	go agent.Run()

	// Ожидание сигнала завершения
	sig := <-sigChan
	log.Printf("Received signal: %v, shutting down...", sig)

	agent.Stop()
	log.Println("Agent stopped gracefully")
}

func NewAgent(cfg *config.Config, hostID, hostname string) *Agent {
	agent := &Agent{
		config:   cfg,
		hostID:   hostID,
		hostname: hostname,
		done:     make(chan bool),
	}

	// Инициализация коллекторов
	agent.initCollectors()

	// Инициализация отправителей
	if err := agent.initSenders(); err != nil {
		log.Fatal("Failed to initialize senders:", err)
	}

	log.Printf("Agent initialized with %d collectors and %d senders",
		len(agent.collectors), len(agent.senders))

	return agent
}

func (a *Agent) initCollectors() {
	if a.config.Collector.System {
		a.collectors = append(a.collectors, &SystemCollector{
			hostID:   a.hostID,
			hostname: a.hostname,
		})
		log.Println("System collector enabled")
	}

	if a.config.Collector.Network {
		a.collectors = append(a.collectors, &NetworkCollector{
			hostID:   a.hostID,
			hostname: a.hostname,
		})
		log.Println("Network collector enabled")
	}

	if a.config.Collector.Processes {
		a.collectors = append(a.collectors, &ProcessCollector{
			hostID:   a.hostID,
			hostname: a.hostname,
		})
		log.Println("Process collector enabled")
	}

	if a.config.Collector.Security {
		a.collectors = append(a.collectors, &SecurityCollector{
			hostID:   a.hostID,
			hostname: a.hostname,
		})
		log.Println("Security collector enabled")
	}

	if a.config.Collector.Docker {
		// Docker collector можно добавить позже
		log.Println("Docker collector not implemented yet")
	}
}

func (a *Agent) initSenders() error {
	// Kafka sender
	if a.config.Kafka.Enabled && len(a.config.Kafka.Brokers) > 0 {
		kafkaSender, err := sender.NewKafkaSender(&a.config.Kafka, a.hostID)
		if err != nil {
			log.Printf("Warning: Failed to init Kafka sender: %v", err)
		} else {
			a.senders = append(a.senders, kafkaSender)
			log.Printf("Kafka sender initialized to %v", a.config.Kafka.Brokers)
		}
	}

	// HTTP sender (ваш SIEM)
	if a.config.HTTP.Enabled && a.config.HTTP.URL != "" {
		httpSender := sender.NewHTTPSender(&a.config.HTTP, a.hostID)
		a.senders = append(a.senders, httpSender)
		log.Printf("HTTP sender initialized to %s", a.config.HTTP.URL)
	}

	// Elasticsearch sender
	if a.config.Elastic.Enabled && a.config.Elastic.URL != "" {
		elasticSender, err := sender.NewElasticSender(&a.config.Elastic, a.hostID)
		if err != nil {
			log.Printf("Warning: Failed to init Elasticsearch sender: %v", err)
		} else {
			a.senders = append(a.senders, elasticSender)
			log.Printf("Elasticsearch sender initialized to %s", a.config.Elastic.URL)
		}
	}

	if len(a.senders) == 0 {
		return fmt.Errorf("no senders configured or initialized")
	}

	return nil
}

func (a *Agent) Run() {
	log.Printf("Starting collection loop with interval %v", a.config.Collector.Interval)

	ticker := time.NewTicker(a.config.Collector.Interval)
	defer ticker.Stop()

	// Первый сбор сразу
	a.collectAndSend()

	for {
		select {
		case <-ticker.C:
			a.collectAndSend()
		case <-a.done:
			return
		}
	}
}

func (a *Agent) collectAndSend() {
	start := time.Now()
	totalMetrics := 0

	for _, collector := range a.collectors {
		if !collector.Enabled() {
			continue
		}

		data, err := collector.Collect()
		if err != nil {
			log.Printf("Collector %s failed: %v", collector.Name(), err)
			continue
		}

		// Отправка через все senders
		for _, sender := range a.senders {
			if err := sender.Send(data); err != nil {
				log.Printf("Sender failed for %s: %v", collector.Name(), err)
			} else {
				totalMetrics++
			}
		}
	}

	log.Printf("Collection completed in %v, sent %d metrics",
		time.Since(start), totalMetrics)
}

func (a *Agent) Stop() {
	close(a.done)

	// Закрываем все senders
	for _, sender := range a.senders {
		if err := sender.Close(); err != nil {
			log.Printf("Error closing sender: %v", err)
		}
	}

	time.Sleep(2 * time.Second)
}

// Реализации коллекторов
type SystemCollector struct {
	hostID   string
	hostname string
}

func (s *SystemCollector) Name() string  { return "system" }
func (s *SystemCollector) Enabled() bool { return true }
func (s *SystemCollector) Collect() (interface{}, error) {
	return collector.CollectSystemInfo(s.hostID, s.hostname)
}

type NetworkCollector struct {
	hostID   string
	hostname string
}

func (n *NetworkCollector) Name() string  { return "network" }
func (n *NetworkCollector) Enabled() bool { return true }
func (n *NetworkCollector) Collect() (interface{}, error) {
	return collector.CollectNetworkInfo(n.hostID, n.hostname)
}

type ProcessCollector struct {
	hostID   string
	hostname string
}

func (p *ProcessCollector) Name() string  { return "process" }
func (p *ProcessCollector) Enabled() bool { return true }
func (p *ProcessCollector) Collect() (interface{}, error) {
	return collector.CollectProcessInfo(p.hostID, p.hostname)
}

type SecurityCollector struct {
	hostID   string
	hostname string
}

func (s *SecurityCollector) Name() string  { return "security" }
func (s *SecurityCollector) Enabled() bool { return true }
func (s *SecurityCollector) Collect() (interface{}, error) {
	return collector.CollectSecurityInfo(s.hostID, s.hostname)
}

// Генерация уникального ID хоста
func generateHostID(hostname string) string {
	// Читаем machine-id если есть
	machineID, err := os.ReadFile("/etc/machine-id")
	if err == nil && len(machineID) > 0 {
		return string(machineID[:32])
	}

	// Или генерируем на основе hostname и времени
	return fmt.Sprintf("%s-%d", hostname, time.Now().UnixNano())
}
