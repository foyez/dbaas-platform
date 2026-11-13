# Important differences in K8s

## Components vs Resources

<details>
<summary><strong>View Contents</strong></summary>

Components vs Resources is an important distinction in Kubernetes.

* **Components** are the **software processes** that make Kubernetes work.
* **Resources** are the **objects** that Kubernetes manages.

Think of it this way:

* **Components = workers** in a factory.
* **Resources = work orders** or items the workers process.

## Kubernetes components

These are binaries or services running on the control plane or worker nodes.

### Control plane components

| Component                | Purpose                                                       |
| ------------------------ | ------------------------------------------------------------- |
| API Server               | Entry point for all Kubernetes requests.                      |
| etcd                     | Stores the cluster state.                                     |
| Scheduler                | Chooses which node should run a Pod.                          |
| Controller Manager       | Runs controllers like Deployment, ReplicaSet, Node, Job, etc. |
| Cloud Controller Manager | Integrates with cloud providers (optional).                   |

### Worker node components

| Component                             | Purpose                                        |
| ------------------------------------- | ---------------------------------------------- |
| Kubelet                               | Runs Pods assigned to the node.                |
| Container Runtime (containerd, CRI-O) | Creates and manages containers.                |
| kube-proxy                            | Handles Service networking and load balancing. |

---

## Kubernetes resources

Resources are objects stored in the Kubernetes API (and persisted in etcd).

Examples:

| Resource              | Purpose                                      |
| --------------------- | -------------------------------------------- |
| Pod                   | Runs one or more containers.                 |
| Deployment            | Manages ReplicaSets and rolling updates.     |
| ReplicaSet            | Maintains a desired number of Pod replicas.  |
| Service               | Provides stable networking for Pods.         |
| ConfigMap             | Stores configuration data.                   |
| Secret                | Stores sensitive data.                       |
| Namespace             | Organises resources into logical groups.     |
| Job                   | Runs a task to completion.                   |
| CronJob               | Runs Jobs on a schedule.                     |
| StatefulSet           | Manages stateful applications.               |
| DaemonSet             | Ensures one Pod runs on each node.           |
| PersistentVolume      | Represents storage available to the cluster. |
| PersistentVolumeClaim | Requests storage for workloads.              |
| Ingress               | Defines HTTP/HTTPS routing rules.            |

---

## How they work together

Suppose you run:

```bash
kubectl apply -f deployment.yaml
```

The flow looks like this:

```
Deployment resource
        │
        ▼
API Server
        │
        ▼
etcd (stores Deployment)
        │
        ▼
Controller Manager
        │
Creates ReplicaSet resource
        │
        ▼
ReplicaSet Controller
        │
Creates Pod resources
        │
        ▼
Scheduler
        │
Assigns Pod to Node-1
        │
        ▼
Kubelet
        │
Starts container using containerd
```

Notice that:

* **Deployment**, **ReplicaSet**, and **Pod** are **resources**.
* **API Server**, **Scheduler**, **Controller Manager**, **Kubelet**, and **containerd** are **components** acting on those resources.

## Quick comparison

| Components                               | Resources                                   |
| ---------------------------------------- | ------------------------------------------- |
| Software processes                       | API objects                                 |
| Execute logic                            | Describe desired state                      |
| Examples: kubelet, scheduler, API server | Examples: Pod, Deployment, Service          |
| Run on nodes                             | Stored in etcd                              |
| Continuously reconcile state             | Represent the desired cluster configuration |

### Easy way to remember

* **Components** answer: **"Who does the work?"**
* **Resources** answer: **"What should exist?"**

For example:

* **Deployment** (resource): "I want 3 replicas of this application."
* **ReplicaSet controller** (part of the controller manager component): "I'll make sure there are 3 Pods."
* **Scheduler** (component): "I'll choose a node for each Pod."
* **Kubelet** (component): "I'll start and keep those Pods running on my node."

</details>

## Resource vs Controller

<details>
<summary><strong>View Contents</strong></summary>

Resource vs Controller is one of the most confusing parts of Kubernetes because **controllers and resources often have the same name**. They are related, but they are **not the same thing**.

## Resource vs Controller

