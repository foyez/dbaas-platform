package controller

import (
	"context"
	"fmt"
	"strconv"

	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ingressNamespace     = "ingress-nginx"
	tcpServicesConfigMap = "tcp-services"
	ingressServiceName   = "ingress-nginx-controller"
)

// exposeTCP makes port reachable externally by (1) mapping it to this
// instance's CNPG -rw service in ingress-nginx's tcp-services ConfigMap and
// (2) opening that port on the shared ingress-nginx-controller LoadBalancer
// Service. Both must exist for the port to actually route anywhere - the
// ConfigMap entry alone does nothing without the Service also listening on
// that port. Idempotent.
func (r *PostgreSQLReconciler) exposeTCP(
	ctx context.Context,
	pg *databasev1.PostgreSQL,
	port int32,
) error {
	backend := fmt.Sprintf("%s/%s-rw:5432", pg.Namespace, pg.Name)

	if err := r.patchTCPServicesConfigMap(ctx, port, backend); err != nil {
		return fmt.Errorf("patching tcp-services configmap: %w", err)
	}

	if err := r.patchIngressServicePort(ctx, port); err != nil {
		return fmt.Errorf(
			"patching %s service: %w", ingressServiceName, err,
		)
	}

	return nil
}

// unexposeTCP reverses exposeTCP. Safe to call even if nothing was ever
// exposed (e.g. the CR failed before reconcileExternalExposure ran).
func (r *PostgreSQLReconciler) unexposeTCP(
	ctx context.Context,
	port int32,
) error {
	if err := r.removeTCPServicesEntry(ctx, port); err != nil {
		return fmt.Errorf("removing tcp-services entry: %w", err)
	}

	if err := r.removeIngressServicePort(ctx, port); err != nil {
		return fmt.Errorf(
			"removing port from %s service: %w", ingressServiceName, err,
		)
	}

	return nil
}

func (r *PostgreSQLReconciler) patchTCPServicesConfigMap(
	ctx context.Context,
	port int32,
	backend string,
) error {
	var cm corev1.ConfigMap

	key := client.ObjectKey{
		Namespace: ingressNamespace,
		Name:      tcpServicesConfigMap,
	}

	if err := r.Get(ctx, key, &cm); err != nil {
		return err
	}

	portKey := strconv.Itoa(int(port))

	if cm.Data != nil && cm.Data[portKey] == backend {
		return nil // already correct
	}

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}

	cm.Data[portKey] = backend

	return r.Update(ctx, &cm)
}

func (r *PostgreSQLReconciler) removeTCPServicesEntry(
	ctx context.Context,
	port int32,
) error {
	var cm corev1.ConfigMap

	key := client.ObjectKey{
		Namespace: ingressNamespace,
		Name:      tcpServicesConfigMap,
	}

	if err := r.Get(ctx, key, &cm); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	portKey := strconv.Itoa(int(port))

	if _, ok := cm.Data[portKey]; !ok {
		return nil
	}

	delete(cm.Data, portKey)

	return r.Update(ctx, &cm)
}

func (r *PostgreSQLReconciler) patchIngressServicePort(
	ctx context.Context,
	port int32,
) error {
	var svc corev1.Service

	key := client.ObjectKey{
		Namespace: ingressNamespace,
		Name:      ingressServiceName,
	}

	if err := r.Get(ctx, key, &svc); err != nil {
		return err
	}

	for _, p := range svc.Spec.Ports {
		if p.Port == port {
			return nil // already open
		}
	}

	svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
		Name:       fmt.Sprintf("pg-%d", port),
		Port:       port,
		TargetPort: intstr.FromInt32(port),
		Protocol:   corev1.ProtocolTCP,
	})

	return r.Update(ctx, &svc)
}

func (r *PostgreSQLReconciler) removeIngressServicePort(
	ctx context.Context,
	port int32,
) error {
	var svc corev1.Service

	key := client.ObjectKey{
		Namespace: ingressNamespace,
		Name:      ingressServiceName,
	}

	if err := r.Get(ctx, key, &svc); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	kept := svc.Spec.Ports[:0]
	changed := false

	for _, p := range svc.Spec.Ports {
		if p.Port == port {
			changed = true
			continue
		}

		kept = append(kept, p)
	}

	if !changed {
		return nil
	}

	svc.Spec.Ports = kept

	return r.Update(ctx, &svc)
}
