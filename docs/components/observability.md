# Observability

The observability stack runs in its own `observability` namespace in production. It's defined in [`gitops/observability`](https://github.com/foyez/dbaas-platform/tree/main/gitops/observability) as pinned Helm charts rendered by Kustomize.

| Piece | Chart | Role |
|---|---|---|
| Prometheus | `kube-prometheus-stack` | Metrics, scraped from the API's `/metrics` through a `ServiceMonitor` |
| Loki | `loki` | Log storage, 7 days of retention |
| Grafana Alloy | `alloy` | Ships container logs from every node to Loki |
| Grafana | `grafana` | Dashboards, including an API overview dashboard |

## What the API emits

- **Metrics:** request counts and latencies per route, from a Gin middleware.
- **Request logs:** one structured JSON line per request with method, route, status, duration and client IP.
- **Audit events:** a separate `component=audit` log stream with who did what to which instance. The admin-only `GET /api/v1/audit-logs` endpoint reads them back from Loki. [Design decision 7](../design-decisions/audit-in-loki.md) explains why audit lives in logs instead of a database.

Instance logs shown in the UI come from the same place: the API queries Loki for the PostgreSQL pods of that instance.

## Locally

The local `make dev` setup doesn't install this stack, to keep it light. In production, `scripts/observability-forward-start.sh` port-forwards Grafana and Prometheus to your machine.
