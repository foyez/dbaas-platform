package ports

import (
	"context"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
)

// Querier executes a single ad-hoc SQL statement against a specific
// instance's database, using connection details already resolved by
// InstanceClient.GetInstanceInternalCredentials. Implementations own
// connection pooling/eviction per instance.
type Querier interface {
	ExecuteQuery(
		ctx context.Context,
		instanceID string,
		creds *domain.InstanceCredentials,
		sql string,
	) (*domain.QueryResult, error)

	// Close releases all pooled connections. Call once on server shutdown.
	Close()
}
