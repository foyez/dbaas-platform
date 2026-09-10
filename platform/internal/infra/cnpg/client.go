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
	k8sClient    ctrlclient.Client
	publicDBHost string
}

func NewClient(
	k8sClient ctrlclient.Client,
	publicDBHost string,
) ports.InstanceClient {
	return &client{
		k8sClient:    k8sClient,
		publicDBHost: publicDBHost,
	}
}

func (c *client) CreateInstance(
	ctx context.Context,
	input domain.CreateInstanceInput,
) (*domain.Instance, error) {
	resourceName := crName(input.UserID, input.IdempotencyKey)

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

			// Defense in depth: crName is now scoped by user, so this
			// should be unreachable for a different user in practice -
			// but verify explicitly rather than trusting the name alone.
			// Report the same "not found" a genuinely-absent resource
			// would give, rather than a 409/conflict that would confirm
			// something exists here at all.
			if existing.Labels[UserIDLabel] != input.UserID {
				return nil, domain.ErrNotFound
			}

			return toDomainInstance(existing), nil
		}

		return nil, err
	}

	return toDomainInstance(pg), nil
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
	pg, creds, err := c.getInternalCredentials(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// The secret's host/port/uri are only reachable from inside the
	// cluster (they point at CNPG's ClusterIP -rw Service). If the
	// operator has allocated and wired up an external port for this
	// instance, prefer that instead so credentials returned to a client
	// outside the cluster actually work. Falls back to the internal
	// values when no external port is set yet (e.g. local dev, which
	// deliberately doesn't run the ingress-nginx TCP passthrough - see
	// gitops/ingress-nginx-tcp-access) - those are still correct for
	// someone using `kubectl port-forward` directly to the -rw service.
	if pg.Status.ExternalPort != 0 && c.publicDBHost != "" {
		creds.Host = c.publicDBHost
		creds.Port = int(pg.Status.ExternalPort)
		creds.URI = fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s",
			creds.Username, creds.Password, creds.Host, creds.Port,
			creds.Database,
		)
	}

	return creds, nil
}

// GetInstanceInternalCredentials always returns the cluster-internal
// host/port - see the ports.InstanceClient doc comment for why callers
// running inside the cluster (the query executor) should use this instead
// of GetInstanceCredentials.
func (c *client) GetInstanceInternalCredentials(
	ctx context.Context,
	id, userID string,
) (*domain.InstanceCredentials, error) {
	_, creds, err := c.getInternalCredentials(ctx, id, userID)
	return creds, err
}

// getInternalCredentials fetches and validates ownership, then reads the
// CNPG app secret. Always returns the cluster-internal host/port - callers
// needing the externally-reachable ones apply that override themselves
// (see GetInstanceCredentials), since it only makes sense for one of the
// two public methods.
func (c *client) getInternalCredentials(
	ctx context.Context,
	id, userID string,
) (*databasev1.PostgreSQL, *domain.InstanceCredentials, error) {
	pg, err := c.getOwnedInstance(ctx, id, userID)
	if err != nil {
		return nil, nil, err
	}

	if pg.Status.Phase != databasev1.PostgreSQLPhaseReady {
		return nil, nil, domain.ErrInstanceNotReady
	}

	secret := &corev1.Secret{}

	secretKey := ctrlclient.ObjectKey{
		Namespace: pg.Namespace,
		Name:      appSecretName(pg.Name),
	}

	if err := c.k8sClient.Get(ctx, secretKey, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil, domain.ErrInstanceNotReady
		}

		return nil, nil, err
	}

	// CNPG's secret stores a bare short name (e.g. "<name>-rw") for
	// "host", which only resolves via Kubernetes' default DNS search path
	// for pods in the *same* namespace as the cluster (database-system).
	// The API runs in a different namespace (dbaas-api), so it needs the
	// fully-qualified in-cluster name instead - same naming convention
	// tcpexpose.go's backend string already uses, just as an FQDN here
	// since this is a real client-side connection, not a kube-proxy
	// Service reference.
	host := fmt.Sprintf("%s-rw.%s.svc.cluster.local", pg.Name, pg.Namespace)

	portRaw, err := getSecretValue(secret.Data, "port")
	if err != nil {
		return nil, nil, err
	}

	port, err := parsePort([]byte(portRaw))
	if err != nil {
		return nil, nil, err
	}

	database, err := getSecretValue(secret.Data, "dbname")
	if err != nil {
		return nil, nil, err
	}

	username, err := getSecretValue(secret.Data, "username")
	if err != nil {
		return nil, nil, err
	}

	password, err := getSecretValue(secret.Data, "password")
	if err != nil {
		return nil, nil, err
	}

	uri := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		username, password, host, port, database,
	)

	return pg, &domain.InstanceCredentials{
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		Password: password,
		URI:      uri,
	}, nil
}

