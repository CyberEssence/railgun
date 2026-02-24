package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"linux-agent/internal/core/domain"
	"linux-agent/internal/core/ports"
)

// CollectorService сервис для сбора метрик
type CollectorService struct {
	collectors map[string]ports.Collector
	mu         sync.RWMutex
}

// NewCollectorService создает новый сервис коллектора
func NewCollectorService() *CollectorService {
	return &CollectorService{
		collectors: make(map[string]ports.Collector),
	}
}

// RegisterCollector регистрирует коллектор
func (s *CollectorService) RegisterCollector(c ports.Collector) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collectors[c.Name()] = c
	log.Printf("Registered collector: %s", c.Name())
}

// CollectAll собирает данные со всех коллекторов
func (s *CollectorService) CollectAll(ctx context.Context) (*domain.MetricBatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	batch := &domain.MetricBatch{
		Timestamp: time.Now().UTC(),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for name, collector := range s.collectors {
		if !collector.Enabled() {
			continue
		}

		wg.Add(1)
		go func(name string, c ports.Collector) {
			defer wg.Done()

			data, err := c.Collect(ctx)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("collector %s failed: %w", name, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			switch name {
			case "system":
				if sys, ok := data.(*domain.SystemInfo); ok {
					batch.HostID = sys.HostID
					batch.Hostname = sys.Hostname
					batch.System = sys
				}
			case "network":
				if net, ok := data.(*domain.NetworkInfo); ok {
					batch.HostID = net.HostID
					batch.Hostname = net.Hostname
					batch.Network = net
				}
			case "processes":
				if proc, ok := data.(*domain.ProcessInfo); ok {
					batch.HostID = proc.HostID
					batch.Hostname = proc.Hostname
					batch.Processes = proc
				}
			case "security":
				if sec, ok := data.(*domain.SecurityInfo); ok {
					batch.HostID = sec.HostID
					batch.Hostname = sec.Hostname
					batch.Security = sec
				}
			}
			mu.Unlock()
		}(name, collector)
	}

	wg.Wait()

	if len(errs) > 0 {
		return batch, fmt.Errorf("collection errors: %v", errs)
	}

	return batch, nil
}

// GetCollector возвращает коллектор по имени
func (s *CollectorService) GetCollector(name string) (ports.Collector, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.collectors[name]
	return c, ok
}

// ListCollectors возвращает список всех коллекторов
func (s *CollectorService) ListCollectors() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.collectors))
	for name := range s.collectors {
		names = append(names, name)
	}
	return names
}
