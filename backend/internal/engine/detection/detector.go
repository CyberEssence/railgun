package engine

import (
	"context"
	"fmt"
	"log"
	"net"
	"railgun-core/internal/config"
	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"
	"sync"
	"time"
)

type Detector struct {
	// Здесь в будущем будут правила корреляции (например, из YAML)
	// Память для хранения последних событий (IP -> список таймстампов)
	// В продакшене будем использовать Redis
	history map[string][]time.Time
	mu      sync.Mutex
	config  config.DetectionConfig
	repo    domain.IncidentRepository
}

func NewDetector(cfg config.DetectionConfig, repo domain.IncidentRepository) *Detector {
	return &Detector{
		history: make(map[string][]time.Time),
		config:  cfg,
		repo:    repo,
	}
}

// AddEvent - основной метод, куда стекаются все логи
func (e *Detector) AddEvent(ctx context.Context, event models.EventCorrelation) {
	if event.Type == "login_attempt" && !event.Success {
		e.checkBruteForce(ctx, event.SourceIP)
	}
}

func (e *Detector) checkBruteForce(ctx context.Context, ip string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	e.history[ip] = append(e.history[ip], now)

	// Очищаем старые записи, используя окно из конфига
	var recent []time.Time
	for _, t := range e.history[ip] {
		if now.Sub(t) < e.config.BruteForceWindow {
			recent = append(recent, t)
		}
	}
	e.history[ip] = recent

	// Проверяем порог из конфига
	if len(e.history[ip]) >= e.config.BruteForceThreshold {
		log.Printf("[ALERT] IP %s exceeded threshold of %d attempts in %v",
			ip, e.config.BruteForceThreshold, e.config.BruteForceWindow)

		e.RespondToThreat(ip, 5)
		delete(e.history, ip)
	}

	if len(e.history[ip]) >= e.config.BruteForceThreshold {
		incident := &models.Incident{
			Type:        "brute_force",
			SourceIP:    ip,
			ThreatLevel: 5,
			Description: fmt.Sprintf("Detected %d failed logins in %v", len(e.history[ip]), e.config.BruteForceWindow),
		}

		// Сохраняем в БД
		_ = e.repo.SaveIncident(ctx, incident)

		e.RespondToThreat(ip, 5)
		delete(e.history, ip)
	}
}

// RespondToThreat — это "мозг" системы, который решает, как именно наказать нарушителя
func (e *Detector) RespondToThreat(targetIP string, threatLevel int) error {
	if net.ParseIP(targetIP) == nil {
		return fmt.Errorf("invalid IP: %s", targetIP)
	}

	// Логика выбора контрмеры в зависимости от уровня угрозы (SIEM Logic)
	switch {
	case threatLevel >= 5:
		return e.setupNullRoute(targetIP) // Полная блокировка
	case threatLevel >= 3:
		return e.startTarpit(targetIP, threatLevel) // Замедление (Tarpit)
	default:
		return e.deployHoneypot(targetIP, threatLevel) // Наблюдение (Honeypot)
	}
}

// Приватные методы реализации (бывший lowlevel)
func (e *Detector) startTarpit(ip string, intensity int) error {
	log.Printf("[ENGINE] Activating Tarpit for %s (Intensity: %d)", ip, intensity)
	time.Sleep(time.Duration(intensity) * time.Second)
	return nil
}

func (e *Detector) setupNullRoute(ip string) error {
	log.Printf("[ENGINE] Executing Blackhole/NullRoute for %s", ip)
	return nil
}

func (e *Detector) deployHoneypot(ip string, intensity int) error {
	log.Printf("[ENGINE] Redirecting %s to Honeypot", ip)
	return nil
}
