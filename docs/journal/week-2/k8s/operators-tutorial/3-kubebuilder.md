# Kubebuilder Scaffolding
*(Covers: Kubebuilder vs Operator SDK → `init` → project structure → `create api` → lab → what the generator actually wrote)*

---

## 1. Section Overview

Parts 1 and 2 covered CRDs by hand and what an Operator conceptually is. Writing that controller reconcile loop by hand — client setup, informers, work queues, leader election, RBAC YAML, CRD YAML — is a lot of boilerplate before you write a single line of *your* logic.

**Kubebuilder** is a CLI + scaffolding framework, built on top of `controller-runtime` (the same library underpinning most Go-based Operators), that generates that boilerplate for you. This section covers:

- How Kubebuilder compares to the other major choice, Operator SDK
- Scaffolding a brand-new project with `kubebuilder init`
- Reading the generated project structure
- Adding a new API/CR type with `kubebuilder create api`
- A hands-on lab building a `WebApp` operator scaffold
- A deep read-through of exactly what the generator wrote and why

By the end, you'll have a real, compiling (if not yet functional) Operator project structure, and understand every generated file well enough to know where your reconcile logic actually goes.

---

## 2. Kubebuilder vs Operator SDK

Both are legitimate, widely-used choices for scaffolding Go Operators, and both are actually built on the **same underlying library**: `controller-runtime`. The difference is in scope and opinionation, not in the runtime engine.

| | **Kubebuilder** | **Operator SDK** |
|---|---|---|
| Maintained by | Kubernetes SIG (`sigs.k8s.io/kubebuilder`) | Operator Framework / Red Hat |
| Underlying library | `controller-runtime` | `controller-runtime` (same) |
| Scope | Go-only, minimal, "just scaffolding" | Go, Ansible, or Helm-based operators; adds packaging/distribution tooling (OLM bundles) |
| Packaging/distribution | Not built-in — you handle it yourself | Built-in support for OLM (Operator Lifecycle Manager) bundles, for publishing to OperatorHub |
| Learning curve | Slightly leaner, fewer concepts up front | More concepts (bundles, CSVs, channels) if you use the OLM features |
| CLI command style | `kubebuilder init`, `kubebuilder create api` | `operator-sdk init`, `operator-sdk create api` (deliberately near-identical) |
| Best fit | You're writing a Go controller and want the leanest path | You want non-Go language options, or plan to publish via OperatorHub/OLM |

**Practical takeaway:** for a pure-Go Operator with no plans to publish through OLM, the two are close enough that project structure and generated code are nearly interchangeable — Operator SDK's Go plugin *is* Kubebuilder under the hood. This tutorial uses Kubebuilder directly since it's the leaner, more direct path and everything transfers if you later move to Operator SDK.

---

## 3. Demo: Kubebuilder Init

### Prerequisites

```bash
# Go 1.21+, and a configured cluster (kind/minikube is fine for scaffolding/testing)
go version
kubectl version --client
```

### Install Kubebuilder

```bash
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder
sudo mv kubebuilder /usr/local/bin/
kubebuilder version
```

### Scaffold a new project

```bash
mkdir webapp-operator && cd webapp-operator

kubebuilder init --domain example.com --repo github.com/yourname/webapp-operator
```

Breaking down the flags:
- `--domain example.com` — becomes the **API group suffix** for every CRD you create in this project (e.g. `webapp.example.com`). This matches the `group` field from the raw CRD YAML in Part 1 — Kubebuilder is just generating that for you based on this one flag.
- `--repo` — the Go module path, used for import paths throughout the generated code. Should match wherever this will actually live (GitHub, internal git host, etc.).

Running this generates a working (but API-less) Go module: a `main.go` that can already build and run as a controller-runtime manager, plus supporting config, RBAC, and CI scaffolding. Nothing here is CR-specific yet — this step just sets up the *harness* your future API types and controllers plug into.

### Sanity check it builds

```bash
go build ./...
```

If that succeeds, the scaffold is valid and ready for an API type — the next step.

---

