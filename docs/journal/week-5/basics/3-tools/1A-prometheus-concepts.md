# Prometheus concepts

Before starting, setup a small prometheus app to practice:

<details>
<summary>View setup</summary>

### 1. Create this project

```text
prometheus-demo/
├── docker-compose.yml
└── prometheus.yml
```

### 2. `prometheus.yml`

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: prometheus
    static_configs:
      - targets: ["prometheus:9090"]

  - job_name: node_exporter
    static_configs:
      - targets: ["node-exporter:9100"]

  # Add your application targets here
  - job_name: 'my-app'
    scrape_interval: 30s
    static_configs:
      - targets: ['my-app:8080']
    metrics_path: /metrics
```

### 3. `docker-compose.yml`

```yaml
services:
  prometheus:
    image: prom/prometheus:v3.13.0-distroless
    container_name: prometheus
    ports:
      - 9090:9090
    restart: unless-stopped
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prom_data:/var/prometheus

  node-exporter:
    image: prom/node-exporter:latest
    container_name: node-exporter
    ports:
      - "9100:9100"

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - 3000:3000
    restart: unless-stopped
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=grafana
    depends_on:
      - prometheus
  #   volumes:
  #     - ./grafana.yml:/etc/grafana/provisioning/datasources/prometheus.yml
volumes:
  prom_data:
```

Start everything:

```bash
docker compose up -d
```

Check that both containers are running:

```bash
docker compose ps
```

You should see something like:

```text
prometheus
node-exporter
```

### 4. Test Prometheus

Open:

[http://localhost:9090](http://localhost:9090?utm_source=chatgpt.com)

Go to **Status → Targets**.

You should see:

```text
prometheus       UP
node_exporter    UP
```

You can also query metrics directly in the Prometheus UI.

For example:

```promql
up
```

You should get something similar to:

```text
up{job="prometheus"}       1
up{job="node_exporter"}    1
```

`1` means the target is reachable. `0` means Prometheus cannot scrape it.

### 5. Open Grafana

Go to:

[Grafana at localhost:3000](http://localhost:3000?utm_source=chatgpt.com)

The default login is usually:

```text
Username: admin
Password: admin
```

Grafana will ask you to change the password after logging in.

---

### 6. Connect Grafana to Prometheus

In Grafana:

**Connections → Data sources → Add new data source → Prometheus**

For the URL, **do not use**:

```text
http://localhost:9090
```

Use:

```text
http://prometheus:9090
```

Why?

From inside the Grafana container:

```text
localhost:9090
```

means:

> Grafana container, port 9090

Whereas:

```text
prometheus:9090
```

means:

> Prometheus container, port 9090

Then click **Save & test**.

You should get something like:

```text
Successfully queried the Prometheus API.
```

---

### 7. Test a metric

Go to:

**Explore → select Prometheus**

Enter:

```promql
up
```

You should see:

```text
up{job="prometheus"}       1
up{job="node_exporter"}    1
```

Now Grafana is querying Prometheus.

The architecture is:

```text
                    Docker Compose network

┌─────────────┐       ┌─────────────┐
│   Grafana   │──────▶│ Prometheus  │
│   :3000     │       │    :9090    │
└─────────────┘       └──────┬──────┘
                             │
                             │ scrape
                             ▼
                      ┌──────────────┐
                      │ Node Exporter│
                      │    :9100     │
                      └──────────────┘
```

So the flow is:

**Node Exporter → Prometheus → Grafana**

* **Node Exporter** exposes metrics.
* **Prometheus** collects and stores them.
* **Grafana** queries Prometheus and creates dashboards.

### 8. A useful first Grafana query

Try:

```promql
rate(node_cpu_seconds_total[5m])
```

Or, to see whether your targets are working:

```promql
up
```

Once this works, the next useful step is to create a **CPU/RAM/Disk dashboard for Node Exporter** in Grafana.

</details>

## 1. Targets

A **target** is something Prometheus monitors.

For example:

```text id="9r2m6k"
Prometheus
   │
   ├── web-server:8080
   ├── api-server:8080
   ├── database-exporter:9100
   └── node-exporter:9100
