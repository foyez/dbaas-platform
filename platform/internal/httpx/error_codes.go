// Package httpx defines API-level error codes returned to HTTP clients.
// These codes are part of the external API contract and should remain
// stable even if internal implementation details change.
package httpx

type ErrorCode string

const (
	CodeInvalidRequest        ErrorCode = "INVALID_REQUEST"
	CodeUnauthorized          ErrorCode = "UNAUTHORIZED"
	CodeForbidden             ErrorCode = "FORBIDDEN"
	CodeNotFound              ErrorCode = "NOT_FOUND"
	CodeInternalError         ErrorCode = "INTERNAL_ERROR"
	CodeInstanceNotFound      ErrorCode = "INSTANCE_NOT_FOUND"
	CodeInstanceAlreadyExists ErrorCode = "INSTANCE_ALREADY_EXISTS"
	CodeMissingIdempotencyKey ErrorCode = "MISSING_IDEMPOTENCY_KEY"
)
