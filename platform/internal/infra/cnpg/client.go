package cnpg

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/ports"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	UserIDLabel     = "app.example.com/user-id"
	Namespace       = "database-system"
	appSecretSuffix = "-app"
)

type client struct {
	k8sClient ctrlclient.Client
}

func NewClient(
	k8sClient ctrlclient.Client,
) ports.InstanceClient {
	return &client{
		k8sClient: k8sClient,
	}
}

func (c *client) CreateInstance(
	ctx context.Context,
	input domain.CreateInstanceInput,
) (*domain.Instance, error) {
	resourceName := crName(input.IdempotencyKey)

	instances := input.Instances
	if instances == 0 {
		instances = 1
	}

	pg := &databasev1.PostgreSQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: Namespace,
			Labels: map[string]string{
				UserIDLabel: input.UserID,
			},
		},
		Spec: databasev1.PostgreSQLSpec{
			Version:   strconv.Itoa(input.Version),
			Storage:   input.Storage,
			Instances: int32(instances),
			Database:  input.Name,
			Username:  input.Username,
		},
	}

	if err := c.k8sClient.Create(ctx, pg); err != nil {
		if apierrors.IsAlreadyExists(err) {
			existing := &databasev1.PostgreSQL{}

			if err := c.k8sClient.Get(
				ctx,
				ctrlclient.ObjectKey{
					Namespace: Namespace,
					Name:      resourceName,
				},
				existing,
			); err != nil {
				return nil, err
			}

			return toDomainInstance(existing), nil
		}

		return nil, err
	}

	return toDomainInstance(pg), nil
}

func (c *client) GetInstanceByIdempotencyKey(
	ctx context.Context,
	key string,
) (*domain.Instance, error) {
	// Lookup existing instance.

	return nil, nil
}

// getOwnedInstance fetches a PostgreSQL CR by name and verifies it belongs
// to userID.
func (c *client) getOwnedInstance(
	ctx context.Context,
	id, userID string,
) (*databasev1.PostgreSQL, error) {
	pg := &databasev1.PostgreSQL{}

	if err := c.k8sClient.Get(ctx, ctrlclient.ObjectKey{
		Namespace: Namespace,
		Name:      id,
	}, pg); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if pg.Labels[UserIDLabel] != userID {
		return nil, domain.ErrNotFound
	}

	return pg, nil
}

func (c *client) GetInstance(ctx context.Context, id, userID string) (*domain.Instance, error) {
	pg, err := c.getOwnedInstance(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return toDomainInstance(pg), nil
}

func (c *client) GetInstanceCredentials(
	ctx context.Context,
	id, userID string,
) (*domain.InstanceCredentials, error) {
	pg, err := c.getOwnedInstance(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if pg.Status.Phase != databasev1.PostgreSQLPhaseReady {
		return nil, domain.ErrInstanceNotReady
	}

	secret := &corev1.Secret{}

	secretKey := ctrlclient.ObjectKey{
		Namespace: pg.Namespace,
		Name:      appSecretName(pg.Name),
	}

	if err := c.k8sClient.Get(ctx, secretKey, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, domain.ErrInstanceNotReady
		}

		return nil, err
	}

	host, err := getSecretValue(secret.Data, "host")
	if err != nil {
		return nil, err
	}

	portRaw, err := getSecretValue(secret.Data, "port")
	if err != nil {
		return nil, err
	}

	port, err := parsePort([]byte(portRaw))
	if err != nil {
		return nil, err
	}

	database, err := getSecretValue(secret.Data, "dbname")
	if err != nil {
		return nil, err
	}

	username, err := getSecretValue(secret.Data, "username")
	if err != nil {
		return nil, err
	}

	password, err := getSecretValue(secret.Data, "password")
	if err != nil {
		return nil, err
	}

	uri, err := getSecretValue(secret.Data, "uri")
	if err != nil {
		return nil, err
	}

	return &domain.InstanceCredentials{
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		Password: password,
		URI:      uri,
	}, nil
}

func (c *client) ListInstances(ctx context.Context, userID string) ([]*domain.Instance, error) {
	pgList := &databasev1.PostgreSQLList{}

	if err := c.k8sClient.List(
		ctx,
		pgList,
		ctrlclient.InNamespace(Namespace),
		ctrlclient.MatchingLabels{
			UserIDLabel: userID,
		},
	); err != nil {
		return nil, err
	}

	instances := make([]*domain.Instance, 0, len(pgList.Items))
	for i := range pgList.Items {
		instances = append(instances, toDomainInstance(&pgList.Items[i]))
	}

	return instances, nil
}

func (c *client) UpdateInstance(
	ctx context.Context,
	input domain.UpdateInstanceInput,
) (*domain.Instance, error) {
	pg, err := c.getOwnedInstance(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, err
	}

	if input.Version != nil {
		pg.Spec.Version = strconv.Itoa(*input.Version)
	}
	if input.Storage != nil {
		pg.Spec.Storage = *input.Storage
	}

	if err := c.k8sClient.Update(ctx, pg); err != nil {
		return nil, err
	}

	return toDomainInstance(pg), nil
}

func (c *client) DeleteInstance(ctx context.Context, id, userID string) error {
	// pgList := &databasev1.PostgreSQLList{}

	// if err := c.k8sClient.List(
	// 	ctx,
	// 	pgList,
	// 	ctrlclient.InNamespace(Namespace),
	// ); err != nil {
	// 	return err
	// }

	// for i := range pgList.Items {
	// 	pg := &pgList.Items[i]

	// 	if string(pg.UID) == id {
	// 		return c.k8sClient.Delete(ctx, pg)
	// 	}
	// }

	// return domain.ErrNotFound

	pg, err := c.getOwnedInstance(ctx, id, userID)
	if err != nil {
		return err
	}

	return c.k8sClient.Delete(ctx, pg)
}

func toDomainInstance(pg *databasev1.PostgreSQL) *domain.Instance {
	version, _ := strconv.Atoi(pg.Spec.Version)

	status := domain.InstanceStatusPending

	switch pg.Status.Phase {
	case databasev1.PostgreSQLPhaseReady:
		status = domain.InstanceStatusRunning

	case databasev1.PostgreSQLPhaseFailed:
		status = domain.InstanceStatusFailed

	case databasev1.PostgreSQLPhaseCreating:
		status = domain.InstanceStatusPending
	}

	return &domain.Instance{
		ID:             pg.Name,
		Name:           pg.Name,
		Version:        version,
		Storage:        pg.Spec.Storage,
		Instances:      int(pg.Spec.Instances),
		ReadyInstances: int(pg.Status.ReadyInstances),
		Status:         status,
		CreatedAt:      pg.CreationTimestamp.Time.UTC(),
	}
}

func crName(idempotencyKey string) string {
	sum := sha256.Sum256([]byte(idempotencyKey))
	return "instance-" + hex.EncodeToString(sum[:8])
}

func parsePort(value []byte) (int, error) {
	port, err := strconv.Atoi(string(value))
	if err != nil {
		return 0, fmt.Errorf("invalid port: %w", err)
	}

	return port, nil
}

func getSecretValue(data map[string][]byte, key string) (string, error) {
	value, ok := data[key]
	if !ok || len(value) == 0 {
		return "", fmt.Errorf("secret missing required key %q", key)
	}

	return string(value), nil
}

func appSecretName(clusterName string) string {
	return clusterName + appSecretSuffix
}
