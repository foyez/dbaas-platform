# Kubernetes CRDs
*(Covers: CRD fundamentals → schema validation → writing a CRD by hand → status/scale/printer columns/short names, with labs)*

---

## 1. Section Overview

This tutorial walks through everything you need to **write, apply, and extend a CustomResourceDefinition (CRD) by hand** — no scaffolding tools, no generators. By the end you'll be able to:

- Explain what a CRD and a CR actually are, and how they relate
- Write an OpenAPI v3 validation schema for a custom resource
- Hand-write a complete CRD YAML file from scratch
- Add a `status` subresource, a `scale` subresource, `additionalPrinterColumns`, and `shortNames`
- Apply CRs against your CRD and inspect them with `kubectl`

This is deliberately the "manual" path — understanding the raw YAML is what makes tools like Kubebuilder/Operator SDK (which generate this for you) make sense later, when you move on to Operators.

---

## 2. Custom Resources and CRDs — Explained

**Resource** = an API endpoint that stores a collection of objects of one kind (`pods`, `deployments`, `services`...).

**Custom Resource (CR)** = a resource type that isn't built into Kubernetes by default — you define it yourself.

**CustomResourceDefinition (CRD)** = the Kubernetes object you create that *registers* a new custom resource type with the API server. Think of it as a `class` definition; a CR is an `instance` of that class.

```
CRD  →  "Database is a valid kind, here's what fields it can have"
CR   →  an actual Database object: name=my-db, engine=postgres, replicas=3
```

Once you `kubectl apply` a CRD, the API server immediately starts accepting requests for that new type — `kubectl get databases`, `kubectl apply -f my-db.yaml`, `kubectl describe database my-db` all just work, the same as for built-in types.

**Why this matters:** a CRD *by itself* just gives you structured storage + validation + a nice `kubectl` interface. It does **nothing** on its own — nothing watches it, nothing reacts to it. That's the job of a *controller* (covered in later Operator material). This section is entirely about the data side: defining the shape of your custom object correctly.

**Key building blocks of a CRD**, which the rest of this tutorial covers one at a time:
| Field | Purpose |
|---|---|
| `group` / `version` / `kind` | Identity of your new API type |
| `scope` | Namespaced or Cluster-wide |
| `names` (plural, singular, shortNames) | How `kubectl` refers to it |
| `schema.openAPIV3Schema` | Validation rules for `spec`/`status` |
| `subresources.status` | Separates status updates from spec edits |
| `subresources.scale` | Lets `kubectl scale` work on your CR |
| `additionalPrinterColumns` | Custom columns shown in `kubectl get` |

---

## 3. OpenAPI Schema Validation

Every CRD needs a `schema.openAPIV3Schema` block. This is what makes the API server *reject bad objects before they're ever stored* — no controller code required for basic checks.

### Core validation keywords you'll use constantly

```yaml
type: object | string | integer | number | boolean | array
required: ["field1", "field2"]        # fields that must be present
properties:                            # nested field definitions
  fieldName:
    type: string
    enum: ["a", "b", "c"]              # allowed values only
    minimum: 1                         # for numbers
    maximum: 10
    minLength: 3                       # for strings
    maxLength: 20
    pattern: "^[a-z0-9-]+$"            # regex match
    default: "b"                       # default if omitted
```

### A minimal validated `spec`

```yaml
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
```

**What you get for free**, purely from this schema, at `kubectl apply` time:
- `engine: mongo` → rejected, `not in enum`
- `replicas: 0` or `replicas: 50` → rejected, out of range
- `storageSize: 10gigs` → rejected, doesn't match pattern
- missing `engine` entirely → rejected, `required` field missing

### Important rule: `x-kubernetes-preserve-unknown-fields`

By default (structural schemas, the only kind allowed since CRD `v1`), any field **not declared in the schema is pruned** — silently dropped, not rejected. If you genuinely want free-form/unstructured data somewhere, you must explicitly opt in:

