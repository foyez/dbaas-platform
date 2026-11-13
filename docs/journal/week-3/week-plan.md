# Week 3 – Developing a Platform-as-a-Service (PaaS) Product: API Layer

**Stack:** Go + Gin · kubebuilder operator (already built) · CNPG under the hood · SKE · ArgoCD GitOps (already live)

---

## 0. Ground Rules Before You Start

- **Provisioning is delegated, not implemented.** Your API never talks to CNPG directly. It creates/patches a Custom Resource; your operator reconciles it. The API is a stateless façade over a control loop.
- **Async by default.** Cluster creation/deletion takes time. `POST` and `DELETE` return immediately with `202 Accepted`, not a final state.
- **No custom state machine.** `CR.status.phase` (plus CNPG's own status conditions) *is* your instance lifecycle. Don't build a second source of truth.
- **GitOps is done.** Just add the API's manifests to the repo ArgoCD already watches — no new pipeline.

---

## 1. Domain Model

Define `ProductInstance` as the resource your API exposes, and how it maps onto your CRD:

| API field | Source |
|---|---|
| `id`, `name`, `plan`, `region` | Request → CR `.spec` |
| `status` | CR `.status.phase` (+ CNPG conditions) |
| `connection` | CNPG-generated Service/endpoint |
| `credentials` | CNPG-generated Secret |
| `createdAt`, `owner` | CR metadata / annotations |

Sketch the relationship:

```
ProductInstance
      │
      ├── spec        → drives CR.spec
      ├── status       ← CR.status.phase / CNPG conditions
      ├── connection   ← CNPG Service
      └── credentials  ← CNPG Secret
```

---

## 2. REST API Design

**Endpoints:**

```
POST   /v1/instances
GET    /v1/instances
GET    /v1/instances/{id}
DELETE /v1/instances/{id}
GET    /v1/instances/{id}/connection
GET    /v1/instances/{id}/credentials
PATCH  /v1/instances/{id}          (bonus)
```

**HTTP semantics:**

| Action | Method | Response |
|---|---|---|
| Create | POST | `202 Accepted` (async — reconciliation pending) |
| List | GET | `200 OK` |
| Read | GET | `200 OK` |
| Update | PATCH | `200 OK` (or `202` if it triggers reconciliation) |
| Delete | DELETE | `202 Accepted` (finalizer cleanup is async) |

**Status codes to use deliberately:**

```
200 OK                — successful read/update
201 Created           — synchronous creation only (not your case)
202 Accepted          — async creation/deletion accepted
204 No Content         — sync delete only (not your case)
400 Bad Request        — malformed request
401 Unauthorized       — no/invalid credentials
403 Forbidden          — authenticated, not allowed
404 Not Found          — instance doesn't exist
409 Conflict           — e.g. instance already exists / delete in progress
410 Gone                — existed, now deleted (optional, nice touch)
422 Unprocessable       — semantically invalid (bad plan name, etc.)
500 Internal Server Error
```

**Naming & versioning:** plural nouns (`instances`, not `instanceList`), no verbs in paths, always version (`/v1/...`) even with one version live.

---

## 3. Lifecycle & Flowcharts

Lifecycle is just your CR's phase, surfaced through `GET`:

```
Client → POST (apply CR) → 202
                              │
                    Operator reconciles CR
                              │
        CR.status.phase: Pending → Provisioning → Running
                              │
             GET /instances/{id} reflects current phase
```

Draw three flowcharts, each showing the **API → CR → operator → CNPG** handoff explicitly — this is the part worth showing off:

1. **Create:** validate → build CR spec → `CreateOrUpdate` → 202 → (operator reconciles async)
2. **Delete:** find instance → delete CR → 202 → (operator/finalizer tears down CNPG cluster async)
3. **Get Connection:** find CR → check phase == Running → read Service/Secret → return endpoint

---

## 4. OpenAPI Specification

Write `openapi.yaml` before implementation. Include:

- `servers`, `tags`, `security` schemes
- Schemas: `CreateInstanceRequest`, `Instance`, `ConnectionInfo`, `Credentials`, `ErrorResponse`
- Explicit `202` responses with a `Location`/status-check header documented
- Examples for every request/response

---

## 5. Implementation (Gin, layered)

```
Router → Middleware → Handler → Service → controller-runtime Client → CR
```

Handlers never touch the Kubernetes client directly — always through the service layer, so tests can mock it.

- `POST /instances` → service builds a CR object, calls `client.Create` (or `CreateOrUpdate` for idempotency), returns 202 + instance ID + status URL.
- `GET /instances/{id}` → service fetches CR, maps `.status.phase` into the API response shape.
- `DELETE /instances/{id}` → service calls `client.Delete`, returns 202.

---

## 6. Validation & Error Handling

Validate before anything touches the service layer: required fields, enum values (plan, region), name format/length, UUID format.

Standard error envelope:

```json
{
  "error": {
    "code": "INSTANCE_NOT_FOUND",
    "message": "Product instance does not exist."
  }
}
```

---

## 7. Auth/Security Scaffolding

- Bearer token / API key middleware (even if simple for now — structure it so JWT/OAuth2 can slot in later)
- Never log credentials or raw secrets
- `Cache-Control: no-store` on `/credentials` and `/connection` responses
- Input size limits, CORS policy if applicable

---

## 8. Unit Tests

Mock the service layer (or the `controller-runtime` fake client) so tests exercise API behavior, not real Kubernetes.

```
POST    ✓ valid request  ✓ invalid request  ✓ duplicate (409)  ✓ validation (422)
GET     ✓ exists  ✓ missing (404)  ✓ unauthorized (401)
DELETE  ✓ success (202)  ✓ already deleted  ✓ not found
```

---

## 9. Docker

Multi-stage build → minimal runtime image (e.g. `distroless` or `scratch` for a static Go binary) → non-root user → only required port exposed → config via env vars → push to STACKIT Container Registry.

---

## 10. Kubernetes Deployment (SKE, via existing ArgoCD)

- `Deployment`, `Service`, `ConfigMap`/`Secret`
- Health probes: `/healthz`, `/readyz`
- Resource `requests`/`limits`
- Labels: `app`, `version`, `component`
- Manifests land in the same repo/path ArgoCD already watches — no new pipeline needed

---

## Bonus (only after core is solid)

```
11. HPA for the API deployment (CPU or custom metrics)
        ↓
12. Performance tests (k6 / JMeter) — verify HPA scaling, measure
    latency, throughput, error rate, pod count over time
        ↓
13. PATCH /v1/instances/{id} — idempotent full-field update
    (avoid increment-style semantics; same body twice = same end state)
```

---

## Deliverables Checklist

- [ ] Domain model mapped to CRD spec/status
- [ ] Versioned REST design, consistent naming
- [ ] 3 flowcharts (Create / Delete / Connection) showing API→CR→operator handoff
- [ ] Complete `openapi.yaml`
- [ ] Layered Gin implementation (router→middleware→handler→service→k8s client)
- [ ] Consistent validation & error responses
- [ ] Unit tests (mocked k8s client), covering success + failure paths
- [ ] Docker image in STACKIT Container Registry
- [ ] K8s manifests deployed to SKE via existing ArgoCD pipeline
- [ ] *(Bonus)* HPA config + performance test results
- [ ] *(Bonus)* Idempotent `PATCH /v1/instances/{id}`

---

## Appendix — REST Interview Review

### Statelessness
No session state on the server between requests. Auth travels with every request (JWT/API key). Workflow state lives in a resource (your CR's `.status`), not in server memory — this is exactly why an API in front of Kubernetes is naturally stateless. Any pod can handle any request; no sticky sessions needed.

**Trap:** "stateless API" doesn't mean "no state exists anywhere" — the CRs/database hold state; the *request-handling process* doesn't.

### Idempotency

Same request repeated N times → same end state as doing it once.

| Method | Safe | Idempotent |
|---|---|---|
| GET / HEAD / OPTIONS | Yes | Yes |
| PUT | No | Yes |
| DELETE | No | Yes |
| PATCH | No | Depends on semantics |
| POST | No | Not by default |

- **PUT** = replace resource at URL with this exact representation → same result every time.
- **POST** = create a new subordinate resource → repeated calls often create duplicates, unless engineered otherwise.
- **Idempotency-Key header** (Stripe pattern): client sends a UUID per logical operation; server caches key→result and replays it instead of duplicating work. Directly relevant to your `POST /instances` — a client retry after a timeout shouldn't spin up two clusters.
- **Your operator gives this to you almost for free**: `CreateOrUpdate` against a CR is idempotent by construction — same desired state in, same desired state out, no matter how many times it's applied. Good line for an interview.
- **PATCH nuance:** `{ "replicas": 3 }` is idempotent; `{ "increment_replicas_by": 1 }` is not.

### Safe vs. Idempotent
- **Safe** = no state change at all (GET, HEAD).
- **Idempotent** = may change state, but repeating doesn't change it further (PUT, DELETE).
- All safe methods are idempotent; not all idempotent methods are safe.

### Resource Orientation & HATEOAS
Resources are nouns; state transitions happen via representations, not RPC verbs. HATEOAS = responses include links telling the client what it can do next:

```json
{
  "id": "abc123",
  "status": "Provisioning",
  "_links": {
    "self": { "href": "/v1/instances/abc123" },
    "connection": { "href": "/v1/instances/abc123/connection" }
  }
}
```

Rarely fully implemented in practice — know it exists and why most APIs stop short (client coupling to link structure, SDKs hardcode paths anyway).

### Richardson Maturity Model
- **Level 0:** one endpoint, everything via POST (RPC-over-HTTP)
- **Level 1:** multiple resource URIs, one verb
- **Level 2:** proper verbs + status codes ← most real APIs, including yours, target this
- **Level 3:** HATEOAS

### Status Code Nuances
- **200 vs 201 vs 202** — 200: generic success with body. 201: sync creation, include `Location` header. 202: accepted, processing is async (your case).
- **401 vs 403** — 401: not authenticated. 403: authenticated, not authorized.
- **409 vs 422** — 409: conflicts with current state (already exists, delete in progress). 422: well-formed but semantically invalid.
- **404 vs 410** — 410 (Gone): existed, now deleted — distinct from never having existed.

### Caching
- `ETag` + `If-None-Match` for conditional GETs — useful if instance list is polled often.
- `Cache-Control: no-store` on anything returning secrets.

### Versioning Strategies
- **URI versioning** (`/v1/instances`) — simple, discoverable, what you're using. Downside: URL proliferation over time.
- **Header versioning** (`Accept: application/vnd.api+json;version=1`) — cleaner URLs, harder to debug/curl.
- Defensible answer: URI versioning chosen for discoverability and simplicity, consistent with how most PaaS control planes (Stripe, AWS-style) do it.