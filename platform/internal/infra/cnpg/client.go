package cnpg

import (
	"context"
	"strconv"

	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
	"github.com/foyez/dbaas-platform/platform/internal/domain"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type client struct {
	k8sClient ctrlclient.Client
}

func NewClient(
	k8sClient ctrlclient.Client,
) domain.InstanceClient {
	return &client{
		k8sClient: k8sClient,
	}
}

func (c *client) CreateInstance(
	ctx context.Context,
	input domain.CreateInstanceInput,
) (*domain.Instance, error) {
	// Build your custom CNPG CR.
	pg := &databasev1.PostgreSQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: "database-system",
		},
		Spec: databasev1.PostgreSQLSpec{
			Version:  strconv.Itoa(input.Version),
			Storage:  input.Storage,
			Database: input.Name,
			Username: input.Username,
		},
	}

	if err := c.k8sClient.Create(ctx, pg); err != nil {
		return nil, err
	}

	// Call Kubernetes API.

	// Wait/read status if required.
	version, err := strconv.Atoi(pg.Spec.Version)
	if err != nil {
		return nil, err
	}

	return &domain.Instance{
		ID:        string(pg.UID),
		Name:      pg.Name,
		Version:   version,
		Storage:   pg.Spec.Storage,
		Status:    domain.InstanceStatusPending,
		CreatedAt: pg.CreationTimestamp.Time.UTC(),
	}, nil
}

func (c *client) GetInstanceByIdempotencyKey(
	ctx context.Context,
	key string,
) (*domain.Instance, error) {
	// Lookup existing instance.

	return nil, nil
}