```

Each target normally exposes a `/metrics` endpoint.

Example:

```text id="4v8x2p"
http://api-server:8080/metrics
```

Prometheus periodically scrapes it.

**Think:**

> Target = "What should Prometheus monitor?"

---

## 2. Scrape configuration

Prometheus needs to know **which targets to scrape and how often**.

A simplified configuration looks like:

```yaml
scrape_configs:
  - job_name: "api"
    scrape_interval: 15s
    static_configs:
      - targets:
          - "api-server:8080"
```

This says:

> Every 15 seconds, scrape `api-server:8080`.

Important concepts:

* `job_name`
* `targets`
* `scrape_interval`
* `labels`
* `static_configs`

---

## 3. Jobs and instances

Prometheus automatically adds useful labels.

For example:

```text id="6p3k8n"
up{
    job="api",
    instance="api-server:8080"
} 1
```

**Job** = logical group of targets.

**Instance** = individual target.

For example:

```text id="m8q2vc"
job = api

instances:
    api-01:8080
    api-02:8080
    api-03:8080
```

This becomes very useful when you have multiple servers.

---

## 4. `up` metric

One of the first metrics you should know:

```text id="t5n7rx"
up
```

Prometheus automatically creates it for scraped targets.

```text id="y3c8mw"
up = 1  → target is reachable
up = 0  → scrape failed
```

For example:

```text id="q9v4kp"
up{job="api",instance="api-01:8080"} 1

up{job="api",instance="api-02:8080"} 0
```

You immediately know that `api-02` isn't being successfully scraped.

---

# 5. Exporters

Not everything exposes Prometheus metrics natively.

An **exporter** converts metrics from another system into Prometheus format.

For example:

```text id="f2m7qx"
Linux server
     │
     ▼
Node Exporter
     │
     ▼
Prometheus
```

Node Exporter exposes things like:

```text id="x8k3nz"
node_cpu_seconds_total
node_memory_MemAvailable_bytes
node_filesystem_avail_bytes
```

Other exporters exist for databases, hardware, networking devices, etc.

**Think:**

> Exporter = translator that exposes metrics in Prometheus format.

---

# 6. Prometheus data model

This is extremely important.

A Prometheus time series consists essentially of:

```text id="k7p4ms"
metric name
+
set of labels
+
timestamp
+
value
```

For example:

```text id="c4x9vn"
http_requests_total{
    job="api",
    method="GET",
    status="200"
}
```

That combination identifies a particular time series.

This connects directly to your earlier topic:

**Labels → cardinality → number of time series.**

---

# 7. PromQL

**PromQL = Prometheus Query Language.**

This is how you actually get information from Prometheus.

Start with simple queries.

### Select a metric

```text id="w6r2qa"
up
```

### Filter

```text id="n8c5mz"
up{job="api"}
```

### Regex filtering

```text id="p3v7kx"
http_requests_total{status=~"5.."}
```

Meaning:

> HTTP 500–599 responses.

### Rate

```text id="j9m4xc"
rate(http_requests_total[5m])
```

Meaning:

> Average per-second increase over the last 5 minutes.

---

# 8. `rate()` vs `increase()`

This is a very useful distinction.

Suppose:

```text id="a7q2vz"
http_requests_total
```

increased from:

```text id="r5m8nk"
10,000 → 10,600
```

over 5 minutes.

### `increase()`

```text id="d3x9pw"
increase(http_requests_total[5m])
```

Approximately:

```text id="q8k4ms"
600 requests
```

### `rate()`

```text id="v2n7cz"
rate(http_requests_total[5m])
```

Approximately:

```text id="h6p3rx"
600 / 300
= 2 requests/sec
```

So:

> `increase()` → how much did it increase?

> `rate()` → how fast is it increasing?

---

# 9. Aggregation

Suppose you have three API servers:

```text id="m4q8zn"
api-01 → 100 req/s
api-02 → 150 req/s
api-03 → 200 req/s
```

You can aggregate:

```text id="x7c2kp"
sum(rate(http_requests_total[5m]))
```

Result:

```text id="b9v5mw"
450 req/s
```

Other important aggregation functions/operators include:

```text id="p3k8rx"
sum()
avg()
min()
max()
count()
```

You can also aggregate **by label**:

```text id="n6q2vc"
sum by (status) (
    rate(http_requests_total[5m])
)
```

---

# 10. Histogram + PromQL

Now connect this to what you learned about **p95**.

Suppose your application exposes:

```text id="z5m8qx"
http_request_duration_seconds_bucket
```

You can calculate a p95 latency approximately like:

```text id="w3c7pn"
histogram_quantile(
    0.95,
    sum by (le) (
        rate(http_request_duration_seconds_bucket[5m])
    )
)
```

This looks complicated initially.

Don't worry about memorizing it yet.

The important progression is:

```text id="q8v3mz"
Request duration
       ↓
