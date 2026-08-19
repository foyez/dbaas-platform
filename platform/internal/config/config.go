// Package config provides application configuration management.
// It is responsible for loading, parsing, and exposing configuration
// values required by different parts of the application.
package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Env string

const (
	Development Env = "development"
	Production  Env = "production"
)

// Config represents the complete application configuration.
type Config struct {
	App     AppConfig
	Server  ServerConfig
	Zitadel ZitadelConfig
}

type AppConfig struct {
	Env      Env    `envconfig:"APP_ENV" default:"development"`
	LogLevel string `envconfig:"APP_LOG_LEVEL" default:"info"`
}

type ServerConfig struct {
	Address      string   `envconfig:"SERVER_ADDRESS" default:":8080"`
	AllowOrigins []string `envconfig:"SERVER_ALLOW_ORIGINS"`
	LokiURL      string   `envconfig:"LOKI_URL" default:"http://dbaas-loki:3100"`
}

type ZitadelConfig struct {
	Domain  string `envconfig:"ZITADEL_DOMAIN"`
	KeyPath string `envconfig:"ZITADEL_KEY_PATH"`
}

// Load loads configuration from enviroment variables
func Load() (*Config, error) {
	var cfg Config

	// Loads .env if present.
	// Existing environment variables always take precedence.
	_ = godotenv.Load()

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (e Env) Valid() bool {
	switch e {
	case Development, Production:
		return true
	default:
		return false
	}
}

func (e Env) IsProduction() bool {
	return e == Production
}

func (cfg *Config) validate() error {
	if !cfg.App.Env.Valid() {
		return fmt.Errorf(
			"invalid APP_ENV %q (must be %q or %q)",
			cfg.App.Env,
			Development,
			Production,
		)
	}

	if cfg.Zitadel.Domain == "" {
		return fmt.Errorf("ZITADEL_DOMAIN is required")
	}

	if cfg.Zitadel.KeyPath == "" {
		return fmt.Errorf("ZITADEL_KEY_PATH is required")
	}

	return nil
}
