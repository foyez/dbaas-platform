package cnpg

import (
	"context"
	"testing"

	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestUpdateInstance_UsesResourceNameWhenProvided(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, databasev1.AddToScheme(scheme))

	pg := &databasev1.PostgreSQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance-abc",
			Namespace: namespace,
			UID:       types.UID("uid-123"),
		},
		Spec: databasev1.PostgreSQLSpec{
			Version: "16",
			Storage: "5Gi",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pg).Build()
	c := &client{k8sClient: k8sClient}

	newStorage := "20Gi"
	updated, err := c.UpdateInstance(context.Background(), domain.UpdateInstanceInput{
		ID:      pg.Name,
		Storage: &newStorage,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, "20Gi", updated.Storage)

	updatedPG := &databasev1.PostgreSQL{}
	err = k8sClient.Get(context.Background(), ctrlclient.ObjectKey{Name: pg.Name, Namespace: namespace}, updatedPG)
	require.NoError(t, err)
	require.Equal(t, "20Gi", updatedPG.Spec.Storage)
}
