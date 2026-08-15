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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls/finalizers,verbs=update
// +kubebuilder:rbac:groups=postgresql.cnpg.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

const (
	PostgreSQLConditionReady databasev1.PostgreSQLConditionType = "Ready"
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
		return ctrl.Result{}, fmt.Errorf(
			"reconciling CNPG cluster: %w",
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
