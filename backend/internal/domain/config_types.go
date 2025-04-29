package domain

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
	MaxFileSizeMB    int
}

type SecurityConfig struct {
	WhitelistIPs []string
}

type AuthConfig struct {
	Secret    string
	TokenTTL  int
	IssuerURL string
}

type AgentConfig struct {
	ServerURL  string
	APIKey     string
	HostID     string
	Interval   int
	LogLevel   string
	MaxRetries int
}
