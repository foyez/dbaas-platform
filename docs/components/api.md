# REST API

The API (`platform/`) is a Go service built with [Gin](https://gin-gonic.com/). It's the only thing the UI and SDK talk to: it authenticates the caller, checks ownership and translates requests into Kubernetes operations on `PostgreSQL` resources. The full contract is in [`platform/docs/openapi.yaml`](https://github.com/foyez/dbaas-platform/blob/main/platform/docs/openapi.yaml): browse it in the [API reference](../api-reference.md), or at `/docs` on a running API (Swagger UI, with "Try it out").

## Endpoints

All routes live under `/api/v1` and require a valid ZITADEL access token.

| Method | Path | What it does |
|---|---|---|
| `POST` | `/instances` | Create an instance (requires an `Idempotency-Key` header) |
| `GET` | `/instances` | List the caller's instances (cursor pagination) |
| `GET` | `/instances/{id}` | Get one instance |
| `PATCH` | `/instances/{id}` | Change the PostgreSQL version or storage size |
| `DELETE` | `/instances/{id}` | Delete an instance |
| `GET` | `/instances/{id}/credentials` | Connection details and password |
| `GET` | `/instances/{id}/logs` | The instance's PostgreSQL logs, from Loki |
| `POST` | `/instances/{id}/query` | Run SQL through the console |
| `GET` | `/audit-logs` | Audit trail, `dbaas.admin` only |

`/health`, `/metrics` (Prometheus), `/docs` (Swagger UI) and `/openapi.yaml` are unauthenticated.

## Structure

The code follows a pragmatic clean architecture: dependencies point inwards and infrastructure sits behind interfaces.

| Package | Role |
|---|---|
| `api/handler`, `api/router` | HTTP: binding, validation, status codes |
| `service` | Use cases: ownership checks, idempotency, orchestration |
| `domain`, `ports` | Core types and the interfaces the service depends on |
| `infra/cnpg`, `infra/k8s` | Kubernetes client for `PostgreSQL` resources and their secrets |
| `infra/loki` | LogQL queries for instance logs and the audit trail |
| `infra/pgquery` | Connection pools for the SQL console |
| `auth` | ZITADEL token introspection and role checks |
| `observability` | Request logging, metrics and audit middleware |

## Behaviour worth knowing

- **Ownership** is a label on each resource (`app.example.com/user-id`). Every read and write filters on it, so users only ever see their own instances. See [design decision 1](../design-decisions/kubernetes-as-state-store.md).
- **Idempotent create.** The resource name is derived from `sha256(userID + Idempotency-Key)`, so retries can't create duplicates. See [design decision 4](../design-decisions/idempotent-creation.md).
- **Credentials** point at the public database host and the instance's allocated port when one exists, and otherwise at the in-cluster service.
- **SQL console limits.** Each query runs for at most 10 seconds and returns at most 500 rows, with a small per-instance connection pool. See [design decision 5](../design-decisions/sql-console-through-api.md).
- **SQL errors reach the user.** When PostgreSQL rejects a statement, the API returns `422 QUERY_FAILED` with PostgreSQL's message, SQLSTATE code, detail, hint and error position, and the console shows them like `psql` does. A statement that runs past the limit returns `422 QUERY_TIMEOUT`. The API logs only the SQLSTATE, since messages can contain row values.

## Configuration

| Variable | Purpose |
|---|---|
| `APP_ENV` | `development` or `production` |
| `SERVER_ALLOW_ORIGINS` | CORS origins, i.e. the UI's URL |
| `ZITADEL_DOMAIN` | ZITADEL instance used for introspection |
| `ZITADEL_KEY_PATH` | Path to the API application's key file |
| `ZITADEL_INSECURE_PORT` | Plain-HTTP ZITADEL port, for the local setup only |
| `PUBLIC_DB_HOST` | Host returned in credentials for external connections |
| `LOKI_URL` | Loki endpoint for logs and audit queries |

## Develop and test

```sh
cd platform
make test
```
