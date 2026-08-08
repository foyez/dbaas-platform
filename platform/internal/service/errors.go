// Package service defines errors returned by business service operations.
// These errors represent expected business failures and allow callers
// to handle failures without depending on implementation details.
package service

import (
	"errors"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
)

var (
	ErrNotFound      = domain.ErrNotFound
	ErrAlreadyExists = domain.ErrAlreadyExists

	ErrInvalidInput          = errors.New("invalid input")
	ErrMissingIdempotencyKey = errors.New("missing idempotency key")
	ErrNoUpdateFields        = errors.New("no fields provided to update")
)
