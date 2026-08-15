package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
)

func NewClient(cfg *rest.Config) (ctrlclient.Client, error) {
	// cfg, err := ctrl.GetConfig()
	// if err != nil {
	// 	return nil, err
	// }

	scheme, err := newScheme()
	if err != nil {
		return nil, err
	}

	return ctrlclient.New(cfg, ctrlclient.Options{
		Scheme: scheme,
	})
}

func newScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	if err := databasev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return scheme, nil
}