A **resource** is a Kubernetes object stored in the API server (and persisted in etcd).

A **controller** is a program running inside the **kube-controller-manager** that watches resources and reconciles their actual state with the desired state.

For many resources, there is a controller with the same name.

| Resource   | Controller            | Responsibility                                      |
| ---------- | --------------------- | --------------------------------------------------- |
| Deployment | Deployment Controller | Creates and updates ReplicaSets.                    |
| ReplicaSet | ReplicaSet Controller | Creates and deletes Pods.                           |
| Job        | Job Controller        | Creates Pods until the Job completes.               |
| Node       | Node Controller       | Monitors node health and handles unreachable nodes. |

---

## Example: Deployment

### Step 1: You create a Deployment resource

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 3
```

This YAML becomes a **Deployment resource** in the API server.

Think of it as data:

```text
Deployment Resource
-------------------
Name: nginx
Replicas: 3
Image: nginx
```

---

### Step 2: Deployment Controller notices it

The Deployment Controller continuously watches Deployment resources.

It sees:

```text
Desired ReplicaSets: 1
Actual ReplicaSets: 0
```

So it creates a **ReplicaSet resource**.

---

### Step 3: ReplicaSet Controller notices

The ReplicaSet Controller watches ReplicaSet resources.

It sees:

```text
Desired Pods: 3
Actual Pods: 0
```

So it creates **three Pod resources**.

---

### Step 4: Scheduler

The Scheduler watches unscheduled Pods and assigns them to nodes.

---

### Step 5: Kubelet

The kubelet on each assigned node sees the Pod resource and starts its containers.

---

## Visual flow

```text
Deployment Resource
        │
        ▼
Deployment Controller
        │
Creates
        ▼
ReplicaSet Resource
        │
        ▼
ReplicaSet Controller
        │
Creates
        ▼
Pod Resources
        │
        ▼
Scheduler
        │
Assigns node
        ▼
Kubelet
        │
Runs containers
```

Notice the pattern:

* **Resources** are objects describing the desired state.
* **Controllers** watch those objects and take action to achieve that state.

---

## Why do they share the same name?

Because each controller is responsible for reconciling a particular resource.

For example:

* **Deployment resource** → "I want version 2 of my app with 3 replicas."
* **Deployment Controller** → "I'll create or update ReplicaSets to make that happen."

Similarly:

* **ReplicaSet resource** → "I want 3 Pods."
* **ReplicaSet Controller** → "I'll create or delete Pods until there are exactly 3."

---

## Think of it like a to-do list

Imagine a task manager application:

* **Task (resource):** "Buy milk."
* **Person (controller):** Reads the task and goes to the shop.

The task is just data. The person performs the work.

In Kubernetes:

* **Deployment resource** = the instruction.
* **Deployment Controller** = the worker that follows the instruction.

This separation is a core design principle of Kubernetes: **resources declare the desired state, and controllers continuously reconcile the actual state to match it.**

</details>

## Kubelet vs ReplicaSet Controller

<details>
<summary><strong>View Contents</strong></summary>

The **kubelet** and the **ReplicaSet Controller** both reconcile desired state, but at **different scopes**.

| ReplicaSet Controller                    | Kubelet                                |
| ---------------------------------------- | -------------------------------------- |
| Cluster-level controller                 | Node-level agent                       |
| Watches ReplicaSet and Pod resources     | Watches Pods assigned to its node      |
| Ensures the correct number of Pods exist | Ensures assigned Pods are running      |
| Creates/deletes Pod resources            | Starts, stops, and restarts containers |
| Doesn't run containers                   | Doesn't create replacement Pods        |

## Their responsibilities

### ReplicaSet Controller

Suppose you create:

```yaml
replicas: 3
```

The ReplicaSet Controller checks:

* Desired: **3 Pods**
* Actual: **2 Pods**

It creates another **Pod resource**.

It does **not** start containers.

---

### Kubelet

On a worker node, the kubelet checks the Pods assigned to it.

Suppose `Pod-A` is assigned to `worker-1`.

Desired:

```text
Pod-A should be running.
```

Actual:

```text
Container is not running.
```

The kubelet asks the container runtime (such as containerd) to start the container.

If the container crashes, the kubelet restarts it according to the Pod's restart policy.

---

## Example workflow

You create a Deployment with 3 replicas.

### 1. ReplicaSet Controller

Creates three Pod resources:

```text
Pod-A
Pod-B
Pod-C
```

### 2. Scheduler

Assigns them:

```text
Pod-A → worker-1
Pod-B → worker-2
Pod-C → worker-1
```

### 3. Kubelets

Each kubelet starts the Pods assigned to its node.

```text
worker-1 kubelet
    ├── Start Pod-A
    └── Start Pod-C

