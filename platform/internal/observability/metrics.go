package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

// var (
// 	HTTPRequestsTotal = prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Name: "paas_http_requests_total",
// 			Help: "Total number of HTTP requests.",
// 		},
// 		[]string{"method", "route", "status"},
// 	)

// 	HTTPRequestDuration = prometheus.NewHistogramVec(
// 		prometheus.HistogramOpts{
// 			Name: "paas_http_request_duration_seconds",
// 			Help: "HTTP request duration in seconds.",
// 			Buckets: []float64{
// 				0.005,
// 				0.01,
// 				0.025,
// 				0.05,
// 				0.1,
// 				0.25,
// 				0.5,
// 				1,
// 				2.5,
// 				5,
// 				10,
// 			},
// 		},
// 		[]string{"method", "route", "status"},
// 	)

// 	InstanceOperationTotal = prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Name: "paas_instance_operations_total",
// 			Help: "Total number of instance operations.",
// 		},
// 		[]string{"operation", "result"},
// 	)
// )

// func RegisterMetrics() {
// 	prometheus.MustRegister(HTTPRequestsTotal)
// 	prometheus.MustRegister(HTTPRequestDuration)
// 	prometheus.MustRegister(InstanceOperationTotal)
// }

// func Middleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		start := time.Now()

// 		c.Next()

// 		duration := time.Since(start).Seconds()

// 		route := c.FullPath()

// 		// If Gin couldn't resolve the route, avoid
// 		// creating path lables.
// 		if route == "" {
// 			route = "unknown"
// 		}

// 		method := c.Request.Method
// 		status := strconv.Itoa(c.Writer.Status())

// 		HTTPRequestsTotal.WithLabelValues(
// 			method,
// 			route,
// 			status,
// 		).Inc()

// 		HTTPRequestDuration.WithLabelValues(
// 			method,
// 			route,
// 			status,
// 		).Observe(duration)
// 	}
// }

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "dbaas",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed.",
			},
			[]string{"method", "route", "status"},
		),

		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "dbaas",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
			},
			[]string{"method", "route"},
		),
	}
}

func (m *Metrics) Register(registry prometheus.Registerer) error {
	if err := registry.Register(m.HTTPRequestsTotal); err != nil {
		return err
	}

	if err := registry.Register(m.HTTPRequestDuration); err != nil {
		return err
	}

	return nil
}
