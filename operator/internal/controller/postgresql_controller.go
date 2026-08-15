package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=postgresqls/finalizers,verbs=update
// +kubebuilder:rbac:groups=postgresql.cnpg.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var pg databasev1.PostgreSQL
	if err := r.Get(ctx, req.NamespacedName, &pg); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("reconciling PostgreSQL",
		"name", pg.Name,
		"version", pg.Spec.Version,
		"instances", pg.Spec.Instances,
	)

	if err := r.reconcileCNPGCluster(ctx, &pg); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling CNPG cluster: %w", err)
	}

	// if pg.Spec.ConnectionPooler.Enabled {
	// 	if err := r.reconcilePgBouncer(ctx, &pg); err != nil {
	// 		return ctrl.Result{}, fmt.Errorf("reconciling pgbouncer: %w", err)
	// 	}
	// }

	if err := r.updateStatus(ctx, &pg); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PostgreSQLReconciler) updateStatus(ctx context.Context, pg *databasev1.PostgreSQL) error {
	var cluster cnpgv1.Cluster

	if err := r.Get(ctx, types.NamespacedName{Name: pg.Name, Namespace: pg.Namespace}, &cluster); err != nil {
		return client.IgnoreNotFound(err)
	}

	pg.Status.ReadyInstances = int32(cluster.Status.ReadyInstances)

	if cluster.Status.ReadyInstances == int(pg.Spec.Instances) {
		pg.Status.Phase = databasev1.PostgreSQLPhaseReady
	} else {
		pg.Status.Phase = databasev1.PostgreSQLPhaseCreating
	}

	return r.Status().Update(ctx, pg)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.PostgreSQL{}).
		Named("postgresql").
		Owns(&cnpgv1.Cluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
