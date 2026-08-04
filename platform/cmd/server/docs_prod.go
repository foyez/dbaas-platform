//go:build !dev

package main

import "github.com/gin-gonic/gin"

func setupDocs(r *gin.Engine) {
	// Swagger disabled in production
}
