package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"linux-agent/internal/adapters/senders"

	"gopkg.in/yaml.v2"
)

// Config главная конфигурация
type Config struct {
	HostID    string        `yaml:"host_id"`
	Hostname  string        `yaml:"hostname"`
	Collector CollectorConf `yaml:"collector"`
	Elastic   ElasticConf   `yaml:"elastic"`
	LogLevel  string        `yaml:"log_level"`
}

// CollectorConf конфигурация коллектора
type CollectorConf struct {
	System    bool          `yaml:"system" default:"true"`
	Network   bool          `yaml:"network" default:"true"`
	Processes bool          `yaml:"processes" default:"false"`
	Security  bool          `yaml:"security" default:"true"`
	Docker    bool          `yaml:"docker" default:"false"`
	Interval  time.Duration `yaml:"interval" default:"60s"`
	BatchSize int           `yaml:"batch_size" default:"50"`
}

// ElasticConf конфигурация Elasticsearch
type ElasticConf struct {
	Enabled  bool     `yaml:"enabled" default:"false"`
	URLs     []string `yaml:"urls"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Index    string   `yaml:"index" default:"siem-logs-%{+yyyy.MM.dd}"`
}

// Load загружает конфигурацию из файла
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.setDefaults()
	return &config, nil
}

// LoadFromEnv загружает конфигурацию из переменных окружения
func LoadFromEnv() (*Config, error) {
	config := &Config{}

	// Базовые настройки
	config.HostID = os.Getenv("HOST_ID")
	config.Hostname = os.Getenv("HOSTNAME")
	config.LogLevel = getEnv("LOG_LEVEL", "info")

	// Настройки коллектора
	config.Collector.System = getEnvBool("COLLECTOR_SYSTEM", true)
	config.Collector.Network = getEnvBool("COLLECTOR_NETWORK", true)
	config.Collector.Processes = getEnvBool("COLLECTOR_PROCESSES", false)
	config.Collector.Security = getEnvBool("COLLECTOR_SECURITY", true)
	config.Collector.Docker = getEnvBool("COLLECTOR_DOCKER", false)
	config.Collector.BatchSize = getEnvInt("COLLECTOR_BATCH_SIZE", 50)

	if interval := os.Getenv("COLLECTOR_INTERVAL"); interval != "" {
		if dur, err := time.ParseDuration(interval); err == nil {
			config.Collector.Interval = dur
		}
	}

	// Настройки Elasticsearch
	config.Elastic.Enabled = getEnvBool("ELASTIC_ENABLED", false)
	if urls := os.Getenv("ELASTIC_URLS"); urls != "" {
		config.Elastic.URLs = splitEnv(urls)
	} else {
		// Для обратной совместимости
		if url := os.Getenv("ELASTIC_URL"); url != "" {
			config.Elastic.URLs = []string{url}
		}
	}
	config.Elastic.Username = os.Getenv("ELASTIC_USERNAME")
	config.Elastic.Password = os.Getenv("ELASTIC_PASSWORD")
	config.Elastic.Index = getEnv("ELASTIC_INDEX", "siem-logs-%{+yyyy.MM.dd}")

	config.setDefaults()
	return config, nil
}

func (c *Config) setDefaults() {
	// Hostname
	if c.Hostname == "" {
		hostname, err := os.Hostname()
		if err == nil {
			c.Hostname = hostname
		} else {
			c.Hostname = "unknown"
		}
	}

	// Host ID
	if c.HostID == "" {
		c.HostID = generateHostID(c.Hostname)
	}

	// Collector defaults
	if c.Collector.Interval == 0 {
		c.Collector.Interval = 60 * time.Second
	}
	if c.Collector.BatchSize == 0 {
		c.Collector.BatchSize = 50
	}

	// Elastic defaults
	if len(c.Elastic.URLs) == 0 {
		c.Elastic.URLs = []string{"http://localhost:9200"}
	}
	if c.Elastic.Index == "" {
		c.Elastic.Index = "siem-logs-%{+yyyy.MM.dd}"
	}

	// Log level
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
}

// ToElasticConfig преобразует в конфигурацию Elasticsearch отправителя
func (c *Config) ToElasticConfig() senders.ElasticsearchConfig {
	return senders.ElasticsearchConfig{
		URLs:     c.Elastic.URLs,
		Username: c.Elastic.Username,
		Password: c.Elastic.Password,
		Index:    c.Elastic.Index,
	}
}

func generateHostID(hostname string) string {
	// Пробуем прочитать machine-id
	if id, err := os.ReadFile("/etc/machine-id"); err == nil && len(id) > 0 {
		return strings.TrimSpace(string(id[:32]))
	}
	// Fallback
	return fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
}

// Вспомогательные функции
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func splitEnv(value string) []string {
	return strings.Split(value, ",")
}
