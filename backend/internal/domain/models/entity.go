package models

import (
	"encoding/json"
	"railgun-core/internal/domain/dto"
	"time"

	"github.com/uptrace/bun"
)

// Config содержит конфигурацию приложения
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Elastic     ElasticConfig
	Integration IntegrationConfig
	Security    SecurityConfig
	Auth        AuthConfig
}

type ServerConfig struct {
	Port             string
	CORSAllowOrigins []string
}

type DatabaseConfig struct {
	DSN string
}

type ElasticConfig struct {
	URL string
}

type IntegrationConfig struct {
	VirusTotalAPIKey string
	MaxFileSize      int
}

type SecurityConfig struct {
	WhitelistIPs []string
}

type AuthConfig struct {
	Secret    string
	TokenTTL  int
	IssuerURL string
}

// AgentConfig содержит конфигурацию агента
type AgentConfig struct {
	ServerURL  string
	APIKey     string
	HostID     string
	Interval   int
	LogLevel   string
	MaxRetries int
}

// Event представляет собой базовое событие в системе
type Event struct {
	ID        int64                  `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Severity  string                 `json:"severity"`
	HostID    string                 `json:"host_id"`
}

// Host представляет систему Windows
type Host struct {
	bun.BaseModel `bun:"table:hosts,alias:h"`
	ID            string                 `bun:"id,pk" json:"id"`
	HostID        string                 `bun:"host_id,unique" json:"host_id"`
	Hostname      string                 `bun:"hostname" json:"hostname"`
	IPAddresses   []string               `bun:"ip_addresses,type:text[]" json:"ip_addresses"`
	OSVersion     string                 `bun:"os" json:"os_version"`
	Platform      string                 `bun:"platform" json:"platform"`
	Kernel        string                 `bun:"kernel" json:"kernel"`
	AgentVersion  string                 `bun:"agent_version" json:"agent_version"`
	Status        string                 `bun:"status" json:"status"`
	LastSeen      time.Time              `bun:"last_seen" json:"last_seen"`
	Description   string                 `bun:"-" json:"description"`
	CreatedAt     time.Time              `bun:"created_at" json:"created_at"`
	UpdatedAt     time.Time              `bun:"updated_at" json:"updated_at"`
	Labels        map[string]interface{} `bun:"labels,type:jsonb" json:"labels"`
}

// NetworkTraffic представляет запись о сетевом трафике
type NetworkTraffic struct {
	bun.BaseModel `bun:"table:network_traffic"`
	ID            string    `bun:"id,pk,default:gen_random_uuid()"`
	HostID        string    `json:"host_id"`
	Timestamp     time.Time `json:"timestamp"`
	SrcIP         string    `json:"src_ip"`
	DstIP         string    `json:"dst_ip"`
	SrcPort       int       `json:"src_port"`
	DstPort       int       `json:"dst_port"`
	Protocol      string    `json:"protocol"`
	BytesSent     int64     `json:"bytes_sent"`
	BytesRecv     int64     `json:"bytes_recv"`
	PacketsSent   int64     `json:"packets_sent"`
	PacketsRecv   int64     `json:"packets_recv"`
	Duration      float64   `json:"duration"`
}

// WindowsArtifact представляет системный артефакт Windows
type WindowsArtifact struct {
	bun.BaseModel `bun:"table:windows_artifacts,alias:wa"`

	ID          int64     `json:"-" bun:"id,pk,autoincrement"`
	UUID        string    `json:"id" bun:"uuid,notnull"`
	HostID      string    `json:"host_id" bun:"host_id"`
	Type        string    `json:"type" bun:"type"`
	Path        string    `json:"path" bun:"path"`
	Size        int64     `json:"size" bun:"size"`
	Hash        string    `json:"hash" bun:"hash"`
	Value       string    `json:"value,omitempty" bun:"value"`
	Owner       string    `json:"owner,omitempty" bun:"owner"`
	Permissions string    `json:"permissions,omitempty" bun:"permissions"`
	Timestamp   time.Time `json:"timestamp" bun:"timestamp"`
	ThreatLevel int       `json:"threat_level" bun:"threat_level"`
}

type IsolationTask struct {
	bun.BaseModel `bun:"table:isolation_tasks,alias:t"`

	ID          int64      `bun:"id,pk,autoincrement" json:"id"`
	HostID      string     `bun:"host_id" json:"host_id"`
	Action      string     `bun:"action" json:"action"`
	Status      string     `bun:"status" json:"status"`
	Output      string     `bun:"output" json:"output"`
	CreatedAt   time.Time  `bun:"created_at" json:"created_at"`
	CompletedAt *time.Time `bun:"completed_at" json:"completed_at"`
}

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           string `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Username     string `bun:"username,unique,notnull"`
	Email        string `bun:"email,unique,notnull"`
	PasswordHash string `bun:"password_hash,notnull"`

	TOTPSecret  string `bun:"totp_secret"`
	TOTPEnabled bool   `bun:"totp_enabled,default:false"`

	TOTPBackupCodes json.RawMessage `bun:"totp_backup_codes,type:jsonb"`

	IsActive  bool      `bun:"is_active,default:true"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	LastLogin time.Time `bun:"last_login"`
}

// TwoFAToken модель для хранения токенов 2FA
type TwoFAToken struct {
	bun.BaseModel `bun:"table:two_fa_tokens,alias:t"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID    string    `bun:"user_id,notnull" json:"user_id"`
	TokenHash string    `bun:"token_hash,notnull" json:"-"`
	ExpiresAt time.Time `bun:"expires_at,notnull" json:"expires_at"`
	Used      bool      `bun:"used,notnull,default:false" json:"used"`
}