```yaml
metadata:
  type: object
someMap:
  type: object
  x-kubernetes-preserve-unknown-fields: true
```

Without this, typos in field names don't error — they just vanish, which is a common gotcha the first time you hand-write a CRD.

### `status` gets its own schema block too

`status` should almost always be defined as its own sibling to `spec` under `properties`, since (as covered in section 6) it's usually meant to be written only by a controller, not the end user.

---

## 4. Demo: Write a CRD by Hand

Let's build a complete, working CRD from an empty file, field by field. We'll model a `Website` custom resource — something a hypothetical "Website Operator" would later manage.

### Step 1 — the skeleton

Every CRD starts with this shape:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: <plural>.<group>      # MUST match plural.group exactly
spec:
  group: <group>
  names:
    kind: <Kind>
    plural: <plural>
    singular: <singular>
  scope: Namespaced
  versions: []
```

For our example:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: websites.stable.example.com
spec:
  group: stable.example.com
  names:
    kind: Website
    plural: websites
    singular: website
  scope: Namespaced
  versions: []
```

> ⚠️ **Common mistake:** `metadata.name` is not arbitrary — it must be exactly `<spec.names.plural>.<spec.group>`. Get this wrong and the CRD is rejected outright.

### Step 2 — fill in the `versions` list

Each entry under `versions` needs `name`, `served`, `storage`, and a `schema`:

```yaml
  versions:
    - name: v1
      served: true      # is this version reachable via the API right now?
      storage: true      # is this the version actually persisted in etcd?
                         # exactly ONE version must have storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required: ["domain", "replicas"]
              properties:
                domain:
                  type: string
                  pattern: "^[a-z0-9.-]+\\.[a-z]{2,}$"
                replicas:
                  type: integer
                  minimum: 1
                  default: 1
                tlsEnabled:
                  type: boolean
                  default: false
```

`served` vs `storage` matters once you have multiple API versions (e.g. `v1alpha1`, `v1`) — you can serve several at once for backward compatibility, but only one can be the actual storage format.

### Step 3 — put it all together

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: websites.stable.example.com
spec:
  group: stable.example.com
  names:
    kind: Website
    plural: websites
    singular: website
  scope: Namespaced
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
              required: ["domain", "replicas"]
              properties:
                domain:
                  type: string
                  pattern: "^[a-z0-9.-]+\\.[a-z]{2,}$"
                replicas:
                  type: integer
                  minimum: 1
                  default: 1
                tlsEnabled:
                  type: boolean
                  default: false
```

### Step 4 — apply it and verify

```bash
kubectl apply -f website-crd.yaml
kubectl get crd websites.stable.example.com
kubectl explain website.spec        # confirms your schema is live
```

`kubectl explain` reading back your fields correctly is the fastest sanity check that the schema was parsed the way you intended.

### Step 5 — create a CR against it

```yaml
apiVersion: stable.example.com/v1
kind: Website
metadata:
  name: my-blog
spec:
  domain: myblog.com
  replicas: 2
```

```bash
kubectl apply -f my-blog.yaml
kubectl get websites
kubectl get website my-blog -o yaml
```

Try breaking validation on purpose to see it in action:

```bash
kubectl patch website my-blog --type merge -p '{"spec":{"replicas":0}}'
# error: spec.replicas in body should be greater than or equal to 1
```

---

## 5. Lab: Create and Apply Your First CRD

**Try this yourself before reading the solution below.**

**Task:**
1. Create a CRD for a resource called `CronTask` in group `batch.mycompany.com`, version `v1`, scope `Namespaced`.
2. `spec` must require: `schedule` (string, e.g. cron format) and `command` (string).
3. `spec` should optionally accept `retries` (integer, default `3`, minimum `0`, maximum `10`).
4. Apply the CRD, then create a `CronTask` named `cleanup-job` with:
   - `schedule: "0 * * * *"`
   - `command: "cleanup.sh"`
5. Verify it with `kubectl get crontasks` and `kubectl get crontask cleanup-job -o yaml`.
6. Bonus: try creating one with `retries: 99` and confirm it's rejected.

---

## 6. Lab Solution: Create and Apply Your First CRD

**`crontask-crd.yaml`:**

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crontasks.batch.mycompany.com
spec:
  group: batch.mycompany.com
  names:
    kind: CronTask
    plural: crontasks
    singular: crontask
  scope: Namespaced
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
              required: ["schedule", "command"]
              properties:
                schedule:
                  type: string
                command:
                  type: string
                retries:
                  type: integer
                  default: 3
                  minimum: 0
                  maximum: 10
```

