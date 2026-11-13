Once DevStack is installed, the best way to learn OpenStack is by performing a complete hands-on lab. The exercises below cover almost every core OpenStack service you'll use daily.

This version adds the **"why"** behind each step — the engineering reasoning a cloud admin or SRE actually goes through, not just the commands to type.

---

# OpenStack Hands-on Lab (Annotated)

## Lab Flow

```text
Keystone (Login)
       │
       ▼
Create Network
       │
       ▼
Create Subnet
       │
       ▼
Create Router
       │
       ▼
Connect Router to External Network
       │
       ▼
Upload Ubuntu Image
       │
       ▼
Create Security Group
       │
       ▼
Create SSH Key Pair
       │
       ▼
Create Flavor
       │
       ▼
Launch VM
       │
       ▼
Assign Floating IP
       │
       ▼
SSH into VM
       │
       ▼
Create Cinder Volume
       │
       ▼
Attach Volume
       │
       ▼
Create Snapshot
       │
       ▼
Detach/Delete VM
       │
       ▼
Launch VM from Snapshot
```

**Why this order?** It mirrors how a real data center network is built, from the outside in: you can't launch a VM without a network to plug it into, you can't reach that network without a router, and you can't reach the router's private side from your laptop without a floating IP. Every layer depends on the one before it — this is the same dependency chain you'd hit in AWS (VPC → subnet → route table → EC2) or any other cloud.

---

# Lab 1 – Verify OpenStack Services

```bash
source openrc admin admin

openstack service list
```

Expected services:

* Keystone
* Nova
* Neutron
* Glance
* Cinder
* Placement

**Why we do this first:** Before touching any resource, an engineer checks that the control plane is actually healthy. `source openrc` isn't just "logging in" — it injects auth tokens and the Keystone endpoint into your shell environment so every subsequent `openstack` CLI call knows *who* you are and *where* to send the request. If you skip this, every later command fails with an auth error, and it's the single most common "why isn't this working" support ticket in real OpenStack deployments. Checking `service list` first is the same instinct as checking `kubectl get nodes` before deploying a pod — confirm the platform is up before you build on top of it.

---

# Lab 2 – Create a Private Network (Neutron)

```bash
openstack network create private-net
```

Verify

```bash
openstack network list
```

**Why a network comes before anything else:** In OpenStack (like in physical data centers), a VM is just a compute resource sitting on a rack — it's useless until it has a network to talk on. A **private network** is the software-defined equivalent of a VLAN or a switch segment: it isolates your tenant's traffic from every other tenant's traffic on the same physical hardware. Engineers create this first because it's the "container" everything else (subnets, ports, VMs) will attach to. Use case: this is what lets Company A and Company B run VMs on the same physical compute node without ever seeing each other's packets.

---

# Lab 3 – Create a Subnet

```bash
openstack subnet create private-subnet-2 \
--network private-net \
--subnet-range 192.168.10.0/24
```

Verify

```bash
openstack subnet list
```

**Why a network alone isn't enough:** A network is just a broadcast domain — it has no concept of IP addressing, DHCP, or gateway until you attach a subnet. The subnet is what actually hands out IPs (192.168.10.x here) to instances via DHCP and defines the gateway address. Think of the network as "the room" and the subnet as "the floor plan with electrical outlets numbered." Use case: this is where you'd plan out your private IP scheme — e.g., 192.168.10.0/24 for app servers, 192.168.20.0/24 for databases — exactly like designing VPC subnets in AWS.

---

# Lab 4 – Create a Router

```bash
openstack router create router2
```

Verify

```bash
openstack router list
```

**Why we need a router:** A private network is isolated by design — that's the whole point of tenant isolation. But isolated also means *unreachable*. A router is the Neutron object that can bridge two networks (e.g., your private tenant network and the shared external network) and perform NAT. Without it, your VMs would be permanently stuck talking only to each other. This is the same role a home router plays between your private LAN and your ISP.

---

# Lab 5 – Connect Router to Private Network

```bash
openstack router add subnet router2 private-subnet-2
```