func (c *client) ListInstances(
	ctx context.Context,
	userID string,
	limit int,
	cursor string,
) ([]*domain.Instance, string, error) {
	pgList := &databasev1.PostgreSQLList{}

	opts := []ctrlclient.ListOption{
		ctrlclient.InNamespace(Namespace),
		ctrlclient.MatchingLabels{
			UserIDLabel: userID,
		},
		ctrlclient.Limit(int64(limit)),
	}

	if cursor != "" {
		opts = append(opts, ctrlclient.Continue(cursor))
	}

	if err := c.k8sClient.List(ctx, pgList, opts...); err != nil {
		return nil, "", err
	}

	instances := make([]*domain.Instance, 0, len(pgList.Items))
	for i := range pgList.Items {
		instances = append(instances, toDomainInstance(&pgList.Items[i]))
	}

	// pgList.Continue is empty once the API server has no more pages to
	// give back - the Kubernetes List API only supports this kind of
	// forward-only cursor (via Limit/Continue against etcd), not
	// arbitrary offsets, so that's what's exposed all the way up through
	// the REST API too.
	return instances, pgList.Continue, nil
}

func (c *client) UpdateInstance(
	ctx context.Context,
	input domain.UpdateInstanceInput,
) (*domain.Instance, error) {
	pg, err := c.getOwnedInstance(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, err
	}

	changed := false

	if input.Version != nil && pg.Spec.Version != strconv.Itoa(*input.Version) {
		pg.Spec.Version = strconv.Itoa(*input.Version)
		changed = true
	}
	if input.Storage != nil && pg.Spec.Storage != *input.Storage {
		pg.Spec.Storage = *input.Storage
		changed = true
	}

	if err := c.k8sClient.Update(ctx, pg); err != nil {
		return nil, err
	}

	instance := toDomainInstance(pg)

	// pg.Status here is whatever it was BEFORE this edit - a spec update
	// doesn't touch status, and the operator's reconcile (which would
	// actually move phase off Ready while it applies the change) runs
	// asynchronously off the watch, after this call has already
	// returned. Reporting the stale Ready/Running status would make the
	// UI think the update already fully settled.
	if changed {
		instance.Status = domain.InstanceStatusPending
	}

	return instance, nil
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

	case databasev1.PostgreSQLPhaseDeleting:
		status = domain.InstanceStatusDeleting
	}

	// Belt-and-suspenders: DeletionTimestamp is set by the API server the
	// instant a delete is requested, before the operator's own reconcile
	// even runs. Checking it directly here means the platform reflects
	// "deleting" immediately rather than however many milliseconds it
	// takes the operator to observe and persist the phase itself.
	if !pg.DeletionTimestamp.IsZero() {
		status = domain.InstanceStatusDeleting
	}

	return &domain.Instance{
		ID:             pg.Name,
		Name:           pg.Name,
		Database:       pg.Spec.Database,
		Version:        version,
		Storage:        pg.Spec.Storage,
		Instances:      int(pg.Spec.Instances),
		ReadyInstances: int(pg.Status.ReadyInstances),
		Status:         status,
		CreatedAt:      pg.CreationTimestamp.Time.UTC(),
		ExternalPort:   int(pg.Status.ExternalPort),
	}
}

func crName(userID, idempotencyKey string) string {
	sum := sha256.Sum256([]byte(userID + ":" + idempotencyKey))
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
