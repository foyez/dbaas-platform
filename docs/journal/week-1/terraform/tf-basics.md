# Terraform (TF) Fundamentals

## What is Terraform (TF)?

**Terraform (TF)** is an **Infrastructure as Code (IaC)** tool developed by **HashiCorp**. It allows you to create, manage, update, and destroy infrastructure using configuration files instead of manually clicking through cloud provider consoles.

With Terraform, you write infrastructure in code, and Terraform automatically provisions it on cloud platforms such as AWS, Azure, Google Cloud, Kubernetes, VMware, and many others.

Example:
Instead of manually creating an EC2 instance in AWS Console, you write a Terraform configuration, and Terraform creates it for you.

---

# Why Terraform?

Without Terraform:

* Create resources manually through cloud consoles.
* Difficult to recreate the same infrastructure.
* Human errors are common.
* Hard to manage large environments.
* No version control.

With Terraform:

* Infrastructure is stored as code.
* Easy to recreate environments.
* Version control with Git.
* Automation through CI/CD.
* Consistent deployments.
* Supports multiple cloud providers.

---

# Benefits of Terraform

* Infrastructure as Code (IaC)
* Multi-cloud support
* Version controlled infrastructure
* Reusable modules
* Automatic dependency management
* Execution plan before changes
* Easy rollback using Git
* Team collaboration
* Automation friendly

---

# Terraform Workflow

```
Write Code
      ↓
terraform init
      ↓
terraform plan
      ↓
terraform apply
      ↓
Infrastructure Created
```

To remove infrastructure:

```
terraform destroy
```

---

# Important Terraform Terms

## 1. Provider

A **Provider** is a plugin that allows Terraform to communicate with a cloud platform or service.

Examples:

* AWS
* Azure
* Google Cloud
* Kubernetes
* Docker
* GitHub

Example:

```hcl
provider "aws" {
  region = "us-east-1"
}
```

Meaning:
Terraform will use the AWS API in the `us-east-1` region.

---

## 2. Resource

A **Resource** is any infrastructure object that Terraform creates or manages.

Examples:

* EC2 Instance
* S3 Bucket
* VPC
* Load Balancer
* Security Group
* Database

Example:

```hcl
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}
```

This creates one EC2 instance.

---

## 3. State

The **Terraform State** is a file (`terraform.tfstate`) that records the infrastructure Terraform manages.

It stores:

* Resource IDs
* Current configuration
* Metadata
* Resource relationships

Terraform compares:

```
Configuration
      vs
Terraform State
      vs
Real Infrastructure
```

This comparison determines what needs to be created, changed, or destroyed.

Example:

If the state file says:

```
1 EC2 Instance exists
```

and your configuration now defines:

```
2 EC2 Instances
```

Terraform plans to create **1 additional EC2 instance**.

---

## 4. Data Source

A **Data Source** lets Terraform **read existing infrastructure** without creating it.

Example:
Suppose an existing VPC already exists.

Instead of creating another VPC:

```hcl
data "aws_vpc" "existing" {
  default = true
}
```

Terraform only reads information about the existing VPC.

**Difference**

Resource:

* Creates or manages infrastructure.

Data Source:

* Reads existing infrastructure.

---

## 5. Variables

Variables make configurations reusable.

Example:

```hcl
variable "instance_type" {
  default = "t2.micro"
}
```

Usage:

```hcl
instance_type = var.instance_type
```

---

## 6. Output

Outputs display useful information after deployment.

Example:

```hcl
output "public_ip" {
  value = aws_instance.web.public_ip
}
```

After `terraform apply`:

```
Public IP:
52.16.xx.xx
```

---

## 7. Module

A **Module** is a reusable collection of Terraform configurations.

Instead of writing the same code repeatedly:

```
Project A
Project B
Project C
```

Use one module:

```
EC2 Module
```

and call it multiple times.

---

## 8. Backend

A **Backend** determines where Terraform stores its state.

Common backends:

* Local
* AWS S3
* Azure Storage
* Google Cloud Storage
* Terraform Cloud

Example:

```
Local:
terraform.tfstate

Remote:
S3 Bucket
```

Remote backends are preferred for teams because everyone shares the same state.

---

## 9. Workspace

A **Workspace** allows multiple copies of the same infrastructure using one configuration.

Examples:

* Development
* Testing
* Production

