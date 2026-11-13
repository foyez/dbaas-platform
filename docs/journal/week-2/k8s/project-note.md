# PostgreSQL PaaS Operator — Build Guide (Part 3 → Finish)

Continuing from your finished infrastructure (SKE cluster via Terraform). Uses **Podman** instead of Docker.

Every step below follows the same shape: **why → code → file it goes in → run it → verify it worked**, before moving to the next step. Nothing is deferred to a big test at the end.

## Table of Contents

1. [Step 1 — Install CloudNativePG](#step-1--install-cloudnativepg)
2. [Step 2 — Scaffold the Operator](#step-2--scaffold-the-operator)
3. [Step 3 — Define the CRD](#step-3--define-the-crd)
4. [Step 4 — Minimal Reconciler: Read and Log](#step-4--minimal-reconciler-read-and-log)
5. [Step 5 — Create the CNPG Cluster](#step-5--create-the-cnpg-cluster)
6. [Step 6 — Add PgBouncer](#step-6--add-pgbouncer)
7. [Step 7 — Owner References & Cascade Delete](#step-7--owner-references--cascade-delete)
8. [Step 8 — Status Reporting](#step-8--status-reporting)
9. [Step 9 — Switch Polling to Watches](#step-9--switch-polling-to-watches)
10. [Step 10 — Build & Deploy with Podman](#step-10--build--deploy-with-podman)
11. [Step 11 — GitOps with ArgoCD (Bonus)](#step-11--gitops-with-argocd-bonus)

---

## Step 1 — Install CloudNativePG

You're not writing Postgres replication logic — CNPG already does that. Install it as a plain operator.

```bash
kubectl apply --server-side -f \
  https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/main/releases/cnpg-1.24.0.yaml
```

`--server-side` avoids the annotation-size limit that `kubectl apply` hits on large third-party CRDs.

**✅ Verify before moving on:**

```bash
kubectl get pods -n cnpg-system
# cnpg-controller-manager   1/1   Running
```

Nothing else in this guide will work if this isn't green — everything downstream depends on the CNPG CRD (`Cluster`) already existing in the cluster.

---

## Step 2 — Scaffold the Operator

```bash
mkdir postgres-platform && cd postgres-platform
kubebuilder init --domain example.com --repo github.com/you/postgres-platform
kubebuilder create api --group database --version v1 --kind PostgreSQL
# Create Resource? y
# Create Controller? y
```

This generates the files you'll be editing for the rest of this guide:

```
api/v1/postgresql_types.go                       ← Step 3
internal/controller/postgresql_controller.go      ← Steps 4, 8, 9
internal/controller/cnpg.go                       ← Step 5 (new file, you create it)
internal/controller/pgbouncer.go                  ← Step 6 (new file, you create it)
cmd/main.go                                       ← Step 2 (edit now), Step 5 (edit again)
```

Because your controller will build CNPG's `Cluster` object in Go, pull in CNPG's API types now:

```bash
go get github.com/cloudnative-pg/cloudnative-pg/api/v1@v1.24.0
```

**File: `cmd/main.go`** — register CNPG's types alongside your own, in the existing `init()` function:

```go
import (
	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme)) // Pod, Service, StatefulSet, Secret...
	utilruntime.Must(databasev1.AddToScheme(scheme))     // your PostgreSQL CRD (already generated here)
	utilruntime.Must(cnpgv1.AddToScheme(scheme))          // add this line — CNPG's Cluster CRD
}
```

Skip this now and `r.Create()` on a CNPG object in Step 5 will fail with a scheme error — worth getting right before you need it.

**✅ Verify before moving on:**

```bash
go build ./...
```

No cluster interaction yet — this step is purely "does the project compile with the new dependency wired in." If it builds clean, move on.

---

## Step 3 — Define the CRD

**File: `api/v1/postgresql_types.go`** — replace the generated placeholder `Spec`/`Status` with:

<details>
<summary>Full file</summary>

```go
package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type PostgreSQLSpec struct {
	// +kubebuilder:validation:Required
	Version string `json:"version"`

	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Required
	Storage string `json:"storage"`

	// +kubebuilder:validation:Required
	Database string `json:"database"`

	// +kubebuilder:validation:Required
	Username string `json:"username"`

	ConnectionPooler ConnectionPoolerSpec `json:"connectionPooler,omitempty"`
}

type ConnectionPoolerSpec struct {
	Enabled  bool  `json:"enabled,omitempty"`
	PoolSize int32 `json:"poolSize,omitempty"`
}

type PostgreSQLStatus struct {
	Phase              string `json:"phase,omitempty"`
	ReadyInstances     int32  `json:"readyInstances,omitempty"`
	ConnectionEndpoint string `json:"connectionEndpoint,omitempty"`
	PgBouncerStatus    string `json:"pgBouncerStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=".status.connectionEndpoint"
type PostgreSQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgreSQLSpec   `json:"spec,omitempty"`
	Status PostgreSQLStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type PostgreSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgreSQL{}, &PostgreSQLList{})
}
```
</details>

```bash
make manifests   # regenerates the CRD YAML from your +kubebuilder markers
make install     # applies the CRD to your cluster
```

**✅ Verify before moving on:**

```bash
kubectl get crd postgresqls.database.example.com
kubectl explain postgresql.spec
```

You should see your fields (`version`, `replicas`, `storage`, `database`, `username`, `connectionPooler`) listed. At this point you can create a bare object and Kubernetes will accept it — there's just no controller running yet to do anything with it:

```bash
kubectl apply -f - <<'EOF'
apiVersion: database.example.com/v1
kind: PostgreSQL
metadata:
  name: customer-db
spec:
  version: "16"
  replicas: 1
  storage: 5Gi
  database: customer
  username: customer-admin
EOF

kubectl get postgresql customer-db
```

Leave this object in place — you'll reuse it in every following step instead of recreating it.

---

## Step 4 — Minimal Reconciler: Read and Log

Before writing any create/update logic, confirm the wiring works end-to-end: object created → your controller notices → your controller logs it. This is the fastest way to catch scheme/RBAC/watch problems before they're buried under real logic.

**File: `internal/controller/postgresql_controller.go`** — replace the generated `Reconcile` body:

```go
func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var pg databasev1.PostgreSQL
	if err := r.Get(ctx, req.NamespacedName, &pg); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("reconciling PostgreSQL",
		"name", pg.Name,
		"version", pg.Spec.Version,
		"replicas", pg.Spec.Replicas,
	)

	return ctrl.Result{}, nil
}
```

`client.IgnoreNotFound(err)` is a controller-runtime helper — returns `nil` if the object was deleted, passes through any other error unchanged.

**✅ Run it and verify:**

```bash
make run
```

This runs the operator locally against your current `kubeconfig` — no image build needed yet, fastest possible feedback loop. In another terminal:

```bash
kubectl delete postgresql customer-db --ignore-not-found
kubectl apply -f - <<'EOF'
apiVersion: database.example.com/v1
kind: PostgreSQL
metadata:
  name: customer-db
spec:
  version: "16"
  replicas: 1
  storage: 5Gi
  database: customer
  username: customer-admin
EOF
```

Back in the `make run` terminal you should see:

```
INFO    reconciling PostgreSQL    {"name": "customer-db", "version": "16", "replicas": 1}
```

If you see this, your scheme registration, RBAC markers, and watch are all correct. Everything from here is additive logic inside this same loop. Leave `make run` running in its terminal — you'll keep re-triggering it by editing `customer-db` in the steps below (Kubebuilder's `make run` doesn't hot-reload code, so you'll `Ctrl+C` and re-run it each time you change Go code, but not when you just change the CR).

---

## Step 5 — Create the CNPG Cluster

**Why a separate file:** `postgresql_controller.go` should stay an orchestrator. Put CNPG-specific object construction in its own file so it's easy to find and doesn't turn `Reconcile` into a wall of code.

**File: `internal/controller/cnpg.go`** (new file):

```go
package controller

import (
	"context"
	"fmt"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	databasev1 "github.com/you/postgres-platform/api/v1"
)

func (r *PostgreSQLReconciler) reconcileCNPGCluster(ctx context.Context, pg *databasev1.PostgreSQL) error {
	cluster := &cnpgv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pg.Name,
			Namespace: pg.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, cluster, func() error {
		cluster.Spec.Instances = int(pg.Spec.Replicas)
		cluster.Spec.ImageName = fmt.Sprintf("ghcr.io/cloudnative-pg/postgresql:%s", pg.Spec.Version)
		cluster.Spec.StorageConfiguration = cnpgv1.StorageConfiguration{
			Size: pg.Spec.Storage,
		}
		cluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
			InitDB: &cnpgv1.BootstrapInitDB{
				Database: pg.Spec.Database,
				Owner:    pg.Spec.Username,
			},
		}
		return controllerutil.SetControllerReference(pg, cluster, r.Scheme)
	})
	return err
}
```

`CreateOrUpdate` fetches the object first, mutates it inside your callback, then diffs and patches only if something actually changed — this is what keeps the operator idempotent instead of writing on every reconcile regardless of drift.

`SetControllerReference` here is doing double duty: it also makes CNPG's `Cluster` a child of your `PostgreSQL` CR — you'll see the payoff of that in Step 7, but it costs nothing to add it now.

**File: `internal/controller/postgresql_controller.go`** — call it from `Reconcile`:

```go
	log.Info("reconciling PostgreSQL", "name", pg.Name, "version", pg.Spec.Version)

	if err := r.reconcileCNPGCluster(ctx, &pg); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling CNPG cluster: %w", err)
	}

	return ctrl.Result{}, nil
```

**✅ Run it and verify:**

```bash
# Ctrl+C the previous `make run`, then:
make run
```

```bash
kubectl delete postgresql customer-db --ignore-not-found
kubectl apply -f - <<'EOF'
apiVersion: database.example.com/v1
kind: PostgreSQL
metadata:
  name: customer-db
spec:
  version: "16"
  replicas: 1
  storage: 5Gi
  database: customer
  username: customer-admin
EOF

kubectl get cluster customer-db
kubectl get pods -l cnpg.io/cluster=customer-db -w
```

You should see CNPG's own `Cluster` object appear, and shortly after, a real `customer-db-1` Postgres pod come up and go `Running`. This confirms your operator successfully delegates to CNPG — the actual database is now alive before you've written a single line of PgBouncer code.

---

## Step 6 — Add PgBouncer

This is the part where you own a Kubernetes object directly instead of delegating — the reason for using a StatefulSet (stable pod identity + its own small volume) rather than a Deployment.

**File: `internal/controller/pgbouncer.go`** (new file):

```go
package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	databasev1 "github.com/you/postgres-platform/api/v1"
)

func (r *PostgreSQLReconciler) reconcilePgBouncer(ctx context.Context, pg *databasev1.PostgreSQL) error {
	name := pg.Name + "-pgbouncer"

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: pg.Namespace},
	}
	replicas := int32(1)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.Spec.Replicas = &replicas
		sts.Spec.ServiceName = name
		sts.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}}
		sts.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "pgbouncer",
					Image: "edoburu/pgbouncer:latest",
					Env: []corev1.EnvVar{
						{Name: "DB_HOST", Value: pg.Name + "-rw"}, // CNPG's read-write Service, always the current primary
						{Name: "DB_NAME", Value: pg.Spec.Database},
						{Name: "POOL_SIZE", Value: fmt.Sprintf("%d", pg.Spec.ConnectionPooler.PoolSize)},
					},
					Ports: []corev1.ContainerPort{{ContainerPort: 6432}},
				}},
			},
		}
		sts.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{{
			ObjectMeta: metav1.ObjectMeta{Name: "pgbouncer-data"},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")},
				},
			},
		}}
		return controllerutil.SetControllerReference(pg, sts, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("reconciling pgbouncer statefulset: %w", err)
	}

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: pg.Namespace}}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec.Selector = map[string]string{"app": name}
		svc.Spec.Ports = []corev1.ServicePort{{Port: 6432, TargetPort: intstr.FromInt(6432)}}
		return controllerutil.SetControllerReference(pg, svc, r.Scheme)
	})
	return err
}
```

**File: `internal/controller/postgresql_controller.go`** — call it, guarded by the CR's own toggle:

```go
	if err := r.reconcileCNPGCluster(ctx, &pg); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling CNPG cluster: %w", err)
	}

	if pg.Spec.ConnectionPooler.Enabled {
		if err := r.reconcilePgBouncer(ctx, &pg); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling pgbouncer: %w", err)
		}
	}

	return ctrl.Result{}, nil
