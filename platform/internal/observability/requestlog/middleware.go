package requestlog

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware replaces gin.Logger() with structured slog output, so access
// logs are queryable in Loki via `| json` instead of regex over plain text.
// Logging with the request context adds trace_id when the request is traced.
func Middleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		level := slog.LevelInfo
		switch {
		case route == "/health" || route == "/metrics":
			// Probes and scrapes arrive every few seconds; only log them
			// when debugging.
			level = slog.LevelDebug
		case c.Writer.Status() >= 500:
			level = slog.LevelError
		}

		log.Log(c.Request.Context(), level, "http_request",
			"method", c.Request.Method,
			"route", route,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"error", c.Errors.String(), // populated if a handler called c.Error(err)
		)
	}
}
