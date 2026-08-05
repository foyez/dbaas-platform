package k8s

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
)

func NewClient() (ctrlclient.Client, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()

	if err := databasev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return ctrlclient.New(cfg, ctrlclient.Options{
		Scheme: scheme,
	})
}
