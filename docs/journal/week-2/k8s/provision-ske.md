# Provision an STACKIT Kubernetes Engine (SKE) cluster with the STACKIT Terraform Provider

## Prerequisites

Before you begin, ensure you have:

* Terraform installed (v1.5 or later recommended)
* A STACKIT account
* A STACKIT project
* A STACKIT Service Account with sufficient permissions (for example, `project.owner` or permissions to create SKE resources)
* Your **Project ID**
* Your preferred **region** (for example, `eu01`)

---

## Step 1. Create a Service Account <sup>[ref](https://www.youtube.com/watch?v=2W6JE8BVTBk)</sup>

1. Log in to the STACKIT Portal.
2. Open your project.
3. Navigate to **IAM & Security**.
4. Create a **Service Account**.
5. Assign a role that allows managing Kubernetes resources, such as:

   * `project.owner`
   * or a custom role with SKE permissions.

---

## Step 2. Create a Service Account Key

Inside the Service Account:

1. Click **Create Key**.
2. Download the JSON key.

Example:

```
stackit-sa-key.json
```

Keep this file secure.

---

## Step 3. Create a Terraform project

Example directory:

```
ske-tf/
│
├── provider.tf
├── main.tf
├── variables.tf
├── terraform.tfvars
└── outputs.tf
```

---

## Step 4. Configure the STACKIT Provider

Create `provider.tf`.

```hcl
terraform {
  required_version = ">= 1.7.0"

  required_providers {
    stackit = {
      source  = "stackitcloud/stackit"
      version = "~> 0.104" # pin to a minor version
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.31"
    }
  }
}

provider "stackit" {
  region = "eu01"

  service_account_key_path = "./stackit-sa-key.json"
}
```

The provider searches for credentials in the following order:

1. `service_account_key_path` in the provider block
2. Environment variables
3. Credentials file (`$HOME/.stackit/credentials.json`)

### Alternative authentication (Environment Variables)

Instead of specifying the key path in the provider block, you can authenticate with environment variables:

```bash
export STACKIT_SERVICE_ACCOUNT_KEY_PATH=./stackit-sa-key.json
```

Then simplify the provider configuration:

```hcl
provider "stackit" {
  region = "eu01"
}
```

The provider also supports supplying the key contents directly via `STACKIT_SERVICE_ACCOUNT_KEY` or using a credentials file at `$HOME/.stackit/credentials.json`.

---

## Step 6. Variables

`variables.tf`

```hcl
variable "project_id" {
  description = "Stackit project id"
  type        = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "service_account_key_path" {
  description = "Stackit service accoutn key path"
  type        = string
  default     = "./stackit-sa-key.json"
}
```

`terraform.tfvars`

```hcl
project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
```

---

## Step 7. Create the SKE Cluster

The `stackit_ske_cluster` resource provisions a Kubernetes cluster in STACKIT. An SKE cluster consists of:

* A **managed control plane**, operated by STACKIT.
* One or more **node pools**, which define the worker nodes (virtual machines) where your applications run.

A node pool is a group of worker nodes that share the same configuration, such as machine type, operating system, storage, and autoscaling settings.

`main.tf`

```hcl
resource "stackit_ske_cluster" "cluster" {
  project_id = var.project_id
  name       = "paas"

  node_pools = [
    {
      name               = "node-pool-1"
      machine_type       = "g1a.2d"
      minimum            = 1
      maximum            = 2
      availability_zones = ["eu01-1"]

      os_version_min = "4593.2.2"
      os_name        = "flatcar"
      volume_size    = 32
      volume_type    = "storage_premium_perf6"
    }
  ]
}
```

> **Note:** The cluster name must be between **1 and 11 characters**. For example, `ske-demo` is valid, while `demo-cluster` exceeds the maximum length.

### Node pool configuration

