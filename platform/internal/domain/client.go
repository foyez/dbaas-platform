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

	// Future operations:
	// GetInstance(...)
	// ListInstances(...)
	// DeleteInstance(...)
}
