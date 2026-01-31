package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the UI server
type Config struct {
	Server ServerConfig
	Broker BrokerConfig
	UI     UIConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string
	Host string
}

// BrokerConfig holds broker connection configuration
type BrokerConfig struct {
	Type  string // redis, nats-js, or in-memory
	Redis RedisConfig
	NATS  NATSConfig
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NATSConfig holds NATS JetStream configuration
type NATSConfig struct {
	URL    string
	Stream string
}

// UIConfig holds UI-specific settings
type UIConfig struct {
	RefreshInterval time.Duration
	MaxJobsDisplay  int
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Broker: BrokerConfig{
			Type: "redis",
			Redis: RedisConfig{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			},
			NATS: NATSConfig{
				URL:    "nats://localhost:4222",
				Stream: "TASQUEUE",
			},
		},
		UI: UIConfig{
			RefreshInterval: 3 * time.Second,
			MaxJobsDisplay:  100,
		},
	}
}

// LoadFromEnv loads configuration from environment variables
// This is a simple implementation - can be enhanced with viper or similar
func LoadFromEnv() (Config, error) {
	cfg := DefaultConfig()
	// TODO: Add environment variable parsing if needed
	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Broker.Type != "redis" && c.Broker.Type != "nats-js" && c.Broker.Type != "in-memory" {
		return fmt.Errorf("invalid broker type: %s (must be redis, nats-js, or in-memory)", c.Broker.Type)
	}

	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	return nil
}
