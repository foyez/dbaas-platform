# Traces

**Traces** show the **journey of a request through a system**. They're especially useful in distributed systems where one user action may involve many services.

For example, when you click **“Buy”**, the request might travel through:

```text
Browser
   ↓
API Gateway
   ↓
Order Service
   ↓
Payment Service
   ↓
Database
```

A trace records that whole journey, including how long each part took:

```text
Trace: checkout_request
Total: 850 ms

├── API Gateway       20 ms
├── Order Service    150 ms
├── Payment Service  500 ms
└── Database         180 ms
```

Each individual operation within the trace is usually called a **span**.

### Traces vs. logs vs. metrics vs. events

| Type       | Main question                                               | Example                    |
| ---------- | ----------------------------------------------------------- | -------------------------- |
| **Event**  | What happened?                                              | `payment_completed`        |
| **Metric** | How much/how often?                                         | `500 payments/minute`      |
| **Log**    | What happened, with details?                                | `Payment failed: timeout`  |
| **Trace**  | Where did this request go, and how long did each step take? | API → Order → Payment → DB |

The key idea is:

> **Metrics tell you that something is wrong. Logs give you details. Traces show you where in the request flow the problem occurred.**

These four—**events, metrics, logs, and traces**—are common building blocks of modern **observability/telemetry**.
