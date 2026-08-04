// Package logger provides application-wide logging setup and utilities.
// It is responsible for creating and configuring the logger used by
// application components.
package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/foyez/dbaas-platform/platform/internal/config"
)

// New creates a configured application logger.
func New(logLevel string, env config.Env) *slog.Logger {
	var level slog.Level

	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	if env.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts.AddSource = true
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
