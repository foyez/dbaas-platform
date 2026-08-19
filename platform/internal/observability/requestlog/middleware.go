package requestlog

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware replaces gin.Logger() with structured slog output, so access
// logs are queryable in Loki via `| json` instead of regex over plain text.
func Middleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		log.Info("http_request",
			"method", c.Request.Method,
			"route", route,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"error", c.Errors.String(), // populated if a handler called c.Error(err)
		)
	}
}
