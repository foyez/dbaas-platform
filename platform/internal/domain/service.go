// Package domain defines business-facing service contracts.
// Interfaces in this package describe what the application needs,
// while concrete implementations live in the service package.
package domain

import "context"

// InstanceService defines business operations for managing PostgreSQL instances.
type InstanceService interface {
	CreateInstance(ctx context.Context, input CreateInstanceInput) (*CreateInstanceResult, error)
	ListInstances(ctx context.Context) (*ListInstancesResult, error)
	DeleteInstance(ctx context.Context, id string) error

	// Future APIs:
	// GetInstance(...)
	// ListInstances(...)
	// DeleteInstance(...)
	// UpdateInstance(...)``
}
