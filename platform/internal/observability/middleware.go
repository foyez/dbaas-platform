package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		status := strconv.Itoa(c.Writer.Status())

		metrics.HTTPRequestsTotal.
			WithLabelValues(
				c.Request.Method,
				route,
				status,
			).Inc()

		metrics.HTTPRequestDuration.
			WithLabelValues(
				c.Request.Method,
				route,
			).Observe(duration)
	}
}
