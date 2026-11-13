# Week 2 – Developing a Platform-as-a-Service (PaaS) Product on STACKIT Kubernetes Engine (SKE)

This looks like the requirements for a project where you need to build a **simple Platform as a Service (PaaS)** on top of Kubernetes. The goal is not to build something as complex as Heroku or OpenShift, but to demonstrate that you understand Kubernetes Operators, Custom Resources, and how to automate application deployment.

Here's what each requirement means.

### 1. PaaS Product Implementation

**"Design and technical implementation of a simple PaaS service."**

This is the actual product you're building.

A PaaS lets users deploy or manage applications without manually dealing with Kubernetes resources.

For example, your PaaS could allow a user to create a simple application by writing:

```yaml
apiVersion: mypaas.com/v1
kind: WebApp
metadata:
  name: my-app
spec:
  image: nginx
  replicas: 2
```

Instead of the user creating Deployments, Services, ConfigMaps, etc., your PaaS creates them automatically.

You need to:

* Decide what your product does.
* Define the features.
* Implement it.

Examples:

* Web application hosting platform
* Database provisioning service
* Redis deployment service
* Minecraft server deployment
* Blog hosting platform

---

### 2. Operator Deployment

**"Provisioning of a Kubernetes Operator."**

A Kubernetes Operator is a controller that watches Custom Resources.

You'll probably use:

* Kubebuilder
* Operator SDK

The operator should:

* Watch your custom resource
* Create Kubernetes resources
* Update them when the CR changes
* Delete resources when the CR is deleted

Example:

User creates:

```yaml
kind: WebApp
```

Operator creates:

* Deployment
* Service
* ConfigMap
* Secret

---

### 3. Product Component Management

**"Utilization of Custom Resources (CRs) for provisioning and managing product components."**

This is the heart of the assignment.

You define your own Kubernetes object.

Example:

```yaml
apiVersion: mypaas.com/v1
kind: Database
metadata:
  name: postgres
spec:
  version: "16"
  storage: 10Gi
```

The Operator reads this CR and provisions:

* StatefulSet
* PersistentVolumeClaim
* Service
* Secret

The user only edits the CR.

If they change

```yaml
storage: 20Gi
```

the Operator updates the database.

---

### 4. Connectivity

**"Documentation and demonstration of connecting to and using the PaaS product."**

You need to show someone can actually use your platform.

For example:

1. Install operator

```bash
kubectl apply -f operator.yaml
```

2. Create a CR

```bash
kubectl apply -f webapp.yaml
```

3. Operator creates everything automatically

4. Connect

```bash
kubectl port-forward service/my-app 8080:80
```

Visit

```
http://localhost:8080
```

Or expose it using an Ingress.

---

# What your final project should contain

## 1. Kubernetes Operator

For example:

```
SimplePaaS Operator
```

Built using:

* Operator SDK
* Kubebuilder

---

## 2. Custom Resource Definition (CRD)

Example:

```yaml
kind: WebApp
```

or

```yaml
kind: Database
```

---

## 3. Controller Logic

The controller watches CRs and creates resources like:

```
WebApp
   │
   ▼
Deployment
Service
ConfigMap
Secret
```

---

## 4. Example Custom Resource

Example:

```yaml
apiVersion: mypaas.com/v1
kind: WebApp
metadata:
  name: hello
spec:
  image: nginx
  replicas: 3
```

---

## 5. Demonstration

Show:

* Install operator
* Create CR
* Kubernetes resources appear
* Access the application
* Delete CR
* Resources disappear

---

## 6. Documentation

Explain:

* Architecture
* Installation
* Usage
* Example CR
* How to connect to the application

---

## A possible architecture

```
             User

               │

kubectl apply -f webapp.yaml

               │

        Custom Resource

               │

        Kubernetes API

               │

           Operator

               │

      ┌────────┴────────┐
      │                 │
 Deployment         Service
      │                 │
      └────────┬────────┘
               │
          Running App
```

---

## A good project idea

A **Simple Web App Hosting Platform** is a manageable choice that covers all the requirements.

* Users create a `WebApp` CR.
* The Operator deploys a container image.
* It creates a Service (and optionally an Ingress).
* Users can scale the app by changing `replicas` in the CR.
* The Operator reconciles the changes automatically.
* You demonstrate access to the application through port-forwarding or an Ingress.

This project is straightforward to implement, clearly demonstrates Operator and CRD concepts, and satisfies all four assignment requirements without becoming overly complex.


## Stateless vs Stateful Projects

### Stateless Project (Recommended for beginners)

A stateless application does **not** keep data internally. If the pod dies, Kubernetes can simply start another one.

Examples:

* Nginx web server
* React frontend
* Flask REST API
* Node.js API
* Static website
* URL shortener (without persistence)

Kubernetes resources:

* Deployment
* Service
* ConfigMap
* Ingress

Example architecture:

```
User
   |
Ingress
   |
Service
   |
Deployment
   |
Pods
```

### Advantages

* Easier Operator logic
* No Persistent Volumes
* No StatefulSets
* Easier scaling
* Easier debugging

---

### Stateful Project

A stateful application stores important data.

Examples:

