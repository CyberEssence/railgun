package models

import (
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
	ID          string    `json:"id"`
	Hostname    string    `json:"hostname"`
	IPAddress   string    `json:"ip_address"`
	LastSeen    time.Time `json:"last_seen"`
	OSVersion   string    `json:"os_version"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
}

// NetworkTraffic представляет запись о сетевом трафике
type NetworkTraffic struct {
	ID          int64     `json:"id"`
	HostID      string    `json:"host_id"`
	Timestamp   time.Time `json:"timestamp"`
	SrcIP       string    `json:"src_ip"`
	DstIP       string    `json:"dst_ip"`
	SrcPort     int       `json:"src_port"`
	DstPort     int       `json:"dst_port"`
	Protocol    string    `json:"protocol"`
	BytesSent   int64     `json:"bytes_sent"`
	BytesRecv   int64     `json:"bytes_recv"`
	PacketsSent int64     `json:"packets_sent"`
	PacketsRecv int64     `json:"packets_recv"`
	Duration    float64   `json:"duration"`
}

// WindowsArtifact представляет системный артефакт Windows
type WindowsArtifact struct {
	bun.BaseModel `bun:"table:windows_artifacts,alias:wa"`

	ID          int64                  `bun:"id,pk,autoincrement" json:"id"`
	HostID      string                 `bun:"host_id,notnull" json:"host_id"`
	Timestamp   time.Time              `bun:"timestamp,notnull" json:"timestamp"`
	Type        string                 `bun:"type,notnull" json:"type"`
	Path        string                 `bun:"path" json:"path"`
	Value       string                 `bun:"value" json:"value"`
	Size        int64                  `bun:"size" json:"size"`
	Hash        string                 `bun:"hash" json:"hash"`
	Owner       string                 `bun:"owner" json:"owner"`
	Permissions string                 `bun:"permissions" json:"permissions"`
	Metadata    map[string]interface{} `bun:"metadata,type:jsonb" json:"metadata"`
	// Добавьте threat_level, если он нужен
	ThreatLevel int `bun:"threat_level" json:"threat_level"`
}

// User представляет пользователя системы
type User struct {
	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	Username     string    `bun:"username,unique,notnull" json:"username"`
	Email        string    `bun:"email,unique,notnull" json:"email"`
	PasswordHash string    `bun:"password_hash,notnull" json:"-"`
	IsActive     bool      `bun:"is_active,default:true" json:"is_active"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	LastLogin    time.Time `bun:"last_login" json:"last_login"`
}

// TwoFAToken модель для хранения токенов 2FA
type TwoFAToken struct {
	bun.BaseModel `bun:"table:two_fa_tokens,alias:t"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID    int64     `bun:"user_id,notnull" json:"user_id"`
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
	ID            int64     `json:"id"`
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

// ScanResult результат сканирования файла
type ScanResult struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			Status  string `json:"status"`
			Results map[string]struct {
				Category string `json:"category"`
				Result   string `json:"result"`
			} `json:"results"`
			Stats struct {
				Malicious  int `json:"malicious"`
				Suspicious int `json:"suspicious"`
				Undetected int `json:"undetected"`
			} `json:"stats"`
		} `json:"attributes"`
	} `json:"data"`
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

// AnalysisResult результат анализа данных
type AnalysisResult struct {
	ThreatLevel     int       `json:"threat_level"`
	Confidence      float64   `json:"confidence"`
	DetectedThreats []string  `json:"detected_threats"`
	Recommendations []string  `json:"recommendations"`
	Timestamp       time.Time `json:"timestamp"`
}

type Artifact struct {
	ID          int64     `json:"id"`
	HostID      string    `json:"host_id"`
	Type        string    `json:"type"` // file, registry, process и т.д.
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Hash        string    `json:"hash"`
	Timestamp   time.Time `json:"timestamp"`
	ThreatLevel int       `json:"threat_level"`
	Content     []byte    `json:"content,omitempty"`
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
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}