**`cleanup-job.yaml`:**

```yaml
apiVersion: batch.mycompany.com/v1
kind: CronTask
metadata:
  name: cleanup-job
spec:
  schedule: "0 * * * *"
  command: "cleanup.sh"
```

**Commands:**

```bash
kubectl apply -f crontask-crd.yaml
kubectl apply -f cleanup-job.yaml
kubectl get crontasks
kubectl get crontask cleanup-job -o yaml
```

Note `retries` will show up as `3` automatically even though we didn't set it — that's the `default` kicking in.

**Bonus check:**

```bash
kubectl patch crontask cleanup-job --type merge -p '{"spec":{"retries":99}}'
# rejected: spec.retries in body should be less than or equal to 10
```

---

## 7. Demo: Status, Scale, Printer Columns, Short Names

Four polish features that make a hand-written CRD behave like a real, first-class Kubernetes type.

### 7.1 — The `status` subresource

By default, `status` is just another field inside your object — anyone with write access to the object can overwrite it, including end users, and a single `kubectl apply` touches both `spec` and `status` together. Enabling the **status subresource** splits this:

```yaml
versions:
  - name: v1
    served: true
    storage: true
    subresources:
      status: {}          # <-- enables it
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              replicas:
                type: integer
          status:
            type: object
            properties:
              readyReplicas:
                type: integer
              phase:
                type: string
```

Once enabled:
- `kubectl apply -f my-obj.yaml` (editing `spec`) **no longer touches `status`**, and vice versa.
- Regular users get RBAC'd out of `/status` separately — you can grant `update` on the main resource but not on the `status` subresource, so only your controller (using its own ServiceAccount) can write status.
- You update it via a dedicated endpoint: `PUT /apis/<group>/<version>/.../<name>/status`. In client-go this is `UpdateStatus()`; via `kubectl` you'd typically only touch this through automation, not by hand.

This is the mechanism that enforces the convention: **users own `spec`, controllers own `status`.**

### 7.2 — The `scale` subresource

Lets `kubectl scale` (and the Horizontal Pod Autoscaler) work against your custom resource, the same way it does against a Deployment.

```yaml
subresources:
  status: {}
  scale:
    specReplicasPath: .spec.replicas
    statusReplicasPath: .status.readyReplicas
    labelSelectorPath: .status.labelSelector   # optional, needed for HPA
```

With this enabled:

```bash
kubectl scale website my-blog --replicas=5
```

...will patch `spec.replicas` to `5` directly, without you needing to know the field name — and tools that expect a generic `Scale` object (like HPA) can now target your CR.

### 7.3 — `additionalPrinterColumns`

By default, `kubectl get mycrd` shows just `NAME` and `AGE` — not very useful. Printer columns pull specific fields (via JSONPath) into the table output.

```yaml
versions:
  - name: v1
    ...
    additionalPrinterColumns:
      - name: Domain
        type: string
        jsonPath: .spec.domain
      - name: Replicas
        type: integer
        jsonPath: .spec.replicas
      - name: Phase
        type: string
        jsonPath: .status.phase
      - name: Age
        type: date
        jsonPath: .metadata.creationTimestamp
```

Now:

```bash
kubectl get websites
NAME      DOMAIN        REPLICAS   PHASE     AGE
my-blog   myblog.com    2          Running   3m
```

This is purely cosmetic/informational but makes a hand-rolled CRD feel like a real built-in resource at a glance — you can debug state without `-o yaml` every time.

