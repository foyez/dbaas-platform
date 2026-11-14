// Package router configures the HTTP routes for the application.
// It is responsible for registering endpoints, applying middleware,
// and connecting routes to their corresponding handlers.
package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/foyez/dbaas-platform/platform/docs"
	"github.com/foyez/dbaas-platform/platform/internal/api/handler"
	"github.com/foyez/dbaas-platform/platform/internal/auth"
	"github.com/foyez/dbaas-platform/platform/internal/config"
	"github.com/foyez/dbaas-platform/platform/internal/observability"
	"github.com/foyez/dbaas-platform/platform/internal/observability/audit"
	"github.com/foyez/dbaas-platform/platform/internal/observability/requestlog"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// New creates and configures a Gin HTTP server with all application routes
// and middleware registered.
func New(
	h *handler.InstanceHandler,
	log *slog.Logger,
	cfg config.ServerConfig,
	authmw *auth.Middleware,
	metrics *observability.Metrics,
	registry *prometheus.Registry,
	auditLogger *audit.AuditLogger,
) *gin.Engine {
	// Release mode: request logs come from requestlog, not gin's debug output.
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Tracing first, so every later middleware and handler runs inside the
	// request's span (and logs with its trace_id). Probes and scrapes run
	// every few seconds and aren't worth a trace each.
	r.Use(otelgin.Middleware("dbaas-api", otelgin.WithFilter(func(req *http.Request) bool {
		return req.URL.Path != "/health" && req.URL.Path != "/metrics"
	})))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(gin.Recovery())
	r.Use(requestlog.Middleware(log))
	r.Use(observability.MetricsMiddleware(metrics))
	r.Use(audit.Middleware(auditLogger, authmw.UserID))

	// Health check for Kubernetes probes.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Prometheus scrape target
	r.GET("/metrics", gin.WrapH(
		promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{},
		),
	))

	// API reference: the OpenAPI description and a Swagger UI page for it.
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", docs.OpenAPI)
	})
	r.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", docs.SwaggerUI)
	})

	v1 := r.Group("/api/v1")

	// Authentication boundary.
	//
	// Everything under `protected` requires a valid
	// ZITADEL access token.
	protected := v1.Group("")
	protected.Use(auth.Gin(authmw.RequireAuth()))

	// Authenticated users
	{
		protected.POST("/instances", h.CreateInstance)
		protected.GET("/instances", h.ListInstances)
		protected.GET("/instances/:id", h.GetInstance)
		protected.GET("/instances/:id/credentials", h.GetCredentials)
		protected.GET("/instances/:id/logs", h.GetInstanceLogs)
		protected.POST("/instances/:id/query", h.ExecuteQuery)
		protected.PATCH("/instances/:id", h.UpdateInstance)
		protected.DELETE("/instances/:id", h.DeleteInstance)
	}

	// Admin users
	admin := protected.Group("")
	admin.Use(auth.Gin(authmw.RequireRole(auth.RoleAdmin)))

	admin.GET("/audit-logs", h.GetAuditLogs)
	// admin.GET("/instances/:id", h.GetInstance)
	// admin.DELETE("/instances/:id", h.DeleteInstance)

	return r
}
