# Observability

!!! abstract "Where it's used in this project"
    Prometheus, Grafana, Loki and Alloy in [`gitops/observability`](https://github.com/foyez/dbaas-platform/tree/main/gitops/observability), the API's metrics and audit middleware in [`platform/internal/observability`](https://github.com/foyez/dbaas-platform/tree/main/platform/internal/observability). See [Observability](../components/observability.md).

## Key ideas

### Signals

| Signal | Answers | Here |
|---|---|---|
| **Metrics** | How much, how fast, how often? Numbers over time | Prometheus |
| **Logs** | What exactly happened? Timestamped records | Loki |
| **Traces** | Where did this request spend its time across services? | Not yet (OpenTelemetry + Tempo is the next step) |
| **Events** | Something happened at a point in time | Kubernetes events, audit events |

### Prometheus

Prometheus **pulls** (scrapes) `/metrics` from targets on an interval and stores time series: a metric name plus labels, with timestamped samples.

| Type | Behaviour | Example |
|---|---|---|
| Counter | Only goes up (resets on restart) | `requests_total` |
| Gauge | Goes up and down | memory in use, queue length |
| Histogram | Counts observations in buckets | request duration; gives percentiles |
| Summary | Client-side quantiles | rarely the right choice |

Counters are read with `rate()` (per second) or `increase()` (over a window), never raw. Percentiles come from histograms: `histogram_quantile(0.95, sum by (le) (rate(..._bucket[5m])))`.

**Cardinality** is the number of unique label combinations. Each one is a separate time series, so a label like `user_id` or a raw URL path can create millions and overload Prometheus. Labels must have small, bounded value sets.

### Loki and Alloy

Loki stores logs indexed **only by labels** (namespace, pod, app); the log content isn't indexed and is filtered at query time with LogQL. That keeps it cheap, but the same cardinality rule applies to labels. Grafana Alloy runs on every node, tails container logs and ships them to Loki with labels attached.

Structured JSON logs make LogQL filtering possible without making every field a label: `{namespace="dbaas-api"} | json | component="audit"`.

### SLI, SLO, SLA

- **SLI**: what you measure, e.g. the fraction of requests that succeed in under 500 ms.
- **SLO**: the target for it, e.g. 99.5 % over 30 days. The gap to 100 % is the error budget.
- **SLA**: a contract with consequences if the SLO is missed.

## How this project uses it

- **API metrics** are labelled by method, **route template** (`c.FullPath()`, e.g. `/api/v1/instances/:id`) and status, never by raw path or user.
- **Request logs** are structured JSON, one line per request.
- **Audit events** are log lines with `component=audit`, filtered at query time rather than made a Loki label, and read back by the admin audit endpoint. See [design decision 7](../design-decisions/audit-in-loki.md).
- **Instance logs** shown in the UI come from Loki by instance label.
- Retention is 7 days for both Prometheus and Loki.

## What I learned building it

- **Use the route template as a label, not the URL.** `/instances/instance-3f36…` would create a new series per instance; `/instances/:id` stays bounded.
- **Don't scrape yourself into noise.** The metrics middleware skips `/metrics`, and health checks shouldn't be logged at INFO.
- **Logs-as-audit has limits.** It reused the existing stack, but retention equals log retention and nothing is tamper-evident; a compliance-grade audit trail needs append-only storage.
- **Build it in phases.** Metrics and dashboards for the API first, then platform logs, then user-facing instance logs, then audit, each on top of the last.

## Review questions

??? question "Why is `user_id` a bad Prometheus label?"
    Every distinct value creates a new time series. With many users that's an unbounded number of series, which uses memory and slows queries. Put such detail in logs instead.

??? question "How do you get p95 latency from Prometheus?"
    Record a histogram, then `histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))`.

??? question "Why can't you read a counter's raw value to know the request rate?"
    It's cumulative since the process started and resets on restart. `rate()` computes the per-second increase over a window and handles resets.

??? question "What does Loki index, and what doesn't it?"
    It indexes only the label set of each log stream. Log lines themselves are stored compressed and scanned at query time, which is why label filters should narrow the search first.

??? question "SLO vs SLA?"
    An SLO is an internal reliability target. An SLA is an external commitment, usually looser than the SLO, with penalties if it's broken.
