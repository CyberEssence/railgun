package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type CollectorConfig struct {
	System    bool          `yaml:"system" default:"true"`
	Network   bool          `yaml:"network" default:"true"`
	Processes bool          `yaml:"processes" default:"true"`
	Security  bool          `yaml:"security" default:"true"`
	Docker    bool          `yaml:"docker" default:"false"`
	Interval  time.Duration `yaml:"interval" default:"30s"`
}

type KafkaConfig struct {
	Enabled  bool     `yaml:"enabled" default:"true"`
	Brokers  []string `yaml:"brokers"`
	Topic    string   `yaml:"topic" default:"siem-logs"`
	SASL     bool     `yaml:"sasl" default:"false"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
}

type HTTPConfig struct {
	Enabled   bool   `yaml:"enabled" default:"true"`
	URL       string `yaml:"url" default:"http://localhost:8080"`
	Token     string `yaml:"token"`
	BatchSize int    `yaml:"batch_size" default:"100"`
}

type ElasticConfig struct {
	Enabled  bool   `yaml:"enabled" default:"false"`
	URL      string `yaml:"url" default:"http://localhost:9200"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Index    string `yaml:"index" default:"siem-logs-%{+yyyy.MM.dd}"`
}

type Config struct {
	HostID    string          `yaml:"host_id"`
	Hostname  string          `yaml:"hostname"`
	Collector CollectorConfig `yaml:"collector"`
	Kafka     KafkaConfig     `yaml:"kafka"`
	HTTP      HTTPConfig      `yaml:"http"`
	Elastic   ElasticConfig   `yaml:"elastic"`
	LogLevel  string          `yaml:"log_level" default:"info"`
}

// Load загружает конфигурацию из файла
func Load(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Устанавливаем дефолтные значения
	config.setDefaults()

	return &config, nil
}

// LoadFromEnv загружает конфигурацию из переменных окружения
func LoadFromEnv() (*Config, error) {
	config := &Config{}

	// Загружаем базовые настройки
	config.HostID = os.Getenv("HOST_ID")
	config.Hostname = os.Getenv("HOSTNAME")

	// Настройки коллектора
	config.Collector.System = getEnvBool("COLLECTOR_SYSTEM", true)
	config.Collector.Network = getEnvBool("COLLECTOR_NETWORK", true)
	config.Collector.Processes = getEnvBool("COLLECTOR_PROCESSES", true)
	config.Collector.Security = getEnvBool("COLLECTOR_SECURITY", true)
	config.Collector.Docker = getEnvBool("COLLECTOR_DOCKER", false)

	if interval := os.Getenv("COLLECTOR_INTERVAL"); interval != "" {
		if dur, err := time.ParseDuration(interval); err == nil {
			config.Collector.Interval = dur
		}
	}

	// Настройки Kafka
	config.Kafka.Enabled = getEnvBool("KAFKA_ENABLED", false)
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		config.Kafka.Brokers = splitEnv(brokers)
	}
	config.Kafka.Topic = getEnv("KAFKA_TOPIC", "siem-logs")
	config.Kafka.SASL = getEnvBool("KAFKA_SASL", false)
	config.Kafka.Username = os.Getenv("KAFKA_USERNAME")
	config.Kafka.Password = os.Getenv("KAFKA_PASSWORD")

	// Настройки HTTP
	config.HTTP.Enabled = getEnvBool("HTTP_ENABLED", true)
	config.HTTP.URL = getEnv("HTTP_URL", "http://localhost:8080")
	config.HTTP.Token = os.Getenv("HTTP_TOKEN")

	if batchSize := os.Getenv("HTTP_BATCH_SIZE"); batchSize != "" {
		if size, err := strconv.Atoi(batchSize); err == nil {
			config.HTTP.BatchSize = size
		}
	}

	// Настройки Elasticsearch
	config.Elastic.Enabled = getEnvBool("ELASTIC_ENABLED", false)
	config.Elastic.URL = getEnv("ELASTIC_URL", "http://localhost:9200")
	config.Elastic.Username = os.Getenv("ELASTIC_USERNAME")
	config.Elastic.Password = os.Getenv("ELASTIC_PASSWORD")
	config.Elastic.Index = getEnv("ELASTIC_INDEX", "siem-logs-%{+yyyy.MM.dd}")

	config.LogLevel = getEnv("LOG_LEVEL", "info")

	config.setDefaults()

	return config, nil
}

func (c *Config) setDefaults() {
	// Если hostname не установлен, получаем из системы
	if c.Hostname == "" {
		hostname, err := os.Hostname()
		if err == nil {
			c.Hostname = hostname
		} else {
			c.Hostname = "unknown"
		}
	}

	// Если host_id не установлен, генерируем из hostname
	if c.HostID == "" {
		c.HostID = generateHostID(c.Hostname)
	}

	// Устанавливаем дефолтный интервал если не задан
	if c.Collector.Interval == 0 {
		c.Collector.Interval = 30 * time.Second
	}

	// Дефолтный batch size
	if c.HTTP.BatchSize == 0 {
		c.HTTP.BatchSize = 100
	}
}

func generateHostID(hostname string) string {
	// Генерация простого ID на основе hostname и времени
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
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

func splitEnv(value string) []string {
	return strings.Split(value, ",")
}
