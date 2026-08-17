package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "paas_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "paas_http_request_duration_seconds",
			Help: "HTTP request duration in seconds.",
			Buckets: []float64{
				0.005,
				0.01,
				0.025,
				0.05,
				0.1,
				0.25,
				0.5,
				1,
				2.5,
				5,
				10,
			},
		},
		[]string{"method", "route", "status"},
	)

	InstanceOperationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "paas_instance_operations_total",
			Help: "Total number of instance operations.",
		},
		[]string{"operation", "result"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(InstanceOperationTotal)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		route := c.FullPath()

		// If Gin couldn't resolve the route, avoid
		// creating path lables.
		if route == "" {
			route = "unknown"
		}

		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		HTTPRequestsTotal.WithLabelValues(
			method,
			route,
			status,
		).Inc()

		HTTPRequestDuration.WithLabelValues(
			method,
			route,
			status,
		).Observe(duration)
	}
}
