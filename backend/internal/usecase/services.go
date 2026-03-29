package usecase

import (
	"log"
	"railgun-core/internal/config"
	"railgun-core/internal/domain"
	repository "railgun-core/internal/domain/repository"
	engine "railgun-core/internal/engine/detection"
	services "railgun-core/internal/infrastructure/collectors"

	"github.com/elastic/go-elasticsearch/v8"
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

	detEngine := engine.NewDetector(config.Detection, incidentRepo)

	integrationCfg := services.IntegrationConfig{
		VirusTotalAPIKey: config.VirusTotal.VirusTotalAPIKey,
		MaxFileSize:      config.VirusTotal.MaxFileSizeMB * 1024 * 1024,
		PollTimeout:      config.VirusTotal.PollTimeout,
		PollInterval:     config.VirusTotal.PollInterval,
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{config.Elastic.URL},
	})
	if err != nil {
		log.Fatal("Failed to create ES client: ", err)
	}

	timelineRepo := repository.NewESTimelineRepository(esClient, "railgun-logs-*")

	analysisRepo := repository.NewAnalysisRepository(db)

	return &Services{
		AIService:          services.NewAIService(analysisRepo, timelineRepo),
		IntegrationService: services.NewIntegrationService(integrationCfg),
		TwoFAService:       services.NewTwoFAService(userRepo, config),
		DetectionEngine:    detEngine,
	}
}
