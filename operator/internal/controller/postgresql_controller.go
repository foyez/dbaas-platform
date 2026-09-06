package controller

import (
	"context"
	"fmt"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls/finalizers,verbs=update
// +kubebuilder:rbac:groups=postgresql.cnpg.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//
// Deliberately NOT using +kubebuilder:rbac markers with namespace= for the
// port-allocator ConfigMap (dbaas-operator ns) or the ingress-nginx TCP
// wiring (ingress-nginx ns): controller-gen picks those up and regenerates
// config/rbac/role.yaml with per-namespace Role objects that config/default's
// blanket namespace transform then collides into a single namespace. Those
// two permission sets are hand-maintained instead - see
// config/rbac/port_allocator_role.yaml and
// gitops/ingress-nginx-tcp-access/base/rbac.yaml (kept outside this
// kustomize tree specifically to avoid the same collision).

const (
	PostgreSQLConditionReady databasev1.PostgreSQLConditionType = "Ready"

	// tcpExposeFinalizer ensures the allocated port and the ingress-nginx
	// ConfigMap/Service entries it created are cleaned up before the CR is
	// actually removed - garbage collection alone won't do this, since
	// those objects aren't owned by the PostgreSQL CR.
	tcpExposeFinalizer = "postgresql.database.example.com/tcp-expose"
)

type PostgreSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// RawClient is a direct (uncached) client used only for the port
	// allocator's ConfigMap and the ingress-nginx ConfigMap/Service.
	// The default cached client (embedded above) starts a cluster-wide
	// List+Watch the first time it Gets any object of a type - which
	// requires cluster-wide "list"/"watch" RBAC, since Kubernetes can't
	// scope those verbs by resourceName. That's incompatible with the
	// deliberately narrow, resourceNames-scoped Roles those two touch
	// points use (see config/rbac/port_allocator_role.yaml and
	// gitops/ingress-nginx-tcp-access/base/rbac.yaml), so those calls use
	// this direct client instead - single-object API calls, no caching,
	// no need for list/watch permissions at all.
	RawClient client.Client
}

func (r *PostgreSQLReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var pg databasev1.PostgreSQL

	if err := r.Get(ctx, req.NamespacedName, &pg); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info(
		"reconciling PostgreSQL",
		"name", pg.Name,
		"version", pg.Spec.Version,
		"instances", pg.Spec.Instances,
	)

	if !pg.DeletionTimestamp.IsZero() {
		if pg.Status.Phase != databasev1.PostgreSQLPhaseDeleting {
			if err := r.setStatus(
				ctx,
				&pg,
				databasev1.PostgreSQLPhaseDeleting,
				pg.Status.ReadyInstances,
				"Deleting",
				"Instance is being deleted",
			); err != nil {
				return ctrl.Result{}, err
			}
		}

		if controllerutil.ContainsFinalizer(&pg, tcpExposeFinalizer) {
			if err := r.cleanupExternalExposure(ctx, &pg); err != nil {
				return ctrl.Result{}, fmt.Errorf(
					"cleaning up external exposure: %w", err,
				)
			}

			controllerutil.RemoveFinalizer(&pg, tcpExposeFinalizer)

			if err := r.Update(ctx, &pg); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&pg, tcpExposeFinalizer) {
		controllerutil.AddFinalizer(&pg, tcpExposeFinalizer)

		if err := r.Update(ctx, &pg); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.reconcileCNPGCluster(ctx, &pg); err != nil {
		return ctrl.Result{}, fmt.Errorf(
			"reconciling CNPG cluster: %w",
			err,
		)
	}

	if err := r.reconcileExternalExposure(ctx, &pg); err != nil {
		return ctrl.Result{}, fmt.Errorf(
			"reconciling external exposure: %w",
			err,
		)
	}

	if err := r.updateStatus(ctx, &pg); err != nil {
		return ctrl.Result{}, fmt.Errorf(
			"updating PostgreSQL status: %w",
			err,
		)
	}

	return ctrl.Result{}, nil
}

// reconcileExternalExposure ensures pg has a stable external port and that
// the port is actually wired up through ingress-nginx. Sets
// pg.Status.ExternalPort in-memory; the write to the API server happens
// together with the rest of the status in updateStatus.
//
// External exposure is treated as best-effort: environments that don't run
// ingress-nginx (e.g. local dev, which deliberately skips it - see
// scripts/local-forward.sh) shouldn't have every PostgreSQL CR stuck
// permanently failing because of it. A missing ingress-nginx ConfigMap/
// Service is logged and skipped rather than returned as an error; anything
// else (a real RBAC problem, a conflict that won't resolve, etc.) still
// fails the reconcile so it isn't silently swallowed.
func (r *PostgreSQLReconciler) reconcileExternalExposure(
	ctx context.Context,
	pg *databasev1.PostgreSQL,
) error {
	log := logf.FromContext(ctx)
	owner := ownerKey(pg)

	port, err := r.allocatePort(ctx, owner)
	if err != nil {
		return err
	}

	if err := r.exposeTCP(ctx, pg, port); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info(
				"skipping external TCP exposure: ingress-nginx not found "+
					"in this cluster",
				"port", port,
			)

			return nil
		}

		return err
	}

	pg.Status.ExternalPort = port

	return nil
}

