// Package httpx provides error handling helpers for the HTTP layer.
// It maps internal service errors into appropriate HTTP status codes
// and API error responses.
package httpx

import (
	"errors"
	"net/http"

	"github.com/foyez/dbaas-platform/platform/internal/service"
	"github.com/gin-gonic/gin"
)

type errorMapping struct {
	err    error
	status int
	code   ErrorCode
}

var errorMappings = []errorMapping{
	{
		err:    service.ErrNotFound,
		status: http.StatusNotFound,
		code:   CodeNotFound,
	},
	{
		err:    service.ErrAlreadyExists,
		status: http.StatusConflict,
		code:   CodeInstanceAlreadyExists,
	},
}

// RespondError converts application errors into standardized HTTP responses.
func RespondError(c *gin.Context, err error) {
	for _, mapping := range errorMappings {
		if errors.Is(err, mapping.err) {
			Error(
				c,
				mapping.status,
				mapping.code,
				err.Error(),
			)
			return
		}
	}

	Error(
		c,
		http.StatusInternalServerError,
		CodeInternalError,
		"something went wrong",
	)
}
