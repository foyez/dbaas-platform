package cnpg

import (
	"context"
	"time"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
)

type client struct {
	// k8sClient ctrlclient.Client
}

func NewClient(
// k8sClient ctrlclient.Client,
) domain.InstanceClient {
	return &client{
		// k8sClient: k8sClient,
	}
}

func (c *client) CreateInstance(
	ctx context.Context,
	input domain.CreateInstanceInput,
) (*domain.Instance, error) {
	// Build your custom CNPG CR.

	// Call Kubernetes API.

	// Wait/read status if required.

	return &domain.Instance{
		ID:        "inst-123",
		Name:      input.Name,
		Version:   input.Version,
		Storage:   input.Storage,
		Status:    domain.InstanceStatusPending,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (c *client) GetInstanceByIdempotencyKey(
	ctx context.Context,
	key string,
) (*domain.Instance, error) {

	// Lookup existing instance.

	return nil, nil
}