// TrafficStats содержит статистику по трафику
type TrafficStats struct {
	TotalBytesSent   int64   `json:"total_bytes_sent"`
	TotalBytesRecv   int64   `json:"total_bytes_recv"`
	TotalPacketsSent int64   `json:"total_packets_sent"`
	TotalPacketsRecv int64   `json:"total_packets_recv"`
	AverageDuration  float64 `json:"average_duration"`
}

// ThreatReport содержит результаты анализа угроз
type ThreatReport struct {
	ID                   int64     `json:"id"`
	Timestamp            time.Time `json:"timestamp"`
	AnalysisType         string    `json:"analysis_type"`
	MaliciousProbability float64   `json:"malicious_probability"`
	DetectedPatterns     []string  `json:"detected_patterns"`
	Confidence           float64   `json:"confidence"`
	RawData              []byte    `json:"-"`
	ThreatType           string    `json:"threat_type"`
	Indicators           []string  `json:"indicators"`
}

// AttackPattern модель шаблона атаки
type AttackPattern struct {
	bun.BaseModel `bun:"table:attack_patterns,alias:ap"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	Name        string    `bun:"name,unique" json:"name"`
	Description string    `bun:"description" json:"description"`
	MITREID     string    `bun:"mitre_id" json:"mitre_id"`
	Category    string    `bun:"category" json:"category"` // Добавлено поле category
	Severity    string    `bun:"severity" json:"severity"`
	Indicators  []string  `bun:"indicators,type:jsonb" json:"indicators"`
	CreatedAt   time.Time `bun:"created_at,nullzero,default:current_timestamp" json:"created_at"`
}

// NetworkLog представляет сетевой лог
type NetworkLog struct {
	ID            string    `bun:"id,pk,default:gen_random_uuid()"`
	SourceIP      string    `json:"source_ip"`
	DestinationIP string    `json:"destination_ip"`
	Protocol      string    `json:"protocol"`
	LogType       string    `json:"log_type"` // Например: firewall, ids, proxy
	RawData       string    `json:"raw_data"`
	Timestamp     time.Time `json:"timestamp"`
	Severity      string    `json:"severity"` // info, warning, critical
}

// RealtimeDetectionRequest запрос на анализ в реальном времени
type RealtimeDetectionRequest struct {
	TrafficData  []NetworkTraffic `json:"traffic_data"`
	AnalysisType string           `json:"analysis_type"` // например: "network", "behavior"
	TimeWindow   time.Duration    `json:"time_window"`
	Timestamp    time.Time        `json:"timestamp"`
}

// IsolationRequest запрос на изоляцию хоста
type IsolationRequest struct {
	HostID string `json:"host_id" binding:"required"`
}

// IsolationEvent хранит историю изоляций для аудита
type IsolationEvent struct {
	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	HostID    string    `bun:"host_id,notnull" json:"host_id"`
	Reason    string    `bun:"reason" json:"reason"`
	Duration  int       `bun:"duration" json:"duration"`
	Status    string    `bun:"status,default:'active'" json:"status"`
	CreatedAt time.Time `bun:"created_at,default:current_timestamp" json:"created_at"`
}

// TableName задает имя таблицы в БД для библиотеки bun
func (IsolationEvent) TableName() string {
	return "isolation_events"
}

// CounterAttackRequest запрос на контратаку
type CounterAttackRequest struct {
	TargetIP   string `json:"target_ip"`
	AttackType string `json:"attack_type"`
	Intensity  int    `json:"intensity"`
}

// DecoyRequest запрос на создание приманки
type DecoyRequest struct {
	Type      string   `json:"type"`
	TargetNet string   `json:"target_net"`
	Services  []string `json:"services"`
}

// ModelUpdateRequest запрос на обновление модели
type ModelUpdateRequest struct {
	ModelType  string                 `json:"model_type" binding:"required"`
	DatasetURL string                 `json:"dataset_url" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
}