**Why this specific step:** Creating a router doesn't automatically attach it to anything — you have to explicitly plug one of its "legs" into your private subnet. This gives the router an interface (and IP) inside 192.168.10.0/24, so it can now receive traffic from your VMs and decide where to forward it. Skipping this step is a common gotcha: the router exists, but nothing can actually use it as a gateway.

---

# Lab 6 – Connect Router to External Network

List external networks

```bash
openstack network list --external
```

Example

```bash
openstack router set router2 --external-gateway public
```

Now your private network can reach the Internet.

**Why:** This plugs the router's *other* leg into the shared "public" network that DevStack pre-creates, which usually maps to real, routable IP space on the host. This is the step that enables SNAT (source NAT) — your VMs' private 192.168.10.x addresses get translated to a real address when they reach out to the internet, the same way your home devices share your ISP's single public IP. Without this, `apt update` inside your VM later would simply hang.

---

# Lab 7 – Upload Ubuntu Image (Glance)

Download Ubuntu Cloud Image

```bash
wget https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img
```

Upload

```bash
openstack image create ubuntu-22 \
--file /opt/stack/jammy-server-cloudimg-amd64.img \
--disk-format qcow2 \
--container-format bare \
--public
```

Verify

```bash
openstack image list
```

**Why images matter:** Nova can't boot a VM from nothing — it needs a base disk image to clone from, the same way you can't launch an EC2 instance without picking an AMI. Glance is OpenStack's image registry: it stores the qcow2/raw disk templates and metadata (OS type, min disk, min RAM) that Nova reads when scheduling a boot. Uploading it once and marking it `--public` means every project in the cloud can reuse the same golden image instead of each team uploading their own copy — this is exactly how enterprises manage a small set of hardened, patched "golden images" for compliance.

**Why `--container-format bare`?** The container format describes whether the disk image is wrapped inside another container. `bare` means the image is stored exactly as it is (for example, a qcow2 or raw disk) with **no additional packaging or metadata wrapper**. This is the most common choice because Nova expects a plain virtual disk image that it can boot directly. Other container formats (such as `ovf` or `ami`) are used for specialized image packaging or compatibility with other virtualization ecosystems, but for standard OpenStack VM images, `bare` is the correct and recommended option.


---

# Lab 8 – Create a Flavor (Nova)

```bash
openstack flavor create small \
--ram 2048 \
--vcpus 2 \
--disk 20
```

Verify

```bash
openstack flavor list
```

**Why:** A flavor is just a named hardware profile — vCPU count, RAM, disk size — the OpenStack equivalent of an EC2 instance type (`t3.medium`, etc.). Engineers define flavors up front so users request compute by *tier* ("small", "medium", "gpu-large") instead of arbitrary numbers, which makes capacity planning and billing predictable across the whole cloud. It also lets an admin cap how much any one VM can consume of a shared physical host.

---

# Lab 9 – Create SSH Key Pair

Generate key

```bash
ssh-keygen
```

Import

```bash
openstack keypair create \
--public-key ~/.ssh/id_rsa.pub \
mykey
```

Verify

```bash
openstack keypair list
```

**Why:** Cloud images ship with no password login enabled — it's a security default, not an oversight. Instead, cloud-init injects your **public** key into the VM's `authorized_keys` at first boot. Importing the key pair into Nova beforehand means you never have to bake secrets into the image itself, and you can rotate access per-user without touching the image at all. This is the same model as AWS EC2 key pairs.

---

# Lab 10 – Create Security Group

List existing groups

```bash
openstack security group list
```

Create

```bash
openstack security group create web-secgroup
```

**Why:** By default, Neutron security groups deny almost all inbound traffic — this is the cloud's stateful firewall layer, applied per-port at the hypervisor, before packets even reach the VM's own iptables. Creating a purpose-named group (`web-secgroup`) instead of reusing `default` is good practice because it documents *intent*: anyone looking at it later knows this group is meant for web-facing instances, not databases.

---

## Allow SSH

```bash
openstack security group rule create \
--proto tcp \
--dst-port 22 \
web-secgroup
```

**Use case:** Without this rule, you'd successfully assign a floating IP and still get `connection timed out` — the most common first debugging trap for OpenStack beginners. Opening 22/tcp is the minimum needed for admin access to the box.