* PostgreSQL
* MongoDB
* Redis (with persistence)
* Kafka
* Elasticsearch

Resources involved:

* StatefulSet
* PersistentVolumeClaim (PVC)
* StorageClass
* Headless Service
* Secrets

Architecture:

```
User
   |
Service
   |
StatefulSet
   |
PVC
   |
Persistent Disk
```

---

### Why is it harder?

Imagine PostgreSQL.

If Pod A dies:

```
Pod dies
↓
New Pod starts
↓
Database files must still exist
↓
PVC must be mounted
↓
Same hostname
↓
Same identity
```

The Operator has to manage much more than just creating pods.

---

# Difficulty Comparison

| Feature              | Stateless | Stateful     |
| -------------------- | --------- | ------------ |
| Operator logic       | ⭐ Easy    | ⭐⭐⭐⭐ Hard    |
| Kubernetes resources | Few       | Many         |
| Storage              | No        | Yes          |
| Scaling              | Easy      | Complex      |
| Debugging            | Easy      | Hard         |
| Good first Operator? | ✅ Yes     | ⚠️ Not ideal |

---

# Which teaches you more?

A stateful Operator teaches:

* StatefulSets
* Persistent Volumes
* StorageClasses
* Data persistence
* Backup considerations
* More reconciliation logic

But it also takes considerably more time.

---

# Projects from easiest to hardest

### 1. Web App Hosting Platform ⭐⭐

User creates:

```yaml
kind: WebApp
```

Operator creates:

* Deployment
* Service
* Ingress
* ConfigMap

Features:

* image
* replicas
* port
* environment variables

Learning:

* Deployments
* Services
* Reconciliation
* Scaling

Difficulty: **2/5**

---

### 2. Static Website Hosting ⭐⭐

User provides:

* image
* domain
* replicas

Operator deploys:

```
Nginx
Service
Ingress
```

Similar to a simplified Heroku or Netlify.

Difficulty: **2/5**

---

### 3. Redis Platform ⭐⭐⭐

User writes:

```yaml
kind: RedisInstance
```

Operator creates:

* StatefulSet
* PVC
* Service

Learning:

* Storage
* StatefulSets
* Persistent volumes

Difficulty: **3.5/5**

---

### 4. PostgreSQL Platform ⭐⭐⭐⭐

Operator provisions:

* PostgreSQL
* PVC
* Secrets
* Service

Learning:

* Database operators
* StatefulSets
* Credentials
* Storage

Difficulty: **4/5**

---

### 5. Minecraft Server Platform ⭐⭐⭐

Very uncommon.

User creates:

```yaml
kind: MinecraftServer
```

Operator provisions:

* Deployment (or StatefulSet)
* Service
* PVC
* ConfigMap

Users connect with the Minecraft client.

Interesting because you manage a real application instead of a typical web server.

Difficulty: **3.5/5**

---

### 6. Jenkins Platform ⭐⭐⭐

Operator deploys Jenkins automatically.

Creates:

* Deployment
* Service
* PVC
* Admin Secret

Good learning because Jenkins requires persistent storage but is easier than a database.

Difficulty: **3/5**

---

### 7. ML Notebook Platform ⭐⭐⭐⭐

User creates:

```yaml
kind: Notebook
```

Operator deploys a Jupyter Notebook.

Features:

* CPU
* RAM
* Image
* Storage

Very practical and less common.

Difficulty: **4/5**

---

### 8. API Hosting Platform ⭐⭐⭐

Users deploy REST APIs.

Example:

```yaml
kind: APIService
spec:
  image: my-api:v1
  replicas: 3
  autoscaling: true
```

Operator creates:

* Deployment
* Service
* HPA
* ConfigMap
* Ingress

Learning:

* Autoscaling
* Resource limits
* Ingress
* Services

Difficulty: **3/5**

---

# My recommendation

If your goals are to **learn a lot**, keep the project **manageable**, and avoid something that every student builds, I'd suggest one of these:

| Project                   | Learning | Difficulty | Uniqueness |
| ------------------------- | -------- | ---------- | ---------- |
| API Hosting Platform      | ⭐⭐⭐⭐     | ⭐⭐⭐        | ⭐⭐⭐⭐       |
| Minecraft Server Platform | ⭐⭐⭐⭐     | ⭐⭐⭐        | ⭐⭐⭐⭐⭐      |
| Redis Platform            | ⭐⭐⭐⭐⭐    | ⭐⭐⭐⭐       | ⭐⭐⭐        |
| Jupyter Notebook Platform | ⭐⭐⭐⭐⭐    | ⭐⭐⭐⭐       | ⭐⭐⭐⭐⭐      |
| Web App Hosting Platform  | ⭐⭐⭐      | ⭐⭐         | ⭐⭐         |

For a first Kubernetes Operator project, I'd recommend an **API Hosting Platform**. It remains mostly stateless—so you avoid the complexity of persistent storage—but introduces more advanced Kubernetes concepts such as ConfigMaps, Secrets, Ingress, resource requests and limits, and optionally Horizontal Pod Autoscalers. That gives you a richer learning experience than a basic web-app deployment while staying achievable within a typical course project. Once you're comfortable with Operators, you can extend it later with stateful components like databases or Redis.
