package repository

import (
	"context"
	"railgun-core/internal/domain/models"

	"github.com/uptrace/bun"
)

type IncidentRepository struct {
	db *bun.DB
}

func NewIncidentRepository(db *bun.DB) *IncidentRepository {
	return &IncidentRepository{db: db}
}

func (r *IncidentRepository) SaveIncident(ctx context.Context, incident *models.Incident) error {
	_, err := r.db.NewInsert().Model(incident).Exec(ctx)
	return err
}

func (r *IncidentRepository) GetLatestIncidents(ctx context.Context, limit int) ([]models.Incident, error) {
	var incidents []models.Incident
	err := r.db.NewSelect().Model(&incidents).Order("created_at DESC").Limit(limit).Scan(ctx)
	return incidents, err
}