// TrainingRequest запрос на обучение модели
type TrainingRequest struct {
	ModelType  string                 `json:"model_type" binding:"required"`
	DatasetURL string                 `json:"dataset_url" binding:"required,url"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ThreatHeatmap данные для тепловой карты угроз
type ThreatHeatmap struct {
	Regions    []string          `json:"regions"`
	Intensity  map[string]int    `json:"intensity"`
	Timestamps []int64           `json:"timestamps"`
	Metadata   map[string]string `json:"metadata"`
}

// APTTimeline временная шкала APT атак
type APTTimeline struct {
	Events []APTEpoch `json:"events"`
}

// APTEpoch эпоха в временной шкале APT
type APTEpoch struct {
	Timestamp int64  `json:"timestamp"`
	Stage     string `json:"stage"`
	Indicator string `json:"indicator"`
}

// DashboardStats статистика для дашборда
type DashboardStats struct {
	TotalEvents        int    `json:"totalEvents"`
	ActiveConnections  int    `json:"activeConnections"`
	SuspiciousActivity int    `json:"suspiciousActivity"`
	SystemHealth       string `json:"systemHealth"`
}

type ThreatAnalysis struct {
	ID            int64     `json:"id"`
	ReportID      int64     `json:"report_id"`
	AnalysisDate  time.Time `json:"analysis_date"`
	Conclusion    string    `json:"conclusion"`
	Severity      int       `json:"severity"`
	Mitigation    string    `json:"mitigation"`
	Analyst       string    `json:"analyst"`
	FalsePositive bool      `json:"false_positive"`
}

type HeatmapPoint struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Weight    int     `json:"weight"`
	IP        string  `json:"ip"`
	Country   string  `json:"country"`
}

type GeoInfo struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Country   string  `json:"country"`
}

// AnalysisResult представляет запись анализа лога
type AnalysisResult struct {
	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	HostID       string    `bun:"host_id,type:varchar(64)" json:"host_id"`              // ID хоста или агента
	InputData    string    `bun:"input_data,type:text" json:"input_data"`               // Сам лог
	PredictLabel string    `bun:"predict_label,type:varchar(128)" json:"predict_label"` // Метка (Malicious/Normal)
	Score        float64   `bun:"score,type:decimal(10,4)" json:"score"`                // Уверенность (0.0 - 1.0)
	IsMalicious  bool      `bun:"is_malicious,type:boolean" json:"is_malicious"`        // Быстрый флаг
	CreatedAt    time.Time `bun:"created_at,type:timestamptz,default:current_timestamp" json:"created_at"`
}

// TableName задает имя таблицы в БД для Bun
func (AnalysisResult) TableName() string {
	return "analysis_results"
}

type Artifact struct {
	UUID        string    `json:"id" bun:"uuid,pk"` // UUID как первичный ключ
	HostID      string    `json:"host_id" bun:"host_id"`
	Type        string    `json:"type" bun:"type"`
	Name        string    `json:"name" bun:"name"`
	Path        string    `json:"path" bun:"path"`
	Size        int64     `json:"size" bun:"size"`
	Hash        string    `json:"hash" bun:"hash"`
	Timestamp   time.Time `json:"timestamp" bun:"timestamp"`
	ThreatLevel int       `json:"threat_level" bun:"threat_level"`
	Content     []byte    `json:"content,omitempty" bun:"content,omitempty"`
}

type ThreatStats struct {
	Total      int // Общее количество угроз
	Critical   int // Количество критических угроз
	High       int // Высокие угрозы
	Medium     int // Средние угрозы
	Low        int // Низкие угрозы
	NewLast24h int // Новые угрозы за последние 24 часа
	Resolved   int // Устраненные угрозы
	// Добавьте другие поля по необходимости
}

type Threat struct {
	ID        string    `bun:"id,pk"`
	Severity  string    `bun:"severity,notnull"` // "critical", "high", "medium", "low"
	CreatedAt time.Time `bun:"created_at,nullzero"`
	Resolved  bool      `bun:"resolved,default:false"`
}

type AIModel struct {
	ID          string
	Name        string
	Version     string
	Description string
	LoadedAt    time.Time
}

// LoginRequest - запрос на вход
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse - ответ на запрос входа с 2FA
type LoginResponse struct {
	RequiresTwoFA bool   `json:"requires_2fa"`
	UserID        int64  `json:"user_id,omitempty"`
	Message       string `json:"message"`
	TwoFAToken    string `json:"two_fa_token,omitempty"`
}

// TokenResponse - ответ с JWT токенами
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

type EventCorrelation struct {
	Type      string                 `json:"type"`      // например, "login_attempt" или "network_flow"
	SourceIP  string                 `json:"source_ip"` // IP атакующего или источника трафика
	HostID    string                 `json:"host_id"`   // Идентификатор хоста (ДОБАВЛЕНО)
	Success   bool                   `json:"success"`   // успешно или нет
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type Incident struct {
	bun.BaseModel `bun:"table:incidents,alias:i"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	Type        string    `bun:"type,notnull" json:"type"` // brute_force, ai_anomaly
	SourceIP    string    `bun:"source_ip" json:"source_ip"`
	ThreatLevel int       `bun:"threat_level" json:"threat_level"`
	Description string    `bun:"description" json:"description"`
	CreatedAt   time.Time `bun:"created_at,default:current_timestamp" json:"created_at"`
}