// cleanupExternalExposure releases everything reconcileExternalExposure set
// up, in the reverse order, so a partially-provisioned instance can never
// leak a claimed port.
func (r *PostgreSQLReconciler) cleanupExternalExposure(
	ctx context.Context,
	pg *databasev1.PostgreSQL,
) error {
	if pg.Status.ExternalPort != 0 {
		if err := r.unexposeTCP(ctx, pg.Status.ExternalPort); err != nil {
			return err
		}
	}

	return r.releasePort(ctx, ownerKey(pg))
}

func ownerKey(pg *databasev1.PostgreSQL) string {
	return pg.Namespace + "/" + pg.Name
}

func (r *PostgreSQLReconciler) updateStatus(
	ctx context.Context,
	pg *databasev1.PostgreSQL,
) error {
	var cluster cnpgv1.Cluster

	err := r.Get(
		ctx,
		client.ObjectKey{
			Name:      pg.Name,
			Namespace: pg.Namespace,
		},
		&cluster,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.setStatus(
				ctx,
				pg,
				databasev1.PostgreSQLPhaseCreating,
				0,
				"ClusterNotReady",
				"CNPG cluster has not been created yet",
			)
		}

		return err
	}

	readyInstances := int32(cluster.Status.ReadyInstances)
	desiredInstances := pg.Spec.Instances

	if desiredInstances <= 0 {
		return r.setStatus(
			ctx,
			pg,
			databasev1.PostgreSQLPhaseFailed,
			readyInstances,
			"InvalidInstances",
			"desired instance count must be greater than zero",
		)
	}

	if readyInstances < desiredInstances {
		return r.setStatus(
			ctx,
			pg,
			databasev1.PostgreSQLPhaseCreating,
			readyInstances,
			"ClusterNotReady",
			fmt.Sprintf(
				"waiting for PostgreSQL cluster: %d/%d instances ready",
				readyInstances,
				desiredInstances,
			),
		)
	}

	if !isClusterHealthy(&cluster) {
		return r.setStatus(
			ctx,
			pg,
			databasev1.PostgreSQLPhaseCreating,
			readyInstances,
			"ClusterNotHealthy",
			"PostgreSQL cluster has the expected number of ready instances but is not healthy",
		)
	}

	return r.setStatus(
		ctx,
		pg,
		databasev1.PostgreSQLPhaseReady,
		readyInstances,
		"ClusterReady",
		"PostgreSQL cluster is healthy and all requested instances are ready",
	)
}

func (r *PostgreSQLReconciler) setStatus(
	ctx context.Context,
	pg *databasev1.PostgreSQL,
	phase databasev1.PostgreSQLPhase,
	readyInstances int32,
	reason string,
	message string,
) error {
	pg.Status.Phase = phase
	pg.Status.ReadyInstances = readyInstances

	meta.SetStatusCondition(
		&pg.Status.Conditions,
		metav1.Condition{
			Type:               string(PostgreSQLConditionReady),
			Status:             conditionStatus(phase),
			Reason:             reason,
			Message:            message,
			ObservedGeneration: pg.Generation,
		},
	)

	return r.Status().Update(ctx, pg)
}

func isClusterHealthy(cluster *cnpgv1.Cluster) bool {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == string(cnpgv1.ConditionClusterReady) {
			return condition.Status == metav1.ConditionTrue
		}
	}

	return false
}

func conditionStatus(
	phase databasev1.PostgreSQLPhase,
) metav1.ConditionStatus {
	if phase == databasev1.PostgreSQLPhaseReady {
		return metav1.ConditionTrue
	}

	return metav1.ConditionFalse
}

func (r *PostgreSQLReconciler) SetupWithManager(
	mgr ctrl.Manager,
) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.PostgreSQL{}).
		Named("postgresql").
		Owns(&cnpgv1.Cluster{}).
		Complete(r)
}
