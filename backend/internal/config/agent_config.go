package config

type AgentConfig struct {
	ServerURL  string
	APIKey     string
	HostID     string
	Interval   int
	LogLevel   string
	MaxRetries int
}
