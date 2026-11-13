# Loki + Grafana Alloy

## 1. What are Loki and Alloy?

They work together in the **log collection and storage pipeline**:

```text
Application / Container
        ↓
      Alloy
        ↓
       Loki
        ↓
     Grafana
```

* **Alloy** → collects, processes, and forwards logs.
* **Loki** → stores and queries logs.
* **Grafana** → visualizes and explores logs.

**Remember:**

> **Alloy collects → Loki stores → Grafana shows**

---

## 2. The Log Pipeline

An application produces logs:

```text
2026-08-21 ERROR database connection failed
```

Alloy can:

* Discover the log source
* Collect the log
* Add metadata
* Parse/process it
* Filter unwanted logs
* Send it to Loki

Then Loki stores it and Grafana can query it.

```text
┌─────────────┐
│ Application │
└──────┬──────┘
       │ logs
       ▼
┌─────────────┐
│    Alloy    │
│ collect     │
│ process     │
│ filter      │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    Loki     │
│ store/query │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Grafana   │
│ explore     │
│ visualize   │
└─────────────┘
```

---

## 3. Loki: Labels, Streams & LogQL

Loki's most important concept is **labels**.

Example:

```text
{service="api", environment="prod"}
```

These labels identify a **log stream**.

You can then query it using **LogQL**:

```logql id="m9b1v8"
{service="api"}
```

Filter logs:

```logql id="1z6q5c"
{service="api"} |= "ERROR"
```

Exclude logs:

```logql id="h7r2pk"
{service="api"} != "healthcheck"
```

For structured JSON logs:

```logql id="q3j8nf"
{service="api"} | json
```

The mental model is:

```text
Loki
 ├── Labels → identify/narrow log streams
 ├── Log data → actual messages
 └── LogQL → query/filter logs
```

---

## 4. Loki Labels & Cardinality

This is **very important**.

Use relatively stable labels:

```text
service
environment
cluster
namespace
container
```

Avoid putting highly dynamic values into labels:

```text
user_id
request_id
trace_id
email
IP address
```

For example, prefer:

```text
{service="api", environment="prod"}
```

with the log containing:

```json
{
  "user_id": "12345",
  "trace_id": "abc123",
  "message": "payment failed"
}
```

rather than making `user_id` or `trace_id` a label.

Why?

> Too many unique label combinations create **high cardinality**, which can make Loki expensive and inefficient.

This connects directly to the **cardinality concept** you already learned.

---

## 5. Alloy: Collection & Processing

Alloy is the **collector/agent** in front of your backends.

Its general pipeline is:

```text
Collect
   ↓
Process
   ↓
Filter / Transform
   ↓
Forward
```

For example:

```text
Container logs
      ↓
    Alloy
      ↓
add service metadata
      ↓
filter unwanted logs
      ↓
     Loki
```

Alloy isn't the place where you normally search your logs.

That's Loki's job.

---

## 6. Structured Logging

Loki works particularly well with structured logs.

Instead of:

```text
Payment failed for user 123 because timeout
```

prefer:

```json
{
  "level": "error",
  "service": "payments",
  "user_id": "123",
  "request_id": "abc123",
  "error": "timeout",
  "message": "payment failed"
}
```

This makes it easier to parse and investigate logs.

High-cardinality information can remain **inside the log**, rather than becoming Loki labels.

---

## 7. Grafana + Loki

Grafana connects to Loki as a data source.

```text
Grafana
   │
   │ LogQL
   ▼
 Loki
   │
   ▼
 Logs
```

You can use **Explore** to investigate logs:

```logql id="qf8s4m"
{service="payments"} |= "ERROR"
```

You can also create dashboards containing things like:

```text
Error logs
Request failures
Authentication failures
Log volume
```

---

## 8. Metrics → Logs

This is where Loki becomes really useful.

Suppose your Grafana dashboard shows:

```text
Error Rate
       │
       │       ╭────
       │  ╭────╯
───────┴────────────────
              time
```

You notice an error spike.

You investigate Loki:

```logql id="c9x2kd"
{service="api"} |= "ERROR"
```

You discover:

```text
database connection failed
database connection failed
database connection failed
```

Now you have:

```text
Prometheus
    ↓
"Something is wrong."
    ↓
Loki
    ↓
"Here's what happened."
```

Later, when you learn traces:

```text
Metrics
   ↓
Logs
   ↓
Traces
```

becomes your main troubleshooting workflow.

---

## 9. Loki vs Alloy vs Grafana

| Tool        | Main responsibility               |
| ----------- | --------------------------------- |
| **Alloy**   | Collect/process/forward telemetry |
| **Loki**    | Store/query logs                  |
| **Grafana** | Visualize/explore telemetry       |

The simplest picture:

```text
             Application
                  │
                  │ logs
                  ▼
                Alloy
             "collect"
                  │
                  ▼
                 Loki
             "store/query"
                  │
                  ▼
               Grafana
             "visualize"
```

---

## 10. What You Should Know

Before moving on, make sure you understand:

* Why Loki exists
* What Alloy does
* **Alloy → Loki → Grafana** pipeline
* Loki labels
* Log streams
* Basic LogQL
* Structured logs
* Loki cardinality
* Why dynamic values shouldn't usually be labels
* Alloy's role in collecting and processing logs
* Using Grafana Explore to investigate Loki logs
* **Metrics → Logs** troubleshooting workflow

### Core sentence to memorize

> **Grafana Alloy collects and processes logs, Loki stores and queries them, and Grafana visualizes and explores them.**
