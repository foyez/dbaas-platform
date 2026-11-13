# Grafana

## 1. What is Grafana?

**Grafana is a visualization and observability platform.** It connects to different data sources, queries them, and presents the data through dashboards, panels, and Explore.

For your stack:

```text
Prometheus → Metrics ──┐
Loki       → Logs ─────┤
Tempo      → Traces ───┤
                       ▼
                    Grafana
```

**Remember:**

> Prometheus/Loki/Tempo store and provide telemetry; Grafana visualizes and explores it.

---

## 2. Data Sources, Dashboards & Panels

A **data source** is where Grafana gets its data.

Examples:

```text
Prometheus → metrics
Loki       → logs
Tempo      → traces
```

A **dashboard** contains multiple **panels**.

```text
Dashboard
├── Request Rate
├── Error Rate
├── CPU Usage
├── Memory Usage
└── P95 Latency
```

Each panel typically has:

```text
Data source + Query + Visualization
```

For example:

```promql
rate(http_requests_total[5m])
```

can be displayed as a time-series graph.

---

## 3. Grafana + Prometheus

Since you've already learned Prometheus, this is the most important connection.

Grafana sends **PromQL queries** to Prometheus:

```text
Grafana
   │
   │ PromQL
   ▼
Prometheus
   │
   ▼
Metrics
```

Example:

```promql
rate(http_requests_total{status="500"}[5m])
```

Grafana visualizes the result as a graph, stat, table, etc.

So:

> **PromQL is still the query language. Grafana is the interface for exploring and visualizing the result.**

---

## 4. Explore vs Dashboard

**Dashboard** → continuously monitor known important information.

**Explore** → investigate something interactively.

For example, you notice an error spike in a dashboard:

```text
Error Rate
     │
     │       ╭───
     │  ╭────╯
─────┴──────────────── time
```

You go to **Explore** and investigate:

```promql
rate(http_requests_total{status=~"5.."}[5m])
```

Then filter by service:

```promql
sum by (service) (
  rate(http_requests_total{status=~"5.."}[5m])
)
```

Think:

> **Dashboard = monitoring. Explore = investigation.**

---

## 5. Variables and Labels

Grafana can use dashboard variables to make one dashboard work for multiple services.

Example:

```text
Service: [ api ▼ ]
```

Query:

```promql
rate(http_requests_total{service="$service"}[5m])
```

Changing the variable to `payments` changes the data shown.

This works together with your Prometheus knowledge of **labels**.

---

## 6. Useful Visualizations

Choose the visualization based on the question:

| Visualization | Good for                         |
| ------------- | -------------------------------- |
| Time series   | Rate, CPU, latency over time     |
| Stat          | One important number             |
| Gauge         | Current percentage/threshold     |
| Table         | Per-service/per-instance details |
| Heatmap       | Distributions such as latency    |

Don't create dashboards just because you can. A good dashboard should answer useful operational questions.

For an API, a good starting dashboard is:

```text
Traffic
Error rate
Latency
CPU / Memory
```

---

## 7. Alerting

Grafana can also evaluate alert rules and send notifications.

Conceptually:

```text
Metric
  ↓
Query
  ↓
Alert rule
  ↓
Notification
```

Example:

```text
IF error rate > 5%
FOR 5 minutes
THEN alert
```

But remember the distinction from **Alertmanager**:

```text
Prometheus
   ↓
Alertmanager
   ↓
routing / grouping / silencing
```

Grafana also has its own alerting system, so don't assume "Grafana alert" and "Prometheus + Alertmanager alert" are the same thing.

---

## 8. Grafana + Loki + Tempo

Later your observability workflow becomes:

```text
             Grafana
          /      |      \
         /       |       \
Prometheus      Loki     Tempo
 Metrics        Logs     Traces
```

Example troubleshooting:

```text
Grafana dashboard
       ↓
Error rate increased
       ↓
Explore Loki
       ↓
Find error logs
       ↓
Find trace
       ↓
Identify slow/failing service
```

This **metrics → logs → traces** workflow is one of the most important things you'll eventually practice.

---

## 9. Important Things to Remember

* **Grafana does not primarily collect your telemetry.**
* It connects to **data sources**.
* **Dashboard** = collection of panels.
* **Panel** = individual visualization.
* **Explore** = interactive investigation.
* Prometheus is queried with **PromQL**.
* Loki is queried with **LogQL**.
* Dashboard **variables** make dashboards reusable.
* Don't overload dashboards with meaningless graphs.
* Use Grafana to understand **what is happening and when**.

### Core mental model

```text
Prometheus → metrics
Loki       → logs
Tempo      → traces
             ↓
          Grafana
             ↓
      visualize + explore
```

**One sentence to memorize:**

> **Grafana is the visualization and exploration layer that connects to observability data sources such as Prometheus, Loki, and Tempo.**