Histogram
       ↓
Buckets
       ↓
Prometheus
       ↓
PromQL
       ↓
p95 / p99
```

---

# 11. Recording rules

Sometimes you repeatedly execute an expensive query.

Instead of calculating it every time, Prometheus can periodically calculate and store the result.

Example:

```yaml
groups:
  - name: api
    rules:
      - record: api:request_rate_5m
        expr: sum(rate(http_requests_total[5m]))
```

Now you can query:

```text id="v7m2qx"
api:request_rate_5m
```

This is called a **recording rule**.

---

# 12. Alerting rules

Prometheus can evaluate conditions.

For example:

```yaml
alert: HighErrorRate
expr: rate(http_requests_total{status=~"5.."}[5m]) > 1
for: 5m
```

Meaning roughly:

> If the API has more than 1 error/sec continuously for 5 minutes, trigger an alert.

This is where your previous concepts become connected:

```text id="k4x8mp"
Metric
  ↓
Prometheus
  ↓
PromQL
  ↓
Condition
  ↓
Alert
```

---

# 13. Alertmanager

Prometheus detects the alert, but **Alertmanager handles the notification workflow**.

For example:

```text id="r6q3vn"
Prometheus
    │
    │ HighErrorRate
    ▼
Alertmanager
    │
    ├── Email
    ├── Slack
    ├── PagerDuty
    └── etc.
```

Alertmanager also handles things such as:

* Grouping alerts
* Deduplication
* Silencing
* Routing
* Inhibition

---

# 14. Service discovery

In a small environment you can manually configure:

```text id="x3m7qc"
api-01:8080
api-02:8080
api-03:8080
```

But imagine you have **500 containers**.

You don't want to manually list them.

Prometheus can discover targets automatically through mechanisms such as:

* Kubernetes
* Consul
* AWS
* Docker
* DNS
* File-based discovery

Conceptually:

```text id="p8k2vz"
Infrastructure
      ↓
Service Discovery
      ↓
Prometheus
      ↓
Targets
```

This becomes particularly important in Kubernetes.

---

# 15. Grafana

Once you're comfortable with Prometheus and PromQL, bring in Grafana.

You'll learn:

```text id="n4c7mx"
Prometheus
     ↓
   PromQL
     ↓
  Grafana
     ↓
Dashboard
```

Create panels for:

* Request rate
* Error rate
* CPU
* Memory
* Disk
* Network
* p50 latency
* p95 latency
* p99 latency

---

# 16. Then move to Loki

Once metrics make sense, learn logs.

```text id="q5v9kp"
Application
    ↓
Logs
    ↓
Promtail / Alloy
    ↓
Loki
    ↓
Grafana
```

Learn:

* Log labels
* Log streams
* Log queries
* Structured logs
* Log levels
* Correlation with metrics

---

# 17. Then OpenTelemetry + traces

Finally:

```text id="m7x3qc"
Application
     │
     ├── Metrics
     ├── Logs
     └── Traces
          ↓
    OpenTelemetry
          ↓
       Backends
```

Learn:

* Instrumentation
* Spans
* Trace IDs
* Span IDs
* Context propagation
* Distributed tracing
* Trace ↔ log correlation

---

## Recommended learning path

Don't try to learn all of these simultaneously.

I'd do:

```text id="c8n4vy"
1. Prometheus architecture
        ↓
2. Targets & scraping
        ↓
3. Jobs & instances
        ↓
4. Exporters
        ↓
5. Prometheus data model
        ↓
6. PromQL basics
        ↓
7. rate() / increase()
        ↓
8. Aggregation
        ↓
9. Histograms & quantiles
        ↓
10. Recording rules
        ↓
11. Alerting rules
        ↓
12. Alertmanager
        ↓
13. Grafana
        ↓
14. Loki
        ↓
15. OpenTelemetry
        ↓
16. Tracing
        ↓
17. SLI/SLO + error budgets
```

**The best next step is actually to install Prometheus and create a tiny application that exposes 3–4 metrics.** Then you can learn targets, scraping, labels, PromQL, counters, gauges, and Grafana through a real example rather than memorizing concepts.
