// Package router configures the HTTP routes for the application.
// It is responsible for registering endpoints, applying middleware,
// and connecting routes to their corresponding handlers.
package router

import (
	"net/http"
	"time"

	"github.com/foyez/dbaas-platform/platform/internal/api/handler"
	"github.com/foyez/dbaas-platform/platform/internal/auth"
	"github.com/foyez/dbaas-platform/platform/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// New creates and configures a Gin HTTP server with all application routes
// and middleware registered.
func New(h *handler.InstanceHandler, cfg config.ServerConfig, authmw *auth.Middleware) *gin.Engine {
	r := gin.New()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(gin.Recovery(), gin.Logger())

	// Health check for Kubernetes probes.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
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
		protected.PATCH("/instances/:id", h.UpdateInstance)
		protected.DELETE("/instances/:id", h.DeleteInstance)
	}

	// Admin users
	admin := protected.Group("")
	admin.Use(auth.Gin(authmw.RequireRole(auth.RoleAdmin)))

	// admin.GET("/instances/:id", h.GetInstance)
	// admin.DELETE("/instances/:id", h.DeleteInstance)

	return r
}
