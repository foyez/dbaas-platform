# Phase 1: **Prometheus + Grafana + your Gin API + one sample app + basic Loki**.

## 0. Target architecture

```text
k8s-dev
│
├── paas
│   ├── paas-api
│   └── sample-app
│
└── observability
    ├── Prometheus
    ├── Grafana
    └── Loki
```

Flow:

```text
paas-api
   │
   │ /metrics
   ▼
Prometheus
   │
   ▼
Grafana
```

and:

```text
sample-app
   │
   │ stdout/stderr
   ▼
Alloy
   │
   ▼
Loki
   │
   ▼
Grafana
```

For Phase 1, use **Grafana Alloy** as the log collector rather than starting with Promtail.

---

# 1. Create namespaces

Assuming your PaaS is in namespace `paas`:

```bash
kubectl create namespace paas
kubectl create namespace observability
```

If `paas` already exists, that's fine.

---

# 2. Install Prometheus + Grafana

Add the Helm repository:

```bash
helm repo add prometheus-community \
  https://prometheus-community.github.io/helm-charts

helm repo update
```

Create:

```text
observability-values.yaml
```

with:

```yaml
grafana:
  enabled: true

  persistence:
    enabled: true
    size: 5Gi

prometheus:
  prometheusSpec:
    retention: 7d

    storageSpec:
      volumeClaimTemplate:
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 10Gi

    serviceMonitorSelectorNilUsesHelmValues: false

alertmanager:
  enabled: true
```

Install:

```bash
helm install monitoring \
  prometheus-community/kube-prometheus-stack \
  --namespace observability \
  -f observability-values.yaml
```

Check:

```bash
kubectl get pods -n observability
```

You'll see several components, something approximately like:

```
NAME                                                     READY
monitoring-grafana                                       1/1
monitoring-kube-prometheus-prometheus-0                 2/2
monitoring-kube-prometheus-operator-xxxxx               1/1
monitoring-kube-state-metrics-xxxxx                     1/1
monitoring-prometheus-node-exporter-xxxxx               1/1
alertmanager-monitoring-kube-prometheus-alertmanager-0  2/2
```

You should eventually have Prometheus, Grafana, Alertmanager, node-exporter, kube-state-metrics, etc.

```
kube-prometheus-stack
│
├── Prometheus
│   └── Stores/query metrics
│
├── Grafana
│   └── Dashboards
│
├── Alertmanager
│   └── Handles alerts
│
├── Prometheus Operator
│   └── Manages Prometheus configuration
│
├── kube-state-metrics
│   └── Kubernetes object/state metrics
│
└── node-exporter
    └── Node/system metrics
```

---

# 3. Access Prometheus

Run:

```sh
kubectl get svc -n observability
```

You'll see services similar to:

```
NAME                                      TYPE
monitoring-grafana                        ClusterIP
monitoring-kube-prometheus-prometheus    ClusterIP
monitoring-kube-prometheus-alertmanager  ClusterIP
```

For initial testing, use port forwarding rather than exposing anything publicly.

```bash
kubectl port-forward \
  -n observability \
  svc/monitoring-kube-prometheus-prometheus \
  9090:9090
```

Open:

```text
http://localhost:9090
```

In Prometheus, go to:

**Status → Targets**

You should see targets such as:

```
kube-state-metrics
node-exporter
kubelet
prometheus
```

This is one of the most important things to understand.

Prometheus is essentially saying:

```
I know these targets:


node-exporter
kube-state-metrics
kubelet
...


I'll scrape their /metrics endpoints.
```

If they're showing **UP**, congratulations: your first monitoring system is working.

Test:

```promql
up
```

You'll get something like:

```
up{job="node-exporter", instance="10.0.1.10:9100"} 1


up{job="node-exporter", instance="10.0.1.11:9100"} 1


up{job="kube-state-metrics", instance="..."} 1
```

Remember what you learned earlier:

```
up = 1
```

means the target was successfully scraped.

```
up = 0
```

means the scrape failed.

**Explore Kubernetes metrics**

Now try:

```promql
kube_node_info
```

Then:

```promql
kube_pod_info
```

Then:

```promql
kube_deployment_status_replicas
```

And:

```promql
kube_pod_container_status_restarts_total
```

You're now using actual Kubernetes telemetry.

**Look at CPU usage**

Try:

```promql
100 - (
  avg by (instance) (
    rate(node_cpu_seconds_total{mode="idle"}[5m])
  ) * 100
)
```

