package usecase

import (
	"context"
	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"
)

type AgentService struct {
	repo domain.AgentService
}

func NewAgentService(repo domain.AgentService) *AgentService {
	return &AgentService{repo: repo}
}

func (s *AgentService) GetAgentTask(ctx context.Context, hostID string) (*models.IsolationTask, error) {
	return s.repo.GetPendingTask(ctx, hostID)
}

func (s *AgentService) ReportTaskResult(ctx context.Context, taskID int64, status string, output string, hostID string) error {
	// Обновляем статус задачи
	if err := s.repo.UpdateTaskStatus(ctx, taskID, status, output); err != nil {
		return err
	}

	// Если задача выполнена успешно, меняем статус хоста
	if status == "completed" {
		hostStatus := "isolated"

		if err := s.repo.UpdateHostStatus(ctx, hostID, hostStatus); err != nil {
			return err
		}
	}

	return nil
}
