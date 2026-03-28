package repository

import (
	"context"
	"railgun-core/internal/domain/models"

	"github.com/uptrace/bun"
)

type AnalysisRepository struct {
	db *bun.DB
}

func NewAnalysisRepository(db *bun.DB) *AnalysisRepository {
	return &AnalysisRepository{db: db}
}

func (r *AnalysisRepository) Save(ctx context.Context, result *models.AnalysisResult) error {
	_, err := r.db.NewInsert().
		Model(result).
		Exec(ctx)

	return err
}
