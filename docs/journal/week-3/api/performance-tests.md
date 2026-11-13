# Performance Testing of REST API with k6 and Kubernetes HPA

## 1. Purpose

The goal of the performance tests is to verify that the REST API behaves correctly under load and that the Kubernetes **Horizontal Pod Autoscaler (HPA)** automatically scales the API pods when the load increases.

The test should verify both:

* **API performance:** response time, throughput, and errors.
* **HPA behavior:** scale-out when load increases and scale-down when load decreases.

## 2. Performance Testing vs. Load Testing

**Performance testing** is the general term for testing how a system behaves under different workloads.

**Load testing** is one type of performance testing. It verifies that the application can handle the expected/normal workload while maintaining acceptable performance.

For this HPA use case, the most important tests are:

1. **Load test** – verify the API under expected load.
2. **Stress test** – gradually increase the load to determine how the system behaves at higher loads.
3. **Scale-down test** – verify that the HPA reduces the number of pods when the load decreases.
4. **Spike test** – optionally verify how the system behaves when traffic increases suddenly.

## 3. Why k6?

[k6](https://k6.io/) is a load-testing tool that allows you to generate HTTP traffic against the REST API and measure its performance.

A k6 test can define:

* Number of virtual users
* Request rate
* Test duration
* Load increase/decrease
* Performance thresholds
* HTTP checks

The important point is that **k6 generates the traffic; Kubernetes/HPA handles the scaling**.

## 4. Metrics to Monitor

During the tests, monitor both k6 and Kubernetes.

### k6

Important metrics:

* Request rate
* Response time
* p95 latency
* p99 latency
* HTTP error rate
* Failed requests

### Kubernetes

Monitor:

* Current pod count
* Desired pod count
* CPU utilization
* Memory utilization
* HPA events
* Pod startup/readiness time

The most important relationship to observe is:

```text
Load increases
      ↓
CPU utilization increases
      ↓
HPA detects target exceeded
      ↓
Number of pods increases
      ↓
API performance stabilizes
```

## 5. Recommended Tests

### Test 1 — Baseline

Start with a low load to understand normal API performance.

Example:

```text
20 requests/second
2 minutes
```

Record the normal response time, CPU, memory, and error rate.

### Test 2 — Load / HPA Scale-Out

Gradually increase the load.

Example:

```text
20 req/s → 50 req/s → 100 req/s → 200 req/s
```

Hold each level for several minutes.

Verify that:

* API remains available.
* Response times remain within the defined limit.
* Error rate stays within the acceptable limit.
* HPA increases the number of pods when the CPU target is exceeded.

### Test 3 — Scale-Down

After reaching a high load, reduce the traffic significantly.

Example:

```text
200 req/s
     ↓
20 req/s
```

Verify that the HPA eventually reduces the number of pods back toward `minReplicas`.

### Test 4 — Stress / Spike

Increase the load beyond the expected workload or increase it very quickly.

Example:

```text
20 req/s → 300 req/s
```

Observe:

* Maximum sustainable load
* Response-time degradation
* Error rate
* Maximum number of replicas
* Time required for new pods to become ready

## 6. What tool should you use?

For a REST API, common choices are:

* **k6** — excellent for API performance/load testing
* **Apache JMeter** — mature and feature-rich
* **Gatling** — good for code-based performance testing
* **Locust** — Python-based and convenient for custom scenarios

## 7. Simple k6 Test

A basic k6 script could look like:

```javascript
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 10 },
    { duration: '5m', target: 30 },
    { duration: '5m', target: 50 },
    { duration: '5m', target: 10 },
    { duration: '2m', target: 0 },
  ],

  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

export default function () {
  const response = http.get('http://YOUR-API-ENDPOINT/health');

  check(response, {
    'status is 200': (r) => r.status === 200,
  });
}
```

The exact endpoint and thresholds should be adapted to your API's requirements.

**Important:** `target` in this example represents virtual users, not requests per second. The actual request rate depends on what each virtual user does in the script.

## 8. What a Successful Test Means

A successful HPA performance test should demonstrate that:

```text
Normal load
   → API performs normally

Higher load
   → CPU increases
   → HPA increases replicas
   → API continues serving requests

Load decreases
   → HPA decreases replicas
   → API remains healthy
```

The final test report should include at least:

* Test duration and load
* p95/p99 response time
* Request rate
* Error rate
* Minimum/maximum replicas
* CPU/memory utilization
* HPA scale-out and scale-down behavior
* Any performance degradation or errors
