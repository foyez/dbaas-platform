// Package service defines errors returned by business service operations.
// These errors represent expected business failures and allow callers
// to handle failures without depending on implementation details.
package service

import "errors"

var (
	ErrAlreadyExists         = errors.New("instance name already exists")
	ErrNotFound              = errors.New("instance not found")
	ErrInvalidInput          = errors.New("invalid input")
	ErrMissingIdempotencyKey = errors.New("missing idempotency key")
)
