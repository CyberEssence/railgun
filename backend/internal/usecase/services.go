package usecase

import (
	"railgun-core/internal/config"
	"railgun-core/internal/domain"
	repository "railgun-core/internal/domain/repository"
	engine "railgun-core/internal/engine/detection"
	services "railgun-core/internal/infrastructure/collectors"

	"github.com/uptrace/bun"
)

type Services struct {
	AIService          domain.AIService
	IntegrationService domain.IntegrationService
	TwoFAService       domain.TwoFAService
	DetectionEngine    domain.DetectionEngine
}

func SetupServices(db *bun.DB, config *config.Config) *Services {
	userRepo := repository.NewUserRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	// Инициализируем Engine
	detEngine := engine.NewDetector(config.Detection, incidentRepo)

	return &Services{
		AIService: services.NewAIService(db),
		IntegrationService: services.NewIntegrationService(services.IntegrationConfig{
			VirusTotalAPIKey: config.Integration.VirusTotalAPIKey,
			MaxFileSize:      config.Integration.MaxFileSizeMB * 1024 * 1024,
		}),
		TwoFAService:    services.NewTwoFAService(userRepo),
		DetectionEngine: detEngine,
	}
}
