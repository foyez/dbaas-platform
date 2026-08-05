package cnpg

import (
	"context"
	"strconv"

	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/service"
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

const namespace = "database-system"

func (c *client) CreateInstance(
	ctx context.Context,
	input domain.CreateInstanceInput,
) (*domain.Instance, error) {
	// Build your custom CNPG CR.
	pg := &databasev1.PostgreSQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: namespace,
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
	return toDomainInstance(pg), nil
}

func (c *client) GetInstanceByIdempotencyKey(
	ctx context.Context,
	key string,
) (*domain.Instance, error) {
	// Lookup existing instance.

	return nil, nil
}

func (c *client) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	pgList := &databasev1.PostgreSQLList{}

	if err := c.k8sClient.List(
		ctx,
		pgList,
		ctrlclient.InNamespace(namespace),
	); err != nil {
		return nil, err
	}

	instances := make([]*domain.Instance, 0, len(pgList.Items))
	for i := range pgList.Items {
		instances = append(instances, toDomainInstance(&pgList.Items[i]))
	}

	return instances, nil
}

func (c *client) DeleteInstance(ctx context.Context, id string) error {
	pgList := &databasev1.PostgreSQLList{}

	if err := c.k8sClient.List(
		ctx,
		pgList,
		ctrlclient.InNamespace(namespace),
	); err != nil {
		return err
	}

	for i := range pgList.Items {
		pg := &pgList.Items[i]

		if string(pg.UID) == id {
			return c.k8sClient.Delete(ctx, pg)
		}
	}

	return service.ErrNotFound
}

func toDomainInstance(pg *databasev1.PostgreSQL) *domain.Instance {
	version, _ := strconv.Atoi(pg.Spec.Version)

	status := domain.InstanceStatusPending
	if pg.Status.Phase == "Ready" {
		status = domain.InstanceStatusRunning
	}

	return &domain.Instance{
		ID:        string(pg.UID),
		Name:      pg.Name,
		Version:   version,
		Storage:   pg.Spec.Storage,
		Status:    status,
		CreatedAt: pg.CreationTimestamp.Time.UTC(),
	}
}
