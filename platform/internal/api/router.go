// Package api configures the HTTP routes for the application.
// It is responsible for registering endpoints, applying middleware,
// and connecting routes to their corresponding handlers.
package api

import (
	"github.com/gin-gonic/gin"
)

// New creates and configures a Gin HTTP server with all application routes
// and middleware registered.
func NewRouter(h *InstanceHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	v1 := r.Group("/api/v1")
	{
		v1.POST("/instances", h.CreateInstance)
		v1.GET("/instances", h.ListInstances)
	}

	return r
}
