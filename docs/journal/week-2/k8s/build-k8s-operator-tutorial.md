# Building Kubernetes Operators with Go and Kubebuilder

> Source: "Build Your Own Kubernetes Operators with Go and Kubebuilder" (Shubham Ghatara / CubeSimplify) — a hands-on course that builds a real operator managing AWS EC2 instances as native Kubernetes objects.

## Table of Contents

1. [Why This Exists — Kubernetes as an SDK](#1-why-this-exists--kubernetes-as-an-sdk)
2. [Core Concepts](#2-core-concepts)
   - [2.1 Controllers and the Reconcile Loop](#21-controllers-and-the-reconcile-loop)
   - [2.2 Idempotency — the Golden Rule](#22-idempotency--the-golden-rule)
   - [2.3 Operators = Controller + CRD + CR](#23-operators--controller--crd--cr)
   - [2.4 Group, Version, Kind (GVK)](#24-group-version-kind-gvk)
3. [Environment Setup](#3-environment-setup)
4. [Bootstrapping the Project with Kubebuilder](#4-bootstrapping-the-project-with-kubebuilder)
5. [Designing the API — Spec and Status](#5-designing-the-api--spec-and-status)
6. [Anatomy of `main.go` — the Manager](#6-anatomy-of-maingo--the-manager)
7. [Writing the Reconciler](#7-writing-the-reconciler)
   - [7.1 First pass: read and print](#71-first-pass-read-and-print)
   - [7.2 The Reconcile Result contract](#72-the-reconcile-result-contract)
8. [Talking to AWS — Creating a Real EC2 Instance](#8-talking-to-aws--creating-a-real-ec2-instance)
9. [Idempotency in Practice](#9-idempotency-in-practice)
10. [Finalizers and Deletion](#10-finalizers-and-deletion)
11. [What Actually Happens Under the Hood](#11-what-actually-happens-under-the-hood)
12. [Packaging and Shipping: Docker + Helm](#12-packaging-and-shipping-docker--helm)
13. [RBAC — Why Your Operator Needs Permissions](#13-rbac--why-your-operator-needs-permissions)
14. [Full Reference: Example YAML](#14-full-reference-example-yaml)
15. [Command Cheat Sheet](#15-command-cheat-sheet)
16. [Where to Go Next](#16-where-to-go-next)

---

## 1. Why This Exists — Kubernetes as an SDK

Most people learn Kubernetes as a platform to *run* things on: pods, deployments, services. This course flips that framing. Kubernetes is also a **software development kit (SDK)** for building your own control systems — because of four properties it ships with out of the box:

- **Extensibility** — you can teach the API server about brand-new resource kinds (Custom Resource Definitions).
- **API-first design** — every object (pod, service, and your own kinds) is just a versioned API resource with client libraries in Go, Python, Java, etc.
- **Backward compatibility & versioning** — APIs deprecate gracefully (`v1`, `v1beta1`, ...), so you can evolve your own APIs the same way.
- **A pluggable control-loop pattern** — the same reconciliation pattern that runs `kube-controller-manager` is available to you for *anything*, not just built-in resources.

The concrete deliverable of the course is a **cloud controller**: an operator that watches for a custom `EC2Instance` object and creates/deletes a real AWS EC2 instance to match it — Kubernetes as the control plane for infrastructure outside the cluster.

<details>
<summary>Analogy used in the course: the "helper" in a new country</summary>

You move to a new country and rent a furnished house. If something breaks (a tap, a window), you *could* fix it yourself — or you can have a **helper** who already knows the local hardware store, the language, and the process. You just say "get me a key" and the helper handles the how.

- The **helper** = the operator.
- **You** = the developer submitting a YAML.
- The **house** = the underlying resource (a database, an EC2 instance, an S3 bucket).

You (the platform team) encode all the "how" into the operator once. Developers only ever state the "what" via a YAML spec.
</details>

---

## 2. Core Concepts

### 2.1 Controllers and the Reconcile Loop

A **controller** is a forever-running loop with this pseudocode shape:

```
for {
    observe current state of the resource
    compare current state vs desired state (spec)
    if drift:
        act (create/update/delete) to close the drift
        update .status
    else:
        do nothing   // the "happy path"
}
```

This is why Kubernetes is described as **eventually consistent**: a change doesn't apply instantly, but the loop keeps converging cluster state toward desired state.

```
        ┌────────────────────────────────────────────┐
        │                Reconcile Loop               │
        │                                              │
   ┌───►│ 1. Get object (from cache)                   │
   │    │ 2. Read desired state (spec)                 │
   │    │ 3. Read actual state (status / external API) │
   │    │ 4. spec == actual? ──► do nothing, exit       │
   │    │           │no                                │
   │    │           ▼                                  │
   │    │ 5. Create/Update/Delete to close the drift    │
   │    │ 6. Update .status                             │
   └────┤ 7. Return (Result, error)                     │
        └────────────────────────────────────────────┘
```

Every built-in Kubernetes resource (Pod, Deployment, Service, ConfigMap...) has its own controller running this same loop inside `kube-controller-manager`. A **cloud controller** (like the one you're building) is the same pattern pointed at an external system (AWS) instead of purely internal cluster state — this is exactly what EKS/AKS do behind the scenes when you set a Service's type to `LoadBalancer`.

### 2.2 Idempotency — the Golden Rule

> **Only write to the API server when actual state differs from desired state.**

This is stressed repeatedly in the course because of a subtle trap: **any** write to a custom resource — spec, status, labels, or annotations — triggers another reconcile. If your reconciler always writes something (e.g. a "last synced" timestamp) even when nothing needs to change, you create an infinite self-triggering loop:

```
update object → reconcile fires → reconciler updates "last sync" →
update object → reconcile fires → reconciler updates "last sync" → ...
```

The fix: before writing anything, always ask **"is there actually anything for me to do?"** If the current state already matches desired state, exit without touching the API server.

Idempotency also means your reconciler must be **stateless** between runs. It never remembers "I already created this" from memory — it re-derives that fact every single time by checking `.status` (or the external system) fresh. This is precisely what makes an operator resilient to crashes, restarts, and rescheduling: it picks up exactly where reality is, not where it thinks it left off.

### 2.3 Operators = Controller + CRD + CR

| Term | What it is |
|---|---|
| **CRD** (Custom Resource Definition) | The *schema* you register with Kubernetes — tells the API server "there is a kind called `EC2Instance` and here is its shape." Analogous to a table definition. |
| **CR** (Custom Resource) | An *instance* of that schema — the actual YAML object a user submits. Analogous to a table row. |
| **Controller** | The logic that reacts when a CR is created/updated/deleted. Without a controller, a CR is just inert data Kubernetes stores — nothing "does" anything with it. |
| **Operator** | CRD + CR + Controller, packaged together, usually with RBAC and a deployment manifest. |

### 2.4 Group, Version, Kind (GVK)

Every Kubernetes resource — built-in or custom — is uniquely identified by a **Group/Version/Kind**, similar to a URL:

```
apiVersion: compute.cloud.com/v1     ← group = compute.cloud.com, version = v1
kind: EC2Instance                     ← kind
```

Built-in core resources (Pod, Service) have an empty group name (the "core" group), which is why `kubectl explain pod` shows `VERSION: v1` with no group prefix. Your custom API lives in whatever domain/group you choose at project-init time (e.g. `cloud.com`, subdomain `compute`).

---

## 3. Environment Setup

You need a local Kubernetes cluster, Docker, and Go. The course uses **K3D** (a lightweight Kubernetes-in-Docker distribution by Rancher, sibling to K3S).

<details>
<summary>Why K3D</summary>

- Free, fast, runs entirely on your machine.
- Each cluster gets its own isolated Docker network.
- Supports version-controlled cluster definitions via a config file — you can commit your local dev cluster spec to git (GitOps for your dev environment too).
- Automatically wires up your `kubeconfig` context and a load-balancer proxy container for the API server.
</details>

<details>
<summary>Cluster config (example)</summary>

```yaml
# k3d-cluster.yaml
apiVersion: k3d.io/v1alpha5
kind: Simple
servers: 1        # single control-plane, no HA needed locally
agents: 2          # two worker nodes
image: rancher/k3s:v1.31.5-k3s1
```

```bash
k3d cluster create --config k3d-cluster.yaml
kubectl cluster-info
kubectl get nodes
```
</details>

**Smoke test** the cluster before building anything:

```bash
kubectl create deployment my-deployment --image=nginx
kubectl expose deployment my-deployment --port=80
kubectl run curl --image=curlimages/curl --restart=Never -- curl <cluster-ip>
kubectl run google --image=curlimages/curl --restart=Never -- curl https://www.google.com
```
This confirms pod-to-pod networking and external egress — both matter, because your operator will need to reach the AWS API from inside the cluster.

**Toolchain checklist:**

| Tool | Purpose |
|---|---|
| Go | Language the operator/controller is written in |
| Docker (or a Docker-compatible runtime) | Required by K3D and for building the operator's container image |
| `kubectl` | Cluster interaction |
| `kubebuilder` (v4.x) | Scaffolding framework for the whole project |
| AWS credentials (`AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`) | Needed later so the controller can call the EC2 API |

Install kubebuilder from its GitHub releases (pick the binary matching your OS/architecture), then verify:
```bash
kubebuilder version
```

---

## 4. Bootstrapping the Project with Kubebuilder

Kubebuilder is a scaffolding framework: it generates the boilerplate (project layout, Dockerfile, Makefile, RBAC manifests, main entrypoint, test harness) so you only write the business logic — the spec fields and the reconcile function.

```bash
kubebuilder init --domain cloud.com --repo github.com/<you>/ec2-operator
```

This creates, among other things:

```
.
├── cmd/main.go              # entrypoint — starts the manager
├── config/                  # kustomize manifests (RBAC, CRDs, deployment)
├── Dockerfile
├── Makefile                 # manifests, install, run, docker-build, docker-push...
└── PROJECT                  # project metadata (domain, repo, version)
```

Next, generate an API (this is the important step — it's where your custom resource is born):

```bash
kubebuilder create api --group compute --kind EC2Instance
# Create Resource? [y]
# Create Controller? [y]
```

This adds:

```
api/v1/ec2instance_types.go         # your Spec/Status Go structs
api/v1/groupversion_info.go         # registers the GVK with the scheme
internal/controller/ec2instance_controller.go   # your Reconcile() function
```

> A single project (one domain, e.g. `cloud.com`) can host **multiple** APIs — `compute` (EC2Instance, AMI, SecurityGroup), `storage` (EBS, S3Bucket), `network` (VPC, FirewallRule) — each created with its own `kubebuilder create api` call.

---

## 5. Designing the API — Spec and Status

Every Kubernetes object separates **desired state** (`spec`, written by the user) from **observed state** (`status`, written by the controller). Design both deliberately.

<details>
<summary><code>api/v1/ec2instance_types.go</code> — Spec</summary>

```go
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EC2InstanceSpec defines the desired state of an EC2Instance.
type EC2InstanceSpec struct {
	AMIID          string            `json:"amiId"`
	InstanceType   string            `json:"instanceType"`
	SSHKey         string            `json:"sshKey"`
	Region         string            `json:"region"`
	AvailabilityZone string          `json:"availabilityZone"`
	SecurityGroups []string          `json:"securityGroups,omitempty"`
	Subnet         string            `json:"subnet,omitempty"`
	AssociatePublicIP bool           `json:"associatePublicIp,omitempty"`
	UserData       string            `json:"userData,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	Storage        StorageConfig     `json:"storage,omitempty"`
}

type StorageConfig struct {
	Root       VolumeConfig   `json:"root"`
	Additional []VolumeConfig `json:"additional,omitempty"`
}

type VolumeConfig struct {
	Size      int32  `json:"size"`      // GiB
	Type      string `json:"type"`      // e.g. gp2, gp3
	Device    string `json:"device,omitempty"`
	Encrypted bool   `json:"encrypted,omitempty"`
}

// EC2InstanceStatus defines the observed state of an EC2Instance.
type EC2InstanceStatus struct {
	Phase      string `json:"phase,omitempty"`
	InstanceID string `json:"instanceId,omitempty"`
	PublicIP   string `json:"publicIp,omitempty"`
	PrivateIP  string `json:"privateIp,omitempty"`
	PublicDNS  string `json:"publicDns,omitempty"`
	PrivateDNS string `json:"privateDns,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type EC2Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EC2InstanceSpec   `json:"spec,omitempty"`
	Status EC2InstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type EC2InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EC2Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EC2Instance{}, &EC2InstanceList{})
}
```
</details>

**Design notes:**

- **JSON tags are mandatory.** Kubernetes marshals/unmarshals YAML↔struct through these tags; without them the API server can't map `amiId` in a user's YAML to `AMIID` in your Go struct.
- **`omitempty` controls required-vs-optional.** A field *without* `omitempty` becomes a `required` property in the generated CRD schema — the API server will reject a submission missing it. Use this deliberately: AMI ID and instance type should be required; tags and extra storage should not.
- **`+kubebuilder:object:root=true`** and **`+kubebuilder:subresource:status`** are *markers*, not comments — kubebuilder's code generator parses them to produce `DeepCopy` methods and the CRD manifest. Treat every `+kubebuilder:...` line as active configuration.
- **Every resource has `TypeMeta`** (kind + apiVersion) **and `ObjectMeta`** (name, namespace, UID, resourceVersion, creation timestamp...) — this is standard Kubernetes plumbing, not something you invent per-CRD.

Whenever you edit the spec/status structs, **you must regenerate and reinstall the CRD** — the cluster's copy of the schema is a separate artifact from your Go code:

```bash
make manifests   # regenerates config/crd/bases/*.yaml from your +kubebuilder markers
make install     # applies the CRD to whatever cluster your KUBECONFIG points to
```

If you skip this after adding a new field, the field will exist in your Go struct but Kubernetes will reject or silently drop it on `kubectl apply` because the CRD schema hasn't caught up.

---

## 6. Anatomy of `main.go` — the Manager

`cmd/main.go` is the process entrypoint. Conceptually:

```
main.go
  └── builds a Scheme (a "phone book" of known GVKs: core types + your EC2Instance)
  └── parses CLI flags (metrics address, health-probe address, leader-election, TLS)
  └── creates a Manager
        ├── holds a client (talks to the API server)
        ├── holds a cache (in-memory mirror of watched objects)
        ├── hosts one or more Controllers
        │      └── each Controller wraps your Reconciler (your business logic)
        ├── optionally hosts webhooks (validating/mutating — not covered here)
        ├── serves /metrics (Prometheus format, built in for free)
        └── serves /healthz and /readyz (built in for free)
  └── manager.Start(ctx)   // blocks; this is "turning the key"
```

**Scheme registration** — before your controller can do anything with `EC2Instance`, Kubernetes' internal machinery needs to know the type exists:

```go
var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme)) // built-in types: Pod, Deployment, Service...
	utilruntime.Must(computev1.AddToScheme(scheme))      // your custom EC2Instance type
}
```

**Leader election** matters the moment you run more than one replica of your operator for high availability. Without it, N replicas would *all* react to the same event and (in this course's case) create N EC2 instances for a single CR. Kubebuilder wires this up with a single boolean flag; only the elected leader processes reconciles, and a standby takes over automatically if the leader dies.

```go
mgr, err := ctrl.NewManager(cfg, ctrl.Options{
	Scheme:                 scheme,
	Metrics:                metricsServerOptions,
	HealthProbeBindAddress: probeAddr,
	LeaderElection:         enableLeaderElection,
	LeaderElectionID:       "ec2operator.cloud.com",
})
```

**Registering your controller with the manager** (also generated, but worth understanding):

```go
if err = (&controller.EC2InstanceReconciler{
	Client: mgr.GetClient(),
	Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
	setupLog.Error(err, "unable to create controller", "controller", "EC2Instance")
	os.Exit(1)
}
```

Running it locally (no container yet) reuses your local `KUBECONFIG`, exactly as `kubectl` does:

```bash
go run cmd/main.go
```

---

## 7. Writing the Reconciler

### 7.1 First pass: read and print

The reconciler lives in `internal/controller/ec2instance_controller.go`. Start by simply pulling the object and logging what came in — this is the fastest way to confirm your watch/scheme wiring is correct before writing any real logic.

<details>
<summary>Minimal reconcile — read the object, log it, do nothing</summary>

```go
func (r *EC2InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var instance computev1.EC2Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Object was deleted; nothing to reconcile.
			return ctrl.Result{}, nil
		}
		// Some other error reading the object — ask Kubernetes to retry.
		return ctrl.Result{}, err
	}

	log.Info("reconciling EC2Instance",
		"namespace", req.Namespace,
		"name", instance.Name,
		"amiId", instance.Spec.AMIID,
		"instanceType", instance.Spec.InstanceType,
	)

	return ctrl.Result{}, nil
}
```
</details>

Two details worth internalizing here:

1. **`r.Get` reads from the manager's local cache, not directly from etcd via the API server.** This is why reconciles are fast — no network round trip per read.
2. **`apierrors.IsNotFound(err)`** is how you distinguish "the object simply doesn't exist (e.g., it was deleted, and this reconcile is the tail end of that deletion)" from "something is actually broken." Getting this branch wrong is a common source of noisy retry loops.

### 7.2 The Reconcile Result contract

`Reconcile` always returns `(ctrl.Result, error)`. **This return value is a contract with Kubernetes about what happens next**, and getting it wrong is the single biggest source of confusing operator behavior (infinite loops, or an operator that silently stops responding).

| You return | Kubernetes does |
|---|---|
| `error != nil` (Result is **ignored** entirely) | Retries with **exponential backoff**. This is how self-healing works — e.g. a pod stuck Pending due to no matching node keeps quietly retrying until a node becomes available. |
| `error == nil`, `Result.RequeueAfter = 20s` | No error, but re-run this reconcile in exactly 20s regardless of any new events — good for periodic drift detection. |
| `error == nil`, `Result.Requeue = true`, no `RequeueAfter` | No error, but retry again — timing left to Kubernetes' own exponential backoff. |
| `error == nil`, `Result{}` (empty) | Everything is fine. Wait for the *next real event* on this resource. This is the steady state you want once nothing needs to change. |

```
                  err != nil?
                 /           \
              yes             no
               │               │
        ignore Result    RequeueAfter set?
        exponential          /        \
        backoff retry      yes         no
                             │           │
                     requeue at      Requeue=true?
                     fixed delay      /        \
                                    yes          no
                                     │            │
                              exp. backoff   wait for next
                                 retry        watch event
```

---

## 8. Talking to AWS — Creating a Real EC2 Instance

This is where the reconciler stops being Kubernetes-flavored pseudocode and starts calling the AWS SDK for Go. The flow inside your `createEC2Instance` helper:

```
1. Build an EC2 client (reads AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY from env)
2. Build RunInstancesInput  (AMI, instance type, key pair, min/max count...)
3. Call client.RunInstances()   → returns an instance, but likely still "pending"
4. Use a Waiter (NewInstanceRunningWaiter) to block until state == "running"
5. Call client.DescribeInstances()  → now the public IP/DNS are populated
6. Return a small CreatedInstanceInfo struct back to the reconciler
```

<details>
<summary>Go — building the client and creating the instance</summary>

```go
func newEC2Client(ctx context.Context, region string) (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return ec2.NewFromConfig(cfg), nil
}

type CreatedInstanceInfo struct {
	InstanceID string
	PublicIP   string
	PrivateIP  string
	PublicDNS  string
	PrivateDNS string
	State      string
}

func createEC2Instance(ctx context.Context, spec computev1.EC2InstanceSpec) (*CreatedInstanceInfo, error) {
	log := logf.FromContext(ctx).WithName("createEC2Instance")

	client, err := newEC2Client(ctx, spec.Region)
	if err != nil {
		return nil, err
	}

	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(spec.AMIID),
		InstanceType: types.InstanceType(spec.InstanceType),
		KeyName:      aws.String(spec.SSHKey),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	out, err := client.RunInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 instance: %w", err)
	}
	if len(out.Instances) == 0 {
		return nil, fmt.Errorf("no instances returned by RunInstances")
	}

	instanceID := aws.ToString(out.Instances[0].InstanceId)
	log.Info("EC2 instance creation call succeeded", "instanceId", instanceID)

	// The instance exists but may not be "running" yet, and has no public IP yet.
	waiter := ec2.NewInstanceRunningWaiter(client)
	if err := waiter.Wait(ctx,
		&ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}},
		3*time.Minute,
	); err != nil {
		return nil, fmt.Errorf("instance did not reach running state: %w", err)
	}

	// Now that it's running, describe it again to pick up the public IP/DNS.
	desc, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	inst := desc.Reservations[0].Instances[0]
	return &CreatedInstanceInfo{
		InstanceID: instanceID,
		PublicIP:   aws.ToString(inst.PublicIpAddress),
		PrivateIP:  aws.ToString(inst.PrivateIpAddress),
		PublicDNS:  aws.ToString(inst.PublicDnsName),
		PrivateDNS: aws.ToString(inst.PrivateDnsName),
		State:      string(inst.State.Name),
	}, nil
}
```
</details>

**Why a Waiter instead of a manual polling loop:** a hand-rolled `for { poll every 5s }` works, but the AWS SDK's `Waiter` types implement smarter, backing-off polling with a bounded max wait, and they return immediately once the target state is reached rather than blocking for the full timeout. Prefer SDK-provided waiters over hand-written polling wherever the SDK offers one.

**Why `aws.ToString(...)` everywhere:** most AWS SDK response fields are `*string` pointers because a field may be legitimately absent (nil) — e.g. no public IP if the instance has none. Dereferencing a nil pointer directly panics; the SDK's helper safely converts nil → `""`.

---

## 9. Idempotency in Practice

Recall the golden rule from [§2.2](#22-idempotency--the-golden-rule). In this operator, the concrete implementation is a single check at the top of the "create" branch:

```go
if instance.Status.InstanceID != "" {
	// We've already created this instance in a previous reconcile.
	// Do nothing — this is the idempotency guard.
	log.Info("requested object already exists, not creating a new instance")
	return ctrl.Result{}, nil
}

// Status.InstanceID is empty → this is genuinely a new object.
info, err := createEC2Instance(ctx, instance.Spec)
```

Why this works: the reconciler is **stateless** — it never "remembers" that it created an instance five minutes ago. The only durable record of that fact is `status.instanceID`, which lives in etcd. So every reconcile re-derives "have I done this already?" by reading status fresh, rather than trusting in-memory state.

> **Extension idea called out explicitly in the course (left as an exercise):** add *drift detection*. If `status.instanceID` is already set, instead of doing nothing, call `DescribeInstances` again and compare the live AWS state (stopped/running/terminated) against `status.phase`. If they differ, update status to match reality. This turns the operator from "create once" into "continuously reconciling," matching how real controllers behave.

---

## 10. Finalizers and Deletion

Kubernetes has no built-in way to know "this resource change was a deletion" versus a normal update — both are just "the object changed." The signal is a special field: **`.metadata.deletionTimestamp`**.

```
kubectl delete ec2instance my-instance
        │
        ▼
API server sets .metadata.deletionTimestamp (does NOT remove the object yet
        if a finalizer is present)
        │
        ▼
Reconciler fires (this is just another "update" event)
        │
        ▼
Reconciler checks: is DeletionTimestamp set?
        │
       yes
        │
        ▼
Call AWS TerminateInstances, wait for "terminated" state
        │
        ▼
Remove our finalizer from .metadata.finalizers
        │
        ▼
That removal is itself an update → reconciler fires ONE more time
        │
        ▼
This time r.Get() returns NotFound (object is now actually gone) →
return (Result{}, nil) — nothing left to do
```

**Adding a finalizer** (on first creation, alongside creating the instance):

```go
const ec2Finalizer = "ec2instance.compute.cloud.com/finalizer"

if !controllerutil.ContainsFinalizer(&instance, ec2Finalizer) {
	controllerutil.AddFinalizer(&instance, ec2Finalizer)
	if err := r.Update(ctx, &instance); err != nil {
		return ctrl.Result{}, err
	}
}
```

**Handling deletion:**

```go
if !instance.DeletionTimestamp.IsZero() {
	if controllerutil.ContainsFinalizer(&instance, ec2Finalizer) {
		if err := deleteEC2Instance(ctx, instance.Status.InstanceID, instance.Spec.Region); err != nil {
			// Don't remove the finalizer until cleanup actually succeeds —
			// this is what "the finalizer holds the deletion" means in practice.
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(&instance, ec2Finalizer)
		if err := r.Update(ctx, &instance); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}
```

**Why this matters:** without a finalizer, `kubectl delete` removes the Kubernetes object *immediately*, and your reconciler never gets a chance to clean up the real AWS instance — you'd end up with an orphaned, still-billing EC2 instance with no corresponding Kubernetes record to find it by. The finalizer makes deletion **synchronous with real cleanup**.

---

## 11. What Actually Happens Under the Hood

This section demystifies something that looks like a bug the first time you see it: you update the object *twice* in one reconcile (once for the finalizer, once for status) — yet you often only see **one** subsequent reconcile log line, not two.

```
 ┌──────────┐   watch update    ┌───────────┐  add to cache   ┌───────┐
 │API server│ ────────────────► │  Informer  │ ──────────────► │ Cache │
 └──────────┘                   └─────┬─────┘                 └───────┘
                                       │ triggers
                                       ▼
                              ┌─────────────────┐
                              │ Event Handler     │  adds ONLY
                              │ (Add/Update/Del)  │  {namespace, name}
                              └────────┬──────────┘  — not the full object
                                       ▼
                              ┌─────────────────┐
                              │   Work Queue      │◄── de-duplicates:
                              │  (per controller)  │    same key already
                              └────────┬──────────┘    queued? no-op
                                       ▼
                              ┌─────────────────┐
                              │      Worker        │  runs Reconcile(),
                              │  (your Reconcile)  │  reads full object
                              └─────────────────┘   from the Cache
```

Key mechanics:

- **The work queue stores keys (`namespace/name`), not full objects.** The worker fetches the actual spec/status from the **cache** (not etcd, not the API server) when it needs it — this is why reads inside `Reconcile` are cheap.
- **The work queue de-duplicates.** If you update an object twice in quick succession — e.g. once to add a finalizer, once to write status — both updates try to enqueue the same `namespace/name` key. If the worker hasn't picked up the first one yet, the second is coalesced into it: **one** key, **one** subsequent reconcile run, not two. This is precisely why the "expected two log lines, got one" behavior in the course is not a bug — it's the work queue doing its job.
- **The informer keeps the cache fresh** on every Add/Update/Delete event from the API server, so the worker is never reading stale data even though it never talks to etcd directly.
- **`RequeueAfter`** is what lets an operator behave like a cron job even with zero new events — useful for periodic polling/drift-detection independent of anything a user does.

---

## 12. Packaging and Shipping: Docker + Helm

Once the reconciler works when run locally (`go run cmd/main.go` against your `KUBECONFIG`), the next step is running it **as a pod inside the cluster** it manages.

**Build and push the container image** (Makefile targets generated by kubebuilder):

```bash
# Edit IMG in the Makefile first, e.g.:  IMG ?= yourrepo/ec2-operator:latest
make docker-build
make docker-push
```

The generated `Dockerfile` builds a small distroless-style image: compiles a static Go binary (`manager`) and runs it as a non-root user (UID 65532) — standard security practice for controller images.

**Two ways to get deployable manifests:**

| Approach | Command | Result |
|---|---|---|
| Single-file installer | `make build-installer` | One big `dist/install.yaml` with Namespace, CRD, ServiceAccount, RBAC, Deployment, Service — `kubectl apply -f` and done. No templating, no version control friendliness. |
| Helm chart | `kubebuilder edit --plugins=helm/v1-alpha` | Generates a full Helm chart under `dist/chart/`, with a `values.yaml` for image, replica count, resource limits, RBAC toggles, CRD install/keep-on-uninstall behavior, and Prometheus ServiceMonitor wiring. |

The Helm route is what you want for anything beyond a one-off local test — it gives users a single `values.yaml` knob instead of requiring them to understand RBAC internals.

<details>
<summary>Deploying with Helm</summary>

```bash
helm install ec2-operator ./dist/chart \
  --create-namespace \
  --namespace ec2-operator \
  -f values.yaml
```

Minimal `values.yaml` additions for AWS credentials (⚠️ **plaintext env vars are shown in the course purely as a quick demo — use a Kubernetes `Secret` + `envFrom`/`secretKeyRef` in anything real, or better, an external-secrets operator pulling from Vault/AWS Secrets Manager/GCP Secret Manager**):

```yaml
controllerManager:
  container:
    env:
      - name: AWS_ACCESS_KEY_ID
        valueFrom:
          secretKeyRef: { name: aws-credentials, key: access-key-id }
      - name: AWS_SECRET_ACCESS_KEY
        valueFrom:
          secretKeyRef: { name: aws-credentials, key: secret-access-key }
```

```bash
kubectl create secret generic aws-credentials \
  --namespace ec2-operator \
  --from-literal=access-key-id=<key> \
  --from-literal=secret-access-key=<secret>
```
</details>

**Verifying the running pod:**

```bash
kubectl get pods -n ec2-operator
kubectl logs -n ec2-operator deploy/ec2-operator-controller-manager
```

The manager binary exposes two things for free, without you writing any HTTP handlers:

- **Health/readiness probes** at `:8081/healthz` and `:8081/readyz` (used by the Deployment's liveness/readiness probes).
- **Prometheus metrics** at `:8080/metrics` (or `:8443` if TLS-secured) — including `controller_runtime_reconcile_total` broken down by result (`success`/`error`/`requeue`). This is your out-of-the-box signal for "is my operator healthy," before you've instrumented any custom business metrics (e.g. "EC2 instances created").

---

## 13. RBAC — Why Your Operator Needs Permissions

The operator pod runs under a **ServiceAccount**. That ServiceAccount needs a `ClusterRole` + `ClusterRoleBinding` (not a namespaced `Role`) if you want users in *any* namespace to create `EC2Instance` objects and have your single operator deployment see them all — this is the standard pattern (operator in its own dedicated namespace, custom resources created by users in their own namespaces).

Minimum verbs required:

| Resource | Verbs | Why |
|---|---|---|
| `ec2instances` | `get`, `list`, `watch` | To observe spec changes |
| `ec2instances` | `create`, `update`, `patch`, `delete` | To manage finalizers and metadata |
| `ec2instances/status` | `get`, `update`, `patch` | Status is a separate subresource — writing `instanceID`/`publicIP` needs its own permission grant even though it's "the same object" |

Kubebuilder generates all of this automatically from `+kubebuilder:rbac:...` marker comments directly above your `Reconcile` function — you rarely hand-write RBAC YAML for your own CRD.

---

## 14. Full Reference: Example YAML

```yaml
apiVersion: compute.cloud.com/v1
kind: EC2Instance
metadata:
  name: my-instance
  namespace: default
spec:
  amiId: ami-0abcdef1234567890
  instanceType: t3.medium
  sshKey: my-keypair
  region: eu-central-1
  availabilityZone: eu-central-1a
  securityGroups:
    - sg-0123456789abcdef0
  subnet: subnet-0123456789abcdef0
  associatePublicIp: true
  tags:
    environment: dev
    owner: kazi
  storage:
    root:
      size: 20
      type: gp3
```

```bash
kubectl apply -f example/instance.yaml
kubectl get ec2instances
kubectl get ec2instance my-instance -o yaml     # inspect .status
kubectl delete -f example/instance.yaml         # triggers finalizer-gated cleanup
```

Expected `status` once reconciled:

```yaml
status:
  phase: running
  instanceId: i-0123456789abcdef0
  publicIp: 35.159.x.x
  privateIp: 172.31.x.x
  publicDns: ec2-35-159-x-x.eu-central-1.compute.amazonaws.com
  privateDns: ip-172-31-x-x.eu-central-1.compute.internal
```

---

## 15. Command Cheat Sheet

<details>
<summary>Full workflow, start to finish</summary>

```bash
# 1. Local cluster
k3d cluster create --config k3d-cluster.yaml
kubectl cluster-info

# 2. Scaffold project
kubebuilder init --domain cloud.com --repo github.com/you/ec2-operator
kubebuilder create api --group compute --kind EC2Instance

# 3. After ANY change to api/v1/*_types.go
make manifests
make install          # push CRD to whatever cluster KUBECONFIG points at

# 4. Iterate locally (uses local KUBECONFIG + local AWS env vars)
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
go run cmd/main.go
# in another shell:
kubectl apply -f example/instance.yaml --dry-run=client   # validate first
kubectl apply -f example/instance.yaml
kubectl delete -f example/instance.yaml

# 5. Containerize
# (edit IMG in Makefile first)
make docker-build
make docker-push

# 6. Package for distribution
kubebuilder edit --plugins=helm/v1-alpha
helm install ec2-operator ./dist/chart -n ec2-operator --create-namespace -f values.yaml

# 7. Verify in-cluster
kubectl get pods -n ec2-operator
kubectl logs -n ec2-operator deploy/ec2-operator-controller-manager -f
kubectl port-forward -n ec2-operator deploy/ec2-operator-controller-manager 8081:8081
curl localhost:8081/healthz
```
</details>

---

## 16. Where to Go Next

The course explicitly frames this as a foundation, not a finished product. Concrete extension ideas it calls out:

- **Drift detection** — periodically (`RequeueAfter`) re-describe the live AWS instance and reconcile `.status` against reality, not just against your own last write (e.g., someone stops the instance manually in the AWS console).
- **Custom Prometheus metrics** — instrument `createEC2Instance` to increment an "instances created" counter, beyond the built-in controller-runtime reconcile metrics.
- **Validating/mutating webhooks** — reject invalid specs (e.g. bad instance type) before they're even persisted, rather than failing inside `Reconcile`.
- **Different resource kinds in the same project** — an `S3Bucket` or `EFSFileSystem` API alongside `EC2Instance`, all under the same `compute`/`storage` group structure, all running in the same manager process.
- **External secrets** instead of plaintext AWS credentials in Helm values — use `external-secrets-operator` against Vault/AWS Secrets Manager/GCP Secret Manager.
- **Multi-arch images** via `docker buildx` if your dev machine and target cluster nodes differ in CPU architecture (ARM vs AMD64).

The transferable skill from this whole course isn't "how to create EC2 instances from YAML" — it's the **reconcile-loop mental model**: idempotency, status-vs-spec separation, finalizer-gated deletion, and the `(Result, error)` contract. That pattern applies unchanged whether the external system is AWS, Azure, GCP, or an internal company API.