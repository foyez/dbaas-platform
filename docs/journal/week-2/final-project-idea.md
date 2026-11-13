# Week 2 Project Notes — PostgreSQL PaaS Using Custom Operator + CloudNativePG + PgBouncer

## Project Vision

Build a **PostgreSQL Database as a Service (DBaaS)** platform on top of STACKIT Kubernetes Engine (SKE).

The goal is not to build a complete PostgreSQL database operator from scratch. Instead:

* Use **CloudNativePG (CNPG)** for production-grade PostgreSQL management (replication, failover, recovery).
* Build your own **PaaS Operator** that provides a simplified product interface.
* Your Operator manages the lifecycle of PostgreSQL products through Custom Resources.
* Your Operator also manages one Kubernetes component directly (PgBouncer) to demonstrate StatefulSet, PVC, Service, and reconciliation concepts.

This project follows an **operator composition pattern**, where multiple Kubernetes Operators work together, with each operator responsible for a specific layer. Your custom PaaS Operator acts as a higher-level platform layer that creates and manages CNPG resources, while CloudNativePG handles the underlying PostgreSQL database lifecycle. This allows you to build a simplified DBaaS product without reimplementing complex database management features.

Final architecture:

```
                    User / Developer

                           |
                           |
                  PostgreSQL Custom Resource
                  (Your PaaS Product API)

                           |
                           |

              Your PostgreSQL PaaS Operator
              (Kubebuilder Controller)

                    /              \
                   /                \

                  /                  \

        CloudNativePG Cluster       PgBouncer
             CR                      StatefulSet
              |                         |
              |                         |
              v                         v

     CNPG PostgreSQL HA          Connection Pooler
     PostgreSQL Cluster          StatefulSet
                                  PVC
                                  Service

                           |
                           |
                     Kubernetes Cluster

                           |
                           |
                       STACKIT SKE
```

---

# 1. Infrastructure Preparation (SKE + Terraform)

## Objective

Create the Kubernetes foundation where your PaaS platform runs.

## Tasks

### Understand Kubernetes Architecture

Document:

## Control Plane

Responsible for cluster management.

Components:

* API Server
* Scheduler
* Controller Manager
* etcd

Architecture:

```
                 Control Plane

                     API Server

              /        |        \

          Scheduler  Controllers  etcd


                     |

              Worker Nodes

                     |

                   kubelet

                     |

                    Pods
```

---

## Worker Nodes

Responsible for running workloads:

* Operator pod
* CNPG operator
* PostgreSQL pods
* PgBouncer pods

---

# 2. Create SKE Cluster Using Terraform

## Goal

Automate infrastructure provisioning.

Repository:

```
project/

├── terraform/

│   ├── provider.tf
│   ├── cluster.tf
│   ├── variables.tf
│   └── outputs.tf

```

Terraform creates:

```
STACKIT Cloud

        |

        |

SKE Kubernetes Cluster

        |

---------------------
|                   |

Worker Node 1   Worker Node 2
```

---

## Verification

After deployment:

```bash
kubectl get nodes
```

Expected:

```
NAME          STATUS

worker-1      Ready

worker-2      Ready
```

---

# 3. Install CloudNativePG Operator

## Objective

Install the database engine operator.

CNPG responsibilities:

* PostgreSQL lifecycle
* Replication
* Failover
* Recovery
* Backup support

Your Operator will not manage these low-level database operations.

---

Installation flow:

```
Kubernetes Cluster

        |

        |

CloudNativePG Operator

        |

        |

Watches Cluster CR

        |

        |

Creates PostgreSQL Cluster
```

---

Verify:

```bash
kubectl get pods -n cnpg-system
```

Expected:

```
cnpg-controller-manager   Running
```

---

# 4. Design Your PaaS Custom Resource

Your users should not interact directly with CNPG.

They create your resource:

Example:

```yaml
apiVersion: database.example.com/v1
kind: PostgreSQL

metadata:
  name: customer-db

spec:

  version: "16"

  replicas: 3

  storage: 10Gi

  database: customer

  username: customer-admin

  connectionPooler:
    enabled: true
```

---

## Why create your own CR?

Because your CR is your PaaS product interface.

User thinks:

```
I need a PostgreSQL database
```

not:

```
I need a CNPG Cluster resource
```

Your Operator hides complexity.

---

# 5. Build PostgreSQL PaaS Operator Using Kubebuilder

## Objective

Create your own Kubernetes Operator.

Technology:

* Go
* Kubebuilder
* controller-runtime

---

## Initialize Operator

Example:

```bash
kubebuilder init \
--domain example.com \
--repo github.com/user/postgres-platform
```

---

Create API:

```bash
kubebuilder create api \
--group database \
--version v1 \
--kind PostgreSQL
```

Generated structure:

```
postgres-platform/

├── api/

│   └── v1/

│       └── postgresql_types.go


├── controllers/

│   └── postgresql_controller.go


├── config/

```

---

# 6. Define PostgreSQL CRD

File:

```
api/v1/postgresql_types.go
```

Define:

## Spec

Desired state:

```go
type PostgreSQLSpec struct {

    Version string

    Replicas int32

    Storage string

    Database string

    Username string

    ConnectionPooler bool
}
```

---

## Status

Observed state:

```go
type PostgreSQLStatus struct {

    Phase string

    ReadyInstances int32

    ConnectionEndpoint string

    PgBouncerStatus string
}
```

---

Generate CRD:

```bash
make manifests
```

Result:

```
CustomResourceDefinition

postgresqls.database.example.com
```

---

# 7. Implement Reconciler Pattern

