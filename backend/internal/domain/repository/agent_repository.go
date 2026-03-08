package repository

import (
	"context"
	"database/sql"
	"errors"
	"railgun-core/internal/domain/models"
	"time"

	"github.com/uptrace/bun"
)

type agentRepository struct {
	db *bun.DB
}

func NewAgentRepository(db *bun.DB) *agentRepository {
	return &agentRepository{db: db}
}

func (r *agentRepository) GetPendingTask(ctx context.Context, hostID string) (*models.IsolationTask, error) {
	var task models.IsolationTask
	err := r.db.NewSelect().
		Model(&task).
		Where("host_id = ?", hostID).
		Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *agentRepository) UpdateTaskStatus(ctx context.Context, taskID int64, status string, output string) error {
	_, err := r.db.NewUpdate().
		Model((*models.IsolationTask)(nil)).
		Set("status = ?", status).
		Set("output = ?", output).
		Set("completed_at = ?", bun.NullTime{Time: time.Now()}).
		Where("id = ?", taskID).
		Exec(ctx)
	return err
}

func (r *agentRepository) UpdateHostStatus(ctx context.Context, hostID string, status string) error {
	_, err := r.db.NewUpdate().
		Model((*models.Host)(nil)).
		Set("status = ?", status).
		Where("host_id = ?", hostID).
		Exec(ctx)
	return err
}
