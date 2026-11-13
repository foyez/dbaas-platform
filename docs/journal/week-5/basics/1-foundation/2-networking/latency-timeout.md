# Basic latency and timeout concepts

These two concepts are **fundamental for observability**, because many production problems eventually show up as either **high latency** or **timeouts**.

### 1. Latency

**Latency = how long it takes something to happen.**

For an HTTP request:

```text id="e2v3k8"
Client ───── request ─────> Server
Client <──── response ───── Server

        ←── 200 ms ──→
```

The latency is **200 ms**.

You might measure:

* **50 ms** → very fast
* **200 ms** → generally fine
* **2 seconds** → noticeably slow
* **10 seconds** → probably a problem

In observability, you'll often see metrics such as:

```text id="6q1c7s"
HTTP request latency = 250 ms
```

But an important point is that **average latency isn't enough**.

Suppose 100 requests take:

* 95 requests → 50 ms
* 5 requests → 5 seconds

The average is roughly **298 ms**, but 5% of your users are having a terrible experience.

That's why you'll encounter **percentiles**:

* **p50** → median request
* **p95** → 95% of requests are faster than this
* **p99** → 99% are faster than this

For example:

```text id="4m6b2k"
p50 = 80 ms
p95 = 300 ms
p99 = 2.1 seconds
```

This tells you much more than simply:

```text id="q1n8zs"
average = 298 ms
```

**Percentile Calculation:**

Imagine you have **100 requests**, sorted from fastest to slowest.

```text id="h3s8nq"
Request #1    = 20 ms
Request #2    = 25 ms
...
Request #50   = 80 ms
...
Request #95   = 300 ms
...
Request #99   = 2,100 ms
Request #100  = 5,000 ms
```

Then:

* **p50 = 80 ms** → 50% of requests took **80 ms or less**.
* **p95 = 300 ms** → 95% of requests took **300 ms or less**. The slowest 5% took longer.
* **p99 = 2.1 seconds** → 99% of requests took **2.1 seconds or less**. The slowest 1% took longer.

### Very simple visualization

```text id="k2p6wc"
0%                  50%              95%       99%      100%
|-------------------|----------------|---------|---------|
                   80ms            300ms     2.1s      5s
                    ↑                ↑         ↑
                   p50              p95       p99
```

**Important:** p95 does **not** mean "the average of the fastest 95%." It means the **value at the 95th percentile** after sorting the observations.

---

### 2. Timeout

A **timeout is a limit on how long you're willing to wait**.

For example:

```text id="p4x7nm"
Your application
      │
      │ request
      ↓
  Database
      │
      │ ... still processing ...
      │
      │
   5 seconds
      │
      ↓
    TIMEOUT
```

You might configure:

> "If the database doesn't respond within 5 seconds, stop waiting."

The application then gets a timeout error.

### Why do timeouts matter?

Imagine:

```text id="q7c2fz"
User
 ↓
API
 ↓
Payment service
 ↓
Database
```

Suppose the database becomes extremely slow.

Without sensible timeouts:

```text id="3m8x1k"
100 requests
     ↓
100 requests waiting
     ↓
more requests arrive
     ↓
more requests waiting
     ↓
threads/connections get exhausted
     ↓
API becomes unavailable
```

A small problem can therefore **cascade into a much bigger failure**.

---

### Latency vs. timeout

Think of it like ordering food at a restaurant:

**Latency:**

> "How long did my food take to arrive?"

**Timeout:**

> "I'm willing to wait 30 minutes. If it hasn't arrived by then, I'm leaving."

So:

```text
Latency = actual duration
Timeout = maximum acceptable duration
```

---

### Why this matters for Prometheus/Grafana

Eventually you'll create dashboards and alerts around things like:

```text
Request rate
    ↓
1000 requests/sec

Error rate
    ↓
0.5%

Latency
    ↓
p50  = 80 ms
p95  = 300 ms
p99  = 2 sec
```

You might then have an alert such as:

> **p95 HTTP latency has exceeded 1 second for 5 minutes.**

That is much more useful than simply monitoring whether the server is "up."

One important next concept to learn after this is **percentiles (p50/p95/p99) and histograms**, because they connect directly to how Prometheus measures request latency.
