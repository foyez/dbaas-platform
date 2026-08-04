// Package httpx provides shared HTTP transport utilities.
// It contains common response formats and helpers used by API handlers
// to return consistent HTTP responses.
package httpx

import "github.com/gin-gonic/gin"

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

func Error(c *gin.Context, status int, code ErrorCode, message string) {
	c.JSON(status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
