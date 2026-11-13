# OpenStack Provisioning — Manual CLI vs. Terraform

This is the same lab you already ran manually with the `openstack` CLI, rebuilt in Terraform using the [`terraform-provider-openstack/openstack`](https://registry.terraform.io/providers/terraform-provider-openstack/openstack) provider.

The goal isn't just "here's the Terraform syntax" — it's to see **why** engineers move from CLI to IaC once a lab turns into a real environment, by comparing the two side by side at every step.

---

## The Core Mental Model Shift

| | Manual CLI | Terraform |
|---|---|---|
| **Style** | Imperative — you tell it *what to do*, step by step | Declarative — you describe *the end state you want* |
| **Ordering** | You must run commands in the correct dependency order yourself (network → subnet → router...) | You just declare resources; Terraform builds a dependency graph from references and orders creation/deletion automatically |
| **State tracking** | Nothing tracks what you built — `openstack server list` is your only memory | A `terraform.tfstate` file is the source of truth for what exists and how resources relate |
| **Repeatability** | Re-running the same commands can error ("already exists") or create duplicates | Re-running `apply` is idempotent — no changes if nothing drifted |
| **Teardown** | You manually reverse the creation order (Lab 25) | `terraform destroy` walks the dependency graph in reverse automatically |
| **Drift** | If someone changes something in Horizon, the CLI has no idea | `terraform plan` detects and reports drift against the state file |
| **Review** | Commands live in your shell history, if anywhere | Code lives in version control — reviewable, diffable, repeatable across environments |

This is the same reason teams move from clicking around AWS console → CloudFormation/Terraform, or from manual `kubectl` commands → Helm/GitOps. The underlying platform doesn't change; the *engineering discipline* around it does.

---

## Prerequisites

```bash
# Install Terraform (if not already installed)
sudo apt install -y terraform    # or use the HashiCorp apt repo for latest version
terraform version
```

Terraform needs the same credentials your `openrc admin admin` gave the CLI. Instead of sourcing a file every session, you point the provider at your `clouds.yaml` (usually at `~/.config/openstack/clouds.yaml` or generated from Horizon under "API Access → Download OpenStack RC File").

**Why this matters:** `source openrc` injects auth into your *shell environment* — it disappears the moment your terminal closes. Terraform's provider block reads credentials explicitly and consistently every time you run it, which is what makes it safe to run from CI/CD pipelines where there's no interactive shell to `source` anything into.

---

## Project Structure

```
openstack-lab/
├── providers.tf
├── variables.tf
├── network.tf
├── compute.tf
├── storage.tf
├── outputs.tf
└── terraform.tfvars
```

Splitting files by concern (network vs. compute vs. storage) isn't required — Terraform reads every `.tf` file in the directory as one configuration — but it mirrors how the manual labs were grouped (Neutron labs, Nova labs, Cinder labs), which makes the comparison easier to follow.

---

## `providers.tf`

```hcl
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.53"
    }
  }
}

provider "openstack" {
  cloud = "devstack"   # matches a "devstack:" entry in clouds.yaml
}
```

**Why `cloud = "devstack"` instead of hardcoded credentials:** Same reasoning as `openrc` — never put passwords directly in version-controlled code. `clouds.yaml` (or environment variables `OS_*`) keeps secrets out of the `.tf` files that get committed to git.

---

## `variables.tf`

```hcl
variable "private_network_cidr" {
  default = "192.168.10.0/24"
}

variable "backend_network_cidr" {
  default = "10.10.10.0/24"
}

variable "external_network_name" {
  default = "public"
}

variable "image_name" {
  default = "Ubuntu22"
}

variable "flavor_name" {
  default = "small"
}

variable "key_pair_name" {
  default = "mykey"
}

variable "public_key_path" {
  default = "~/.ssh/id_rsa.pub"
}
```

**Why variables instead of hardcoding values in every resource (as the CLI labs did):** In the manual labs, `192.168.10.0/24` or `Ubuntu22` are typed inline every time they're needed. In Terraform, defining them once means changing the CIDR or image name in one place updates every resource that references it — and it lets you spin up a second, identical environment (e.g., `staging`) just by overriding a `.tfvars` file, without touching the resource definitions at all.

---

## `network.tf` — Labs 2 through 6

```hcl
# Lab 2 — Create a Private Network
resource "openstack_networking_network_v2" "private_net" {
  name           = "private-net"
  admin_state_up = true
}

# Lab 3 — Create a Subnet
resource "openstack_networking_subnet_v2" "private_subnet" {
  name       = "private-subnet"
  network_id = openstack_networking_network_v2.private_net.id
  cidr       = var.private_network_cidr
  ip_version = 4
}

# Lab 4 — Create a Router
resource "openstack_networking_router_v2" "router1" {
  name                = "router1"
  admin_state_up      = true
  external_network_id = data.openstack_networking_network_v2.external.id
}

# Lab 5 — Connect Router to Private Network
resource "openstack_networking_router_interface_v2" "router_iface" {
  router_id = openstack_networking_router_v2.router1.id
  subnet_id = openstack_networking_subnet_v2.private_subnet.id
}
```

**Side-by-side comparison:**

| Manual (Labs 2–6) | Terraform |
|---|---|
| Run `network create`, then separately `subnet create --network private-net`, referencing it **by name** | `openstack_networking_subnet_v2` references `openstack_networking_network_v2.private_net.id` **directly** — Terraform resolves the real ID after creation, no risk of typos or name collisions |
| Router external gateway is a *separate* command (Lab 6) run manually after router creation | Folded into the router resource itself via `external_network_name` — one declarative block instead of a two-step imperative sequence |
| You must run these 5 commands in this exact order, or they fail | Terraform reads the `network_id = ...` and `router_id = ...` references and **automatically** builds the create order — you could paste these four blocks in any order in the file and get the same result |

This is the single biggest practical difference: in the CLI labs, *you* are the dependency-resolution engine, manually sequencing 5 commands. In Terraform, the dependency graph is inferred from the code itself.

---

## `compute.tf` — Labs 7 through 13

```hcl
# Lab 7 — Upload Ubuntu Image (rarely done via Terraform in practice — see note below)
resource "openstack_images_image_v2" "ubuntu" {
  name             = var.image_name
  local_file_path  = "${path.module}/jammy-server-cloudimg-amd64.img"
  disk_format      = "qcow2"
  container_format = "bare"
  visibility       = "public"
}

# Lab 8 — Create a Flavor (admin-only operation, same as CLI)
resource "openstack_compute_flavor_v2" "small" {
  name  = var.flavor_name
  ram   = 2048
  vcpus = 2
  disk  = 20
}

# Lab 9 — Create SSH Key Pair
resource "openstack_compute_keypair_v2" "mykey" {
  name       = var.key_pair_name
  public_key = file(var.public_key_path)
}

# Lab 10 — Create Security Group + Rules
resource "openstack_networking_secgroup_v2" "web_secgroup" {
  name        = "web-secgroup"
  description = "Allow SSH, ICMP, HTTP, HTTPS"
}

resource "openstack_networking_secgroup_rule_v2" "allow_ssh" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  security_group_id = openstack_networking_secgroup_v2.web_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "allow_icmp" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  security_group_id = openstack_networking_secgroup_v2.web_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "allow_http" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  security_group_id = openstack_networking_secgroup_v2.web_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "allow_https" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  security_group_id = openstack_networking_secgroup_v2.web_secgroup.id
}

# Lab 11 — Launch VM
resource "openstack_compute_instance_v2" "ubuntu_vm" {
  name            = "ubuntu-vm"
  image_id        = openstack_images_image_v2.ubuntu.id
  flavor_id       = openstack_compute_flavor_v2.small.id
  key_pair        = openstack_compute_keypair_v2.mykey.name
  security_groups = [openstack_networking_secgroup_v2.web_secgroup.name]

  network {
    uuid = openstack_networking_network_v2.private_net.id
  }

  depends_on = [openstack_networking_router_interface_v2.router_iface]
}

# Lab 12 — Assign Floating IP
resource "openstack_networking_floatingip_v2" "fip" {
  pool = var.external_network_name
}

resource "openstack_compute_floatingip_associate_v2" "fip_assoc" {
  floating_ip = openstack_networking_floatingip_v2.fip.address
  instance_id = openstack_compute_instance_v2.ubuntu_vm.id
}
```

**Side-by-side comparison:**

| Manual (Labs 7–12) | Terraform |
|---|---|
| Four separate `security group rule create` commands, each silently assuming `web-secgroup` already exists | Four `openstack_networking_secgroup_rule_v2` resources, each **explicitly** referencing `openstack_networking_secgroup_v2.web_secgroup.id` — if the group definition ever changes name, every rule updates automatically |
| `server create` lists `--image`, `--flavor`, `--network`, `--security-group`, `--key-name` **by name string** | `openstack_compute_instance_v2` references the actual resource objects — Terraform guarantees they exist and are correctly wired before the VM is created |
| Floating IP creation and assignment are two manual commands you run back-to-back, easy to forget the second one | Two resources (`floatingip_v2` + `floatingip_associate_v2`) — `terraform plan` will visibly show a floating IP with no association if you forget to link them, catching the mistake before `apply` |
| Nothing enforces that the router is actually connected before the VM boots | The explicit `depends_on` block forces Terraform to wait for `router_interface_v2` — the same ordering the manual lab flow diagram showed, now encoded as code, not memory |

**Note on Lab 7 (image upload):** In real production Terraform, uploading base OS images via `openstack_images_image_v2` is uncommon — images are usually managed centrally by a platform team (or built via Packer) and simply referenced by `data "openstack_images_image_v2" "ubuntu" { name = "Ubuntu22" }` as a **data source** rather than a resource you create. It's included here as a resource only to mirror Lab 7 exactly; in practice you'd write:

```hcl
data "openstack_images_image_v2" "ubuntu" {
  name = var.image_name
}
```

and reference `data.openstack_images_image_v2.ubuntu.id` instead — this avoids Terraform trying to re-upload a multi-GB image file every time someone runs `apply`.

---

## `storage.tf` — Labs 14, 15, 17, 18

```hcl
# Lab 14 — Create Cinder Volume
resource "openstack_blockstorage_volume_v3" "volume1" {
  name = "volume1"
  size = 10
}

# Lab 15 — Attach Volume
resource "openstack_compute_volume_attach_v2" "attach1" {
  instance_id = openstack_compute_instance_v2.ubuntu_vm.id
  volume_id   = openstack_blockstorage_volume_v3.volume1.id
}

# Lab 17 — Create Volume Snapshot
resource "openstack_blockstorage_snapshot_v3" "snapshot1" {
  name      = "snapshot1"
  volume_id = openstack_blockstorage_volume_v3.volume1.id
}
```

**Why Lab 16 (format/mount) and Lab 18 (instance snapshot) have no Terraform blocks:**

- **Lab 16** (`mkfs.ext4`, `mount`) happens *inside the guest OS*, not against the OpenStack API — Terraform manages infrastructure, not what runs inside a VM's operating system. The idiomatic way to automate this in Terraform is `user_data` / cloud-init passed to the instance, e.g.:

```hcl
resource "openstack_compute_instance_v2" "ubuntu_vm" {
  # ...
  user_data = <<-EOF
    #cloud-config
    bootcmd:
      - mkfs.ext4 /dev/vdb || true
    mounts:
      - [ /dev/vdb, /data, ext4, "defaults", "0", "2" ]
  EOF
}
```

  This is a meaningful lesson in itself: **Terraform's job ends at the API boundary.** Anything that needs to happen *inside* the guest belongs to cloud-init/user_data (or a follow-up tool like Ansible), not to the OpenStack provider.

- **Lab 18** (instance snapshot via `openstack server image create`) has no dedicated Terraform resource in the OpenStack provider — creating a bootable image *from a running instance* is an imperative, one-off operational action, not a piece of long-lived declared infrastructure. This is a good example of where Terraform genuinely isn't the right tool, and the manual CLI command (or a small script) is still the correct approach even in an otherwise fully Terraform-managed environment.

---

## `outputs.tf`

```hcl
output "instance_floating_ip" {
  value = openstack_networking_floatingip_v2.fip.address
}

output "instance_private_ip" {
  value = openstack_compute_instance_v2.ubuntu_vm.access_ip_v4
}

output "volume_id" {
  value = openstack_blockstorage_volume_v3.volume1.id
}
```

**Why:** In the manual labs, you re-ran `openstack server list` or `openstack floating ip list` any time you needed to remember an IP. Terraform outputs surface exactly the values you'll need repeatedly (for SSH, for handing to another team, for feeding into a different Terraform module) right after `apply` finishes, without a separate lookup command.

---

## Running It

```bash
terraform init      # downloads the openstack provider plugin
terraform plan       # shows exactly what will be created, in dependency order — review before touching anything
terraform apply      # builds everything: network → subnet → router → security group → keypair → VM → floating IP → volume → attach → snapshot
```

**Why `plan` before `apply` matters so much here:** This is the step with no CLI equivalent at all. None of the manual `openstack ... create` commands let you preview what *would* happen before it happens — you just run them and see the result. `terraform plan` gives you a dry-run diff (`+ create`, `~ update`, `- destroy`) against the current state, which is exactly the safety net you'd want before running anything against a shared or production cloud.

Verification (same idea as **Lab 13b**, now with Terraform outputs feeding the commands directly):

```bash
FLOATING_IP=$(terraform output -raw instance_floating_ip)

ping -c 4 "$FLOATING_IP"
ssh ubuntu@"$FLOATING_IP" -- cloud-init status
ssh ubuntu@"$FLOATING_IP" -- 'uptime && df -h'
```

Teardown:

```bash
terraform destroy
```

**Why this one command replaces all of Lab 25:** Terraform reads its own state file, builds the *reverse* of the dependency graph it used to create everything, and deletes VM → volume → floating IP association → router interface → router → subnet → network in the correct order automatically. The manual lab made you memorize and hand-execute that exact reverse sequence; here it's one command, and it will refuse to proceed (with a clear error) if something external still depends on a resource it's trying to delete — the same protection Neutron's API gives you manually, just surfaced earlier and more clearly.

---

## Summary Table — Lab-by-Lab Mapping

| Lab # | Manual Task | Terraform Resource |
|---|---|---|
| 2 | Create private network | `openstack_networking_network_v2` |
| 3 | Create subnet | `openstack_networking_subnet_v2` |
| 4 | Create router | `openstack_networking_router_v2` |
| 5 | Attach router to private subnet | `openstack_networking_router_interface_v2` |
| 6 | Set router external gateway | `external_network_name` on the router resource |
| 7 | Upload image | `openstack_images_image_v2` (or `data` source in practice) |
| 8 | Create flavor | `openstack_compute_flavor_v2` |
| 9 | Create SSH key pair | `openstack_compute_keypair_v2` |
| 10 | Create security group + rules | `openstack_networking_secgroup_v2` + `_secgroup_rule_v2` |
| 11 | Launch VM | `openstack_compute_instance_v2` |
| 12 | Floating IP + assign | `openstack_networking_floatingip_v2` + `_floatingip_associate_v2` |
| 13 / 13b | SSH in / verify reachable | not a resource — a `terraform output` feeding your `ssh`/`ping`/`curl` checks |
| 14 | Create volume | `openstack_blockstorage_volume_v3` |
| 15 | Attach volume | `openstack_compute_volume_attach_v2` |
| 16 | Format/mount inside VM | `user_data` (cloud-init) — not a Neutron/Nova API call |
| 17 | Volume snapshot | `openstack_blockstorage_snapshot_v3` |
| 18 | Instance snapshot | not supported as a resource — stays a manual/scripted CLI action |
| 19 | Boot VM from snapshot | new `openstack_compute_instance_v2` with `image_id` pointed at the snapshot's resulting image |
| 20–21 | Stop/start/reboot/resize | generally left to operational tooling (or `openstack` CLI / Horizon), not something you'd put in a Terraform desired-state file, since these are transient operational actions rather than declared infrastructure |
| 22 | Create port | `openstack_networking_port_v2` |
| 23–24 | Second network + attach | second `openstack_networking_network_v2`/`_subnet_v2`, plus a second `network {}` block in the instance resource |
| 25 | Delete everything | `terraform destroy` |

---

## The Real Takeaway

Doing the same lab twice — once by hand, once in Terraform — is meant to expose one thing clearly: **OpenStack's API doesn't change between the two approaches.** Every Terraform resource above is calling the exact same Nova/Neutron/Cinder/Glance endpoints the CLI commands were calling. What changes is *who tracks the dependency order, who remembers what exists, and who reverses the process safely on teardown* — you, or the tool. That's the whole argument for Infrastructure as Code in one sentence.