## 4. Demo: Project Structure Walkthrough

After `init`, the directory looks roughly like this:

```
webapp-operator/
├── Makefile                  # build/test/deploy targets — your main interface
├── PROJECT                   # Kubebuilder's own metadata about this project
├── go.mod / go.sum
├── main.go                   # entrypoint: sets up and starts the controller manager
├── cmd/                       # (in some versions) may hold main.go instead
├── config/
│   ├── crd/                  # generated CRD YAMLs land here (empty until create api)
│   ├── rbac/                 # generated Role/ClusterRole/Binding YAMLs
│   ├── manager/              # Deployment manifest for the controller itself
│   ├── samples/              # example CR YAMLs for each API you add
│   ├── default/               # kustomize overlay wiring the above together
│   └── ...
├── hack/
│   └── boilerplate.go.txt    # license header template stamped onto generated files
└── bin/                       # downloaded tool binaries (controller-gen, etc.) after `make`
```

**Key things to internalize about this layout:**

- **`config/` is entirely Kustomize-based**, not Helm. Every subdirectory is a Kustomize base/overlay; `config/default` is the top-level overlay that composes CRDs + RBAC + the manager Deployment into one thing you can `kubectl apply -k config/default`.
- **Nothing is CRD-specific yet.** `config/crd/` exists but is empty — because we haven't told Kubebuilder about any API types. That's the next step.
- **`main.go`** already contains a working `controller-runtime` `Manager` setup — leader election config, metrics/health endpoints, scheme registration — even before you've added a single API. This is the harness mentioned in section 3.
- **The `Makefile`** is your primary interface from here on: `make manifests` (regenerate CRD/RBAC YAML from code markers), `make generate` (regenerate deepcopy code), `make install` (apply CRDs to your current cluster), `make run` (run the controller locally against your current kubeconfig, useful for fast dev loops without building an image).

This structure is the direct, literal implementation of the "Anatomy of an Operator" diagram from Part 2 — `config/crd` is piece 1, `main.go`+ your future `controllers/` package is piece 2/3, and `config/rbac` is piece 4.

---

## 5. Demo: Kubebuilder `create api` for WebApp

Now we tell Kubebuilder about an actual custom resource — a `WebApp` type, similar in spirit to the `Website` example from Part 1.

```bash
kubebuilder create api --group webapp --version v1 --kind WebApp
```

You'll be prompted:

```
Create Resource [y/n]
y
Create Controller [y/n]
y
```

- **"Create Resource"** → generates the Go struct that defines `WebApp`'s `Spec`/`Status` fields (`api/v1/webapp_types.go`) — this is the programmatic equivalent of the `openAPIV3Schema` block you hand-wrote in Part 1. Say `n` only if you're adding a *controller* for a type that already exists elsewhere (e.g. watching a built-in or third-party CR).
- **"Create Controller"** → generates a controller skeleton (`internal/controller/webapp_controller.go`) with an empty `Reconcile()` function wired up and ready. Say `n` only if you're generating the API type alone, with no reconciliation logic (rare).

### What gets added to the project

```
api/v1/
├── webapp_types.go        # NEW — Spec/Status Go structs, +kubebuilder markers
├── groupversion_info.go   # NEW — registers webapp.example.com/v1 with the scheme
└── zz_generated.deepcopy.go  # NEW — auto-generated, do not hand-edit

internal/controller/
├── webapp_controller.go   # NEW — Reconcile() skeleton + SetupWithManager()
└── suite_test.go          # NEW — envtest-based test harness

config/crd/bases/
└── webapp.example.com_webapps.yaml   # NEW — generated CRD YAML (once you run `make manifests`)

config/samples/
└── webapp_v1_webapp.yaml   # NEW — a starter CR YAML you can apply immediately

config/rbac/
└── (role.yaml gains new rules)  # UPDATED — RBAC for the new resource
```

`main.go` also gets updated automatically — a line is added wiring `WebAppReconciler` into the manager's `SetupWithManager` call, so the new controller actually starts when the binary runs.

### Regenerate the CRD YAML and RBAC from your Go markers