---

## Allow Ping

```bash
openstack security group rule create \
--proto icmp \
web-secgroup
```

**Use case:** ICMP is opened mainly for operational sanity-checking — `ping` is the fastest way to confirm L3 reachability before debugging anything deeper like SSH or HTTP. Many production security groups actually restrict this to internal ranges only, since ICMP can be used for reconnaissance if left open to the world.

---

## Allow HTTP

```bash
openstack security group rule create \
--proto tcp \
--dst-port 80 \
web-secgroup
```

**Use case:** This is the rule that turns the VM into an actual web server from the network's point of view. If you're planning to run nginx/Apache on this box, this port has to be explicitly opened — nothing is allowed in by default.

---

## Allow HTTPS

```bash
openstack security group rule create \
--proto tcp \
--dst-port 443 \
web-secgroup
```

**Use case:** Same idea as port 80, but for TLS traffic. In a real deployment you'd typically only open 443 (and redirect 80→443), but both are included here so the lab VM can serve either.

Verify

```bash
openstack security group rule list web-secgroup
```

---

# Lab 11 – Launch Virtual Machine (Nova)

```bash
openstack server create \
--image Ubuntu22 \
--flavor small \
--network private-net \
--security-group web-secgroup \
--key-name mykey \
ubuntu-vm
```

Verify

```bash
openstack server list
```

**Why this is the "assembly" step:** Notice this single command references *everything* built so far — image, flavor, network, security group, keypair. This is the moment you can see why the earlier steps existed: Nova's scheduler needs all five inputs to decide which compute node to place the VM on and how to configure its virtual NIC, disk, and firewall. This is conceptually identical to an EC2 `RunInstances` call, which also needs an AMI, instance type, subnet, security group, and key pair.

---

# Lab 12 – Assign Floating IP

Create Floating IP

```bash
openstack floating ip create public
```

List Floating IPs

```bash
openstack floating ip list
```

Assign to VM

```bash
openstack server add floating ip ubuntu-vm <Floating-IP>
```

Example

```bash
openstack server add floating ip ubuntu-vm 172.24.4.25
```

Verify

```bash
openstack server list
```

**Why:** The VM already has a private IP (192.168.10.x) from Lab 3, but that address is meaningless outside the tenant network. A **floating IP** is a real, routable address from the external pool that gets 1:1 NAT-mapped to the VM's private IP via the router from Lab 6. This decouples "which VM has internet-facing access" from "which VM exists" — you can unassign a floating IP from a broken VM and instantly reassign it to a replacement, without changing DNS or client-facing config. It's the same pattern as an AWS Elastic IP.

---

# Lab 13 – SSH into VM

```bash
ssh ubuntu@<Floating-IP>
```

Example

```bash
ssh ubuntu@172.24.4.25
```

**Why this is the checkpoint:** If this works, it silently proves every prior layer is correct end-to-end: Keystone auth, Neutron networking, the router's NAT, the security group rule, and the SSH key injection. Engineers treat "can I SSH in" as the go/no-go gate before doing any real work on the instance — it's the fastest single test of the whole stack.

---

# Lab 13b – Verify That the Instance Is Reachable and Operational

SSH getting you a prompt only proves the network path works — it doesn't prove the *instance itself* finished booting cleanly or is healthy. A careful engineer checks both.

From your local machine (network reachability):

```bash
ping -c 4 <Floating-IP>
```

```bash
nc -zv <Floating-IP> 22
```

From inside the VM, once connected (operational health):

```bash
# Confirm cloud-init finished successfully (not still running / not errored)
cloud-init status

# Basic system health
uptime
df -h
free -m

# Confirm the instance can reach the outside world (proves SNAT/router path works)
ping -c 4 8.8.8.8
curl -I https://ubuntu.com

# Check for boot-time errors in the system log
sudo dmesg | tail -n 50
sudo journalctl -p err -b
```

If you opened HTTP/HTTPS in Lab 10 and plan to run a web service, confirm the port is actually listening once the service is up:

```bash
sudo ss -tulpn | grep -E ':80|:443'
curl -I http://<Floating-IP>
```

