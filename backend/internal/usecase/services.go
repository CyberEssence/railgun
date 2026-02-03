package usecase

import (
	"railgun-core/internal/config"
	"railgun-core/internal/domain"

	"railgun-core/internal/infrastructure/persistence"
	"railgun-core/internal/infrastructure/services"

	"github.com/uptrace/bun"
)

type Services struct {
	AIService          domain.AIService
	IntegrationService domain.IntegrationService
	TwoFAService       domain.TwoFAService
}

func SetupServices(db *bun.DB, config *config.Config) *Services {
	userRepo := persistence.NewUserRepository(db)

	return &Services{
		AIService: services.NewAIService(db),
		IntegrationService: services.NewIntegrationService(services.IntegrationConfig{
			VirusTotalAPIKey: config.Integration.VirusTotalAPIKey,
			MaxFileSize:      config.Integration.MaxFileSizeMB * 1024 * 1024,
		}),
		TwoFAService: services.NewTwoFAService(userRepo),
	}
}
