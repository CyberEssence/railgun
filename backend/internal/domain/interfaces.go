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
	ProcessNetworkLog(ctx context.Context, hostID, logData, logType string) ([]models.NetworkLog, error)
}

// AnalyticsRepository - для графиков и дашбордов
type AnalyticsRepository interface {
	GetThreatHeatmap(ctx context.Context, from, to time.Time) ([]models.HeatmapPoint, error)
	GetDashboardStats(ctx context.Context, from, to time.Time) (map[string]interface{}, error)
}

// HostActionRepository - для управления состоянием агентов (Active Response)
type HostActionRepository interface {
	IsolateHost(ctx context.Context, hostID, reason string, duration int) error
}

type NetworkLogRepository interface {
	ProcessNetworkLog(ctx context.Context, hostID, logData, logType string) ([]models.NetworkLog, error)
}

type AnalysisRepository interface {
	Save(ctx context.Context, result *models.AnalysisResult) error
}

// ArtifactRepository интерфейс для работы с артефактами
type ArtifactRepository interface {
	GetArtifactsByHost(ctx context.Context, hostID string, page, perPage int) ([]*models.Artifact, int, error)
	GetArtifactByID(ctx context.Context, id int64) (*models.Artifact, error)
	SaveArtifact(ctx context.Context, artifact *models.WindowsArtifact) error
	SearchArtifacts(ctx context.Context, query, artifactType, severity string, page, perPage int) ([]*models.Artifact, int, error)
	GetArtifactByUUID(ctx context.Context, uuid string) (*models.Artifact, error)
}

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	SaveTwoFAToken(ctx context.Context, token models.TwoFAToken) error
	GetTwoFAToken(ctx context.Context, tokenHash string, userID string) (*models.TwoFAToken, error)
	MarkTokenAsUsed(ctx context.Context, tokenID int64) error
	GetTOTPSecret(ctx context.Context, userID string) (string, error)
	Enable2FA(ctx context.Context, userID string, secret string, backupCodes []string) error
	Disable2FA(ctx context.Context, userID string) error
	SaveTOTPSecret(ctx context.Context, userID string, secret string) error
}

type AIService interface {
	AnalyzeAndSave(ctx context.Context, logLines []string, hostID string) ([]*models.AnalysisResult, error)
	//AnalyzeData(ctx context.Context, data []string, dataType, hostID string) (*models.AnalysisResult, error)
	/*GetAttackPatterns(ctx context.Context, category, severity string, page, perPage int) ([]*models.AttackPattern, int, error)
	GetAPTTimeline(ctx context.Context, aptID string, startTime time.Time, endTime time.Time) (*models.APTTimeline, error)
	UpdateModels(ctx context.Context, modelIDs []string) (map[string]string, error)
	TrainModel(ctx context.Context, modelID, datasetPath string, epochs int) (string, error)
	ListModels(ctx context.Context, modelType string) ([]*models.AIModel, error)
	GetThreatStats(ctx context.Context, from time.Time, to time.Time) (*models.ThreatStats, error)*/
}

// IntegrationService - независим от протоколов (HTTP/gRPC)
type IntegrationService interface {
	// Принимает поток байт, что позволяет сканировать файлы из любых источников
	ScanWithVirusTotal(ctx context.Context, file io.Reader, size int64) (*models.ScanResult, error)
}

// TwoFAService интерфейс для работы с двухфакторной аутентификацией
type TwoFAService interface {
	GenerateToken(ctx context.Context, userID string) (string, error)
	Validate2FAToken(ctx context.Context, token string, userID string) (bool, error)
	Enable2FA(ctx context.Context, userID string, username string) (map[string]interface{}, error)
	Disable2FA(ctx context.Context, userID string) error
	GenerateNewBackupCodes(ctx context.Context, userID string) ([]string, error)
	VerifySetup(ctx context.Context, token string, userID string) (bool, error)
	ValidateTOTP(ctx context.Context, token string, userID string, secret string) (bool, error)
	Activate2FA(ctx context.Context, userID string) ([]string, error)
	ValidateBackupCode(ctx context.Context, code string, userID string) (bool, error)
}

type DetectionEngine interface {
	// Анализирует угрозу и выбирает стратегию защиты
	AddEvent(ctx context.Context, event models.EventCorrelation)
	RespondToThreat(targetIP string, threatLevel int) error
}

type IncidentRepository interface {
	SaveIncident(ctx context.Context, incident *models.Incident) error
	GetLatestIncidents(ctx context.Context, limit int) ([]models.Incident, error)
}

type AgentService interface {
	// Получить ожидающую задачу для хоста
	GetPendingTask(ctx context.Context, hostID string) (*models.IsolationTask, error)
	// Обновить статус задачи
	UpdateTaskStatus(ctx context.Context, taskID int64, status string, output string) error
	// Обновить статус хоста
	UpdateHostStatus(ctx context.Context, hostID string, status string) error
}