### 7.4 — `shortNames`

Small quality-of-life addition under `spec.names`:

```yaml
names:
  kind: Website
  plural: websites
  singular: website
  shortNames:
    - ws
```

```bash
kubectl get ws          # same as kubectl get websites
kubectl describe ws my-blog
```

### Full CRD with all four features combined

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: websites.stable.example.com
spec:
  group: stable.example.com
  names:
    kind: Website
    plural: websites
    singular: website
    shortNames: ["ws"]
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      subresources:
        status: {}
        scale:
          specReplicasPath: .spec.replicas
          statusReplicasPath: .status.readyReplicas
      additionalPrinterColumns:
        - name: Domain
          type: string
          jsonPath: .spec.domain
        - name: Replicas
          type: integer
          jsonPath: .spec.replicas
        - name: Phase
          type: string
          jsonPath: .status.phase
      schema:
        openAPIV3Schema:
          type: object
          required: ["spec"]
          properties:
            spec:
              type: object
              required: ["domain", "replicas"]
              properties:
                domain:
                  type: string
                replicas:
                  type: integer
                  minimum: 1
                  default: 1
            status:
              type: object
              properties:
                readyReplicas:
                  type: integer
                phase:
                  type: string
```

---

## 8. Lab: Apply a CR and Explore

**Try this yourself before reading the solution below.**

Using the combined `Website` CRD from section 7.4:

1. Apply the CRD.
2. Create a `Website` CR named `shop-site` with `domain: shop.example.com`, `replicas: 3`.
3. Confirm `kubectl get ws` shows your custom `DOMAIN`/`REPLICAS`/`PHASE` columns.
4. Manually patch `status.phase` to `"Running"` and `status.readyReplicas` to `3` using the status subresource endpoint (`kubectl patch ... --subresource=status`).
5. Try `kubectl scale ws shop-site --replicas=5` and confirm `spec.replicas` updates.
6. Confirm that patching `status` did **not** touch `spec`, and scaling did **not** touch `status` — they're independent.

---

## 9. Lab Solution: Apply a CR and Explore

```bash
kubectl apply -f website-crd.yaml
```

**`shop-site.yaml`:**

```yaml
apiVersion: stable.example.com/v1
kind: Website
metadata:
  name: shop-site
spec:
  domain: shop.example.com
  replicas: 3
```

```bash
kubectl apply -f shop-site.yaml
kubectl get ws
```

```
NAME        DOMAIN              REPLICAS   PHASE
shop-site   shop.example.com    3
```

(`PHASE` is blank — `status` hasn't been set yet, which is expected; only a controller normally sets it.)

**Patch status directly (requires `--subresource=status`, available in `kubectl` 1.24+):**

```bash
kubectl patch website shop-site --subresource=status --type=merge \
  -p '{"status":{"phase":"Running","readyReplicas":3}}'

kubectl get ws
```

```
NAME        DOMAIN              REPLICAS   PHASE
shop-site   shop.example.com    3          Running
```

**Scale it:**

```bash
kubectl scale ws shop-site --replicas=5
kubectl get website shop-site -o jsonpath='{.spec.replicas}'
# 5
kubectl get website shop-site -o jsonpath='{.status.readyReplicas}'
# 3   <- unchanged! status is independent of spec, exactly as intended
```

**Takeaway:** the split you just observed — `spec` changed by `kubectl scale`, `status` untouched, and vice versa when patching `/status` — is precisely why the status subresource exists. In a real Operator, your controller reconciles the gap between the two: it watches `spec.replicas` change to `5`, does the real work, and only then updates `status.readyReplicas` to match.

---

## What's Next

You now have everything the "CRD fundamentals" stage covers: types, validation, status/scale subresources, printer columns, short names, and hands-on CR creation. The natural next step — which this tutorial deliberately stops short of — is writing the **controller** that watches these CRs and actually reconciles `status` toward `spec`, which is where the Operator pattern proper begins.