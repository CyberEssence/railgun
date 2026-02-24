package usecase

import (
	"context"
	"railgun-core/internal/domain/models"
	"railgun-core/internal/domain/repository"
)

type ArtifactService struct {
	repo *repository.ArtifactRepository
}

func NewArtifactService(repo *repository.ArtifactRepository) *ArtifactService {
	return &ArtifactService{
		repo: repo,
	}
}

func (s *ArtifactService) GetArtifactsByHost(ctx context.Context, hostID string, page, perPage int) ([]*models.Artifact, int, error) {
	return s.repo.GetArtifactsByHost(ctx, hostID, page, perPage)
}

func (s *ArtifactService) GetArtifactByID(ctx context.Context, id int64) (*models.Artifact, error) {
	return s.repo.GetArtifactByID(ctx, id)
}

func (s *ArtifactService) GetArtifactByUUID(ctx context.Context, uuid string) (*models.Artifact, error) {
	return s.repo.GetArtifactByUUID(ctx, uuid)
}

func (s *ArtifactService) SaveArtifact(ctx context.Context, artifact *models.WindowsArtifact) error {
	return s.repo.SaveArtifact(ctx, artifact)
}

func (s *ArtifactService) SearchArtifacts(ctx context.Context, query, artifactType, severity string, page, perPage int) ([]*models.Artifact, int, error) {
	return s.repo.SearchArtifacts(ctx, query, artifactType, severity, page, perPage)
}
