package api

import (
	"railgun-core/internal/domain"
	repository "railgun-core/internal/domain/repository"
)

type DashboardHandler struct {
	trafficRepo   repository.TrafficRepository
	analyticsRepo domain.AnalyticsRepository
	aiRepo        domain.AIService
}

func NewDashboardHandler(trafficRepo repository.TrafficRepository, aiRepo domain.AIService) *DashboardHandler {
	return &DashboardHandler{
		trafficRepo: trafficRepo,
		aiRepo:      aiRepo,
	}
}
