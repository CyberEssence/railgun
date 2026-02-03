package domain

import (
	"context"
	"io"
	"railgun-core/internal/domain/models"

	"time"
)

// TrafficRepository - только хранение и получение
type TrafficRepository interface {
	GetTrafficByHost(ctx context.Context, hostID string, from, to time.Time) ([]models.NetworkTraffic, error)
	SaveTraffic(ctx context.Context, traffic models.NetworkTraffic) error
	GetTrafficStats(ctx context.Context, hostID string, from, to time.Time) (*models.TrafficStats, error)
}

// AnalyticsRepository - для графиков и дашбордов
type AnalyticsRepository interface {
	GetThreatHeatmap(ctx context.Context, from, to time.Time) ([]models.HeatmapPoint, error)
	GetDashboardStats(ctx context.Context, from, to time.Time) (*models.DashboardStats, error)
}

// HostActionRepository - для управления состоянием агентов (Active Response)
type HostActionRepository interface {
	IsolateHost(ctx context.Context, hostID, reason string, duration int) error
}

type NetworkLogRepository interface {
	ProcessNetworkLog(ctx context.Context, hostID, logData, logType string) ([]models.NetworkTraffic, error)
}

// ArtifactRepository интерфейс для работы с артефактами
type ArtifactRepository interface {
	GetArtifactsByHost(ctx context.Context, hostID string, page, perPage int) ([]*models.Artifact, int, error)
	GetArtifactByID(ctx context.Context, id int64) (*models.Artifact, error)
	SaveArtifact(ctx context.Context, artifact *models.WindowsArtifact) error
	SearchArtifacts(ctx context.Context, query, artifactType, severity string, page, perPage int) ([]*models.Artifact, int, error)
}

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, user models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	SaveTwoFAToken(ctx context.Context, token models.TwoFAToken) error
	GetTwoFAToken(ctx context.Context, tokenHash string, userID int64) (*models.TwoFAToken, error)
	MarkTokenAsUsed(ctx context.Context, tokenID int64) error
}

type AIService interface {
	AnalyzeData(ctx context.Context, data []string, dataType, hostID string) (*models.AnalysisResult, error)
	GetAttackPatterns(ctx context.Context, category, severity string, page, perPage int) ([]*models.AttackPattern, int, error)
	GetAPTTimeline(ctx context.Context, aptID string, startTime time.Time, endTime time.Time) (*models.APTTimeline, error)
	UpdateModels(ctx context.Context, modelIDs []string) (map[string]string, error)
	TrainModel(ctx context.Context, modelID, datasetPath string, epochs int) (string, error)
	ListModels(ctx context.Context, modelType string) ([]*models.AIModel, error)
	GetThreatStats(ctx context.Context, from time.Time, to time.Time) (*models.ThreatStats, error)
}

// IntegrationService - независим от протоколов (HTTP/gRPC)
type IntegrationService interface {
	// Принимает поток байт, что позволяет сканировать файлы из любых источников
	ScanWithVirusTotal(ctx context.Context, file io.Reader, size int64) (*models.ScanResult, error)
}

// TwoFAService интерфейс для работы с двухфакторной аутентификацией
type TwoFAService interface {
	GenerateToken(ctx context.Context, userID int64) (string, error)
	Validate2FAToken(ctx context.Context, token string, userID int64) (bool, error)
}
