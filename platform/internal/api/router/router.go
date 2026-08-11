// Package router configures the HTTP routes for the application.
// It is responsible for registering endpoints, applying middleware,
// and connecting routes to their corresponding handlers.
package router

import (
	"time"

	"github.com/foyez/dbaas-platform/platform/internal/api/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// New creates and configures a Gin HTTP server with all application routes
// and middleware registered.
func New(h *handler.InstanceHandler) *gin.Engine {
	r := gin.New()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(gin.Recovery(), gin.Logger())

	v1 := r.Group("/api/v1")
	{
		v1.POST("/instances", h.CreateInstance)
		v1.GET("/instances", h.ListInstances)
		v1.GET("/instances/:id", h.GetInstance)
		v1.PATCH("/instances/:id", h.UpdateInstance)
		v1.DELETE("/instances/:id", h.DeleteInstance)
	}

	return r
}
