package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	HostID    string          `yaml:"host_id"`
	Hostname  string          `yaml:"hostname"`
	Collector CollectorConfig `yaml:"collector"`
	Sender    SenderConfig    `yaml:"sender"`
	SIEM      SIEMConfig      `yaml:"siem"`    // Приоритет 1 - прямо в SIEM
	Kafka     KafkaConfig     `yaml:"kafka"`   // Приоритет 2 - Kafka
	Elastic   ElasticConfig   `yaml:"elastic"` // Приоритет 3 - Elasticsearch
	LogLevel  string          `yaml:"log_level"`
}

type CollectorConfig struct {
	System  bool `yaml:"system" default:"true"`
	Network bool `yaml:"network" default:"true"`
	// Отключили по умолчанию, иначе будет большая нагрузка
	Processes bool          `yaml:"processes" default:"false"`
	Security  bool          `yaml:"security" default:"true"`
	Docker    bool          `yaml:"docker" default:"false"`
	Interval  time.Duration `yaml:"interval" default:"60s"`
}

type SenderConfig struct {
	BatchSize     int           `yaml:"batch_size" default:"50"`
	FlushInterval time.Duration `yaml:"flush_interval" default:"10s"`
	RetryCount    int           `yaml:"retry_count" default:"3"`
	RetryBackoff  time.Duration `yaml:"retry_backoff" default:"1s"`
}

type SIEMConfig struct {
	Enabled bool   `yaml:"enabled" default:"true"`
	URL     string `yaml:"url" default:"http://localhost:8080"`
	Token   string `yaml:"token"`
}

type KafkaConfig struct {
	Enabled  bool     `yaml:"enabled" default:"false"`
	Brokers  []string `yaml:"brokers"`
	Topic    string   `yaml:"topic" default:"siem-logs"`
	SASL     bool     `yaml:"sasl" default:"false"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
}

type ElasticConfig struct {
	Enabled  bool   `yaml:"enabled" default:"false"`
	URL      string `yaml:"url" default:"http://localhost:9200"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Index    string `yaml:"index" default:"siem-logs-%{+yyyy.MM.dd}"`
}

// Load загружает конфигурацию из файла
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
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

	if interval := os.Getenv("COLLECTOR_INTERVAL"); interval != "" {
		if dur, err := time.ParseDuration(interval); err == nil {
			config.Collector.Interval = dur
		}
	}

	// Настройки отправителя
	if batchSize := os.Getenv("SENDER_BATCH_SIZE"); batchSize != "" {
		if size, err := strconv.Atoi(batchSize); err == nil {
			config.Sender.BatchSize = size
		}
	}

	if flushInterval := os.Getenv("SENDER_FLUSH_INTERVAL"); flushInterval != "" {
		if dur, err := time.ParseDuration(flushInterval); err == nil {
			config.Sender.FlushInterval = dur
		}
	}

	// Настройки SIEM (приоритет 1)
	config.SIEM.Enabled = getEnvBool("SIEM_ENABLED", true)
	config.SIEM.URL = getEnv("SIEM_URL", "http://localhost:8080")
	config.SIEM.Token = os.Getenv("SIEM_TOKEN")

	// Настройки Kafka (приоритет 2)
	config.Kafka.Enabled = getEnvBool("KAFKA_ENABLED", false)
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		config.Kafka.Brokers = splitEnv(brokers)
	}
	config.Kafka.Topic = getEnv("KAFKA_TOPIC", "siem-logs")
	config.Kafka.SASL = getEnvBool("KAFKA_SASL", false)
	config.Kafka.Username = os.Getenv("KAFKA_USERNAME")
	config.Kafka.Password = os.Getenv("KAFKA_PASSWORD")

	// Настройки Elasticsearch (приоритет 3)
	config.Elastic.Enabled = getEnvBool("ELASTIC_ENABLED", false)
	config.Elastic.URL = getEnv("ELASTIC_URL", "http://localhost:9200")
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

	// Sender defaults
	if c.Sender.BatchSize == 0 {
		c.Sender.BatchSize = 50
	}
	if c.Sender.FlushInterval == 0 {
		c.Sender.FlushInterval = 10 * time.Second
	}
	if c.Sender.RetryCount == 0 {
		c.Sender.RetryCount = 3
	}
	if c.Sender.RetryBackoff == 0 {
		c.Sender.RetryBackoff = 1 * time.Second
	}

	// SIEM defaults
	if c.SIEM.URL == "" {
		c.SIEM.URL = "http://localhost:8080"
	}

	// Kafka defaults
	if c.Kafka.Topic == "" {
		c.Kafka.Topic = "siem-logs"
	}

	// Elasticsearch defaults
	if c.Elastic.URL == "" {
		c.Elastic.URL = "http://localhost:9200"
	}
	if c.Elastic.Index == "" {
		c.Elastic.Index = "siem-logs-%{+yyyy.MM.dd}"
	}

	// Log level
	if c.LogLevel == "" {
		c.LogLevel = "info"
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

func splitEnv(value string) []string {
	return strings.Split(value, ",")
}
