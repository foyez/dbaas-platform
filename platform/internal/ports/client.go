// Package ports defines interfaces for external dependencies required
// by the application's business services.
package ports

import (
	"context"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
)

// InstanceClient defines operations for interacting with the
// underlying Kubernetes/CNPG infrastructure.
type InstanceClient interface {
	CreateInstance(
		ctx context.Context,
		input domain.CreateInstanceInput,
	) (*domain.Instance, error)

	GetInstanceByIdempotencyKey(
		ctx context.Context,
		key string,
	) (*domain.Instance, error)

	GetInstance(ctx context.Context, id, userID string) (*domain.Instance, error)
	GetInstanceCredentials(
		ctx context.Context,
		id string,
		userID string,
	) (*domain.InstanceCredentials, error)

	ListInstances(ctx context.Context, userID string) ([]*domain.Instance, error)
	UpdateInstance(
		ctx context.Context,
		input domain.UpdateInstanceInput,
	) (*domain.Instance, error)
	DeleteInstance(ctx context.Context, id, userID string) error

	// Future operations:
	// GetInstance(...)
	// ListInstances(...)
	// DeleteInstance(...)
}
