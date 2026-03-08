package services

import (
	"log"
	"time"

	"linux-agent/internal/core/domain"
	"linux-agent/internal/core/ports"
)

type IsolationService struct {
	fetcher  ports.TaskFetcher
	executor ports.IsolationExecutor
	hostID   string
	serverIP string
	ticker   *time.Ticker
	stopChan chan struct{}
}

func NewIsolationService(fetcher ports.TaskFetcher, executor ports.IsolationExecutor, hostID string, serverIP string, interval time.Duration) *IsolationService {
	return &IsolationService{
		fetcher:  fetcher,
		executor: executor,
		hostID:   hostID,
		serverIP: serverIP,
		ticker:   time.NewTicker(interval),
		stopChan: make(chan struct{}),
	}
}

func (s *IsolationService) Start() {
	log.Println("Isolation service started. Polling for tasks...")
	go s.runLoop()
}

func (s *IsolationService) Stop() {
	close(s.stopChan)
	s.ticker.Stop()
	log.Println("Isolation service stopped.")
}

func (s *IsolationService) runLoop() {
	for {
		select {
		case <-s.ticker.C:
			s.processTask()
		case <-s.stopChan:
			return
		}
	}
}

func (s *IsolationService) processTask() {
	task, err := s.fetcher.FetchTask(s.hostID)
	if err != nil {
		log.Printf("Error fetching task: %v", err)
		return
	}

	if task == nil {
		return
	}

	log.Printf("Received task ID %d: Action=%s", task.ID, task.Action)

	var output string
	var status string

	switch task.Action {
	case "isolate":
		// Передаем сохраненный IP сервера в метод Isolate
		if err := s.executor.Isolate(s.serverIP); err != nil {
			status = "failed"
			output = err.Error()
		} else {
			status = "completed"
			output = "Host isolated via iptables"
		}
	case "unisolate":
		if err := s.executor.Unisolate(); err != nil {
			status = "failed"
			output = err.Error()
		} else {
			status = "completed"
			output = "Host unisolated"
		}
	default:
		status = "failed"
		output = "Unknown action"
	}

	report := &domain.TaskReport{
		TaskID: task.ID,
		Status: status,
		Output: output,
	}

	if err := s.fetcher.ReportResult(report); err != nil {
		log.Printf("Error reporting task result: %v", err)
	}
}