// ToIncidentDTO() преобразует модель в DTO для API
func (w *Incident) ToIncidentDTO() dto.IncidentDTO {
	return dto.IncidentDTO{
		ID:          w.ID,
		Type:        w.Type,
		SourceIP:    w.SourceIP,
		ThreatLevel: w.ThreatLevel,
		Description: w.Description,
		CreatedAt:   w.CreatedAt,
	}
}

// ScanResult — корневая структура ответа
type ScanResult struct {
	Data struct {
		ID         string         `json:"id"`
		Type       string         `json:"type"`
		Links      Links          `json:"links"`
		Attributes ScanAttributes `json:"attributes"`
	} `json:"data"`
}

type Links struct {
	Self string `json:"self"`
}

type ScanAttributes struct {
	Status  string            `json:"status"`
	Stats   Stats             `json:"stats"`
	Results map[string]Engine `json:"results"`
}

type Stats struct {
	Malicious  int `json:"malicious"`
	Suspicious int `json:"suspicious"`
	Undetected int `json:"undetected"`
	Harmless   int `json:"harmless"`
	Timeout    int `json:"timeout"`
}

type Engine struct {
	Category string `json:"category"`
	Result   string `json:"result"`
	Method   string `json:"method"`
	Engine   string `json:"engine"`
}

type VirusTotal struct {
	VirusTotalAPIKey string
	MaxFileSize      int64
	PollTimeout      time.Duration
	PollInterval     time.Duration
}
