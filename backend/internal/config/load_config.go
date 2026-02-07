package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lpernett/godotenv"
)

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {

	// Загружаем .env
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:             getEnv("PORT", "8080"),
			CORSAllowOrigins: strings.Split(getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000"), ","),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_URL", ""),
		},
		Elastic: ElasticConfig{
			URL: getEnv("ELASTIC_URL", ""),
		},
		Integration: IntegrationConfig{
			VirusTotalAPIKey: getEnv("VIRUSTOTAL_API_KEY", ""),
			MaxFileSizeMB:    getEnvAsInt("MAX_FILE_SIZE", 32<<20), // 32MB по умолчанию
		},
		Security: SecurityConfig{
			WhitelistIPs: strings.Split(getEnv("WHITELIST_IPS", "127.0.0.1/32,10.0.0.0/8"), ","),
		},
		Auth: AuthConfig{
			Secret:    getEnv("JWT_SECRET", "your-secret-key"),
			TokenTTL:  getEnvAsInt("JWT_TTL", 86400), // 24 часа по умолчанию
			IssuerURL: getEnv("JWT_ISSUER", "railgun-core"),
		},
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("DATABASE_URL is required but not set in .env")
	}
	/*if cfg.Auth.Secret == "" {
		log.Println("WARNING: JWT_SECRET is not set, using default for development only")
	}*/

	cfg.Detection = DetectionConfig{
		BruteForceThreshold: getEnvAsInt("DETECTION_BF_THRESHOLD", 10),
		BruteForceWindow:    time.Duration(getEnvAsInt("DETECTION_BF_WINDOW_SEC", 60)) * time.Second,
	}

	if cfg.JWTConfig.Issuer == "" {
		cfg.JWTConfig.Issuer = "Railgun SIEM"
	}

	return cfg, nil
}

// LoadAgentConfig загружает конфигурацию для агента
func LoadAgentConfig() (*AgentConfig, error) {
	cfg := &AgentConfig{
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