```

**✅ Run it and verify:**

```bash
# Ctrl+C, then:
make run
```

Update the existing CR to turn pooling on (this alone re-triggers a reconcile — no need to delete/recreate):

```bash
kubectl patch postgresql customer-db --type=merge -p '{"spec":{"connectionPooler":{"enabled":true,"poolSize":20}}}'

kubectl get statefulset customer-db-pgbouncer
kubectl get svc customer-db-pgbouncer
kubectl get pods -l app=customer-db-pgbouncer -w
```

You should see `customer-db-pgbouncer-0` come up. If `enabled` is left `false`, confirm the opposite — neither object gets created — which tells you the conditional in `Reconcile` is working.

---

## Step 7 — Owner References & Cascade Delete

You already set `SetControllerReference` in both Step 5 and Step 6 — this step is just proving it works, no new code.

```
PostgreSQL CR ──owns──► CNPG Cluster ──owns──► Postgres pods, PVCs
              ──owns──► PgBouncer StatefulSet ──owns──► pgbouncer pod, PVC
              ──owns──► PgBouncer Service
```

**✅ Verify:**

```bash
kubectl delete postgresql customer-db
kubectl get cluster customer-db          # should be gone
kubectl get statefulset customer-db-pgbouncer   # should be gone
kubectl get svc customer-db-pgbouncer            # should be gone
```

No finalizer is needed here — finalizers exist for cleaning up things *outside* Kubernetes that the garbage collector can't reach (an AWS instance, an external DNS record). Everything this operator owns lives inside the same cluster, so owner references alone are sufficient.

Recreate `customer-db` (with pooling enabled) before continuing — the next steps build on it existing:

```bash
kubectl apply -f - <<'EOF'
apiVersion: database.example.com/v1
kind: PostgreSQL
metadata:
  name: customer-db