worker-2 kubelet
    └── Start Pod-B
```

---

## Scenario 1: A Pod is deleted

Someone runs:

```bash
kubectl delete pod pod-b
```

What happens?

* **ReplicaSet Controller:** "I only have 2 Pods, but I need 3." → Creates a new Pod.
* **Scheduler:** Assigns the new Pod to a node.
* **Kubelet:** Starts the new Pod on that node.

---

## Scenario 2: A container crashes

Suppose the application process inside `Pod-A` exits unexpectedly.

The Pod object still exists in the API server.

* **ReplicaSet Controller:** Does nothing, because there are still 3 Pod resources.
* **Kubelet:** Detects the container has stopped and restarts it (if the restart policy allows).

---

## Scenario 3: The entire node fails

Suppose `worker-2` goes offline.

Initially:

```text
worker-1
  Pod-A
  Pod-C

worker-2 ❌
  Pod-B
```

* The kubelet on `worker-2` is no longer running, so it can't restart or manage `Pod-B`.
* The control plane detects the node failure. After the node is marked unavailable and the old Pod is removed from the active set, the **ReplicaSet Controller** notices there are only 2 healthy Pods and creates a replacement.
* The Scheduler assigns the replacement Pod to another healthy node.
* The kubelet on that node starts it.

---

## The key difference

Think of the ReplicaSet Controller as asking:

> **"Do I have the right number of Pods in the cluster?"**

Think of the kubelet as asking:

> **"Are the Pods assigned to my node actually running?"**

So:

* **ReplicaSet Controller** manages the **existence and count** of Pods across the cluster.
* **Kubelet** manages the **execution and lifecycle** of those Pods on a specific node.

</details>

## Most Common Differnces

<details>
<summary><strong>View Contents</strong></summary>

# 1. Component vs Resource ⭐⭐⭐⭐⭐

**Component** = Software process that runs Kubernetes.

**Resource** = Object stored in the API server.

Example:

```
API Server          ← Component
Scheduler           ← Component
Kubelet             ← Component

Pod                 ← Resource
Deployment          ← Resource
Service             ← Resource
ConfigMap           ← Resource
```

A good rule is:

> **Components do the work. Resources describe the work.**

---

# 2. Resource vs Controller ⭐⭐⭐⭐⭐

This is probably the most confusing one.

Example:

```
Deployment (Resource)
```

is **not** the same as

```
Deployment Controller
```

The resource is just data:

```
Deployment
  replicas = 3
```

The controller watches that data and takes action.

```
Deployment Resource
        │
        ▼
Deployment Controller
        │
Creates
        ▼
ReplicaSet Resource
```

---

# 3. Deployment vs ReplicaSet ⭐⭐⭐⭐⭐

People often think they're interchangeable.

### ReplicaSet

Maintains the number of Pods.

```
Need 3 Pods
```

### Deployment

Manages ReplicaSets.

It provides

* Rolling updates
* Rollbacks
* Revision history

```
Deployment
      │
      ▼
ReplicaSet
      │
      ▼
Pods
```

A Deployment almost always owns a ReplicaSet.

---

# 4. ReplicaSet vs Kubelet ⭐⭐⭐⭐⭐

ReplicaSet asks:

> "Do enough Pods exist?"

Kubelet asks:

> "Are my assigned Pods actually running?"

Example:

Container crashes

ReplicaSet:

```
Still have 3 Pods.

Do nothing.
```

Kubelet:

```
Container exited.

Restart it.
```

---

# 5. ReplicaSet vs Scheduler ⭐⭐⭐⭐

ReplicaSet:

Creates Pods.

Scheduler:

Assigns those Pods to nodes.

```
ReplicaSet
     │