```bash
make manifests
```

This runs `controller-gen` under the hood, reading special comment markers (covered in depth in section 8) in `webapp_types.go` and turning them into `config/crd/bases/webapp.example.com_webapps.yaml` — a full CRD, structurally identical to what you wrote by hand in Part 1, just generated from Go struct tags instead.

---

## 6. Lab: Build the WebApp Operator Scaffold

**Try this yourself before reading the solution below.**

**Task:**
1. Initialize a new Kubebuilder project named `webapp-operator`, domain `example.com`, with an appropriate Go module path.
2. Create an API of group `webapp`, version `v1`, kind `WebApp` — with both Resource and Controller generated.
3. Edit `api/v1/webapp_types.go` so `WebAppSpec` has two fields: `Image` (string, required) and `Replicas` (int32, default `1`, minimum `1`).
4. Regenerate manifests, then install the CRD onto whatever test cluster you have available (`kind`/`minikube` is fine).
5. Apply the sample CR from `config/samples/`, edited to set `image` and `replicas`.
6. Confirm with `kubectl get webapps` and `kubectl get crd | grep webapp`.

---

## 7. Lab Solution: Build the WebApp Operator Scaffold

```bash
mkdir webapp-operator && cd webapp-operator
kubebuilder init --domain example.com --repo github.com/yourname/webapp-operator
kubebuilder create api --group webapp --version v1 --kind WebApp
# answer y / y to the Resource / Controller prompts
```

**Edit `api/v1/webapp_types.go`** — find the generated `WebAppSpec` struct (it starts out with a placeholder `Foo` field) and replace it:

```go
// WebAppSpec defines the desired state of WebApp
type WebAppSpec struct {
	// Image is the container image to deploy.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Replicas is the desired number of pod replicas.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
}
```

Those `+kubebuilder:...` comments are **markers** — this is exactly how `required`, `minimum`, and `default` from the hand-written OpenAPI schema in Part 1 get expressed in Go. `controller-gen` parses these comments directly.

**Regenerate and install:**

```bash
make manifests generate
make install          # applies config/crd to your current kubectl context
kubectl get crd | grep webapp
# webapps.webapp.example.com
```

**Edit and apply the sample:**

`config/samples/webapp_v1_webapp.yaml`:

```yaml
apiVersion: webapp.example.com/v1
kind: WebApp
metadata:
  name: webapp-sample
spec:
  image: nginx:1.25
  replicas: 3
```

```bash
kubectl apply -f config/samples/webapp_v1_webapp.yaml
kubectl get webapps
```

```
NAME             AGE
webapp-sample    4s
```

At this point nothing actually deploys anything yet — `Reconcile()` is still an empty stub (see section 8) — but the full data path (CRD → validated CR → visible via `kubectl`) is working exactly as designed.

---

## 8. Demo: What the Generator Wrote for You

A close read of the four files that matter most, and *why* each line exists.

### 8.1 — `api/v1/webapp_types.go`

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas"

// WebApp is the Schema for the webapps API
type WebApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebAppSpec   `json:"spec,omitempty"`
	Status WebAppStatus `json:"status,omitempty"`
}
```

Map this straight back to Part 1 and Part 2 concepts:
- `+kubebuilder:subresource:status` → generates the exact `subresources: status: {}` block you hand-wrote in the Part 1 CRD demo. Delete this marker and status updates go back to being un-split from spec.
- `+kubebuilder:printcolumn:...` → generates `additionalPrinterColumns` — the same feature from Part 1 section 7.3, just expressed as a Go comment instead of YAML.
- `TypeMeta` / `ObjectMeta` embedding is what gives every CR the standard `apiVersion`/`kind`/`metadata` fields for free — you never define those yourself.
- `WebAppStatus` (generated alongside `WebAppSpec`, initially also just a placeholder field) is where your controller will report `readyReplicas`, `phase`, `conditions`, etc.

```go
// +kubebuilder:object:root=true

