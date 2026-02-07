package config

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Elastic     ElasticConfig
	Integration IntegrationConfig
	Security    SecurityConfig
	Auth        AuthConfig
	Detection   DetectionConfig
	JWTConfig   JWTConfig `yaml:"security"`
}