| Parameter            | Description                                                                                                        |
| -------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `name`               | Name of the node pool.                                                                                             |
| `machine_type`       | The VM flavour used for each worker node (CPU and memory configuration).                                           |
| `minimum`            | Minimum number of worker nodes to maintain in the pool.                                                            |
| `maximum`            | Maximum number of worker nodes when autoscaling is enabled.                                                        |
| `availability_zones` | Availability zone(s) where the worker nodes will be created.                                                       |
| `os_name`            | Operating system used by the worker nodes (for example, `flatcar`).                                                |
| `os_version_min`     | Minimum supported operating system image version for the nodes. Use a currently supported version for your region. |
| `volume_size`        | Size of the root disk attached to each worker node (in GB).                                                        |
| `volume_type`        | Storage class used for the worker node disks.                                                                      |

### Understanding node pools

An SKE cluster contains a **managed control plane** and one or more **node pools**.

```text
SKE Cluster
│
├── Managed Control Plane
│   ├── Kubernetes API Server
│   ├── Scheduler
│   └── Controller Manager
│
└── Node Pool
    ├── Worker Node 1
    └── Worker Node 2
```

The control plane manages the Kubernetes cluster, while the worker nodes in the node pool run your containerised applications.

For development and testing, a single node pool is usually sufficient. As your workloads grow, you can add additional node pools with different machine types, storage configurations, or scaling policies to support different application requirements.

---

## Step 8. Create the cluster

```bash
terraform fmt
terraform init
terraform validate
terraform plan -out=tfplan
terraform apply tfplan
```

Confirm with:

```
yes
```

Terraform will provision the cluster in STACKIT.

---

## Step 9. Retrieve the Kubernetes credentials (kubeconfig)

After the SKE cluster has been created, you need a **kubeconfig file** to connect to the Kubernetes API server using tools such as `kubectl`.

The `stackit_ske_kubeconfig` Terraform resource generates a **temporary kubeconfig** containing the credentials required to authenticate against your SKE cluster.

`main.tf`

```hcl
resource "stackit_ske_kubeconfig" "ske_kubeconfig" {
  project_id   = var.project_id
  cluster_name = stackit_ske_cluster.cluster.name

  refresh        = true
  expiration     = 7200
  refresh_before = 3600
}
```

### Configuration parameters

| Parameter        | Description                                                                         |
| ---------------- | ----------------------------------------------------------------------------------- |
| `project_id`     | The STACKIT project ID where the SKE cluster exists.                                |
| `cluster_name`   | The name of the SKE cluster for which the kubeconfig should be generated.           |
| `refresh`        | Enables automatic renewal of the kubeconfig when it is close to expiration.         |
| `expiration`     | Defines how long the generated kubeconfig remains valid, in seconds.                |
| `refresh_before` | Defines when Terraform should refresh the kubeconfig before it expires, in seconds. |

### Example expiration settings

```hcl id="q0yn1k"
expiration     = 7200
refresh_before = 3600
```

This means:

* The kubeconfig is valid for **7200 seconds (2 hours)**.
* Terraform refreshes it when **3600 seconds (1 hour)** of validity remain.

Example timeline:

```text id="8klx5s"
09:00  kubeconfig generated
       |
       |  Valid for 2 hours
       |
10:00  Refresh triggered (1 hour remaining)
       |
       |  New kubeconfig generated
       |
11:00  Original kubeconfig expires
```

---

## Step 10. Output the kubeconfig

```hcl
output "kubeconfig" {
  value     = stackit_ske_kubeconfig.cluster.kube_config
  sensitive = true
}
```

Retrieve it with:

```bash
terraform output -raw kubeconfig > kubeconfig.yaml
```

Configure kubeconfig:

```bash
export KUBECONFIG=./kubeconfig.yaml

# OR
mkdir -p ~/.kube
mv ./kubeconfig.yaml ~/.kube/config
```

Verify connectivity:

```bash
kubectl get nodes
```

Expected output:

```
NAME          STATUS   ROLES    AGE
worker-001    Ready    <none>   5m
```

---

## Step 11. Create a STACKIT Object Storage bucket to store tf state

The flow will be:

```text
Local terraform.tfstate
          |
          | terraform init -migrate-state
          ▼
STACKIT Object Storage Bucket
          |
          ▼
Remote state + state locking
```

Terraform backend requires an S3-compatible bucket to store the state file.

In the STACKIT Portal:

