# SLI/SLO/SLA

These three concepts are related, but they answer **different questions**:

> **SLI = What are we measuring?**  
> **SLO = What target do we want?**  
> **SLA = What do we promise the customer?**  

### 1. SLI — Service Level Indicator

An **SLI is a measurement of actual service performance**.

For example, suppose you run an API.

You measure:

```text id="8k3m1z"
Successful requests
───────────────────
Total requests
```

If you had:

```text id="6y4p9c"
99,700 successful
100,000 total
```

Then:

```text id="z2r5hx"
SLI = 99,700 / 100,000
    = 99.7%
```

So your **availability SLI = 99.7%**.

Other SLIs could measure:

* Availability
* Request latency
* Error rate
* Throughput
* Data freshness

---

### 2. SLO — Service Level Objective

An **SLO is the target you set for an SLI**.

Suppose you decide:

> "We want 99.9% of requests to succeed."

Then:

```text id="q7v3nk"
SLI = 99.7%   ← actual
SLO = 99.9%   ← target
```

You're **missing your SLO**.

Another example:

```text id="f8k2ms"
SLI: p95 latency = 280 ms

SLO: p95 latency < 500 ms
```

You're meeting the SLO.

---

### 3. SLA — Service Level Agreement

An **SLA is a formal commitment, usually contractual, made to customers**.

For example:

> "Our service will have 99.9% monthly availability."

If you don't meet the SLA, there might be consequences such as **service credits or refunds**, depending on the contract.

So:

```text id="j5c8vp"
SLI
 ↓
99.7% actual availability

SLO
 ↓
99.9% desired availability

SLA
 ↓
99.9% contractual commitment to customers
```

### Simple real-world analogy

Imagine a pizza delivery service:

**SLI:**

> 92% of pizzas arrived within 30 minutes.

That's what actually happened.

**SLO:**

> We aim for 95% of pizzas to arrive within 30 minutes.

That's the internal target.

**SLA:**

> We guarantee 95% delivery within 30 minutes, or you get your money back.

That's the contractual promise.

---

### Why does this matter for monitoring?

This is where observability becomes more than just **"let's collect lots of metrics."**

Suppose your SLO is:

```text id="3v6n8b"
99.9% successful requests
```

Your Prometheus metrics might provide:

```text id="c1m7qa"
Total requests       = 1,000,000
Successful requests  =   999,200
```

Your monitoring can calculate:

```text id="s9d4we"
SLI = 999,200 / 1,000,000
    = 99.92%
```

Therefore:

```text id="0k5x2p"
SLO = 99.9%
SLI = 99.92%

✅ SLO currently being met
```

If it drops to:

```text id="n8q2yr"
SLI = 99.7%

❌ SLO violated
```

You can then configure **Grafana/Alertmanager** to alert your team.

### One more concept: Error budget

This is the natural next step.

If your SLO is **99.9% availability**, you're allowing:

```text id="v4p6ka"
100% - 99.9%
     = 0.1%
```

of failure.

That **0.1% is your error budget**.

For a 30-day month:

```text id="2s7m9d"
30 days × 24 hours × 60 minutes
= 43,200 minutes

0.1% × 43,200
= 43.2 minutes
```

So a **99.9% monthly availability SLO** gives you approximately **43 minutes of allowable unavailability per month**.

This is a very important concept because it connects **observability → monitoring → alerting → reliability → engineering decisions**.