This gives approximate CPU utilization per node.

Don't worry if the query looks complicated.

Break it down:

```text
node_cpu_seconds_total
        ↓
rate(...[5m])
        ↓
CPU usage rate
        ↓
avg by instance
        ↓
100 - idle
        ↓
CPU utilization %
```

This is exactly the type of PromQL you'll eventually use in Grafana.

---

# 4. Access Grafana

In another terminal:

```bash
kubectl port-forward \
  -n observability \
  svc/monitoring-grafana \
  3000:80
```

Open:

```text
http://localhost:3000
```

Get the admin password:

```bash
kubectl get secret monitoring-grafana \
  -n observability \
  -o jsonpath="{.data.admin-password}" | base64 -d
```

If your release name differs, first find the secret:

```bash
kubectl get secrets -n observability
```

Username:

```text
admin
```

The Prometheus datasource should already be configured by `kube-prometheus-stack`.

---

# 5. Your Gin `/metrics`

You already have this:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
        })
    })

    r.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // ...
}
```

That's good.

You should **not put `/metrics` behind your ZITADEL authentication middleware**.

You currently have:

```go
v1 := r.Group("/api/v1")

protected := v1.Group("")
protected.Use(auth.Gin(authmw.RequireAuth()))
```

That's fine.

Your structure should remain:

```text
/health          → Kubernetes
/metrics         → Prometheus

/api/v1/...      → authenticated users
```

---

# 6. Add HTTP metrics to Gin

Now let's add your first useful application metrics.

Create:

```text
internal/observability/metrics.go
```

```go
package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "paas_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "paas_http_request_duration_seconds",
			Help: "HTTP request duration in seconds.",
			Buckets: []float64{
				0.005,
				0.01,
				0.025,
				0.05,
				0.1,
				0.25,
				0.5,
				1,
				2.5,
				5,
				10,
			},
		},
		[]string{"method", "route", "status"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		route := c.FullPath()

		// If Gin couldn't resolve the route, avoid creating
		// arbitrary path labels.
		if route == "" {
			route = "unknown"
		}

		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		HTTPRequestsTotal.WithLabelValues(
			method,
			route,
			status,
		).Inc()

		HTTPRequestDuration.WithLabelValues(
			method,
			route,
			status,
		).Observe(duration)
	}
}
```

Notice this:

```go
route := c.FullPath()
```

That's important.

For:

```text
GET /api/v1/instances/123
GET /api/v1/instances/456
GET /api/v1/instances/789
```

Gin gives you:

```text
/api/v1/instances/:id
```

rather than creating a unique Prometheus label for every ID.

That's how you avoid unnecessary cardinality.

---

# 7. Register the metrics

In your `main.go`:

```go
import (
    // ...
    "your-module/internal/observability"
)
```

Before starting the server:

```go
observability.RegisterMetrics()
```

Then add the middleware:

```go
r := gin.Default()

r.Use(observability.Middleware())
```

Your setup becomes approximately:

```go
r := gin.Default()

observability.RegisterMetrics()
r.Use(observability.Middleware())

r.GET("/health", func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
})

r.GET("/metrics", gin.WrapH(promhttp.Handler()))

v1 := r.Group("/api/v1")

protected := v1.Group("")
protected.Use(auth.Gin(authmw.RequireAuth()))

