// Package httpx provides shared HTTP transport utilities.
// It contains common response formats and helpers used by API handlers
// to return consistent HTTP responses.
package httpx

import "github.com/gin-gonic/gin"

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	// Details carries structured, code-specific data, e.g. QueryErrorDetails.
	Details any `json:"details,omitempty"`
}

// QueryErrorDetails is the Details of a QUERY_FAILED error: PostgreSQL's own
// description of why it rejected the statement.
type QueryErrorDetails struct {
	SQLState string `json:"sqlState"`
	Detail   string `json:"detail,omitempty"`
	Hint     string `json:"hint,omitempty"`
	// Position is the 1-based character offset of the error in the statement.
	Position int `json:"position,omitempty"`
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
