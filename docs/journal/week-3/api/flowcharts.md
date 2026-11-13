# API Flowcharts

## 1. Create Instance

### Goal

Ensure that retrying `POST /instances` (due to timeout, network failure, etc.) **never creates duplicate PostgreSQL clusters**, even when the API is running multiple replicas behind a load balancer.

---

### Design

Use the **Idempotency-Key** supplied by the client to derive a deterministic Kubernetes Custom Resource (CR) name.

```text
Idempotency-Key
        │
        ▼
SHA-256 Hash
        │
        ▼
instance-<hash>
        │
        ▼
PostgreSQL CR Name
```

Also store the original key as an annotation for debugging.

```yaml
metadata:
  name: instance-3b7a9c...
  annotations:
    paas.io/idempotency-key: "550e8400-e29b-41d4-a716-446655440000"
```

* **CR name** → Enforces idempotency.
* **Annotation** → Improves observability (logs, tracing, debugging).

---

### Request Flow

```text
Client
   │
POST /instances
Idempotency-Key: 550e8400...
   │
Validate request
   │
Hash Idempotency-Key
   │
Build PostgreSQL CR
(name = instance-<hash>)
   │
Create()
   │
 ┌──────────────┴──────────────┐
 │                             │
Created                  AlreadyExists
 │                             │
202 Accepted          Get existing CR
 │                             │
 └──────────────┬──────────────┘
                │
Return current resource state
                │
────────── Async ──────────
CNPG Operator reconciles CR
    │
Creates StatefulSet
PVCs
Services
Pods
   │
Cluster Running
```

---

### Why This Works

All API replicas communicate with the **same Kubernetes API server**.

Even if multiple replicas receive the same request simultaneously:

```text
Replica A
      │
      ├──── Create(instance-abc)
      │
Replica B
      │
      └──── Create(instance-abc)
                 │
                 ▼
         Kubernetes API Server
```

The Kubernetes API server guarantees object names are unique.

Exactly one `Create()` succeeds.

The others receive:

```text
AlreadyExists
```

They simply retrieve the existing CR and return its current state.

No distributed lock, shared cache, or external coordination is required.

---

### Benefits

* ✅ Naturally idempotent across retries
* ✅ Safe with multiple API replicas (HPA)
* ✅ No Redis or distributed locks
* ✅ Kubernetes API server provides atomicity
* ✅ Simple implementation (`Create()` + handle `AlreadyExists`)
* ✅ Annotation preserves the original idempotency key for debugging

---

### Interview Talking Point

> I used Kubernetes itself as the concurrency boundary. Instead of coordinating API replicas with Redis or distributed locks, I derived a deterministic CR name from the client's Idempotency-Key. Since the Kubernetes API server guarantees atomic object creation, exactly one request succeeds and concurrent retries receive `AlreadyExists`, making the API naturally idempotent and horizontally scalable. The original Idempotency-Key is also stored as a CR annotation for observability and request tracing.

This explanation demonstrates both an understanding of idempotency and how to leverage Kubernetes' consistency guarantees to simplify distributed system design.

## 2. Delete Instance

```
Client
   │
DELETE /instances/{id}
   │
Find PostgreSQL CR
   │
Delete CR
   │
202 Accepted
   │
────────── Async ──────────
CNPG Operator sees deletion
   │
Runs finalizers
   │
Deletes StatefulSet
Pods
Services
PVCs (if configured)
   │
CR removed
```

### Interview point:

Kubernetes deletion is asynchronous because finalizers allow operators to perform cleanup before the object is actually removed.

## 3. Get Connection

```
Client
   │
GET /instances/{id}/connection
   │
Find PostgreSQL CR
   │
Phase == Running?
   │
   ├── No → 409 / 404
   │
   └── Yes
          │
Read Service
Read Secret
          │
Return
Host
Port
Username
Password
```

## Same visual pattern for all diagrams:

* **Blue** = API layer
* **Orange** = Kubernetes API (Custom Resource)
* **Green** = CNPG Operator
* **Purple** = PostgreSQL resources (Pods, Services, PVCs)

That consistent color coding makes the API → CR → Operator → Kubernetes handoff immediately clear and highlights the reconciliation model, which is often what interviewers are most interested in.