**Why we check both layers separately:** "Reachable" and "operational" are two different failure domains, and conflating them wastes debugging time. `ping`/`nc` only tell you the network path (floating IP → router NAT → security group → port) is open — a VM that's still booting, stuck on cloud-init, or completely hung will still answer ping. Conversely, `cloud-init status` and `journalctl` tell you the guest OS itself came up cleanly, independent of networking. Use case: this exact two-step check (network probe, then in-guest health check) is what a startup script in a CI/CD pipeline or an auto-scaling group's health check does before marking a new instance "in service" — you never want a load balancer routing traffic to a VM that's reachable but not actually ready.

---

# Lab 14 – Create Cinder Volume

```bash
openstack volume create \
--size 10 \
volume1
```

Verify

```bash
openstack volume list
```

**Why a separate volume instead of just using the VM's root disk:** The root disk (from the Glance image) is ephemeral in most OpenStack setups — if the VM is deleted or the compute host fails, that data can be gone. A Cinder volume is persistent, network-attached block storage, independent of any one VM's lifecycle. Use case: this is exactly how you'd store a database's data directory — so you can detach it from a failed VM and reattach it to a fresh one without losing data, the same role EBS plays for EC2.

---

# Lab 15 – Attach Volume

```bash
openstack server add volume ubuntu-vm volume1
```

Inside VM

```bash
lsblk
```

You should see a new disk.

**Why:** Creating a volume just provisions the storage backend — it isn't visible to any VM until it's explicitly attached, which is when Nova/Cinder present it as a new block device (commonly `/dev/vdb`). This explicit attach/detach step is what makes volumes portable between instances.

---

# Lab 16 – Format and Mount Volume

```bash
sudo mkfs.ext4 /dev/vdb
```

Create mount point

```bash
sudo mkdir /data
```

Mount the volume:

```bash
sudo mount /dev/vdb /data
```

Verify the mounted volume:

```bash
df -h
```

Configure automatic mounting after reboot:

Get the UUID of the volume:

```bash
sudo blkid /dev/vdb
```

Example output:

```text
/dev/vdb: UUID="7ca79257-97bb-4141-bef6-d7339008ca24" TYPE="ext4"
```

Edit `/etc/fstab`:

```bash
sudo vi /etc/fstab
```

Add this line:

```text
UUID=7ca79257-97bb-4141-bef6-d7339008ca24 /data ext4 defaults 0 2
```

Test the configuration:

```bash
sudo mount -a
```

Verify:

```bash
df -h
```

This line in `/etc/fstab` has six fields:

```

UUID=<disk-uuid>  /data  ext4  defaults  0  2

```

| Field | Example | Meaning |
|---|---|---|
| 1 | `UUID=7ca79257...` | Identifies the disk by UUID (safer than `/dev/vdb` because device names can change) |
| 2 | `/data` | Directory where the filesystem will be mounted |
| 3 | `ext4` | Filesystem type |
| 4 | `defaults` | Mount options |
| 5 | `0` | Dump backup setting (old `dump` utility). `0` means do not include this filesystem in dump backups |
| 6 | `2` | Filesystem check order during boot (`fsck`). `1` is usually reserved for the root filesystem `/`; `2` checks other filesystems after root |

### What does `defaults` mean?

`defaults` is a shortcut for common mount options:

```

rw,suid,dev,exec,auto,nouser,async

```

Meaning:

- `rw` → mount read/write
- `suid` → allow set-user-ID bits
- `dev` → allow device files
- `exec` → allow execution of binaries
- `auto` → mount automatically at boot
- `nouser` → only root can mount/unmount
- `async` → use asynchronous writes

For a normal data disk like `/data`, `defaults 0 2` is the standard choice.

**Why:** A raw block device has no filesystem yet — attaching it just gives the OS a disk, not something it can write files to. This is identical to plugging in a brand-new physical hard drive: you still have to partition/format it before `/data` means anything. Skipping `mkfs` is a classic mistake that leads to "I/O error" the first time someone tries to write to the mount point.

---

# Lab 17 – Create Volume Snapshot