1. Open your STACKIT project.
2. Go to **Object Storage**.
3. Create an Object Storage instance (if you don't already have one).
4. Create a bucket for Terraform state.

Example:

```text
Bucket name:
ske-terraform-state
```

Recommended naming:

```text
<project>-terraform-state
```

Example:

```text
paas-terraform-state
```

---

## Step 12. Create S3 access credentials

Terraform backend does not use your STACKIT Service Account JSON key.

You need **Object Storage S3 credentials**.

Create:

* Access Key ID
* Secret Access Key

Example:

```text
Access Key:
AKIAxxxxxxxxxxxx

Secret Key:
xxxxxxxxxxxxxxxxxxxxxxxx
```

Store them securely.

---

## Step 13. Configure backend in Terraform

`provider.tf`

```hcl
terraform {
  // ...

  backend "s3" {
    bucket = "ske-terraform-state"
    key    = "ske/paas/terraform.tfstate"

    region = "eu01"

    endpoint = "https://s3.eu01.onstackit.cloud"

    encrypt = true

    use_lockfile = true
  }
}
```

Replace:

```hcl
bucket = "ske-terraform-state"
```

with your actual bucket name.

---

## Step 14. Configure S3 credentials

Do not put credentials inside Terraform files.

Export them in your shell:

```bash
export AWS_ACCESS_KEY_ID="<your-access-key>"
export AWS_SECRET_ACCESS_KEY="<your-secret-key>"
```

Verify:

```bash
echo $AWS_ACCESS_KEY_ID
```

---

### Optional: Make credentials permanent

For local development:

```bash
nano ~/.zshrc
```

Add:

```bash
export AWS_ACCESS_KEY_ID="<your-access-key>"
export AWS_SECRET_ACCESS_KEY="<your-secret-key>"
```

Reload:

```bash
source ~/.zshrc
```

---

## Step 15. Reinitialize Terraform backend

Run:

```bash
terraform init
```

Terraform will detect the backend change:

Example:

```text
Initializing the backend...

Do you want to copy existing state to the new backend?
```

Answer:

```text
yes
```

Terraform will upload:

```text
terraform.tfstate
```

to:

```text
STACKIT Object Storage
 |
 └── ske/paas/terraform.tfstate
```

---

## Step 16. Verify remote state

Run:

```bash
terraform state list
```

You should still see your resources:

Example:

```text
stackit_ske_cluster.cluster
stackit_ske_kubeconfig.ske_kubeconfig
```

This confirms Terraform is reading from the remote state.

---

## Step 17. Test state locking

Open two terminals.

Terminal 1:

```bash
terraform apply
```

Before it finishes, run in Terminal 2:

```bash
terraform apply
```

The second execution should fail:

```text
Error acquiring the state lock
```

Example:

```text
Lock Info:
  ID:        xxxxxxxx
  Operation: OperationTypeApply
```

This prevents two users or pipelines from changing the same SKE cluster simultaneously.

---

## Step 18. Add `.gitignore`

Never commit Terraform state or credentials.

Create:

```bash
touch .gitignore
```

Add:

```gitignore
# Terraform
.terraform/
*.tfstate
*.tfstate.*
.terraform.lock.hcl

# Secrets
*.json
*.pem
*.key

# Environment files
.env
```

---

## Step 19. Recommended final project structure

Your project should now look like:

```text
ske/
│
├── provider.tf
├── main.tf
├── variables.tf
├── outputs.tf
├── terraform.tfvars
├── .gitignore
│
└── Remote State
    |
    └── STACKIT Object Storage
        |
        └── ske/paas/terraform.tfstate
```

---

## Step 20. Normal workflow from now on

After backend setup, your normal Terraform workflow becomes:

```bash
terraform init

terraform plan

terraform apply
```

The state is now:

✅ stored remotely
✅ encrypted
✅ shared between team members
✅ protected by locking
✅ independent of your laptop

Your SKE cluster configuration is now ready for team usage or CI/CD automation.


## Complete workflow

```text
Create Service Account
          │
          ▼
Generate JSON Key
          │
          ▼
Configure Terraform Provider
          │
          ▼
terraform init
          │
          ▼
terraform plan
          │
          ▼
terraform apply
          │
          ▼
Generate kubeconfig
          │
          ▼
Export KUBECONFIG
          │
          ▼
kubectl get nodes
```
