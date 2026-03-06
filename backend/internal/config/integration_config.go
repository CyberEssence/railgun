package config

import "time"

type IntegrationConfig struct {
	VirusTotalAPIKey string        `yaml:"api_key" mapstructure:"api_key"`
	MaxFileSizeMB    int64         `yaml:"max_file_size" mapstructure:"max_file_size"`
	PollInterval     time.Duration `yaml:"poll_interval" mapstructure:"poll_interval"`
	PollTimeout      time.Duration `yaml:"poll_timeout" mapstructure:"poll_timeout"`
}
