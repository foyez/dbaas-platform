package controller

// import (
// 	"context"
// 	"fmt"
//
// 	databasev1 "github.com/foyez/dbaas-platform/operator/api/v1"
// 	appsv1 "k8s.io/api/apps/v1"
// 	corev1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/api/resource"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/util/intstr"
// 	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
// )
//
// func (r *PostgreSQLReconciler) reconcilePgBouncer(ctx context.Context, pg *databasev1.PostgreSQL) error {
// 	name := pg.Name + "-pgbouncer"
//
// 	sts := &appsv1.StatefulSet{
// 		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: pg.Namespace},
// 	}
// 	replicas := int32(1)
//
// 	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
// 		sts.Spec.Replicas = &replicas
// 		sts.Spec.ServiceName = name
// 		sts.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}}
// 		sts.Spec.Template = corev1.PodTemplateSpec{
// 			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
// 			Spec: corev1.PodSpec{
// 				Containers: []corev1.Container{{
// 					Name:  "pgbouncer",
// 					Image: "edoburu/pgbouncer:v1.24.0-p1",
// 					Env: []corev1.EnvVar{
// 						{
// 							Name: "DATABASES",
// 							Value: fmt.Sprintf(
// 								"%s=host=%s-rw port=5432 pool_size=%d",
// 								pg.Spec.Database,
// 								pg.Name,
// 								pg.Spec.ConnectionPooler.PoolSize,
// 							),
// 						},
// 					},
// 				}},
// 			},
// 		}
//
// 		sts.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{{
// 			ObjectMeta: metav1.ObjectMeta{Name: "pgbouncer-data"},
// 			Spec: corev1.PersistentVolumeClaimSpec{
// 				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
// 				Resources: corev1.VolumeResourceRequirements{
// 					Requests: corev1.ResourceList{
// 						corev1.ResourceStorage: resource.MustParse("1Gi"),
// 					},
// 				},
// 			},
// 		}}
//
// 		return controllerutil.SetControllerReference(pg, sts, r.Scheme)
// 	})
// 	if err != nil {
// 		return fmt.Errorf("reconciling pgbouncer statefulset: %w", err)
// 	}
//
// 	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: pg.Namespace}}
// 	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
// 		svc.Spec.Selector = map[string]string{"app": name}
// 		svc.Spec.Ports = []corev1.ServicePort{{Port: 5432, TargetPort: intstr.FromInt(5432)}}
// 		return controllerutil.SetControllerReference(pg, svc, r.Scheme)
// 	})
//
// 	return err
// }
