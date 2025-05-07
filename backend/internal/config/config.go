package config

import (
	"os"
	"strconv"
	"strings"

	"railgun-core/internal/domain"
)

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() (*domain.Config, error) {
	cfg := &domain.Config{
		Server: domain.ServerConfig{
			Port:             getEnv("PORT", "8080"),
			CORSAllowOrigins: strings.Split(getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000"), ","),
		},
		Database: domain.DatabaseConfig{
			DSN: getEnv("DATABASE_URL", "postgres://postgres:(2a+3b=0c+1d)@localhost:5432/siem?sslmode=disable"),
		},
		Elastic: domain.ElasticConfig{
			URL: getEnv("ELASTIC_URL", "http://localhost:9200"),
		},
		Integration: domain.IntegrationConfig{
			VirusTotalAPIKey: getEnv("VIRUSTOTAL_API_KEY", "8951c3c0f005ceac53a4f694723a7071a35ea0a9e2f977c66e6d3ac55c4ef9b2"),
			MaxFileSizeMB:    getEnvAsInt("MAX_FILE_SIZE", 32<<20), // 32MB по умолчанию
		},
		Security: domain.SecurityConfig{
			WhitelistIPs: strings.Split(getEnv("WHITELIST_IPS", "127.0.0.1/32,10.0.0.0/8"), ","),
		},
		Auth: domain.AuthConfig{
			Secret:    getEnv("JWT_SECRET", "your-secret-key"),
			TokenTTL:  getEnvAsInt("JWT_TTL", 86400), // 24 часа по умолчанию
			IssuerURL: getEnv("JWT_ISSUER", "railgun-core"),
		},
	}

	return cfg, nil
}

// LoadAgentConfig загружает конфигурацию для агента
func LoadAgentConfig() (*domain.AgentConfig, error) {
	cfg := &domain.AgentConfig{
		ServerURL:  getEnv("SERVER_URL", "http://localhost:8080"),
		APIKey:     getEnv("API_KEY", ""),
		HostID:     getEnv("HOST_ID", ""),
		Interval:   getEnvAsInt("INTERVAL", 60), // 60 секунд по умолчанию
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		MaxRetries: getEnvAsInt("MAX_RETRIES", 3),
	}

	return cfg, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает числовое значение переменной окружения
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
