// Package domain defines interfaces for external dependencies required
// by the application's business services.
package domain

import "context"

// InstanceClient defines operations for interacting with the
// underlying Kubernetes/CNPG infrastructure.
type InstanceClient interface {
	CreateInstance(
		ctx context.Context,
		input CreateInstanceInput,
	) (*Instance, error)

	GetInstanceByIdempotencyKey(
		ctx context.Context,
		key string,
	) (*Instance, error)

	GetInstance(ctx context.Context, id string) (*Instance, error)
	ListInstances(ctx context.Context) ([]*Instance, error)
	UpdateInstance(
		ctx context.Context,
		input UpdateInstanceInput,
	) (*Instance, error)
	DeleteInstance(ctx context.Context, id string) error

	// Future operations:
	// GetInstance(...)
	// ListInstances(...)
	// DeleteInstance(...)
}
