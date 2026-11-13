# DBaaS Platform

A self-service **Database-as-a-Service** platform: developers create, scale and query their own PostgreSQL databases through a web UI, REST API or Go SDK. It was built end to end, from raw IaaS on OpenStack (Terraform + Ansible) to a product deployed on STACKIT Kubernetes Engine (SKE).

```mermaid
flowchart LR
    User(["Client"]) --> UI["Web UI<br/>(Vue)"]
    UI --> API["API<br/>(Go)"]
    UI --> Auth["ZITADEL<br/>(Auth)"]
    API --> Auth

    API --> Obs["Grafana<br/>(Logs & Metrics)"]
    API --> Operator["dbaas-operator<br/>(PostgreSQL CRD)"]

    Operator --> CNPG["CloudNativePG"]
    CNPG --> PG[("PostgreSQL")]
```

The production cluster has been shut down to save costs. Everything can be recreated from the [repository](https://github.com/foyez/dbaas-platform), and the whole platform runs locally with one command.

New here? [Run it locally](getting-started.md) in about 10 minutes, then read the [architecture](architecture/index.md) and the [design decisions](design-decisions/index.md) behind it. The [learning notes](notes/index.md) cover what I learned building it, from a bare OpenStack install to the finished product, with review questions on every page.

## Features

- **Custom Kubernetes Operator**: A `PostgreSQL` CRD (built with Kubebuilder) that reconciles into a [CloudNativePG](https://cloudnative-pg.io/) `Cluster`, exposing a simplified spec (`version`, `instances`, `storage`, `database`, `username`) and a status (`phase`, `readyInstances`, `externalPort`, conditions) so the API never has to speak CNPG directly.
- **REST API**: Go/Gin service for create / list / get / update / delete of instances, plus credentials, log retrieval, and ad-hoc SQL execution — fully documented with OpenAPI. Idempotent creation via `Idempotency-Key`, per-user isolation via Kubernetes labels, cursor-based pagination on list endpoints (matches the underlying Kubernetes/Loki APIs, which have no offset/skip concept).
- **External database connectivity**: each instance's CNPG cluster is `ClusterIP`-only by default (unreachable outside the cluster) — the operator allocates a port from a shared pool and wires it into ingress-nginx's TCP passthrough, so instances get a real externally-reachable connection string through one shared LoadBalancer instead of one per tenant.
- **In-browser SQL console**: run queries directly against an instance from the UI — proxied through the API with a pooled connection, a per-query timeout, and a row cap, so one query can't take down the API pod.
- **AuthN/AuthZ**: OIDC via [ZITADEL](https://zitadel.com/), role-based access control (`dbaas.user`, `dbaas.admin`), admin-only audit log endpoint.
- **Web UI**: Vue 3 + TypeScript SPA for the full instance lifecycle — create, monitor status (live, via polling — no manual refresh), view logs/credentials, run queries, update, delete. Public landing page and documentation site in front of the authenticated app.
- **GitOps delivery**: ArgoCD-driven deployment of the operator, API, UI, CNPG, and the observability stack via Kustomize overlays.
- **Observability**: Prometheus + Grafana for metrics, Loki + Grafana Alloy for logs, structured JSON application logs, and a per-instance/per-user audit trail queryable through the API.
- **Go SDK**: A typed client library (`sdk/`) with OAuth2 client-credentials auth, wrapping the full instance API for programmatic use.
- **Infrastructure as Code**: Terraform + Ansible provisioning of the underlying Kubernetes cluster — from a single/multi-node cluster on self-hosted OpenStack (DevStack) up to a managed [STACKIT SKE](https://www.stackit.de/en/product/stackit-kubernetes-engine/) cluster.
