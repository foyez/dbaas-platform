// Package httpx provides error handling helpers for the HTTP layer.
// It maps internal service errors into appropriate HTTP status codes
// and API error responses.
package httpx

import (
	"errors"
	"net/http"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
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
	{
		err:    service.ErrMissingIdempotencyKey,
		status: http.StatusBadRequest,
		code:   CodeMissingIdempotencyKey,
	},
	{
		err:    service.ErrNoUpdateFields,
		status: http.StatusBadRequest,
		code:   CodeNoUpdateFields,
	},
	{
		err:    service.ErrInvalidInput,
		status: http.StatusBadRequest,
		code:   CodeInvalidRequest,
	},
	{
		err:    service.ErrQueryTimeout,
		status: http.StatusUnprocessableEntity,
		code:   CodeQueryTimeout,
	},
	{
		err:    service.ErrInstanceNotReady,
		status: http.StatusConflict,
		code:   CodeInstanceNotReady,
	},
}

// RespondError converts application errors into standardized HTTP responses.
func RespondError(c *gin.Context, err error) {
	// The database rejected a SQL console statement: return its own error,
	// so the user can see what's wrong with their SQL.
	var queryErr *domain.QueryError
	if errors.As(err, &queryErr) {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Error: ErrorDetail{
				Code:    CodeQueryFailed,
				Message: queryErr.Message,
				Details: QueryErrorDetails{
					SQLState: queryErr.SQLState,
					Detail:   queryErr.Detail,
					Hint:     queryErr.Hint,
					Position: queryErr.Position,
				},
			},
		})
		return
	}

	for _, mapping := range errorMappings {
		if errors.Is(err, mapping.err) {
			// The sentinel's message, not err.Error(): wrapping adds
			// internal context ("executing query: ...") clients shouldn't see.
			Error(
				c,
				mapping.status,
				mapping.code,
				mapping.err.Error(),
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
