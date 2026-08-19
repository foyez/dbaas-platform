// Package domain defines business-facing service contracts.
// Interfaces in this package describe what the application needs,
// while concrete implementations live in the service package.
package domain

import (
	"context"
)

// InstanceService defines business operations for managing PostgreSQL instances.
type InstanceService interface {
	CreateInstance(ctx context.Context, input CreateInstanceInput) (*CreateInstanceResult, error)
	GetCredentials(ctx context.Context, id, userID string) (*InstanceCredentials, error)
	GetInstance(ctx context.Context, id, userID string) (*Instance, error)
	ListInstances(ctx context.Context, userID string) (*ListInstancesResult, error)
	UpdateInstance(ctx context.Context, input UpdateInstanceInput) (*Instance, error)
	DeleteInstance(ctx context.Context, id, userID string) error

	// Future APIs:
	// GetInstance(...)
	// ListInstances(...)
	// DeleteInstance(...)
	// UpdateInstance(...)``
}
