package controller

import (
	"context"
	"fmt"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const UserIDLabel = "app.example.com/user-id"

func (r *PostgreSQLReconciler) reconcileCNPGCluster(
	ctx context.Context,
	pg *databasev1.PostgreSQL,
) error {
	cluster := &cnpgv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pg.Name,
			Namespace: pg.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(
		ctx,
		r.Client,
		cluster,
		func() error {
			if cluster.Labels == nil {
				cluster.Labels = map[string]string{}
			}

			if userID, ok := pg.Labels[UserIDLabel]; ok {
				cluster.Labels[UserIDLabel] = userID
			}

			cluster.Spec.Instances = int(pg.Spec.Instances)

			cluster.Spec.ImageName = fmt.Sprintf(
				"ghcr.io/cloudnative-pg/postgresql:%s",
				pg.Spec.Version,
			)

			cluster.Spec.StorageConfiguration = cnpgv1.StorageConfiguration{
				Size: pg.Spec.Storage,
			}

			cluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
				InitDB: &cnpgv1.BootstrapInitDB{
					Database: pg.Spec.Database,
					Owner:    pg.Spec.Username,
				},
			}

			return controllerutil.SetControllerReference(
				pg,
				cluster,
				r.Scheme,
			)
		},
	)

	return err
}
