# Configuring a Kubernetes API with a Kubebuilder Operator

## Overview

The platform consists of two components:

* **API**: Receives user requests and manages custom resources.
* **Operator**: Watches those custom resources and reconciles them into Kubernetes objects (for example, CloudNativePG clusters).

```
Client
   |
   v
API
   |
   | Create/List/Update/Delete
   v
PostgreSQL (Custom Resource)
(database.example.com/v1)
   |
   v
Kubebuilder Operator
   |
   v
CloudNativePG Cluster
   |
   v
PostgreSQL Database
```

The API never manages Pods, StatefulSets, or Services directly. It only interacts with the Kubernetes API by creating and reading custom resources.

---

# Required Kubernetes Resources

## 1. Namespace

Deploy the API into its own namespace.

Example:

```
dbaas-api
```

---

## 2. ServiceAccount

Create a dedicated ServiceAccount for the API.

Avoid using the `default` ServiceAccount.

Example:

```
ServiceAccount
Name: dbaas-api
Namespace: dbaas-api
```

Configure the Deployment to use it:

```yaml
spec:
  template:
    spec:
      serviceAccountName: dbaas-api
```

---

## 3. RBAC

The API requires Kubernetes permissions to access the custom resources.

The Role must include permissions such as:

* get
* list
* watch
* create
* update
* patch
* delete

for the custom resource.

Example:

```
API Group:
database.example.com

Resource:
postgresqls
```

---

## 4. Role Namespace

A common mistake is placing the Role in the API namespace.

RBAC is evaluated in the namespace that owns the resource.

If the API accesses:

```
database-system
```

then the Role and RoleBinding must exist in:

```
database-system
```

even though the ServiceAccount lives in:

```
dbaas-api
```

This configuration is valid:

```
ServiceAccount
    namespace: dbaas-api

Role
    namespace: database-system

RoleBinding
    namespace: database-system
```

---

## 5. RoleBinding

Bind the Role to the API ServiceAccount.

The subject references the ServiceAccount in the API namespace.

Example:

```
RoleBinding
    namespace: database-system

Subject:
    ServiceAccount
    namespace: dbaas-api
```

---

# Deployment

The Deployment should include:

* ServiceAccount
* imagePullSecrets (if using a private registry)
* resource requests and limits
* readiness probe
* liveness probe

Example health endpoint:

```
GET /health
```

---

# Service

Expose the application internally using a ClusterIP Service.

Example:

```
Service Port:
8080

Target Port:
8080
```

or expose port 8080 directly if preferred.

---

# Ingress

Expose the API through an Ingress.

Typical flow:

```
Internet
    |
    v
Ingress (8080/443)
    |
    v
Service
    |
    v
API Pod
```

TLS should terminate at the Ingress controller.

---

# Verifying Access

Confirm the API can access the custom resource:

```
kubectl auth can-i \
  list postgresqls.database.example.com \
  --namespace=database-system \
  --as=system:serviceaccount:dbaas-api:dbaas-api
```

Expected output:

```
yes
```

---

# Common Issues

## 403 Forbidden

```
cannot list resource "postgresqls"
```

Usually caused by:

* Role in the wrong namespace
* Missing RoleBinding
* Incorrect API group
* Incorrect resource name

---

## 503 from Ingress

Usually indicates:

* Service has no endpoints
* Pods are not Ready
* Service points to the wrong port

---

## ImagePullBackOff

Usually indicates:

* Missing `imagePullSecrets`
* Invalid registry credentials
* Image does not exist

---

# Recommended Architecture

```
Client
    |
    v
REST API
    |
    v
Kubernetes Custom Resource
(database.example.com/v1)
    |
    v
Kubebuilder Operator
    |
    v
CloudNativePG
    |
    v
PostgreSQL Cluster
```

This keeps responsibilities clearly separated:

* **API** implements business logic and exposes REST endpoints.
* **Operator** reconciles custom resources into Kubernetes resources.
* **CloudNativePG** manages PostgreSQL clusters.
* **Kubernetes** provides scheduling, networking, and RBAC.


## Argo CD application should be:

For your operator:

```yaml
destination:
  namespace: dbaas-operator
```

For your API:

```yaml
destination:
  namespace: dbaas-api
```

For your database resources:

```yaml
destination:
  namespace: database-system
```

---

Your naming should reflect ownership:

* `dbaas-operator` → your operator
* `cnpg-system` → CNPG operator
* `database-system` → database instances
* `dbaas-api` → API layer

That separation will make RBAC and GitOps management much clearer.