spec:
  version: "16"
  replicas: 1
  storage: 5Gi
  database: customer
  username: customer-admin
  connectionPooler:
    enabled: true
    poolSize: 20
EOF
```

---

## Step 8 — Status Reporting

**File: `internal/controller/postgresql_controller.go`** — add a new method in the same file (it needs `r.Client` and `r.Status()`, so it belongs with the reconciler, not in `cnpg.go`):

```go
func (r *PostgreSQLReconciler) updateStatus(ctx context.Context, pg *databasev1.PostgreSQL) error {
	var cluster cnpgv1.Cluster
	if err := r.Get(ctx, types.NamespacedName{Name: pg.Name, Namespace: pg.Namespace}, &cluster); err != nil {
		return client.IgnoreNotFound(err)
	}

	pg.Status.ReadyInstances = cluster.Status.ReadyInstances
	if cluster.Status.ReadyInstances == pg.Spec.Replicas {
		pg.Status.Phase = "Ready"
	} else {
		pg.Status.Phase = "Provisioning"
	}

	if pg.Spec.ConnectionPooler.Enabled {
		pg.Status.ConnectionEndpoint = fmt.Sprintf("%s-pgbouncer.%s.svc:6432", pg.Name, pg.Namespace)
	} else {
		pg.Status.ConnectionEndpoint = fmt.Sprintf("%s-rw.%s.svc:5432", pg.Name, pg.Namespace)
	}

	return r.Status().Update(ctx, pg)
}
```

`r.Status().Update(...)` writes only the `/status` subresource — matching the `+kubebuilder:subresource:status` marker from Step 3 — so status writes never race with a user editing `.spec`.

Call it at the end of `Reconcile`, in the same file, and requeue periodically so status keeps reflecting CNPG's own progress even without a new spec change:

```go
	if pg.Spec.ConnectionPooler.Enabled {
		if err := r.reconcilePgBouncer(ctx, &pg); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling pgbouncer: %w", err)
		}
	}

	if err := r.updateStatus(ctx, &pg); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
