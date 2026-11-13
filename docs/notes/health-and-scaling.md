# Health checks, autoscaling and load testing

!!! abstract "Where it's used in this project"
    The API's `/health` endpoint and probes in [`gitops/api/base/deployment.yaml`](https://github.com/foyez/dbaas-platform/blob/main/gitops/api/base/deployment.yaml), its HPA in [`gitops/api/base/hpa.yaml`](https://github.com/foyez/dbaas-platform/blob/main/gitops/api/base/hpa.yaml), and the k6 load test in [`platform/test/performance`](https://github.com/foyez/dbaas-platform/tree/main/platform/test/performance).

## Key ideas

### Three different health questions

| Probe | Question | On failure |
|---|---|---|
| Liveness | Should this process be restarted? | Container is killed and restarted |
| Readiness | Should this instance get traffic right now? | Pod is removed from Service endpoints |
| Startup | Has initialisation finished? | Liveness and readiness wait until it passes |

Rules that follow from those questions:

- **Liveness must not check dependencies.** If the database is down, restarting every API pod doesn't fix it; it just adds a restart storm.
- **Readiness can check what this instance needs to serve**, and returns `503` when it can't.
- **Health checks are cheap, fast and side-effect free**, and shouldn't be logged at INFO on every call.
- **Health responses aren't diagnostics.** Return a status, not versions, hostnames or dependency details; keep those behind authentication.
- **During shutdown, become unready first**, then stop accepting connections, so the load balancer drains traffic before the pod exits.

### Horizontal Pod Autoscaler

The HPA adjusts a Deployment's replica count from metrics. CPU-based scaling compares usage against the pod's CPU **request**, so requests must be set, and Metrics Server must be installed (`kubectl top pods` should work).

Only the API scales horizontally. The operator runs a fixed number of replicas, because more controllers don't reconcile faster and would need leader election anyway.

### Load testing with k6

k6 scripts describe virtual users and **thresholds** (for example p95 latency under 500 ms, error rate under 1 %), which turn a test run into a pass or fail. Useful runs:

1. **Baseline:** low load, to learn normal latency.
2. **Ramp-up:** increase load until the HPA scales out, and check latency stays within thresholds.
3. **Scale-down:** drop the load and confirm replicas return to the minimum after the stabilisation window.
4. **Spike or stress:** jump past expected load to find the breaking point.

Watch k6 output and the cluster together: `kubectl get hpa -w`, `kubectl top pods`.

## How this project uses it

- The API runs 2–10 replicas, targeting 70 % average CPU, with requests of 250m CPU and 64Mi memory.
- `/health` serves both the liveness and readiness probes and returns `{"status":"ok"}`.

## What I learned building it

- **One endpoint for both probes is a shortcut.** It's fine while `/health` checks nothing external, but a proper readiness check (Kubernetes API reachable, ZITADEL reachable) would need its own endpoint.
- **Health endpoints shouldn't go through the public Ingress.** Probes call the pod directly, so `/health` doesn't need to be routable from the internet.
- **Define pass/fail before running a load test.** Without thresholds, a load test produces numbers but no answer.

## Review questions

??? question "Why is a database check in a liveness probe dangerous?"
    When the database has a problem, every pod fails liveness and gets restarted at once. That doesn't fix the database, drops in-flight requests, and can turn a partial outage into a total one.

??? question "What does an HPA need to scale on CPU?"
    Metrics Server, and CPU requests on the containers, because utilisation is a percentage of the request.

??? question "Why does a pod need to fail readiness before shutting down?"
    So it's removed from the Service endpoints and stops receiving new requests before the process exits. Otherwise requests routed to it during shutdown fail.

??? question "Why doesn't the operator get an HPA?"
    Reconcile throughput isn't CPU-bound per replica, and multiple active controllers would conflict. Extra replicas only help for availability, with leader election so only one is active.
