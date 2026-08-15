package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls/finalizers,verbs=update
// +kubebuilder:rbac:groups=postgresql.cnpg.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

const (
	clusterNameSuffix = "-app"
)

type PostgreSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	if err := r.reconcileCNPGCluster(ctx, &pg); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling CNPG cluster: %w", err)
	}

	// Ensure the desired CNPG cluster exists and matches the
	// PostgreSQL custom resource.
	if err := r.reconcileCNPGCluster(ctx, &pg); err != nil {
		return ctrl.Result{}, fmt.Errorf(
			"reconciling CNPG cluster: %w",
			err,
		)
	}

	// Reflect the observed CNPG state into our CR status.
	if err := r.reconcileStatus(ctx, &pg); err != nil {
		return ctrl.Result{}, fmt.Errorf(
			"reconciling PostgreSQL status: %w",
			err,
		)
	}

	return ctrl.Result{}, nil
}

func (r *PostgreSQLReconciler) reconcileStatus(
	ctx context.Context,
	pg *databasev1.PostgreSQL,
) error {
	var cluster cnpgv1.Cluster

	err := r.Get(
		ctx,
		client.ObjectKey{
			Name:      clusterName(pg.Name),
			Namespace: pg.Namespace,
		},
		&cluster,
	)

	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.updateStatus(
				ctx,
				pg,
				databasev1.PostgreSQLPhaseCreating,
				0,
			)
		}

		return err
	}

	readyInstances := int32(cluster.Status.ReadyInstances)

	phase := databasev1.PostgreSQLPhaseCreating

	if readyInstances >= pg.Spec.Instances {
		phase = databasev1.PostgreSQLPhaseReady
	}

	return r.updateStatus(
		ctx,
		pg,
		phase,
		readyInstances,
	)
}

func (r *PostgreSQLReconciler) updateStatus(
	ctx context.Context,
	pg *databasev1.PostgreSQL,
	phase databasev1.PostgreSQLPhase,
	readyInstances int32,
) error {
	if pg.Status.Phase == phase &&
		pg.Status.ReadyInstances == readyInstances {
		return nil
	}

	pg.Status.Phase = phase
	pg.Status.ReadyInstances = readyInstances

	return r.Status().Update(ctx, pg)
}

func clusterName(pgName string) string {
	return pgName + clusterNameSuffix
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