## Core Operator Logic

Your controller follows:

```
Observe

   |

Compare Desired vs Current State

   |

Create / Update Resources

   |

Update Status

   |

Repeat
```

---

Your Reconcile flow:

```
PostgreSQL CR created

          |

          |

Your Operator detects event

          |

          |

Create CNPG Cluster CR

          |

          |

Create PgBouncer StatefulSet

          |

          |

Create Service

          |

          |

Create Secrets

          |

          |

Update PostgreSQL Status
```

---

# 8. Manage CNPG Cluster Resource

Your Operator creates:

```yaml
kind: Cluster

metadata:

  name: customer-db

spec:

  instances: 3

  storage:
    size: 10Gi

```

Your Operator does:

```
Your PostgreSQL CR

        |

        |

Creates

        |

        |

CNPG Cluster CR

        |

        |

CNPG Operator

        |

        |

PostgreSQL HA Cluster
```

---

# 9. Manage Your Own Stateful Component: PgBouncer

## Purpose

Demonstrate direct Kubernetes resource management.

PgBouncer provides:

* Connection pooling
* Stable database endpoint
* Reduced database connection overhead

---

Your Operator creates:

## StatefulSet

```
pgbouncer-0
```

Properties:

* Stable identity
* Managed lifecycle

---

## PVC

Storage:

```
PgBouncer Pod

       |

       |

PersistentVolumeClaim

       |

       |

Storage
```

---

## Service

Example:

```
customer-db-pgbouncer

        |

        |

Applications connect here
```

---

# 10. Resource Ownership

Important Kubernetes concept.

Your Operator should set:

```
OwnerReference
```

Example:

```
PostgreSQL CR

      |
      |
 owns
      |
      |

CNPG Cluster

      |

owns

      |

PostgreSQL resources
```

and:

```
PostgreSQL CR

      |

owns

      |

PgBouncer StatefulSet

      |

owns

      |

Pods/PVC
```

---

Benefit:

Delete:

```bash
kubectl delete postgresql customer-db
```

Result:

```
PostgreSQL CR deleted

        |

        |

Owned resources removed
```

---

# 11. Connectivity Demonstration

## User workflow

## Step 1: Create database

```bash
kubectl apply -f postgres.yaml
```

---

## Step 2: Check status

```bash
kubectl get postgresql
```

Example:

```
NAME            STATUS

customer-db     Ready
```

---

## Step 3: Get connection information

Your Operator provides:

```
customer-db-connection-secret
```

Containing:

```
host
port
username
password
database
```

---

## Step 4: Connect

Example:

```bash
psql \
-h customer-db-pgbouncer \
-U customer-admin \
-d customer
```

---

# 12. GitOps Automation with ArgoCD (Bonus)

## Goal

Make deployment fully automated.

Architecture:

```
Git Repository

        |

        |

      ArgoCD

        |

        |

 Kubernetes Cluster

        |

        |

Your Operator

        |

        |

PostgreSQL Service
```

---

# Git Repository Structure

Example:

```
gitops/

├── infrastructure/

│   └── terraform/


├── operators/

│   ├── cnpg/

│   └── postgres-platform/


├── applications/

│   └── postgres-cr.yaml


└── argocd/

    └── application.yaml
```

---

# ArgoCD Responsibilities

ArgoCD manages:

* Operator deployment
* CRD deployment
* PostgreSQL CR deployment

---

Example flow:

```
Developer

   |

git push

   |

Git Repository

   |

ArgoCD Sync

   |

Kubernetes

   |

PostgreSQL Instance Created
```

---

# 13. Documentation Structure

Your final documentation should contain:

---

# Chapter 1: Introduction

Explain:

* What is PaaS?
* Why managed database?
* Why Kubernetes Operator?

---

# Chapter 2: Architecture

Include:

```
User

 |

PostgreSQL CR

 |

Your Operator

 |

CNPG + PgBouncer

 |

PostgreSQL Service
```

---

# Chapter 3: Infrastructure

Explain:

* SKE
* Terraform
* Kubernetes cluster

---

# Chapter 4: Operator Implementation

Explain:

* CRD
* Controller
* Reconcile loop
* Owner references

---

# Chapter 5: Product Lifecycle

Show:

```
Create CR

 ↓

Operator Reconciliation

 ↓

CNPG Cluster Created

 ↓

PgBouncer Created

 ↓

Connection Available
```

---

# Chapter 6: Connectivity Demo

Show:

* Creating instance
* Checking status
* Getting credentials
* Connecting to database

---

# Chapter 7: GitOps

Explain:

* Git as source of truth
* ArgoCD synchronization
* Automated deployment

---

# Final Week 2 Deliverable Checklist

## Infrastructure

✅ SKE cluster created with Terraform
✅ Kubernetes environment documented

## PaaS Product

✅ PostgreSQL CRD
✅ PostgreSQL custom resource
✅ Database provisioning workflow

## Operator

✅ Kubebuilder Operator
✅ Reconciler implementation
✅ Status management
✅ Owner references

## Database

✅ CloudNativePG integration
✅ PostgreSQL HA capability delegated to CNPG

## Custom Kubernetes Management

✅ PgBouncer StatefulSet
✅ PgBouncer Service
✅ PVC handling

## Connectivity

✅ Connection secret
✅ Database connection demonstration

## Bonus

✅ ArgoCD GitOps deployment

---

This scope is a very good match for Week 2: you are building a **real PaaS control plane**, not just deploying PostgreSQL. It also creates a clean foundation for Week 3, where your REST API will simply become another interface that creates and manages your `PostgreSQL` Custom Resources.
