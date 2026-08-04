package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	setupDocs(r)

	r.Run(":8080")
}