Creates Pod
     │
Scheduler
     │
Assigns worker-2
```

---

# 6. Pod vs Container ⭐⭐⭐⭐⭐

Very common interview question.

Pod

```
Smallest deployable Kubernetes object.
```

Container

```
Actual application process.
```

One Pod may contain

```
Pod
 ├── nginx container
 └── sidecar container
```

Containers never exist in Kubernetes without a Pod.

---

# 7. Pod vs Node ⭐⭐⭐⭐⭐

Node

```
A machine.
```

Pod

```
Runs on a machine.
```

```
Node
 ├── Pod A
 ├── Pod B
 └── Pod C
```

---

# 8. Service vs Ingress ⭐⭐⭐⭐

Service

```
Internal networking.
```

Ingress

```
External HTTP/HTTPS access.
```

```
Internet
     │
Ingress
     │
Service
     │
Pods
```

---

# 9. ConfigMap vs Secret ⭐⭐⭐⭐

Both store configuration.

ConfigMap

```
Non-sensitive
```

Secret

```
Sensitive
```

Passwords belong in Secrets.

Configuration belongs in ConfigMaps.

---

# 10. StatefulSet vs Deployment ⭐⭐⭐⭐

Deployment

Pods are interchangeable.

```
web-1
web-2
web-3
```

Delete one?

Create another.

No identity matters.

StatefulSet

Each Pod has a stable identity.

```
mysql-0
mysql-1
mysql-2
```

`mysql-0` is always `mysql-0`.

---

# 11. Job vs CronJob ⭐⭐⭐⭐

Job

Run once.

```
Backup database
```

CronJob

Run on a schedule.

```
Every night at 2 AM
```

---

# 12. etcd vs API Server ⭐⭐⭐⭐⭐

API Server

Gateway.

Everyone talks to it.

etcd

Database.

Stores cluster state.

```
kubectl
     │
API Server
     │
etcd
```

Never access etcd directly unless you have a specific administrative reason.

---

# 13. kube-proxy vs Service ⭐⭐⭐⭐

Service

An API object.

```
ClusterIP
```

kube-proxy

The component that programs the networking rules (using mechanisms such as iptables or IPVS) so traffic sent to a Service reaches one of its backing Pods.

```
Service
     │
kube-proxy
     │
iptables/IPVS rules
     │
Pods
```

---

# 14. Desired State vs Actual State ⭐⭐⭐⭐⭐

Desired

```
3 Pods
```

Actual

```
2 Pods
```

Controllers continuously reconcile them.

```
Desired
      │
Compare
      │
Actual
      │
Create missing Pod
```

---

# 15. API Object vs YAML ⭐⭐⭐⭐

Many beginners think YAML is the resource.

Actually:

```
deployment.yaml
```

is only the **definition**.

After

```bash
kubectl apply -f deployment.yaml
```

the API Server creates the real object.

```
YAML
   │
kubectl apply
   │
API Server
   │
Deployment Resource
```

The YAML is a way to declare what you want; the resource is what exists in the cluster.

---

# 16. kubectl vs Kubelet ⭐⭐⭐⭐⭐

People confuse these constantly.

kubectl

```
CLI tool
Runs on your laptop
```

Kubelet

```
Agent
Runs on every worker node
```

```
kubectl
      │
API Server
      │
Kubelet
      │
containerd
```

---

## A mental model that ties everything together

```
YAML
 │
 ▼
kubectl
 │
 ▼
API Server
 │
 ▼
etcd (stores resources)
 │
 ├── Deployment Controller
 ├── ReplicaSet Controller
 ├── Node Controller
 └── Job Controller
        │
        ▼
   Pod Resources
        │
        ▼
   Scheduler
        │
        ▼
   Node
        │
        ▼
   Kubelet
        │
        ▼
   containerd
        │
        ▼
   Containers
```

If you're preparing for certifications like CKA or CKS, these are the distinctions that come up repeatedly. Once you're comfortable with this flow, many Kubernetes troubleshooting scenarios become much easier to reason about because you can identify **which component** is responsible for **which resource** at each stage.

</details>