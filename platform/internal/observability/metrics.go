package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	// SQLQueriesTotal counts SQL console statements by outcome: ok,
	// query_failed (rejected by PostgreSQL), timeout, not_ready or error.
	SQLQueriesTotal *prometheus.CounterVec
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
		SQLQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "dbaas",
				Subsystem: "sql_console",
				Name:      "queries_total",
				Help:      "SQL console statements by outcome.",
			},
			[]string{"outcome"},
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
	if err := registry.Register(m.SQLQueriesTotal); err != nil {
		return err
	}

	if err := registry.Register(collectors.NewGoCollector()); err != nil {
		return err
	}

	if err := registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		return err
	}

	return nil
}
