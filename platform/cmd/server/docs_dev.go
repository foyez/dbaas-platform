//go:build dev

package main

import (
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupDocs(r *gin.Engine) {
	// Prevent browser caching while iterating on spec files in development.
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "GET" && (c.Request.URL.Path == "/openapi.yaml" || strings.HasPrefix(c.Request.URL.Path, "/docs/")) {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	// Backward compatible path.
	r.StaticFile("/openapi.yaml", "./docs/openapi.yaml")

	// Serve nested spec files used by external $ref entries.
	r.Static("/docs", "./docs")

	url := ginSwagger.URL("/docs/openapi.yaml")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
}
