# K8s Basics

Before touching any commands, understand what you are building. A Kubernetes cluster has two roles:

```
┌────────────────────────────────────────────────────┐
│              Control Plane Node                    │
│                                                    │
│  kube-apiserver  ←── all communication goes here   │
│       │                                            │
│  etcd           ←── cluster database (state)       │
│  kube-scheduler ←── decides which node runs a pod  │
│  controller-manager ← reconciliation loops         │
└────────────────────────────┬───────────────────────┘
                             │ watches / instructs
          ┌──────────────────┴──────────────────┐
          │                                     │
┌─────────▼──────────┐             ┌────────────▼────────────┐
│   Worker Node 1    │             │   Worker Node 2         │
│                    │             │                         │
│  kubelet           │             │  kubelet                │
│  kube-proxy        │             │  kube-proxy             │
│  containerd        │             │  containerd             │
│  [your pods]       │             │  [your pods]            │
└────────────────────┘             └─────────────────────────┘
```

**Control Plane components:**

| Component | Responsibility |
|---|---|
| **kube-apiserver** | The front door — all kubectl commands, kubelet calls, and controllers talk to this REST API. |
| **etcd** | Distributed key-value store — the only place Kubernetes persists state. Losing etcd = losing the cluster. |
| **kube-scheduler** | Watches for unscheduled pods and assigns them to a node based on resource availability, taints, affinities. |
| **kube-controller-manager** | Runs controllers (Deployment controller, Node controller, etc.) — reconciliation loops that keep actual state matching desired state. |

**Worker Node components:**

| Component | Responsibility |
|---|---|
| **kubelet** | The agent on every node. Receives pod specs from the API server and ensures containers are running via the container runtime. |
| **kube-proxy** | Maintains network rules (iptables/IPVS) that implement Kubernetes Services — forwards traffic to the right pod. |
| **Container runtime** | Runs the actual containers. You will use **containerd**. |

## Concepts: Pods, Deployments, Services, Namespaces

### Pod

The smallest schedulable unit in Kubernetes. A pod wraps one or more containers that share the same network namespace (same IP) and storage volumes.

> Real-life analogy: A pod is like a shipping container. All goods (containers) inside share the same truck (network/storage). Kubernetes moves whole shipping containers, never individual items.

Key pod properties:
- **Ephemeral** — pods are mortal. They get new IPs on restart.
- **Never reschedule themselves** — that is a higher-level controller's job.

### Deployment

A Deployment manages a set of identical pod replicas and ensures the desired count is always running. It is the standard way to run stateless apps.

```
Deployment (desired: 3 replicas)
      │
      │ manages
      ▼
ReplicaSet
      │
      ├── Pod 1  (running)
      ├── Pod 2  (running)
      └── Pod 3  (crashed → controller creates Pod 4)
```

### Service

Pods have ephemeral IPs. A Service gives a **stable** IP and DNS name that load-balances across matching pods.

```
Client → Service (stable ClusterIP: 10.96.0.1)
              │
              ├── Pod A (10.0.1.5)
              ├── Pod B (10.0.1.6)
              └── Pod C (10.0.1.7)
```

Service types:
- `ClusterIP` — internal only (default)
- `NodePort` — exposes on a port of every node (30000–32767)
- `LoadBalancer` — provisions a cloud load balancer (not available in bare OpenStack without CCM)

### Namespace

A namespace is a logical partition inside a cluster. Resources in different namespaces are isolated by default for access control and quotas.

Built-in namespaces:
- `kube-system` — control plane components
- `kube-public` — publicly readable cluster info
- `default` — where your resources go if you don't specify