// WebAppList contains a list of WebApp
type WebAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebApp `json:"items"`
}
```
This `List` type is what makes `kubectl get webapps` (plural, returning many) work — every custom type needs a paired `List` type registered with the scheme.

### 8.2 — `zz_generated.deepcopy.go`

Never hand-edit this file — it's fully regenerated by `make generate` every time. It exists because `controller-runtime`'s client and cache machinery need to safely **copy** objects (e.g. before mutating a cached object, or handing a copy to your reconciler) without deep-copy logic breaking on nested structs, maps, or slices. Writing this by hand for every type would be tedious and error-prone; `controller-gen` derives it purely from your struct definitions.

### 8.3 — `internal/controller/webapp_controller.go`

The heart of the scaffold — this is where your reconcile loop from Part 2 actually gets implemented:

```go
type WebAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.example.com,resources=webapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.example.com,resources=webapps/status,verbs=get;update;patch

func (r *WebAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your reconcile logic here

	return ctrl.Result{}, nil
}

func (r *WebAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.WebApp{}).
		Complete(r)
}
```

Line by line, tying directly back to Part 2's reconcile-loop pseudocode:

- **`+kubebuilder:rbac:...` markers** — generate the actual `Role`/`ClusterRole` YAML in `config/rbac/role.yaml`. This is where the "controller needs permission to watch/edit its resource" requirement from Part 2, section 4 becomes real, enforced YAML — generated, not hand-written.
- **`Reconcile(ctx, req)`** — this *is* the reconcile loop body from Part 2. `req` gives you just a namespace/name; the convention is you re-fetch the full object fresh every call (never trust a cached copy for logic), compare against real state, and act. Right now it's an empty stub — literally `return ctrl.Result{}, nil`, meaning "nothing to do, don't requeue."
- **`SetupWithManager` / `.For(&webappv1.WebApp{})`** — this is what tells `controller-runtime` *which* resource's changes should trigger `Reconcile()` — i.e., it wires up the watch. Every create/update/delete of a `WebApp` anywhere in the cluster enqueues a request that eventually calls `Reconcile()` for that specific object's namespace/name.

**What you'd fill in here next** (beyond this section's scope, but worth previewing): fetch the `WebApp`, check if its child Deployment exists, create it if not, compare `spec.replicas` to actual replica count, patch if different, then write `.status`. This is precisely the `SampleDB` logic narrated conceptually in Part 2, section 4 — now you can see exactly which function it lives in.

### 8.4 — `main.go`

```go
if err = (&controller.WebAppReconciler{
	Client: mgr.GetClient(),
	Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
	setupLog.Error(err, "unable to create controller", "controller", "WebApp")
	os.Exit(1)
}
```

This is the line `kubebuilder create api` auto-inserted for you — it's the literal moment your new controller gets registered with the running manager process. Delete it, and the CRD would still exist and accept objects, but *nothing* would ever reconcile them — a useful mental check for debugging "why isn't my operator doing anything" later: confirm this wiring exists before debugging reconcile logic itself.

### 8.5 — The Makefile targets that tie it together

| Target | What it does |
|---|---|
| `make manifests` | Runs `controller-gen`, turning your `+kubebuilder:...` markers into `config/crd/` and `config/rbac/` YAML |
| `make generate` | Regenerates `zz_generated.deepcopy.go` |
| `make install` | `kubectl apply -k config/crd` — installs just the CRDs |
| `make run` | Runs the manager binary locally (outside the cluster) against your current kubeconfig — fastest inner dev loop |
| `make docker-build` / `make deploy` | Builds the controller image and applies `config/default` (CRDs + RBAC + Deployment) to actually run it in-cluster |

---

## What's Next

You've now gone from a hand-written CRD (Part 1), to the conceptual Operator pattern (Part 2), to a fully scaffolded, compiling Kubebuilder project with a real (if empty) `Reconcile()` function wired up end-to-end (Part 3). The next natural step is filling in that `Reconcile()` body: fetching child resources, creating/updating a Deployment to match `spec`, handling deletion with finalizers, and writing meaningful `.status` — turning the `WebApp` scaffold into an actually functioning Operator.