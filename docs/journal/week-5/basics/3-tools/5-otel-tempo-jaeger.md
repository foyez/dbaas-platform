# OpenTelemetry + Tempo / Jaeger

## 1. What are they?

These are mainly about **distributed tracing**.

A simple observability stack:

```text
Metrics → Prometheus
Logs    → Loki
Traces  → Tempo / Jaeger
              ↓
           Grafana
```

**OpenTelemetry (OTel)** is used to **instrument applications and collect/transport telemetry**.

**Tempo / Jaeger** are used to **store and query traces**.

Think:

> **OpenTelemetry creates/collects traces → Tempo/Jaeger stores them → Grafana explores them.**

---

## 2. What is a Trace?

A **trace represents one request as it travels through a system.**

Imagine:

```text
User
 ↓
API
 ↓
Auth service
 ↓
Payment service
 ↓
Database
```

A trace lets you see the entire journey:

```text
Trace
│
├── API request        20ms
│    └── Auth          5ms
│    └── Payment      12ms
│         └── DB       9ms
```

This is especially useful in **microservices**, where one user request can involve many services.

---

## 3. Spans

A trace is made up of **spans**.

```text
Trace
│
├── Span: API request
├── Span: Auth service
├── Span: Payment service
└── Span: Database query
```

Each span describes one operation and commonly contains:

* Start/end time
* Service name
* Operation name
* Attributes
* Status
* Parent/child relationship

The relationship is:

```text
Trace
  └── Spans
        ├── parent span
        ├── child span
        └── child span
```

**Remember:**

> **Trace = complete request journey. Span = one operation within that journey.**

---

## 4. OpenTelemetry

**OpenTelemetry is an open-source observability framework for generating, collecting, and exporting telemetry.**

It supports:

```text
Metrics
Logs
Traces
```

For your current topic, focus mainly on **traces**.

A typical flow:

```text
Application
    │
    │ OpenTelemetry instrumentation
    ▼
OTel SDK
    │
    ▼
OpenTelemetry Collector / Alloy
    │
    ▼
Tempo
```

OpenTelemetry helps your application produce standardized telemetry instead of using a backend-specific instrumentation approach.

---

## 5. Automatic vs Manual Instrumentation

OpenTelemetry can instrument applications in different ways.

### Automatic instrumentation

OTel can automatically instrument common frameworks/libraries.

For example:

```text
HTTP request
   ↓
automatic span
```

### Manual instrumentation

You explicitly create spans around important operations:

```text
Process payment
     ↓
create span
     ↓
call payment provider
     ↓
finish span
```

You don't need to manually instrument every tiny function.

The goal is to capture operations useful for troubleshooting.

---

## 6. Context Propagation

This is one of the most important tracing concepts.

Suppose:

```text
Service A
   ↓
Service B
   ↓
Service C
```

How does Service B know that its request belongs to the same trace?

**Trace context is propagated between services.**

Conceptually:

```text
Trace ID: abc123

Service A ─────► Service B ─────► Service C
    │                │                │
  span             span             span
```

Now all these spans can be connected into **one trace**.

Without context propagation, distributed tracing would be much less useful.

---

# 7. Tempo

**Grafana Tempo is a distributed tracing backend.**

Its job is primarily to:

* Receive traces
* Store traces
* Allow traces to be queried
* Integrate with Grafana

Typical flow:

```text
Application
    ↓
OpenTelemetry
    ↓
Alloy / OTel Collector
    ↓
Tempo
    ↓
Grafana
```

Tempo is therefore similar in role to Loki, but for traces:

```text
Loki  → logs
Tempo → traces
```

---

# 8. Jaeger

**Jaeger is another distributed tracing system.**

Conceptually:

```text
OpenTelemetry
      ↓
 ┌────┴─────┐
 ▼          ▼
Tempo     Jaeger
```

Both can act as tracing backends.

You don't need to learn both deeply at the same time.

For your Grafana-focused stack, **Tempo is the natural one to practice first**.

Understand Jaeger mainly as:

> **Another popular distributed tracing backend/tool.**

---

# 9. OpenTelemetry vs Tempo

This distinction is critical.

```text
OpenTelemetry
     ↓
instrument / collect / transport
     ↓
Tempo
     ↓
store / query traces
```

So:

| Component     | Main job                          |
| ------------- | --------------------------------- |
| OpenTelemetry | Generate/collect/export telemetry |
| Alloy         | Collect/process/forward telemetry |
| Tempo         | Store/query traces                |
| Jaeger        | Store/query/visualize traces      |
| Grafana       | Explore/visualize telemetry       |

---

# 10. The Full Observability Flow

Your stack is now coming together:

```text
                     Application
                         │
          ┌──────────────┼──────────────┐
          │              │              │
       Metrics          Logs          Traces
          │              │              │
          ▼              ▼              ▼
     Prometheus         Alloy        OpenTelemetry
                                      │
          │              │            │
          ▼              ▼            ▼
     Prometheus          Loki        Tempo
          │              │            │
          └──────────────┼────────────┘
                         ▼
                      Grafana
```

A more realistic trace path can be:

```text
Application
    ↓
OpenTelemetry SDK
    ↓
Alloy / OTel Collector
    ↓
Tempo
    ↓
Grafana
```

---

# 11. Metrics → Logs → Traces

This is the workflow you ultimately want to master.

Imagine Grafana shows:

```text
P95 latency suddenly increased
```

### Step 1 — Metrics

Prometheus tells you:

```text
API latency ↑
```

### Step 2 — Logs

Loki shows:

```text
database timeout
```

### Step 3 — Trace

Tempo shows:

```text
API request
   ↓
Payment service
   ↓
Database query ← 4.8 seconds
```

Now you can identify **where the request became slow**.

That's the real power of combining the three signals.

---

# 12. Trace IDs

A trace usually has a **Trace ID**.

For example:

```text
trace_id = abc123
```

The same trace ID connects the spans belonging to one request.

This becomes extremely useful when correlating:

```text
Metrics
   ↓
Logs
   ↓
Trace ID
   ↓
Trace
```

For example, a log may contain:

```json
{
  "level": "error",
  "trace_id": "abc123",
  "message": "database timeout"
}
```

You can then jump from the log to the corresponding trace.

---

# 13. What You Should Know

Before moving into the hands-on observability phase, make sure you understand:

* **Trace vs span**
* Parent/child span relationships
* Why distributed tracing is useful
* OpenTelemetry's role
* Instrumentation
* Automatic vs manual instrumentation
* Context propagation
* Trace IDs
* Tempo's role
* Jaeger's role
* OpenTelemetry vs Tempo/Jaeger
* Metrics → Logs → Traces correlation

### Core mental model

```text
OpenTelemetry
     │
     │ creates/collects/exports traces
     ▼
Alloy / OTel Collector
     │
     ▼
Tempo / Jaeger
     │
     ▼
Grafana
```

### Core sentence to memorize

> **OpenTelemetry provides standardized instrumentation and telemetry collection, while Tempo or Jaeger acts as the tracing backend that stores and lets you investigate distributed traces.**
