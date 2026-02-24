package config

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Elastic     ElasticConfig     `yaml:"elastic"`
	Integration IntegrationConfig `yaml:"integration"`
	Security    SecurityConfig    `yaml:"security"`
	Auth        AuthConfig        `yaml:"auth"`
	Detection   DetectionConfig   `yaml:"detection"`
	JWTConfig   JWTConfig         `yaml:"jwt"`
}
