// Package config provides environment configuration management.
package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all environment configuration for the application.
type Config struct {
	DatabaseURL           string        `env:"DATABASE_URL"            envDefault:"postgres://user:password@localhost:5432/outbox_db?sslmode=disable"`
	RedisAddr             string        `env:"REDIS_ADDR"              envDefault:"localhost:6379"`
	Port                  string        `env:"PORT"                    envDefault:"8080"`
	PublisherPollInterval time.Duration `env:"PUBLISHER_POLL_INTERVAL" envDefault:"5s"`
	PublisherBatchSize    int           `env:"PUBLISHER_BATCH_SIZE"    envDefault:"10"`
	ConsumerName          string        `env:"CONSUMER_NAME"           envDefault:"consumer-1"`
	LogLevel              string        `env:"LOG_LEVEL"               envDefault:"info"`
}

// LoadConfig parses environment variables into Config struct.
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
