package config

import (
	"fmt"
	"sync"

	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
)

var (
	instance *Config
	once     sync.Once
)

// LoadConfig loads the configuration from .env and command-line flags.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}
	fp := flags.NewParser(&cfg, flags.Default)
	// Parse flags
	if _, err := fp.Parse(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetConfig returns the singleton instance of Config.
func GetConfig() (*Config, error) {
	var err error
	once.Do(func() {
		instance, err = LoadConfig()
	})
	return instance, err
}

type Config struct {
	Debug bool `long:"debug" env:"DEBUG"`

	Server struct {
		Port int `long:"server-port" env:"SERVER_PORT" default:"8089"`
	}

	Proxy struct {
		Credentials struct {
			User     string `long:"proxy-credentials-user" env:"PROXY_CRED_USER"`
			Password string `long:"proxy-credentials-password" env:"PROXY_CRED_PASSWORD"`
		}
	}
}
