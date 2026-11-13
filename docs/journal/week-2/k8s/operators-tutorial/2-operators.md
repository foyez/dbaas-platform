# Kubernetes Operators

*(Covers: what an Operator is → controllers vs operators → encoding operational knowledge → anatomy → real-world examples → Operator vs Helm → capability levels → cert-manager demo)*

---

## 1. What Is a Kubernetes Operator?

An **Operator** is a piece of software that extends Kubernetes by using **custom resources to manage applications and their components**, following Kubernetes' own principles — most importantly, the **control loop**.

The idea comes directly from the job of a human operator: someone who knows a specific piece of software deeply — how to install it, how to upgrade it safely, how to back it up, how to recover it when it breaks. The Operator pattern tries to **encode that expertise as code**, running continuously inside the cluster, instead of living in a person's head or a runbook.

Mechanically, an Operator is just:

```
Custom Resource(s)  +  Controller(s)
```

- The **CR** is the interface — a declarative "what I want" object (e.g. `SampleDB`) that a human or another system edits.
- The **Controller** is a program, running as a normal Pod in the cluster, that watches those CRs and drives real infrastructure to match.

Nothing here uses special Kubernetes internals — an Operator is written using the same API that `kubectl` itself uses. It authenticates as a client, watches/lists/gets objects, and issues Create/Update/Delete calls just like any other API consumer. This is why Operators can be written in *any* language with an HTTP client, though most use one of the dedicated SDKs (covered in section 5/8).

**What makes it different from "just automation scripts"?** Three properties:

1. **Level-triggered, not one-shot.** It runs forever, continuously reconciling — not "run once and exit."
2. **Declarative interface.** Users state desired end state (`replicas: 3`), not a sequence of steps.
3. **Self-healing.** If reality drifts from the CR's spec — someone deletes something by hand, a node dies — the next reconcile loop notices and fixes it, with no human involved.

**A minimal mental example**, extending the `Website` CRD from Part 1: a `Website` Operator wouldn't just create a Deployment once. It would continuously watch every `Website` object and make sure a matching Deployment, Service, and Ingress exist and match `spec` — forever, automatically, for as long as the cluster runs.

---

## 2. Controllers vs Operators — The Key Distinction

These two terms get used almost interchangeably, but there's a precise distinction worth locking in:

| | **Controller** | **Operator** |
|---|---|---|
| Definition | Any control loop that watches state via the API server and drives it toward desired state | A controller **specifically for a Custom Resource**, that encodes domain-specific application knowledge |
| Built-in examples | Deployment controller, ReplicaSet controller, Node controller — all ship inside `kube-controller-manager` | None — by definition, Operators manage *custom* resources, not built-ins |
| Knowledge encoded | Generic Kubernetes object lifecycle (e.g. "keep N replicas of a Pod template") | Application-specific operational knowledge (e.g. "how to safely fail over a Postgres primary") |
| Where it runs | Often compiled into `kube-controller-manager` (control plane) | Runs as an ordinary workload in the cluster, e.g. a Deployment — outside the control plane |

**The one-line version:** *every Operator is a controller, but not every controller is an Operator.* "Controller" is the general mechanism (the reconcile loop). "Operator" specifically means that mechanism applied to a custom resource, packaged with domain-specific logic, and typically deployed as a standalone application rather than shipped inside Kubernetes itself.

You'll also sometimes hear "Operator = CRD + Controller, plus enough operational logic to replace a human operator's runbook." That last clause is really the differentiator — a trivial controller that just copies a field from `spec` to `status` technically fits the mechanical definition, but nobody would call it an Operator. The term implies it's doing something a skilled human would otherwise do by hand.

---

## 3. The Operator Pattern — Encoding Operational Knowledge

This is the *why*, not the *how*. The operator pattern aims to capture the key aim of a human operator who is managing a service — someone with deep knowledge of how the system ought to behave, how to deploy it, and how to react to problems.

Concretely, "operational knowledge" usually covers things like:

- **Day-1 knowledge (deploying):** what order to create resources in, what config is mandatory, sane defaults.
- **Day-2 knowledge (running it):** how to detect an unhealthy replica, how to safely restart without downtime, how to reshard/rebalance.
- **Failure knowledge:** what a real outage looks like, and the exact sequence of steps a human would run to recover — encoded as code your Operator runs automatically.

The point isn't just "automate deployment" (a Helm chart already does that reasonably well — see section 6). The point is **ongoing, continuous operational logic** that reacts to the live state of a running system, not just a one-time install. A Deployment YAML says nothing about how to take a backup before an upgrade, or how to detect and fix split-brain in a database cluster — that operational knowledge has to live somewhere, and the Operator pattern's answer is: in a controller, as code, running continuously.

---

## 4. Anatomy of an Operator: CRD + Controller

Breaking down exactly what ships when you install an Operator, using the canonical `SampleDB` example:

```
┌─────────────────────────────────────────────┐
│  1. CustomResourceDefinition (SampleDB)     │  ← registers the type
├─────────────────────────────────────────────┤
│  2. Controller container image              │  ← your operator's logic
├─────────────────────────────────────────────┤
│  3. Deployment running that image           │  ← keeps the controller Pod alive
├─────────────────────────────────────────────┤
│  4. RBAC: ServiceAccount + Role + Binding   │  ← permissions to watch/edit resources
└─────────────────────────────────────────────┘
```

Piece by piece:

1. **A CRD** (e.g. `SampleDB`) — the schema for the resource users will actually create. This is exactly what Part 1 covered in depth.
2. **A Deployment** that ensures a Pod is running — this Pod *is* the controller. Note the controller is deployed the same way any application is; it does not run inside the control plane.
3. **A container image** holding the controller's code: written with client-go, Kopf, kubebuilder, etc. (see section 8), containing the watch/reconcile loop.
4. **RBAC objects** — the controller's ServiceAccount needs permission to `watch`/`list`/`get` `SampleDB` objects, and typically to create/manage the child resources it's responsible for (Pods, PVCs, StatefulSets, etc). Without correct RBAC, the controller's `watch` will fail silently or its writes will be `Forbidden`.

**What the controller code actually does**, in order:
1. Queries the API server / watches for `SampleDB` resources.
2. For each one, compares desired (`spec`) vs actual (real cluster resources) state.
3. Issues the Create/Update/Delete calls needed to close the gap — e.g. creating `PersistentVolumeClaims`, a `StatefulSet`, and a `Job` for initial setup when a new `SampleDB` appears.
4. On deletion, it may run cleanup logic first (e.g. take a final snapshot) before letting the StatefulSet and Volumes be removed — this is typically implemented with a **finalizer** on the CR.
5. Writes results back to `.status` (using the status subresource from Part 1) so users can see what happened.

This diagram is the same shape for essentially every Operator you'll encounter — only the specific child resources and reconciliation logic change per application.

---

## 5. Real-World Operator Examples

Concrete, well-known Operators and what operational knowledge each one encodes:

| Operator | Manages | Encodes knowledge about |
|---|---|---|
| **cert-manager** | TLS certificates | ACME/Let's Encrypt issuance, automatic renewal before expiry, challenge solving (HTTP-01/DNS-01) |
| **Prometheus Operator** | Prometheus, Alertmanager | Config reload without downtime, service discovery via `ServiceMonitor` CRs, sharding large deployments |
| **etcd Operator / etcd-druid** | etcd clusters | Safe member add/remove, quorum-aware scaling, backup/restore, disaster recovery |
| **Postgres Operators** (e.g. Zalando, CloudNativePG) | PostgreSQL | Primary/replica failover, connection pooling, point-in-time recovery, safe major-version upgrades |
| **Strimzi / Confluent Operator** | Kafka | Broker rolling restarts without data loss, topic/partition rebalancing, ZooKeeper/KRaft coordination |
| **ArgoCD (via Application CRD)** | GitOps app delivery | Continuous drift detection between Git state and cluster state, automated or manual sync |

The pattern across all of them: each domain has failure modes and procedures too specific and too frequent to leave as a manual runbook, and too dynamic for a static Helm install. That gap is exactly what an Operator fills — this is the general shape every example above is automating what a specialized human operator would otherwise be paged to do at 3am.

The other automation categories mentioned in the Kubernetes docs are worth calling out explicitly, since they show the *breadth* of what "operator" can mean beyond just databases:
- Publishing a Service to non-Kubernetes-aware clients.
- Simulating failure (chaos engineering) to test resilience.
- Leader election for apps with no built-in election mechanism.

---

## 6. When to Build an Operator vs Use Helm

This is one of the most common real-world design questions, so it's worth a clean decision framework rather than vague guidance.

**Helm** is a *templating and release-management* tool: it renders YAML from templates + values, and installs/upgrades/rolls-back that rendered YAML as a unit. It has **no ongoing runtime component** — once `helm install` finishes, Helm is not watching anything.

**An Operator** is an *always-running controller* with live knowledge of application state.

| Question | Leans Helm | Leans Operator |
|---|---|---|
| Is install/config templating the whole problem? | ✅ | |
| Does the app need day-2 logic — failover, backup, self-healing, upgrades that touch data? | | ✅ |
| Do you need the system to react automatically to runtime events (pod crash, primary down)? | | ✅ |
| Is "reinstall/upgrade the chart" an acceptable response to change? | ✅ | |
| Do non-experts need a safe, narrow interface (`kubectl edit mydb`) instead of raw K8s objects? | | ✅ |
| Is the engineering cost of writing+maintaining a controller justified by usage/complexity? | | ✅ |

**Rule of thumb:** if your "operational knowledge" fits entirely in *what gets installed*, use Helm. If it includes *what happens after, continuously, in response to events*, you need a controller — i.e., an Operator. Many real systems actually use **both**: a Helm chart to install the Operator itself (its CRD + Deployment + RBAC), and the Operator then manages the actual application instances via CRs. cert-manager, used in the demo below, is a good example of exactly this — it's Helm-installed, but is itself an Operator once running.

