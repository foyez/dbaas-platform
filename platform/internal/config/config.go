// Package config provides application configuration management.
// It is responsible for loading, parsing, and exposing configuration
// values required by different parts of the application.
package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Env string

const (
	Development Env = "development"
	Production  Env = "production"
)

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

// Config represents the complete application configuration.
type Config struct {
	App struct {
		Env      Env    `envconfig:"APP_ENV" default:"development"`
		LogLevel string `envconfig:"APP_LOG_LEVEL" default:"info"`
	}
	Server struct {
		Address string `envconfig:"SERVER_ADDRESS" default:":8080"`
	}
}

// Load loads configuration from enviroment variables
func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	if !cfg.App.Env.Valid() {
		return nil, fmt.Errorf(
			"invalid APP_ENV %q (must be %q or %q)",
			cfg.App.Env,
			Development,
			Production,
		)
	}

	return &cfg, nil
}