```
Workspace:
dev

Workspace:
test

Workspace:
prod
```

Each workspace has its own state.

---

## 10. Provisioner

A **Provisioner** runs scripts or commands after a resource is created.

Example:

```hcl
provisioner "remote-exec" {
  inline = [
    "sudo apt update"
  ]
}
```

Provisioners are generally used sparingly because Terraform works best for provisioning infrastructure rather than configuring software.

---

## 11. Dependency

Terraform automatically detects dependencies between resources.

Example:

```
VPC
 ↓
Subnet
 ↓
EC2
```

Terraform creates them in the correct order.

---

## 12. Plan

`terraform plan`

Shows what Terraform will do before making changes.

Example:

```
+ Create EC2
~ Modify S3
- Delete Bucket
```

No changes are made at this stage.

---

## 13. Apply

`terraform apply`

Executes the plan and creates, updates, or deletes infrastructure.

---

## 14. Destroy

`terraform destroy`

Deletes all resources managed by the current Terraform configuration.

---

## 15. Init

`terraform init`

Initializes a Terraform project by:

* Downloading provider plugins
* Setting up the backend
* Preparing the working directory

Run this first in any new Terraform project.

---

## 16. HCL (HashiCorp Configuration Language)

Terraform uses **HCL**, a human-readable configuration language.

Example:

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "my-demo-bucket"
}
```

---

# Common Terraform Commands

| Command                    | Purpose                                       |
| -------------------------- | --------------------------------------------- |
| `terraform init`           | Initialize the project and download providers |
| `terraform fmt`            | Format configuration files                    |
| `terraform validate`       | Check configuration syntax                    |
| `terraform plan`           | Preview changes                               |
| `terraform apply`          | Apply changes                                 |
| `terraform destroy`        | Remove infrastructure                         |
| `terraform show`           | Display current state                         |
| `terraform output`         | Show output values                            |
| `terraform state list`     | List resources in the state                   |
| `terraform workspace list` | List workspaces                               |

---

# Terraform File Structure

```
project/
│
├── main.tf          # Main resources
├── variables.tf     # Variables
├── outputs.tf       # Outputs
├── provider.tf      # Provider configuration
├── terraform.tfvars # Variable values
├── backend.tf       # Backend configuration (optional)
└── terraform.tfstate # State file (generated)
```

---

# Resource vs Data Source

| Resource                          | Data Source                   |
| --------------------------------- | ----------------------------- |
| Creates or manages infrastructure | Reads existing infrastructure |
| Can modify infrastructure         | Read-only                     |
| Stored in state                   | Used to retrieve information  |

Example:

```hcl
resource "aws_s3_bucket" "logs" {
  bucket = "my-logs-bucket"
}
```

Creates a new S3 bucket.

```hcl
data "aws_ami" "ubuntu" {
  most_recent = true
}
```

Reads the latest Ubuntu AMI without creating anything.

---

# Complete Terraform Lifecycle

```
Write Terraform Code
          │
          ▼
terraform init
          │
          ▼
terraform validate
          │
          ▼
terraform plan
          │
          ▼
terraform apply
          │
          ▼
Infrastructure Created
          │
          ▼
terraform output
          │
          ▼
Modify Configuration
          │
          ▼
terraform plan
          │
          ▼
terraform apply
          │
          ▼
terraform destroy
```

---

# Quick Revision

* **Terraform (TF):** Infrastructure as Code tool.
* **Provider:** Connects Terraform to a cloud or service (AWS, Azure, GCP, etc.).
* **Resource:** Infrastructure object Terraform creates or manages.
* **Data Source:** Reads existing infrastructure.
* **State:** Tracks managed resources in `terraform.tfstate`.
* **Variable:** Makes configurations reusable.
* **Output:** Displays values after deployment.
* **Module:** Reusable Terraform code.
* **Backend:** Stores the Terraform state.
* **Workspace:** Separate environments (dev, test, prod).
* **Provisioner:** Runs scripts after resource creation (use sparingly).
* **Plan:** Shows proposed changes.
* **Apply:** Makes the changes.
* **Destroy:** Removes managed infrastructure.
* **Init:** Initializes the Terraform project.
* **HCL:** The configuration language used by Terraform.

These concepts form the foundation of Terraform and are the most important topics to understand before working on real-world infrastructure automation.