Also worth noting: Operators cost more to build and maintain — you're writing and testing real software with its own bugs, versioning, and upgrade path — whereas a Helm chart is comparatively cheap to write and reason about. Don't reach for an Operator just because it's the more sophisticated-sounding option.

---

## 7. Operator Capability Levels

A widely used maturity model (popularized by the Operator Framework / OLM) for describing **how much operational knowledge an Operator actually encodes**, from basic installation up to fully autonomous operation. Five levels:

```
Level 1  Basic Install
Level 2  Seamless Upgrades
Level 3  Full Lifecycle
Level 4  Deep Insights
Level 5  Auto Pilot
```

**Level 1 — Basic Install**
Automated application provisioning and configuration management. The Operator can install the application and set config via a CR. Roughly equivalent to what a decent Helm chart already gives you — this is the floor, not really "operator-grade" behavior yet.

**Level 2 — Seamless Upgrades**
Adds the ability to patch and minor-version upgrade the application in-place, handled by the Operator, without manual intervention or downtime beyond what's unavoidable.

**Level 3 — Full Lifecycle**
Adds application lifecycle beyond just upgrades: backup, restore, and handling more complex operations like failure recovery or storage resizing. This is roughly the point where an Operator starts replacing meaningful chunks of a human runbook — matching the `SampleDB` backup/restore/upgrade example from section 4.

**Level 4 — Deep Insights**
Adds observability: the Operator exposes metrics, alerts, log processing, and workload analysis — enough to detect problems (not just fix the ones it already knows about) and expose them for humans or automation to act on.

**Level 5 — Auto Pilot**
Full autonomous operation: horizontal/vertical auto-scaling, auto-healing of complex, multi-component failures, auto-tuning, and abnormality detection — with no human intervention. Very few real Operators reach this level; it's the aspirational top of the model.

**Why this model matters practically:** it gives you a shared vocabulary for scoping "how much Operator do we actually need to build." Many production Operators deliberately stop at Level 2 or 3 — going further has steeply increasing engineering cost, and Level 4/5 behavior (auto-healing, auto-tuning) is genuinely hard to get right and dangerous to get wrong in a stateful system like a database.

---

## 8. Demo: Deploy cert-manager in 60 Seconds

A real, production-grade Operator you can install and see reconciling immediately — no writing code required. cert-manager manages the full lifecycle of TLS certificates in-cluster: issuing, renewing, and rotating them automatically via `Certificate` and `Issuer`/`ClusterIssuer` custom resources.

### Install (Helm — installing the Operator itself)

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update

helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true
```

That's it for the install — Helm here is just deploying the Operator (its CRDs, Deployment, RBAC), exactly matching the "both together" point from section 6.

### Verify the Operator is running

```bash
kubectl get pods -n cert-manager
```

```
NAME                                       READY   STATUS
cert-manager-xxxxxxxxx-xxxxx               1/1     Running
cert-manager-cainjector-xxxxxxxxx-xxxxx    1/1     Running
cert-manager-webhook-xxxxxxxxx-xxxxx       1/1     Running
```

Three controllers ship together: the core cert-manager controller, `cainjector` (injects CA bundles into webhooks/APIServices), and a validating/mutating `webhook`.

### Check the CRDs it registered

```bash
kubectl get crds | grep cert-manager
```

```
certificates.cert-manager.io
certificaterequests.cert-manager.io
clusterissuers.cert-manager.io
issuers.cert-manager.io
challenges.acme.cert-manager.io
orders.acme.cert-manager.io
```

— exactly the "CRD + Controller" anatomy from section 4, at production scale: 6 CRDs, one Operator.

### Create a self-signed Issuer and Certificate to prove it reconciles

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: demo-cert
  namespace: default
spec:
  secretName: demo-cert-tls
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
  dnsNames:
    - demo.example.com
```

```bash
kubectl apply -f issuer-and-cert.yaml
kubectl get certificate demo-cert
```

```
NAME        READY   SECRET          AGE
demo-cert   True    demo-cert-tls   5s
```

`READY: True` means the Operator's control loop detected the new `Certificate` CR, generated a key pair, requested (and in this self-signed case, immediately fulfilled) the certificate, and wrote the resulting TLS secret — all without you touching `openssl` or a cert authority by hand. This is the full loop from Part 1 and section 1 playing out live: CR created → controller reconciles → status updated → real artifact (a Secret) produced.

```bash
kubectl get secret demo-cert-tls -o yaml
```
confirms the actual `tls.crt` / `tls.key` were generated and stored.

---

## What's Next

You now have the full conceptual + hands-on picture: CRDs and validation (Part 1), and what an Operator is, how it differs from a plain controller, why it exists, its typical anatomy, real examples, when to reach for one over Helm, how to gauge its maturity, and a live example reconciling in front of you. The natural next step from here is **writing your own controller's reconcile loop** — typically with a framework like Kubebuilder, Operator SDK, or Kopf — which is where the CRD (Part 1) and the control-loop code (this part) actually get wired together.