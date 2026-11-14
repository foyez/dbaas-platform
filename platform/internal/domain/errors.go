// Package domain defines core business models and rules.
// Errors declared here represent domain-level outcomes that must be
// producible by any InstanceClient implementation (e.g. infra/cnpg),
// not just by the service layer. Keeping them in domain avoids infra
// having to import service, which would invert the dependency direction.
package domain

import "errors"

var (
	ErrNotFound         = errors.New("instance not found")
	ErrAlreadyExists    = errors.New("instance name already exists")
	ErrInstanceNotReady = errors.New("instance not ready")

	// ErrQueryTimeout means a SQL console statement ran past its time
	// limit and was cancelled on the server.
	ErrQueryTimeout = errors.New("query exceeded the time limit and was cancelled")
)