```bash
openstack volume snapshot create snapshot1 \
--volume volume1
```

Verify

```bash
openstack volume snapshot list
```

**Why:** A snapshot is a point-in-time, incremental copy of the volume's data — the standard pre-change safety net. Use case: right before a risky schema migration or OS patch, an engineer snapshots the volume so they can roll back in seconds instead of restoring from a full backup.

---

# Lab 18 – Create Instance Snapshot

```bash
openstack server image create \
ubuntu-vm \
--name ubuntu-snapshot
```

Verify

```bash
openstack image list
```

**Why this is different from Lab 17:** A volume snapshot only captures *data*. An instance snapshot captures the *entire boot disk as a new Glance image* — OS, installed packages, config files, everything. Use case: this is how you turn a manually-configured "golden" VM (say, one with nginx, your app code, and monitoring agents already installed) into a reusable template for auto-scaling or disaster recovery, without repeating all that setup by hand.

---

# Lab 19 – Create New VM from Snapshot

```bash
openstack server create \
--image ubuntu-snapshot \
--flavor small \
--network private-net \
--security-group web-secgroup \
--key-name mykey \
ubuntu-vm2
```

**Why:** This proves the snapshot from Lab 18 is a fully functional, bootable image — not just a backup file. Use case: this exact pattern is how horizontal scaling works in practice — configure one instance perfectly, snapshot it, then boot N identical copies behind a load balancer.

---

# Lab 20 – Stop VM

```bash
openstack server stop ubuntu-vm
```

Start again

```bash
openstack server start ubuntu-vm
```

Reboot

```bash
openstack server reboot ubuntu-vm
```

**Why practice this:** Stop/start/reboot are the day-2 lifecycle operations engineers use constantly — stopping an idle dev VM to save compute capacity, rebooting after a kernel update, or power-cycling a hung instance. Knowing the difference matters: `stop` gracefully shuts down the guest OS and frees the compute resources it was using, while `reboot` just restarts the guest without releasing anything.

---

# Lab 21 – Resize VM

```bash
openstack server resize \
--flavor m1.medium \
ubuntu-vm
```

Confirm

```bash
openstack server resize confirm ubuntu-vm
```

**Why the two-step confirm:** Resize actually migrates the VM to new resource allocations (or even a new compute host) and leaves the old instance in a "verify resize" state until you confirm. This safety window exists so that if the resize breaks something, you can `openstack server resize revert` instead of being stuck. Use case: this is how you'd scale a VM up from `small` to `medium` after realizing your app needs more RAM under load, without rebuilding it from scratch.

---

# Lab 22 – Create Network Port

```bash
openstack port create \
--network private-net \
port1
```

Verify

```bash
openstack port list
```

**Why:** Normally Nova creates a port automatically when you launch a VM, but pre-creating one explicitly gives you control over things like a fixed/reserved IP or a specific MAC address *before* any VM exists. Use case: this matters for scenarios like a floating/virtual IP shared by an HA pair, or when a licensing system is tied to a specific MAC address.

---

# Lab 23 – Create Additional Network

```bash
openstack network create backend-net
```

Subnet

```bash
openstack subnet create backend-subnet \
--network backend-net \
--subnet-range 10.10.10.0/24
```

**Why a second network:** Real architectures rarely put everything on one flat network. A separate `backend-net` (e.g., 10.10.10.0/24) models a common pattern: a "frontend" tier that talks to the internet, and a "backend" tier (databases, internal APIs) that should never be directly reachable externally. This is the same reasoning behind public/private subnet splits in a VPC.

---

# Lab 24 – Attach Second Network

```bash
openstack server add network ubuntu-vm backend-net
```

Verify

```bash
openstack server show ubuntu-vm
```

**Why:** This gives the VM a second NIC, one interface per network — so it can sit in *both* the public-facing tier and the internal tier simultaneously. Use case: this is exactly how a bastion/jump host or an API gateway VM is wired — one leg on the internet-facing network, one leg on the internal network it needs to reach, with security groups controlling what crosses between them.

---

# Lab 25 – Delete Resources

Delete VM

```bash
openstack server delete ubuntu-vm
```

