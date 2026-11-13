# Prometheus Metrics Types

These are the **four Prometheus metric types** you'll encounter most often. The easiest way to understand them is by asking: **"What kind of thing am I measuring?"**

### 1. Counter

A **Counter only goes up** (until the application restarts).

Example: total HTTP requests.

```text id="v7x2qp"
10:00 → 1,000 requests
10:01 → 1,250
10:02 → 1,800
10:03 → 2,100
```

You don't normally ask:

> "How many requests have ever happened?"

Instead, you calculate the **rate**:

> "How many requests per second are happening?"

Prometheus example:

```text id="0u6f9a"
http_requests_total
```

Typical counters:

* Total requests
* Total errors
* Total orders
* Total bytes processed

**Think:** 🧮 **Counter = cumulative total**

---

### 2. Gauge

A **Gauge can go up and down**.

Example: CPU usage:

```text id="5a1z4c"
10:00 → 40%
10:01 → 60%
10:02 → 80%
10:03 → 55%
10:04 → 30%
```

Prometheus example:

```text id="2j8x5m"
cpu_usage_percent 30
```

Typical gauges:

* CPU usage
* Memory usage
* Number of active users
* Queue size
* Temperature
* Number of running processes

**Think:** 📏 **Gauge = current value**

---

### 3. Histogram

A **Histogram measures the distribution of values**, commonly request duration.

Suppose your application receives these requests:

```text id="7c4n1p"
50 ms
80 ms
120 ms
200 ms
300 ms
500 ms
2,000 ms
```

You want to know:

> "How many requests took less than 100 ms, 500 ms, etc.?"

A histogram puts observations into **buckets**:

```text id="3x8q6v"
≤ 100 ms     → 2 requests
≤ 500 ms     → 6 requests
≤ 1 sec      → 6 requests
≤ 5 sec      → 7 requests
```

This lets you calculate things like:

```text id="m5v9ka"
p50 = 120 ms
p95 = 500 ms
p99 = 2 sec
```

Histograms are extremely useful for:

* Request latency
* Response size
* Processing duration

**Think:** 📊 **Histogram = distribution of values**

---

### 4. Summary

A **Summary also measures distributions**, but it calculates **quantiles** such as p50, p95, and p99 on the client/application side.

For example:

```text id="q8w2de"
request_duration:
    p50 = 80 ms
    p95 = 300 ms
    p99 = 1.2 sec
```

So why have both Histogram and Summary?

The important practical distinction is:

|                                | Histogram                     | Summary                            |
| ------------------------------ | ----------------------------- | ---------------------------------- |
| Measures distribution          | ✅                             | ✅                                  |
| Can calculate p50/p95/p99      | Prometheus can calculate them | Client calculates them             |
| Uses buckets                   | ✅                             | ❌                                  |
| Can aggregate across instances | ✅                             | Generally difficult/not meaningful |
| Common choice in Prometheus    | **Very common**               | Less common                        |

For example, if you have:

```text id="1d5y6z"
Server 1 → p95 = 200 ms
Server 2 → p95 = 800 ms
Server 3 → p95 = 300 ms
```

You generally **cannot simply average those p95 values** to get the overall p95.

Histograms solve this by allowing Prometheus to aggregate the underlying buckets across servers.

---

### The easiest way to remember

```text id="2n6z4c"
COUNTER
"What is the total?"
        ↓
Requests = 1,000,000


GAUGE
"What is the value right now?"
        ↓
CPU = 72%


HISTOGRAM
"How are the values distributed?"
        ↓
Request latency:
50% < 100ms
95% < 500ms
99% < 2s


SUMMARY
"Give me quantiles of the distribution"
        ↓
p50 = 100ms
p95 = 500ms
p99 = 2s
```

### One important practical point

When you start working with Prometheus, you'll probably use **Counter, Gauge, and Histogram most frequently**.

A very common real-world setup is:

```text id="j4p7sx"
http_requests_total       → Counter
active_users               → Gauge
http_request_duration     → Histogram
```

And then Grafana can turn those measurements into dashboards and alerts.