```

**✅ Run it and verify:**

```bash
# Ctrl+C, then:
make run
```

```bash
kubectl get postgresql customer-db -w
```

Within ~15–30s you should see the `Phase` column (from the `+kubebuilder:printcolumn` marker) flip from `Provisioning` to `Ready`, and `Endpoint` populate:

```
NAME          PHASE   ENDPOINT
customer-db   Ready   customer-db-pgbouncer.default.svc:6432
```

---

## Step 9 — Switch Polling to Watches

`RequeueAfter: 15s` works but wastes cycles polling even when nothing changed. Once status reporting is confirmed working (Step 8), replace the poll with real event-driven watches.

**File: `internal/controller/postgresql_controller.go`** — replace the `SetupWithManager` function kubebuilder generated:

```go
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.PostgreSQL{}).
		Owns(&cnpgv1.Cluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
```

`Owns(...)` tells controller-runtime to also watch these kinds and, on any change, look up the object's owner reference and re-run `Reconcile` for the parent `PostgreSQL` — this only works because Steps 5 and 6 already set owner references. You can now drop the `RequeueAfter` line from Step 8 (or lower it to something much longer, e.g. 5 minutes, purely as a safety net):

```go
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
```

**✅ Run it and verify:**

```bash
# Ctrl+C, then:
make run
```

Force a CNPG-owned change directly (not through your CR) and confirm your operator reacts without you touching `customer-db` at all:

```bash
kubectl delete pod -l cnpg.io/cluster=customer-db,cnpg.io/instanceRole=primary
```

Watch the `make run` terminal — a new reconcile should fire almost immediately (from the `Owns(&cnpgv1.Cluster{})` watch reacting to CNPG updating cluster status after failover), well before the old 15s poll would have caught it.

---

## Step 10 — Build & Deploy with Podman

Everything above ran via `make run` against your local kubeconfig. Now package it to actually live in the cluster.

```bash
export CONTAINER_TOOL=podman
```

Kubebuilder's Makefile respects `CONTAINER_TOOL` for `docker-build`/`docker-push` on recent scaffolds. If your generated Makefile predates that, override the two targets directly in the `Makefile`:

```makefile
.PHONY: docker-build
docker-build:
	podman build -t ${IMG} .

.PHONY: docker-push
docker-push:
	podman push ${IMG}
```

```bash
export IMG=registry.example.com/you/postgres-platform-operator:v0.1.0
podman login registry.example.com   # if your registry needs auth
make docker-build
make docker-push
make deploy IMG=${IMG}
```

**✅ Verify before moving on:**

```bash
kubectl get pods -n postgres-platform-system
kubectl logs -n postgres-platform-system deploy/postgres-platform-controller-manager -f
```

You should see the same reconcile log lines you watched locally in Step 4 onward — confirming the in-cluster deployment behaves identically to `make run`. Re-run the Step 7 delete/recreate test against this in-cluster instance as a final sanity check.

---

## Step 11 — GitOps with ArgoCD (Bonus)

Once the operator itself is stable, stop `kubectl apply`-ing by hand and let ArgoCD reconcile from Git — the same idea one layer up: ArgoCD is a controller whose desired state is your repo.

```
gitops/
├── infrastructure/terraform/     # SKE cluster
├── operators/
│   ├── cnpg/                     # CNPG install manifest (Step 1)
│   └── postgres-platform/        # your operator's kustomize deploy (Step 10)
├── applications/
│   └── customer-db.yaml          # PostgreSQL CRs — your actual "databases"
└── argocd/
    └── application.yaml
```

<details>
<summary><code>argocd/application.yaml</code></summary>

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: postgres-platform
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/you/gitops.git
    targetRevision: main
    path: operators/postgres-platform
  destination:
    server: https://kubernetes.default.svc
    namespace: postgres-platform-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```
</details>

`selfHeal: true` is what makes this GitOps rather than "apply once from a repo": hand-edit a live resource, and ArgoCD's own loop reverts it to match Git — the same "desired state wins" principle your operator applies to `PostgreSQL` CRs, now one layer up applied to the whole platform.

**✅ Verify:**

```bash
kubectl apply -f argocd/application.yaml
kubectl get application postgres-platform -n argocd
```

---

## File Map (Quick Reference)

| File | What lives there |
|---|---|
| `cmd/main.go` | Scheme registration (Step 2) |
| `api/v1/postgresql_types.go` | `PostgreSQLSpec` / `PostgreSQLStatus` (Step 3) |
| `internal/controller/postgresql_controller.go` | `Reconcile`, `updateStatus`, `SetupWithManager` (Steps 4, 8, 9) |
| `internal/controller/cnpg.go` | `reconcileCNPGCluster` (Step 5) |
| `internal/controller/pgbouncer.go` | `reconcilePgBouncer` (Step 6) |