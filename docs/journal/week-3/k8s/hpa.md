An **Horizontal Pod Autoscaler (HPA)** is exactly the right tool for your API. It automatically increases or decreases the number of API pods based on metrics like CPU or memory usage.

For your DBaaS platform, I would scale **only the API**. Your operator should typically run with a fixed number of replicas (usually 1, or 2 with leader election enabled).

---

## 1. Prerequisites

Make sure the Metrics Server is installed:

```bash
kubectl get deployment metrics-server -n kube-system
```

You already showed earlier that it exists, so you're good.

Also verify it works:

```bash
kubectl top pods -n dbaas-api
```

If you get CPU and memory values, HPA will work.

---

## 2. Ensure Deployment has resource requests

HPA based on CPU uses the CPU **requests** as the reference.

Your deployment already has:

```yaml
resources:
  requests:
    cpu: "250m"
    memory: "64Mi"
  limits:
    cpu: "500m"
    memory: "128Mi"
```

Perfect.

---

## 3. Create an HPA

Create:

```text
gitops/api/base/hpa.yaml
```

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler

metadata:
  name: dbaas-api
  namespace: dbaas-api

spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: dbaas-deploy

  minReplicas: 2
  maxReplicas: 10

  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

This means:

* minimum 2 pods
* maximum 10 pods
* keep CPU around 70%

---

## 4. Add it to Kustomize

`gitops/api/base/kustomization.yaml`

```yaml
resources:
  - namespace.yaml
  - serviceaccount.yaml
  - deployment.yaml
  - service.yaml
  - ingress.yaml
  - hpa.yaml
```

Argo CD will deploy it automatically.

---

## 5. Check it

```bash
kubectl get hpa -n dbaas-api
```

Example:

```text
NAME        REFERENCE                   TARGETS   MINPODS   MAXPODS   REPLICAS
dbaas-api   Deployment/dbaas-deploy     22%/70%   2         10        2
```

---

## 6. Watch scaling

```bash
watch kubectl get hpa,pods -n dbaas-api
```

When CPU increases:

```text
2 pods
↓

3 pods
↓

5 pods
↓

8 pods
```

When traffic drops:

```text
8

↓

5

↓

3

↓

2
```

---

# Best practices

For a production REST API, I'd configure:

```yaml
minReplicas: 2
maxReplicas: 20

behavior:
  scaleUp:
    stabilizationWindowSeconds: 0
    policies:
      - type: Percent
        value: 100
        periodSeconds: 30

  scaleDown:
    stabilizationWindowSeconds: 300
    policies:
      - type: Percent
        value: 20
        periodSeconds: 60
```

This allows the API to:

* scale up quickly under load
* scale down gradually to avoid oscillation

---

## Architecture

```
Internet
     │
Ingress
     │
Service
     │
+-----------+
| API Pod 1 |
+-----------+
| API Pod 2 |
+-----------+
     ▲
     │
 HorizontalPodAutoscaler
     │
 CPU > 70%
     │
     ▼
+-----------+
| API Pod 3 |
+-----------+
| API Pod 4 |
+-----------+
| API Pod 5 |
+-----------+
```

---

### One suggestion for your project

Since you're building a DBaaS platform, I'd put the HPA in its own file:

```text
gitops/
└── api/
    ├── base/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   ├── ingress.yaml
    │   ├── hpa.yaml
    │   └── kustomization.yaml
    └── overlays/
        └── production/
```

That keeps scaling configuration separate from the deployment and makes it easy to tune independently for different environments (for example, `minReplicas: 1` in development and `minReplicas: 2` or `3` in production).
