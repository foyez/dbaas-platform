# Getting started

One command brings up the whole platform on a local [kind](https://kind.sigs.k8s.io/) cluster, including its own ZITADEL instance, so you don't need any cloud account.

## Requirements

- Docker (Docker Desktop, OrbStack or Colima) with about 6 GB of memory
- `make`, `curl`, `unzip` and `openssl`

kind, kubectl, Helm and Terraform are downloaded at pinned versions into `./bin`, so nothing is installed globally.

## Start the platform

```sh
git clone https://github.com/foyez/dbaas-platform.git
cd dbaas-platform
make dev
```

The first run takes about 10 minutes, most of it building images and initialising ZITADEL. Re-running is safe and only applies changes. It:

1. creates the `dbaas-dev` kind cluster with ingress-nginx on port 80
2. installs CloudNativePG and a local ZITADEL backed by it
3. applies [`infra/zitadel/tf`](https://github.com/foyez/dbaas-platform/tree/main/infra/zitadel/tf) to create the project, roles and apps (the same Terraform used for production)
4. builds the operator, API and frontend images and deploys them through the `gitops/*/overlays/dev` overlays

When it prints **Platform is up.**, these are available:

| What | URL |
|---|---|
| App | <http://app.dbaas.localhost> |
| API | <http://api.dbaas.localhost/api/v1> |
| API docs (Swagger UI) | <http://api.dbaas.localhost/docs> |
| ZITADEL console | <http://auth.dbaas.localhost/ui/console> |
| Grafana (metrics, logs, traces) | <http://grafana.dbaas.localhost>, `admin` / `admin` |

!!! note "Browser support"
    The `*.localhost` hostnames resolve to your machine without any `/etc/hosts` changes in Chrome and Firefox. Safari may not resolve them.

## Try it out

1. Open <http://app.dbaas.localhost> and log in with `admin@dbaas.auth.dbaas.localhost` / `Password1!`. This user has the `dbaas.admin` role.
2. Create an instance and wait until its status is **Running**.
3. Run a query in the instance's SQL console.
4. Connect from your own client, such as `psql` or TablePlus, with the credentials shown in the UI. The host is `127.0.0.1` and the port is the one allocated to the instance, starting at 30101. Port 30100 is taken by the built-in sample database.

To see the Kubernetes side, point kubectl at the cluster:

```sh
kubectl --context kind-dbaas-dev get postgresql -A
```

## See what's happening inside

Open Grafana at <http://grafana.dbaas.localhost> and log in with `admin` / `admin`:

- **Dashboards → DBaaS API Overview** for request rate, errors and latency.
- **Explore → Loki**, query `{namespace="dbaas-api"} | json`, and click **TraceID** on a log line to open its trace.
- **Explore → Tempo**, search for service `dbaas-api` to see each request's spans: token introspection, Kubernetes API calls and SQL console statements.

The observability stack needs about 1 GB of memory. To skip it, run `OBSERVABILITY=false make dev`; the instance Logs tab and the audit log then have no data.

## Everyday commands

| Command | What it does |
|---|---|
| `make dev` | Create or update the local platform |
| `make dev-status` | Show pods, ingresses and PostgreSQL instances |
| `make dev-down` | Delete the cluster and its local state |
| `make docs` | Preview this documentation at <http://localhost:8000/dbaas-platform/> |
| `make help` | List every target |

## How it differs from production

| | Local (`make dev`) | Production |
|---|---|---|
| Cluster | kind | STACKIT SKE |
| HTTPS | No, plain HTTP on `*.dbaas.localhost` | ingress-nginx + Let's Encrypt |
| Identity | Self-hosted ZITADEL in the cluster | ZITADEL Cloud |
| Database access | `127.0.0.1:<port>`, ports 30100–30119 | Shared LoadBalancer, ports 30100–30999 |
| Delivery | Images built and loaded by `make dev` | Forgejo CI + Argo CD |
| Observability | Same stack: Prometheus, Loki, Tempo, Alloy, Grafana | Same |