Delete volume

```bash
openstack volume delete volume1
```

Delete router interface

```bash
openstack router remove subnet router1 private-subnet
```

Delete router

```bash
openstack router delete router1
```

Delete subnet

```bash
openstack subnet delete private-subnet
```

Delete network

```bash
openstack network delete private-net
```

**Why the order matters here too:** Neutron enforces dependency order on teardown just like it did on creation — you can't delete a subnet while a router interface still references it, and you can't delete a network while a subnet still exists on it. This is the same principle as tearing down a Kubernetes namespace or a Terraform stack: dependents must go before their dependencies, or the API will simply reject the delete with a "resource in use" error. Engineers build muscle memory for this exact reverse order (VM → volume → router interface → router → subnet → network) because it's the fastest way to avoid cleanup errors in every teardown script.

---

# Services Covered

| Service       | What You Learn                                         | Why It Exists in the Architecture |
| ------------- | ------------------------------------------------------ | ---------------------------------- |
| **Keystone**  | Authentication using `openrc`                          | Central identity/auth so every other service trusts the same tokens |
| **Glance**    | Upload and manage VM images                             | Reusable, versioned boot templates — no VM boots from nothing |
| **Nova**      | Create, stop, reboot, resize, snapshot, and delete VMs  | The compute scheduler/orchestrator that turns "image + flavor + network" into a running VM |
| **Neutron**   | Networks, subnets, routers, ports, floating IPs         | Software-defined networking that gives tenant isolation *and* controlled external reachability |
| **Cinder**    | Volumes, attaching storage, snapshots                   | Persistent storage decoupled from any single VM's lifecycle |
| **Horizon**   | Perform the same tasks using the web UI                 | Same API calls, visual layer — good for demos, audits, and non-CLI users |
| **Placement** | Used automatically when Nova schedules VMs              | Tracks real resource inventory (CPU/RAM/disk) so Nova doesn't overcommit a host |
| **RabbitMQ**  | Internal communication between services (automatic)     | Message bus so Nova/Neutron/Cinder can coordinate asynchronously without being tightly coupled |
| **MariaDB**   | Stores OpenStack metadata (automatic)                   | Source of truth for every resource's state (what exists, who owns it) |

---

## Additional Practice (Optional but Valuable)

After completing the labs above, try these to build the deeper "why" instincts a cloud admin needs day to day:

* **Create multiple projects (tenants) and users in Keystone** — practice the multi-tenant model that lets one OpenStack cloud safely host several independent teams or customers.
* **Assign roles (e.g., `admin`, `member`) to users** — this is how you enforce least-privilege access, so a developer can launch VMs but not delete another team's network.
* **Create custom flavors (small, medium, large)** — mirrors real capacity planning: matching hardware profiles to workload types (web tier vs. batch processing vs. GPU jobs).
* **Create and apply different security groups (web server, database server)** — reinforces network segmentation: a database security group should almost never allow inbound traffic from the public internet.
* **Reserve and release floating IPs** — models how IP pools are managed at scale, since public IPv4 addresses are a genuinely scarce, billed resource in most real clouds.
* **Create multiple private networks and connect them with routers** — practice designing multi-tier topologies (frontend/backend/DB) instead of one flat network.
* **Attach multiple Cinder volumes to a VM** — common in production for separating OS disk, application data, and logs onto different volumes with different backup policies.
* **Create VM snapshots and boot new instances from them** — the foundation of golden-image workflows and auto-scaling groups.
* **Experiment with quotas (instances, cores, RAM, volumes)** — this is how an admin prevents one team from exhausting shared cloud capacity.
* **Explore the Horizon dashboard and repeat the CLI tasks through the GUI** — useful for understanding what non-CLI stakeholders (auditors, junior admins) actually see and interact with.
* **View logs for Nova, Neutron, Keystone, and Cinder** — this is where real troubleshooting happens; the CLI tells you *that* something failed, the logs tell you *why*.

Completing these exercises gives you not just the commands, but the mental model of *why* a cloud architecture is laid out this way — the same reasoning you'd apply on AWS, Azure, GCP, or any private cloud built on similar primitives.