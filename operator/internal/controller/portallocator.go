package controller

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// portRangeMin/Max bound the external ports handed out for direct
	// Postgres connections through the shared ingress-nginx LoadBalancer.
	portRangeMin = 30100
	portRangeMax = 30999

	allocatorNamespace = "dbaas-operator"
	allocatorConfigMap = "dbaas-port-allocations"
)

// allocatePort claims a free port for owner (namespace/name of the
// PostgreSQL CR), recording the claim in a shared ConfigMap so concurrent
// reconciles across different instances don't race on the same port.
// Idempotent: if owner already holds a port, that same port is returned.
func (r *PostgreSQLReconciler) allocatePort(
	ctx context.Context,
	owner string,
) (int32, error) {
	for {
		cm, err := r.getOrCreateAllocatorConfigMap(ctx)
		if err != nil {
			return 0, err
		}

		for portStr, o := range cm.Data {
			if o == owner {
				port, err := strconv.Atoi(portStr)
				if err != nil {
					return 0, fmt.Errorf(
						"corrupt port allocation entry %q: %w",
						portStr, err,
					)
				}

				return int32(port), nil
			}
		}

		port, err := firstFreePort(cm.Data)
		if err != nil {
			return 0, err
		}

		if cm.Data == nil {
			cm.Data = map[string]string{}
		}

		cm.Data[strconv.Itoa(port)] = owner

		if err := r.Update(ctx, cm); err != nil {
			if apierrors.IsConflict(err) {
				// Another reconcile updated the map concurrently - retry
				// against the latest version.
				continue
			}

			return 0, err
		}

		return int32(port), nil
	}
}

// releasePort frees whatever port owner currently holds, if any.
func (r *PostgreSQLReconciler) releasePort(
	ctx context.Context,
	owner string,
) error {
	for {
		cm, err := r.getOrCreateAllocatorConfigMap(ctx)
		if err != nil {
			return err
		}

		changed := false

		for port, o := range cm.Data {
			if o == owner {
				delete(cm.Data, port)
				changed = true
			}
		}

		if !changed {
			return nil
		}

		if err := r.Update(ctx, cm); err != nil {
			if apierrors.IsConflict(err) {
				continue
			}

			return err
		}

		return nil
	}
}

func firstFreePort(taken map[string]string) (int, error) {
	for p := portRangeMin; p <= portRangeMax; p++ {
		if _, ok := taken[strconv.Itoa(p)]; !ok {
			return p, nil
		}
	}

	return 0, fmt.Errorf(
		"no free ports left in range %d-%d",
		portRangeMin, portRangeMax,
	)
}

func (r *PostgreSQLReconciler) getOrCreateAllocatorConfigMap(
	ctx context.Context,
) (*corev1.ConfigMap, error) {
	var cm corev1.ConfigMap

	key := client.ObjectKey{
		Namespace: allocatorNamespace,
		Name:      allocatorConfigMap,
	}

	if err := r.Get(ctx, key, &cm); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}

		cm = corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      allocatorConfigMap,
				Namespace: allocatorNamespace,
			},
			Data: map[string]string{},
		}

		if err := r.Create(ctx, &cm); err != nil &&
			!apierrors.IsAlreadyExists(err) {
			return nil, err
		}

		// Re-get so we have a clean resourceVersion for the caller to
		// build its update on top of.
		if err := r.Get(ctx, key, &cm); err != nil {
			return nil, err
		}
	}

	return &cm, nil
}
