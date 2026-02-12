package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"linux-agent/config"
	"linux-agent/internal/collector"
	"linux-agent/pkg/models"
	"linux-agent/sender"
)

type Agent struct {
	config     *config.Config
	hostID     string
	hostname   string
	sender     sender.Sender
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	metricsCh  chan *models.MetricBatch
	collectors []collector.Collector
}

func New(cfg *config.Config) (*Agent, error) {
	hostID := cfg.HostID
	if hostID == "" {
		hostID = generateHostID()
	}

	hostname := cfg.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	ctx, cancel := context.WithCancel(context.Background())

	agent := &Agent{
		config:    cfg,
		hostID:    hostID,
		hostname:  hostname,
		ctx:       ctx,
		cancel:    cancel,
		metricsCh: make(chan *models.MetricBatch, 100),
	}

	// Инициализируем коллекторы
	agent.initCollectors()

	// Инициализируем отправитель (только один!)
	s, err := sender.NewSender(cfg, hostID, hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender: %v", err)
	}
	agent.sender = s

	return agent, nil
}

func (a *Agent) initCollectors() {
	if a.config.Collector.System {
		a.collectors = append(a.collectors, &SystemCollector{
			BaseCollector: collector.BaseCollector{
				HostID:   a.hostID,
				Hostname: a.hostname,
			},
		})
		log.Println("System collector enabled")
	}

	if a.config.Collector.Network {
		a.collectors = append(a.collectors, &NetworkCollector{
			BaseCollector: collector.BaseCollector{
				HostID:   a.hostID,
				Hostname: a.hostname,
			},
		})
		log.Println("Network collector enabled")
	}

	if a.config.Collector.Processes {
		a.collectors = append(a.collectors, &ProcessCollector{
			BaseCollector: collector.BaseCollector{
				HostID:   a.hostID,
				Hostname: a.hostname,
			},
		})
		log.Println("Process collector enabled")
	}

	if a.config.Collector.Security {
		a.collectors = append(a.collectors, &SecurityCollector{
			BaseCollector: collector.BaseCollector{
				HostID:   a.hostID,
				Hostname: a.hostname,
			},
		})
		log.Println("Security collector enabled")
	}
}

func (a *Agent) Start() error {
	log.Printf("Starting Linux Agent v1.0.0 on %s (ID: %s)", a.hostname, a.hostID)
	log.Printf("Log level: %s", a.config.LogLevel)

	// Запускаем воркер для отправки метрик
	a.wg.Add(1)
	go a.senderWorker()

	// Запускаем сбор метрик
	a.wg.Add(1)
	go a.collectWorker()

	return nil
}

func (a *Agent) Stop() {
	log.Println("Shutting down agent...")
	a.cancel()
	a.wg.Wait()

	if err := a.sender.Close(); err != nil {
		log.Printf("Error closing sender: %v", err)
	}

	log.Println("Agent stopped")
}

func (a *Agent) collectWorker() {
	defer a.wg.Done()

	ticker := time.NewTicker(a.config.Collector.Interval)
	defer ticker.Stop()

	// Первый сбор сразу
	a.collectAndSend()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.collectAndSend()
		}
	}
}

func (a *Agent) collectAndSend() {
	start := time.Now()

	metrics := &models.MetricBatch{
		HostID:    a.hostID,
		Hostname:  a.hostname,
		Timestamp: time.Now().UTC(),
		Metrics:   make(map[string]interface{}),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, c := range a.collectors {
		if !c.Enabled() {
			continue
		}

		wg.Add(1)
		go func(col collector.Collector) {
			defer wg.Done()

			data, err := col.Collect()
			if err != nil {
				log.Printf("Collector %s failed: %v", col.Name(), err)
				return
			}

			mu.Lock()
			metrics.Metrics[col.Name()] = data
			mu.Unlock()
		}(c)
	}

	wg.Wait()

	// Отправляем батч
	select {
	case a.metricsCh <- metrics:
		log.Printf("Collection completed in %v, metrics queued", time.Since(start))
	case <-time.After(1 * time.Second):
		log.Printf("Warning: metrics channel full, dropping batch")
	}
}

func (a *Agent) senderWorker() {
	defer a.wg.Done()

	batch := make([]*models.MetricBatch, 0, a.config.Sender.BatchSize)
	ticker := time.NewTicker(a.config.Sender.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			if len(batch) > 0 {
				a.sendBatch(batch)
			}
			return

		case metrics := <-a.metricsCh:
			batch = append(batch, metrics)
			if len(batch) >= a.config.Sender.BatchSize {
				a.sendBatch(batch)
				batch = make([]*models.MetricBatch, 0, a.config.Sender.BatchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				a.sendBatch(batch)
				batch = make([]*models.MetricBatch, 0, a.config.Sender.BatchSize)
			}
		}
	}
}

func (a *Agent) sendBatch(batch []*models.MetricBatch) {
	if err := a.sender.SendBatch(batch); err != nil {
		log.Printf("Failed to send batch: %v", err)
	}
}

func generateHostID() string {
	if id, err := os.ReadFile("/etc/machine-id"); err == nil && len(id) > 0 {
		return strings.TrimSpace(string(id[:32]))
	}
	return uuid.New().String()
}

// Коллекторы
type SystemCollector struct {
	collector.BaseCollector
}

func (s *SystemCollector) Name() string  { return "system" }
func (s *SystemCollector) Enabled() bool { return true }
func (s *SystemCollector) Collect() (interface{}, error) {
	return collector.CollectSystemInfo(s.HostID, s.Hostname)
}

type NetworkCollector struct {
	collector.BaseCollector
}

func (n *NetworkCollector) Name() string  { return "network" }
func (n *NetworkCollector) Enabled() bool { return true }
func (n *NetworkCollector) Collect() (interface{}, error) {
	return collector.CollectNetworkInfo(n.HostID, n.Hostname)
}

type ProcessCollector struct {
	collector.BaseCollector
}

func (p *ProcessCollector) Name() string  { return "process" }
func (p *ProcessCollector) Enabled() bool { return true }
func (p *ProcessCollector) Collect() (interface{}, error) {
	return collector.CollectProcessInfo(p.HostID, p.Hostname)
}

type SecurityCollector struct {
	collector.BaseCollector
}

func (s *SecurityCollector) Name() string  { return "security" }
func (s *SecurityCollector) Enabled() bool { return true }
func (s *SecurityCollector) Collect() (interface{}, error) {
	return collector.CollectSecurityInfo(s.HostID, s.Hostname)
}