protected.POST("/instances", h.CreateInstance)
protected.GET("/instances", h.ListInstances)
protected.GET("/instances/:id", h.GetInstance)
protected.GET("/instances/:id/credentials", h.GetCredentials)
protected.PATCH("/instances/:id", h.UpdateInstance)
protected.DELETE("/instances/:id", h.DeleteInstance)
```

One small refinement: you may eventually want to exclude `/metrics` itself from your HTTP application metrics to avoid monitoring noise.

---

# 8. Add PaaS instance metrics

Now add one business metric.

In `metrics.go`:

```go
var (
	InstanceOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "paas_instance_operations_total",
			Help: "Total number of instance operations.",
		},
		[]string{"operation", "result"},
	)
)
```

Register it:

```go
func RegisterMetrics() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(InstanceOperationsTotal)
}
```

Then in your handler:

```go
func (h *Handler) CreateInstance(c *gin.Context) {
	// ...

	instance, err := h.service.CreateInstance(...)
	if err != nil {
		observability.InstanceOperationsTotal.
			WithLabelValues("create", "error").
			Inc()

		// return error
		return
	}

	observability.InstanceOperationsTotal.
		WithLabelValues("create", "success").
		Inc()

	// return instance
}
```

Do the same for:

```text
create
update
delete
```

Don't put:

```text
instance_id
user_id
email
```

into these labels.

---

# 9. Add instance creation latency

This is especially useful for your PaaS.

```go
InstanceCreationDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name: "paas_instance_creation_duration_seconds",
		Help: "Time required to create an instance.",
		Buckets: []float64{
			1,
			2,
			5,
			10,
			20,
			30,
			60,
			120,
		},
	},
)
```

Register it:

```go
prometheus.MustRegister(InstanceCreationDuration)
```

Then:

```go
func (h *Handler) CreateInstance(c *gin.Context) {
	start := time.Now()

	instance, err := h.service.CreateInstance(...)

	observability.InstanceCreationDuration.
		Observe(time.Since(start).Seconds())

	if err != nil {
		// ...
		return
	}

	// ...
}
```

Now you can measure:

```text
p50 instance creation
p95 instance creation
p99 instance creation
```

---

# 10. Deploy your PaaS API

Your Kubernetes Service needs a named port because the `ServiceMonitor` will reference it.

Example:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: paas-api
  namespace: paas
  labels:
    app: paas-api
spec:
  selector:
    app: paas-api

  ports:
    - name: http
      port: 8080
      targetPort: 8080
```

Make sure your Deployment uses the same label:

```yaml
spec:
  template:
    metadata:
      labels:
        app: paas-api
```

---

# 11. Test `/metrics` inside Kubernetes

Before configuring Prometheus, test the endpoint itself.

```bash
kubectl run curl \
  --rm -it \
  --image=curlimages/curl \
  --restart=Never \
  -- sh
```

Inside:

```bash
curl http://paas-api.paas.svc.cluster.local:8080/metrics
```

You should see:

```text
paas_http_requests_total...
paas_http_request_duration_seconds...
paas_instance_operations_total...
go_goroutines...
process_cpu_seconds_total...
...
```

If this doesn't work, **fix it before touching Prometheus**.

---

# 12. Create the ServiceMonitor

Create:

```text
paas-api-servicemonitor.yaml
```

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: paas-api
  namespace: observability
spec:
  namespaceSelector:
    matchNames:
      - paas

  selector:
    matchLabels:
      app: paas-api

  endpoints:
    - port: http
      path: /metrics
      interval: 15s
```

Apply:

```bash
kubectl apply -f paas-api-servicemonitor.yaml
```

Because we set:

```yaml
serviceMonitorSelectorNilUsesHelmValues: false
```

the Prometheus instance should be able to discover this ServiceMonitor.

---

# 13. Verify Prometheus

Go to:

```text
Prometheus → Status → Targets
```

Find your PaaS API.

It should show:

```text
paas-api    UP
```

Then query:

```promql
paas_http_requests_total
```

You should see your API metrics.

---

# 14. Practice useful PromQL

### Request rate

```promql
sum(rate(paas_http_requests_total[5m]))
```

### Requests by route

```promql
sum by (route) (
  rate(paas_http_requests_total[5m])
)
```

### 5xx rate

```promql
sum(rate(paas_http_requests_total{status=~"5.."}[5m]))
```

### Request count

```promql
increase(paas_http_requests_total[1h])
```

### p95 latency

```promql
histogram_quantile(
  0.95,
  sum by (le) (
    rate(paas_http_request_duration_seconds_bucket[5m])
  )
)
```

### Instance creation rate

```promql
sum(
  rate(paas_instance_operations_total{
    operation="create",
    result="success"
  }[5m])
)
```

---

# 15. Grafana dashboard

Create one dashboard:

```text
PaaS API — Dev
```

Start with 4 panels.

### Request rate

```promql
sum(rate(paas_http_requests_total[5m]))
```

### Error rate

```promql
sum(rate(paas_http_requests_total{status=~"5.."}[5m]))
```

### P95 latency

```promql
histogram_quantile(
  0.95,
  sum by (le) (
    rate(paas_http_request_duration_seconds_bucket[5m])
  )
)
```

### Instance creation rate

```promql
sum(
  rate(paas_instance_operations_total{
    operation="create"
  }[5m])
)
```

That's enough for the first dashboard.

---

# 16. Add Loki

Once Prometheus/Grafana works, add Loki.

For Phase 1, keep it simple:

```text
Kubernetes stdout/stderr
          │
          ▼
      Grafana Alloy
          │
          ▼
         Loki
          │
          ▼
       Grafana
