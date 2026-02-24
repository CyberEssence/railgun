package services

import (
	"context"
	"log"
	"sync"
	"time"

	"linux-agent/internal/core/domain"
	"linux-agent/internal/core/ports"
)

// AgentService основной сервис агента
type AgentService struct {
	collectorSvc  *CollectorService
	sender        ports.Sender
	hostID        string
	hostname      string
	interval      time.Duration
	batchSize     int
	flushInterval time.Duration
	metricsCh     chan *domain.MetricBatch
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// AgentConfig конфигурация агента
type AgentConfig struct {
	HostID        string
	Hostname      string
	Interval      time.Duration
	BatchSize     int
	FlushInterval time.Duration
}

// NewAgentService создает новый сервис агента
func NewAgentService(
	collectorSvc *CollectorService,
	sender ports.Sender,
	config AgentConfig,
) *AgentService {
	ctx, cancel := context.WithCancel(context.Background())

	return &AgentService{
		collectorSvc:  collectorSvc,
		sender:        sender,
		hostID:        config.HostID,
		hostname:      config.Hostname,
		interval:      config.Interval,
		batchSize:     config.BatchSize,
		flushInterval: config.FlushInterval,
		metricsCh:     make(chan *domain.MetricBatch, 100),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start запускает агента
func (s *AgentService) Start() error {
	log.Printf("Starting agent service (host: %s, id: %s)", s.hostname, s.hostID)
	log.Printf("Collectors enabled: %v", s.collectorSvc.ListCollectors())

	// Запускаем воркер для отправки
	s.wg.Add(1)
	go s.senderWorker()

	// Запускаем сбор метрик
	s.wg.Add(1)
	go s.collectWorker()

	return nil
}

// Stop останавливает агента
func (s *AgentService) Stop() {
	log.Println("Stopping agent service...")
	s.cancel()
	s.wg.Wait()

	if err := s.sender.Close(); err != nil {
		log.Printf("Error closing sender: %v", err)
	}

	close(s.metricsCh)
	log.Println("Agent service stopped")
}

// collectWorker воркер для сбора метрик
func (s *AgentService) collectWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Первый сбор сразу
	s.collectAndSend()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.collectAndSend()
		}
	}
}

// collectAndSend собирает и отправляет метрики
func (s *AgentService) collectAndSend() {
	start := time.Now()

	ctx, cancel := context.WithTimeout(s.ctx, s.interval/2)
	defer cancel()

	batch, err := s.collectorSvc.CollectAll(ctx)
	if err != nil {
		log.Printf("Collection warning: %v", err)
	}

	// Устанавливаем host информацию если её нет
	if batch.HostID == "" {
		batch.HostID = s.hostID
	}
	if batch.Hostname == "" {
		batch.Hostname = s.hostname
	}

	// Отправляем в канал
	select {
	case s.metricsCh <- batch:
		log.Printf("Collection completed in %v, metrics queued", time.Since(start))
	case <-time.After(1 * time.Second):
		log.Printf("Warning: metrics channel full, dropping batch")
	}
}

// senderWorker воркер для отправки метрик
func (s *AgentService) senderWorker() {
	defer s.wg.Done()

	batch := make([]*domain.MetricBatch, 0, s.batchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			if len(batch) > 0 {
				s.sendBatch(batch)
			}
			return

		case metrics, ok := <-s.metricsCh:
			if !ok {
				return
			}
			batch = append(batch, metrics)
			if len(batch) >= s.batchSize {
				s.sendBatch(batch)
				batch = make([]*domain.MetricBatch, 0, s.batchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.sendBatch(batch)
				batch = make([]*domain.MetricBatch, 0, s.batchSize)
			}
		}
	}
}

// sendBatch отправляет батч метрик
func (s *AgentService) sendBatch(batch []*domain.MetricBatch) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	for _, metrics := range batch {
		if err := s.sender.Send(ctx, metrics); err != nil {
			log.Printf("Failed to send metrics: %v", err)
		}
	}
}
