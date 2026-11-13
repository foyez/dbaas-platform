# Custom Resources (CRs)

## What is a Custom Resource?

A **resource** is just an API endpoint that stores objects of a certain kind (e.g., `pods` stores Pod objects). A **custom resource (CR)** is a resource that isn't part of the default Kubernetes install — it extends the API with your own object type. Many core Kubernetes functions are now built using custom resources, which makes Kubernetes more modular, and CRs can appear and disappear in a running cluster through dynamic registration. Once registered, you can create and access CR objects with `kubectl`, just like built-in resources such as Pods.

## CRs vs. Custom Controllers (why this matters for Operators)

On their own, a CR just lets you store and retrieve structured data. Combine it with a custom controller, and you get a true **declarative API** — you declare desired state, and the controller keeps actual state in sync with it.

**Declarative vs Imperative — quick example:**

| Imperative (RPC-style) | Declarative (Kubernetes-style) |
|---|---|
| `scaleUp(db-service, +2)` → server does it now, returns "done" | `spec.replicas: 5` → controller notices and converges to it over time |
| Client tracks an operation ID to check completion | Client just reads `status` whenever it wants |

Signs an API is *not* declarative (per the K8s docs): synchronous "do this, wait for done" calls, operation IDs you poll, RPC-style talk, storing large blobs of data (>few kB/object), needing high request throughput (10s/sec sustained).

**This is the key concept before Operators:** the **Operator pattern = custom resources + custom controllers**, used to encode application-specific domain knowledge into the Kubernetes API. A CR is the *data model*; the Operator is the *controller logic* that acts on it.

## Two Ways to Add Custom Resources

| | **CRD** | **API Aggregation** |
|---|---|---|
| Programming required | No — can be created without any programming | Yes — requires writing/running your own API server |
| How it works | Registers a new type directly with the main API server | Runs a subordinate API server behind the main one, which proxies requests to it |
| Ongoing maintenance | None — picked up automatically in K8s upgrades | You must maintain the extra service, rebuild for API changes |
| Best for | Simple schemas, a handful of fields, internal/small projects | Full control over storage engine, custom version-conversion logic |

For learning Operators, **CRDs are what you'll use almost always** — this is the mechanism used by tools like Kubebuilder and Operator SDK.

## `spec`, `status`, and `metadata`

Every well-formed custom resource follows this convention (mentioned in the docs as a reason to prefer a CR over a ConfigMap):

- **`metadata`** — name, namespace, labels, annotations, UID. Standard on every K8s object.
- **`spec`** — the *desired state*, written by the user/client. "I want 3 replicas."
- **`status`** — the *observed/actual state*, written by the controller. "Currently 2 are ready."

```yaml
apiVersion: example.com/v1
kind: Database
metadata:
  name: my-db
spec:              # ← user writes this: "what I want"
  engine: postgres
  replicas: 3
status:             # ← controller writes this: "what's actually true"
  readyReplicas: 3
  phase: Running
```

This spec → observe → act → status loop is the engine behind the declarative model described above — the next section shows exactly how a controller implements it.

## The Reconcile Loop (Controllers, in depth)

A controller runs a **loop**, not a one-time action. The core idea is *level-triggered* reconciliation:

```
loop forever:
    observed = get current actual state
    desired  = get object.spec
    if observed != desired:
        take action to move observed → desired
    update object.status with what you observed
```

It does **not** react to individual events like "field X changed to Y" (that would be *edge-triggered*, and fragile — you could miss an event). Instead, whenever *anything* relevant changes (the CR itself, a resource it owns, or just a periodic resync), the controller re-reads the whole `spec` and re-checks whether reality matches it. If it already matches, it does nothing. This is what makes it self-healing: even if someone manually deletes a Pod the controller created, the next reconcile loop notices `status` no longer matches `spec` and recreates it.

**Example flow for the `Database` CR above:**

```
1. User applies: spec.replicas = 3
2. Controller's watch fires → reconcile(Database "my-db") is called
3. Controller checks: how many Pods actually exist for my-db? → finds 1
4. Controller creates 2 more Pods
5. Controller updates status.readyReplicas = 3 (once they're Running)
6. Next reconcile (nothing changed): observed == desired → no-op
```

## CRD Validation Schema (OpenAPI v3)

When you define a CRD, you attach a schema so the API server rejects bad objects *before* your controller ever sees them — no manual validation code needed for basic type/range/required checks.

**Full CRD example:**

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: databases.example.com          # must be <plural>.<group>
spec:
  group: example.com
  names:
    kind: Database
    plural: databases
    singular: database
    shortNames: ["db"]
  scope: Namespaced                     # or Cluster
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required: ["engine", "replicas"]
              properties:
                engine:
                  type: string
                  enum: ["postgres", "mysql"]
                replicas:
                  type: integer
                  minimum: 1
                  maximum: 10
                storageSize:
                  type: string
                  pattern: "^[0-9]+Gi$"
            status:
              type: object
              properties:
                readyReplicas:
                  type: integer
                phase:
                  type: string
```

**What this buys you — all enforced at `kubectl apply` time, at the API server level, before your controller runs:**
- `engine: mongo` → rejected (`not in enum`)
- `replicas: 50` → rejected (`> maximum`)
- missing `engine` → rejected (`required`)

The `Database` CR instance shown earlier in the `spec`/`status` section is what a user creates *against* this CRD — the CRD is the schema/type (like a `class`), the CR is an instance of it (like an `object`). Users normally only set `spec`; `status` is left for the controller to fill in.

## When *not* to use a CR

Avoid using a Custom Resource as data storage for application, end-user, or monitoring data — storing that kind of data inside the Kubernetes API usually signals architecture that's too tightly coupled. CRs are for **configuration/desired-state**, not general data storage. If part of your workload needs a backing service, run it as its own component instead of stuffing its data into a CR.

## CR vs ConfigMap (quick decision guide)

**Use a ConfigMap if:** there's an existing config file format (e.g. `mysql.cnf`), the whole config fits in one key, and a Pod just consumes it as a file/env var.

**Use a CR if:** you want `kubectl get`/`describe` support, you want automation to watch/react to changes, or you want the `spec`/`status`/`metadata` conventions.

## Updated mental model going into Operators

```
CRD        → defines the schema (spec/status fields + validation)
CR         → an instance of that schema, edited by users/clients
Controller → reconcile loop: watches CRs, compares status vs spec, acts, updates status
Operator   = CRD + Controller, packaged for a specific application's domain logic
```

## Checklist before moving to Operators

- ✅ CRD = schema definition that creates a new API type
- ✅ CR object = an *instance* of that type
- ✅ `spec` (desired) vs `status` (observed) vs `metadata`
- ✅ Controller = reconcile loop that watches CRs and drives real state toward `spec` (level-triggered, self-healing)
- ✅ CRD ≠ Operator — CRD is the API, Operator = CRD + controller
- ✅ Basic validation via OpenAPI schema in the CRD (enum, min/max, required, pattern)
- ✅ Comfortable reading/writing `kubectl get/describe/apply` on custom resources

Once solid on all of the above, you're ready to write your first controller (Go + controller-runtime, or scaffolded via Kubebuilder / Operator SDK).