```

Add the Grafana Helm repository:

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

For a dev environment, install Loki with a simple configuration:

```yaml
# loki-values.yaml

deploymentMode: SingleBinary

loki:
  auth_enabled: false

  commonConfig:
    replication_factor: 1

  storage:
    type: filesystem

singleBinary:
  replicas: 1

  persistence:
    enabled: true
    size: 10Gi

gateway:
  enabled: false

read:
  replicas: 0

write:
  replicas: 0

backend:
  replicas: 0
```

Then:

```bash
helm install loki \
  grafana/loki \
  --namespace observability \
  -f loki-values.yaml
```

For **Phase 1/dev**, filesystem storage is fine for learning. Don't copy this configuration directly into production.

---

# 17. Install Grafana Alloy

Alloy will collect Kubernetes container logs.

Create:

```text
alloy-values.yaml
```

```yaml
alloy:
  configMap:
    create: true

    content: |
      discovery.kubernetes "pods" {
        role = "pod"
      }

      discovery.relabel "pod_logs" {
        targets = discovery.kubernetes.pods.targets

        rule {
          source_labels = ["__meta_kubernetes_namespace"]
          target_label  = "namespace"
        }

        rule {
          source_labels = ["__meta_kubernetes_pod_name"]
          target_label  = "pod"
        }

        rule {
          source_labels = ["__meta_kubernetes_pod_container_name"]
          target_label  = "container"
        }

        rule {
          source_labels = ["__meta_kubernetes_pod_label_app"]
          target_label  = "app"
        }
      }

      loki.source.kubernetes "pods" {
        targets    = discovery.relabel.pod_logs.output
        forward_to = [loki.write.default.receiver]
      }

      loki.write "default" {
        endpoint {
          url = "http://loki.observability.svc.cluster.local/loki/api/v1/push"
        }
      }
```

Install:

```bash
helm install alloy \
  grafana/alloy \
  --namespace observability \
  -f alloy-values.yaml
```

Check:

```bash
kubectl get pods -n observability
```

---

# 18. Add Loki to Grafana

In Grafana:

```text
Connections
   ↓
Data sources
   ↓
Add data source
   ↓
Loki
```

Use the internal Loki URL:

```text
http://loki.observability.svc.cluster.local
```

Then test the connection.

---

# 19. Generate and query logs

Your sample app can simply write:

```go
log.Printf("application started")
log.Printf("processing request")
log.Printf("something failed")
```

Kubernetes captures stdout/stderr.

Then in Grafana's **Explore**:

Select Loki and try:

```logql
{namespace="paas"}
```

Or:

```logql
{namespace="paas", app="paas-api"}
```

You should see your container logs.

---

# 20. Phase 1 test

Now deliberately do:

```bash
kubectl delete pod <your-paas-pod> -n paas
```

Then observe:

### Prometheus

```promql
up
```

### Kubernetes metrics

```promql
kube_pod_container_status_restarts_total
```

### Grafana

Look at the dashboard.

### Loki

Search:

```logql
{namespace="paas"}
```

You should be able to answer:

> **What happened to my application?**

using both metrics and logs.

---

# Phase 1 final state

You should end up with:

```text
                    Kubernetes
                         │
          ┌──────────────┼──────────────┐
          │              │              │
       PaaS API       Sample App       CNPG
          │              │
          │              │ logs
          │              ▼
          │            Alloy
          │              │
          │              ▼
          │             Loki
          │              │
          │              │
          ▼              ▼
      Prometheus ──────► Grafana
          │
          ▼
       PromQL
```

### What you should know before moving to Phase 2

You should be comfortable with:

* Prometheus scraping
* ServiceMonitor
* `/metrics`
* Counter/Gauge/Histogram
* Labels and cardinality
* `rate()` and `increase()`
* Basic PromQL
* Grafana dashboards
* Loki
* Basic LogQL
* Metrics vs logs
* Kubernetes pod/container telemetry

**Then stop.** Don't implement user-facing logs or audit logging yet.

Phase 2 will take this foundation and turn it into your **actual internal monitoring system**, including persistent Prometheus/Grafana storage, Kubernetes + PaaS + CNPG dashboards, alerts, and operator-focused health monitoring.
