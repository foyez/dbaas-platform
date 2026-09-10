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

	GetInstance(ctx context.Context, id, userID string) (*domain.Instance, error)
	GetInstanceCredentials(
		ctx context.Context,
		id string,
		userID string,
	) (*domain.InstanceCredentials, error)

	// GetInstanceInternalCredentials always returns the CNPG -rw service's
	// cluster-internal host/port, never the externally-reachable one
	// GetInstanceCredentials prefers. For use by callers running inside
	// the cluster (the query executor) - going out through the public
	// endpoint and back in would be a pointless extra hop and an
	// unnecessary dependency on DNS/ingress-nginx being reachable.
	GetInstanceInternalCredentials(
		ctx context.Context,
		id string,
		userID string,
	) (*domain.InstanceCredentials, error)

	ListInstances(
		ctx context.Context,
		userID string,
		limit int,
		cursor string,
	) (instances []*domain.Instance, nextCursor string, err error)
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
