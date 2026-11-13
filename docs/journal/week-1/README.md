# Week 1 – IaaS Foundations & Provisioning a Kubernetes Cluster on OpenStack

## 1. Understanding the OpenStack Architecture

<details>
<summary><strong>View Contents</strong></summary>

<details>
<summary><strong>0. Cloud Computing & IaaS — Foundations</strong></summary>

**What is Cloud Computing?**
- Delivery of computing resources (VMs, storage, networks, databases, apps) **over the internet**, on demand, instead of owning physical hardware.
- You **rent** resources and pay for what you use; the provider manages the underlying hardware.
- Interview soundbite: *"Traditional = buy 20 servers, install OS, configure networking, maintain hardware. Cloud = spin up 20 VMs in minutes, pay only for usage."*

**Cloud Service Models — IaaS vs PaaS vs SaaS**

| Model | You Manage | Provider Manages | Example |
|---|---|---|---|
| **IaaS** | OS, apps, data | VMs, storage, network, physical HW | OpenStack, AWS EC2, Azure VMs |
| **PaaS** | Apps and data only | OS, runtime, infrastructure | Google App Engine, Heroku |
| **SaaS** | Nothing (just use it) | Everything | Gmail, Microsoft 365, Salesforce |

- Interview angle: This track literally builds **up** the stack — Week 1 is IaaS (OpenStack), Weeks 2–5 build a **PaaS on top of it**. Knowing this ladder cold is a good framing answer for "walk me through what you built."

**Private Cloud vs Public Cloud**
- **Public cloud** — provider-owned, shared among many customers (AWS, Azure, GCP).
- **Private cloud** — infrastructure dedicated to one organization; more control, security, and compliance. **OpenStack is commonly used to build private clouds.**

**Why do organizations use OpenStack specifically?**
- Free/open source (no licensing fees), highly scalable, hypervisor-agnostic, vendor-independent (no lock-in).
- Common in telecoms, research institutions, universities, enterprises that want a private cloud without being tied to AWS/Azure pricing or ecosystem.

**Problem it solves:** replaces manual server/network/storage provisioning with **self-service, API-driven, centrally managed** infrastructure — faster deployment, better utilization, automation-friendly.

</details>

<details>
<summary><strong>1. OpenStack Architecture</strong></summary>

**What is OpenStack?**
- An open-source **Infrastructure-as-a-Service (IaaS)** platform.
- Think of it as the **"Operating System for a Data Center"** — it takes a pool of physical servers, storage, and network hardware and turns them into virtual, on-demand resources (VMs, networks, volumes).
- It's not one program — it's a **collection of independent services** that talk to each other over REST APIs.

**Why is it needed?**
- Without it, provisioning a VM means manually racking servers, configuring hypervisors, wiring networks by hand.
- OpenStack abstracts hardware into an API-driven, self-service, multi-tenant system — the foundation cloud providers (and private clouds) are built on.

**How it's structured (high level):**
- **Modular / service-oriented architecture** — each core function (compute, network, storage, identity, image) is a separate project/service with its own API, database access, and message queue usage.
- Services communicate via:
  - **REST APIs** (synchronous, e.g., "create this VM")
  - **Message Queue** (usually **RabbitMQ**) — asynchronous internal communication between a service's own components (e.g., nova-api → nova-scheduler → nova-compute)
  - **Shared SQL Database** (usually **MySQL/MariaDB**) — each service typically owns its own DB schema to store state
- All services **register themselves with Keystone** (identity service) so they're discoverable and access-controlled centrally.

---

### Physical Layout: The Three Node Types

Beyond "which service does what," OpenStack is also physically organized into **node roles** (in DevStack single-node, these all collapse onto one VM — worth calling out explicitly in your architecture diagram). An OpenStack cloud is typically built using three main types of nodes (servers):

```text
                  Users
                    │
         (Web Dashboard / CLI / API)
                    │
                    ▼
           +-------------------+
           |  Controller Node  |
           |-------------------|
           | Keystone          |
           | Nova API          |
           | Neutron Server    |
           | Glance            |
           | Scheduler         |
           | Horizon           |
           | RabbitMQ          |
           | Database          |
           +-------------------+
                │         │
      ----------         ----------
      │                           │
      ▼                           ▼
+----------------+        +----------------+
| Compute Node 1 |        | Compute Node 2 |
|----------------|        |----------------|
| Nova Compute   |        | Nova Compute   |
| Hypervisor     |        | Hypervisor     |
| Virtual VMs    |        | Virtual VMs    |
+----------------+        +----------------+
              │
              ▼
      +------------------+
      |   Network Node   |
      |------------------|
      | Neutron L3 Agent |
      | DHCP Agent       |
      | Router           |
      | Floating IPs     |
      +------------------+
              │
              ▼
          Internet
```

| Node | Role | Runs |
|---|---|---|
| **Controller Node** | The "brain" — coordinates everything | Keystone, Nova API, Nova Scheduler, Glance, Neutron Server, Horizon, RabbitMQ, Database |
| **Compute Node** | The "muscle" — runs the actual VMs | Nova Compute, Hypervisor (KVM/ESXi/Hyper-V/Xen) |
| **Network Node** | Provides connectivity | Neutron L3 Agent (routing), DHCP Agent, Metadata Agent, Open vSwitch/Linux Bridge |

**Analogy (hotel):** Controller = reception (checks you in, assigns a room) → Compute = the rooms themselves (where you actually stay) → Network = roads/utilities (how you get in/out and connect to the world).

- Interview angle: A **multi-node production OpenStack** separates these physically for scalability/fault isolation. **DevStack** collapses them onto a single host purely for dev/learning — make sure you can say *why* that distinction matters (blast radius, scaling compute independently of control plane, etc.).

</details>

<details>
<summary><strong>2. Core Components & Responsibilities</strong></summary>

**Keystone — Identity Service**
- What: Central authentication & authorization service.
- Why: Every other service needs to know "who is this user/request" and "are they allowed to do this." Keystone issues **tokens** that are passed to every API call.
- Interview angle: It's the **first thing every request touches** — without a valid Keystone token, no other service will act.

**Nova — Compute Service**
- What: Manages the lifecycle of **Virtual Machines** (create, schedule, resize, delete).
- Key sub-parts:
  - `nova-api` — receives requests
  - `nova-scheduler` — decides *which physical host* runs the VM
  - `nova-compute` — runs on each hypervisor host, actually talks to the hypervisor (e.g., KVM/libvirt) to spin up the VM
- Why: This is the actual "give me a VM" engine — the compute equivalent of Kubernetes' scheduler+kubelet, but for VMs instead of containers.

**Neutron — Networking Service**
- What: Manages **virtual networks**, subnets, routers, floating IPs, security groups.
- Why: VMs need isolated, tenant-specific networks (not just one shared flat network) — Neutron provides Software-Defined Networking (SDN) so each tenant can have their own private network topology.
- Backend note: Often implemented via **Open vSwitch (OVS)** or **OVN**.

**Cinder — Block Storage Service**
- What: Manages persistent **volumes** (like virtual hard drives) that can be attached/detached from VMs.
- Why: VM's local disk is ephemeral (lost when VM is deleted) — Cinder volumes persist independently, similar in spirit to a Kubernetes **PersistentVolume**.

**Glance — Image Service**
- What: Stores and manages **VM images/templates** (e.g., an Ubuntu 22.04 base image).
- Why: Nova needs a source image to boot a new VM from — Glance is that catalog/registry.

**Swift — Object Storage Service** *(often not enabled in DevStack by default)*
- What: Scalable, redundant **object storage** (like AWS S3) — good for large unstructured data (backups, images, static files).
- Why worth knowing: Different from Cinder — Swift = objects/files via HTTP API, Cinder = block devices attached to VMs.

**Horizon — Dashboard**
- What: The web-based GUI for OpenStack.
- Why: Optional but common — lets users manage VMs/networks/volumes without the CLI.

**Heat — Orchestration Service** *(good to mention even if not deployed)*
- What: OpenStack's native "Infrastructure as Code" service — takes a template (HOT format) and provisions a full stack of resources.
- Why relevant: Conceptually the *native* alternative to using Terraform against OpenStack — good talking point to contrast IaC approaches.

</details>

<details>
<summary><strong>3. How Services Communicate (the 5 mechanisms)</strong></summary>

| Mechanism | Role | Analogy |
|---|---|---|
| **REST APIs** | Entry point for requests — from Horizon, CLI, SDKs, or other services | Phone call: you tell the waiter (API), waiter tells the kitchen |
| **RabbitMQ (AMQP message queue)** | Async communication *within* a service's internal components (e.g., nova-api → nova-scheduler → nova-compute). Decouples components so a slow/down component doesn't block the whole request | Postal service: drop a letter, it's delivered when possible |
| **MySQL/MariaDB (Database)** | Stores **metadata/state only** — VM name, owner, status, network config — NOT the actual VM or files | Company records office |
| **Nova Scheduler** | Decides *which compute node* should run a new VM, using Placement data (CPU/RAM/disk availability) | Hotel receptionist checking room availability |
| **Hypervisor (KVM etc.)** | The layer that actually creates/starts/stops the VM once told to | The construction worker who does the physical building |

**Full VM creation sequence (good one to have memorized cold):**
1. **Keystone** authenticates the user, issues a token.
2. Request hits **Nova API** (e.g., via Horizon or CLI).
3. Nova records the request in the **Database**.
4. **Nova Scheduler** (using **Placement** data) picks a Compute Node with enough free CPU/RAM/disk.
5. Nova API sends the "create VM" command to that node's `nova-compute` via **RabbitMQ**.
6. `nova-compute` pulls the image from **Glance**, tells the **Hypervisor** (KVM) to create and boot the VM.
7. **Neutron** configures networking — creates/attaches the virtual network, assigns an IP via DHCP, sets up routing and optionally a Floating IP for internet access.
8. **Cinder** attaches a persistent volume if the user requested extra storage.
9. VM is running; status visible via **Horizon**.

**Memory tip:** `API → Database → Scheduler → RabbitMQ → Hypervisor → VM` (network + storage attach around the edges of this core flow).

</details>

<details>
<summary><strong>4. Interview Questions</strong></summary>

- *"Walk me through what happens when you run `openstack server create`."*  
  → CLI → Keystone auth (get token) → Nova API → Nova Scheduler picks a host → Nova Compute on that host talks to libvirt/KVM → Glance provides the image → Neutron assigns network → Cinder attaches volume (if requested).
- *"What's the difference between Cinder and Swift?"*  
  → Block storage (attached like a disk) vs. object storage (accessed via HTTP API, good for unstructured data).
- *"Why does OpenStack use a message queue internally?"*  
  → Decouples components, allows async processing, improves resilience (e.g., scheduler can be down briefly without losing the request).
- *"Is OpenStack a single application?"*  
  → No — it's a collection of independent, API-driven services that together form a cloud platform.
- *"How does OpenStack compare to Kubernetes?"*  
  → OpenStack virtualizes **infrastructure** (VMs, networks, storage) — the "data center OS." Kubernetes orchestrates **containers/applications** on top of infrastructure — the "application OS." Kubernetes assumes underlying compute already exists (often provisioned by something like OpenStack or a cloud provider).
- *"What does the Nova Scheduler actually decide, and what does it use to decide?"*  
  → It decides **which compute node** runs a new VM — not whether/how to create it. It queries **Placement** for real-time CPU/RAM/disk availability across nodes, then applies scheduling policies (e.g., availability zones) to pick one.
- *"What's actually stored in the OpenStack database — the VMs themselves?"*  
  → No. The DB stores **metadata/state** (VM name, owner, status, IP assignments, volume info). The actual VM runs in memory/disk on the compute node via the hypervisor; the actual objects live in Swift; actual images live in Glance's backend storage.
- *"Why separate Controller, Compute, and Network nodes in production instead of running everything on one box (like DevStack does)?"*  
  → Independent scaling (add compute nodes without touching control plane), fault isolation (a compute node failure doesn't take down auth/scheduling), and security segmentation (network-facing components separated from internal management).
- *"Where does OpenStack sit in the IaaS/PaaS/SaaS stack, and why does that matter for this project?"*  
  → OpenStack = IaaS. The project builds a PaaS **on top of** the OpenStack/SKE IaaS layer in later weeks — good framing for explaining the overall track structure to an interviewer.

</details>

</details>

## 2. Installing & Configuring a Single-Node OpenStack Cluster (DevStack)

<details>
<summary><strong>View Contents</strong></summary>

**What is DevStack?**
- An **official OpenStack project** — a set of shell scripts that install a full, working OpenStack environment **from source**, fast.
- Not a production tool. It's built for **developers, learners, and CI testing** — quick to spin up, quick to throw away.

**Why DevStack instead of a "real" production install (e.g., Kolla-Ansible, TripleO, manual packages)?**
- Production installers are heavyweight, containerized/HA-oriented, and slow to iterate on.
- DevStack installs everything on **one VM** in ~15–30 minutes, from git source — ideal for learning the architecture and experimenting without needing real multi-node hardware.
- Trade-off to know: DevStack is **not hardened or persistent-by-design** — it's meant to be disposable (a reboot can even break it unless configured to persist).

**Prerequisites**
- A clean VM (Ubuntu is the most common/well-supported target).
- A **non-root user with sudo privileges** — DevStack explicitly refuses to run as root. Convention is to create a dedicated `stack` user.
- Reasonable resources: interview-safe minimums to quote — at least ~2 vCPUs, 8 GB RAM, 10+ GB disk for a minimal single-node setup (more if you enable extra services).

**Installation Steps (the flow, not just commands)**
1. **Create the `stack` user** and grant it passwordless sudo — DevStack's scripts expect this exact convention.

Create a dedicated `stack` user:

```sh
sudo useradd -s /bin/bash -d /opt/stack -m stack
```

Grant execute permission on the home directory (required on some Ubuntu and RHEL-based distributions):

```sh
sudo chmod +x /opt/stack
```

Grant passwordless sudo privileges:

```sh
echo "stack ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/stack
```

Switch to the `stack` user:

```sh
sudo -u stack -i
```

2. **Clone the DevStack repo**: `git clone https://opendev.org/openstack/devstack`

Clone the official DevStack repository:

```sh
git clone https://opendev.org/openstack/devstack
```

Move into the project directory:

```sh
cd devstack
```

The repository contains:

* Installation scripts
* Configuration templates
* Sample configuration files

3. **Write a `local.conf` file** — this is the single most important step (see below). It's DevStack's configuration entry point.

Example:

```ini
[[local|localrc]]
ADMIN_PASSWORD=secret
DATABASE_PASSWORD=$ADMIN_PASSWORD
RABBIT_PASSWORD=$ADMIN_PASSWORD
SERVICE_PASSWORD=$ADMIN_PASSWORD

HOST_IP=YOUR_VM_IP
```

Replace:

```text
YOUR_VM_IP
```

with your machine's IP address.

This is the minimum required configuration to install DevStack.

4. **Run `./stack.sh`** — this is the actual installer. It:
   - Installs system packages and dependencies
   - Clones the source of each OpenStack service (Nova, Neutron, Glance, Keystone, etc.) from git
   - Sets up the database (MySQL) and message queue (RabbitMQ)
   - Installs and starts every enabled service
   - Prints a summary with Horizon URL + admin credentials at the end
5. **Validate the install** (ties into the next task) — check `systemctl`/`screen` sessions, service status, and logs under `/opt/stack/logs/`.

**The `local.conf` file — what it actually does**
- It's an INI-style file (`[[local|localrc]]` section) that tells `stack.sh`:
  - **Passwords** for admin, database, RabbitMQ, service accounts (`ADMIN_PASSWORD`, `DATABASE_PASSWORD`, `RABBIT_PASSWORD`, `SERVICE_PASSWORD`)
  - **Host IP** (`HOST_IP`) — the address other things use to reach this node
  - **Which services to enable/disable** (`ENABLE_SERVICE` / `disable_service`) — e.g., Swift is often disabled by default; you explicitly opt in
  - **Log settings** (e.g., `LOGFILE`, whether logs persist to disk vs. only in screen/systemd sessions)
- Interview angle: this file explains **why your installed services might differ from a "textbook" OpenStack** — DevStack only installs what `local.conf` tells it to.

**Common Gotchas (good for "what challenges did you face")**
- Running `stack.sh` as root fails intentionally — security-by-convention.
- Re-running `stack.sh` on a broken environment vs. using `unstack.sh` + `clean.sh` to reset properly.
- Networking conflicts if `HOST_IP` / floating IP range overlaps with the host's real network.
- DevStack pulls from **git master by default** unless you pin branches — meaning "stable" isn't guaranteed unless you explicitly target a stable branch (e.g., `stable/2024.1`).

**Real-life framing**
- Think of DevStack as **"docker-compose for OpenStack"** — a fast, disposable way to stand up the whole multi-service stack for dev/test, versus a hand-built production deployment.

### Interview Questions
- *"Why use DevStack instead of installing OpenStack manually?"*
  → Speed and reproducibility for dev/test — installs a full working stack from source in one script run, vs. hours of manual per-service configuration.
- *"What's in `local.conf` and why does it matter?"*
  → It's the config entry point controlling passwords, host IP, which services get enabled, and logging — it directly explains what ends up running on your box.
- *"Is DevStack suitable for production?"*
  → No — it's explicitly a dev/CI tool; not hardened, not HA, often tracks unstable git branches unless pinned.
- *"Why does DevStack refuse to run as root?"*
  → Security convention — forces a dedicated unprivileged `stack` user with sudo, avoiding accidental system-wide damage from an installer script.

</details>

## 3. Validating the Installation (Logs & Service Status/Health)

<details>
<summary><strong>View Contents</strong></summary>

**Why validate at all?**
- `stack.sh` finishing "successfully" only means the script didn't error out — it doesn't guarantee every service is actually healthy, reachable, and correctly wired together.
- Validation is the step that turns "I ran an install script" into "I can prove my cloud works" — this is exactly the evidence an interviewer wants to hear you can produce.

**A. Checking System Logs**

- **Where logs live:** `/opt/stack/logs/` — DevStack writes a **separate log file per service** (e.g., `n-api.log` for Nova API, `n-sch.log` for Nova Scheduler, `q-svc.log` for Neutron, `key.log` for Keystone) when configured with `LOGFILE`/`SCREEN_LOGDIR` in `local.conf`.
- **Live view (older DevStack style):** services historically ran inside a `screen` session (`screen -r stack`), with each window showing one service's live output — useful for watching a service fail in real time.
- **Modern DevStack:** many services run as **systemd units**, so `journalctl -u <service-name>` is the equivalent live-log tool.
- **What to look for:**
  - Stack traces / `ERROR` or `CRITICAL` lines
  - Repeated connection failures (often DB or RabbitMQ auth/connectivity issues — a very common DevStack failure mode)
  - `stack.sh.log` (the master install log) — the first place to check if the install itself failed partway through.

**B. Checking Service Status & Health**

- **From the OpenStack CLI (most interview-relevant):**
  - `openstack service list` — confirms which services are **registered with Keystone** (the source of truth for "what's part of this cloud").
  - `openstack endpoint list` — confirms each service has a reachable API endpoint (public/internal/admin URLs).
  - `openstack compute service list` — Nova-specific: shows `nova-compute`, `nova-scheduler`, etc. and their **State** (`up`/`down`) and **Status** (`enabled`/`disabled`).
  - `openstack network agent list` — Neutron-specific: shows L3 agent, DHCP agent, etc. and whether they're alive.
- **From the OS level:**
  - `systemctl status <service>` (modern DevStack) or `screen -list` (classic DevStack) to confirm processes are actually running.
  - `sudo netstat -tulpn` / `ss -tulpn` to confirm expected ports are listening (e.g., Keystone on 5000, Horizon on 80/443).
- **End-to-end smoke test (the strongest form of validation):** actually **create a test VM** (`openstack server create ...`). If it boots and gets an IP, that's proof the whole chain — Keystone → Nova → Glance → Neutron → hypervisor — is functioning together, not just that individual services are "up."

**Interview framing — the layered validation approach:**
1. Registration layer → `openstack service list` / `endpoint list` (is it part of the catalog?)
2. Process layer → `systemctl status` / `screen -list` (is it actually running?)
3. Log layer → `/opt/stack/logs/` (is it erroring silently even if "running"?)
4. Functional layer → boot a real VM (does the whole system actually work end-to-end?)

### Interview Questions
- *"How do you know your OpenStack install actually worked, beyond `stack.sh` not erroring?"*
  → Check service registration (`openstack service list`), confirm processes are alive (`systemctl`/`screen`), check logs for silent errors, and run an end-to-end test (boot a VM).
- *"A service shows as registered in Keystone but VM creation fails — where do you look?"*
  → Check the relevant service's log under `/opt/stack/logs/` (e.g., `n-cpu.log` for compute-side failures, `n-sch.log` if scheduling fails) — registration only proves the service exists, not that it's functioning correctly.
- *"What's the difference between a service being 'registered' and being 'healthy'?"*
  → Registered = Keystone knows about it and has an endpoint on file. Healthy = the actual process is running, responding, and completing real work — you need process-level and functional checks to confirm that, not just the catalog.

</details>

## 4. Deploying a Virtual Machine via OpenStack

<details>
<summary><strong>View Contents</strong></summary>

**The big picture:** deploying a VM isn't one action — it's **assembling prerequisites first** (networking, keys, security rules), *then* launching. Both Horizon (GUI) and the CLI follow the exact same underlying sequence — the CLI just makes the steps and their dependencies explicit.

**Prerequisite chain (order matters — this is the core concept to internalize):**
1. **Router** — connects your private network(s) to the outside (`public`/external network).
2. **Network + Subnet** — your private, tenant-owned network with a CIDR range (e.g., `192.168.10.0/24`) and gateway.
3. **Plug subnet into router** — without this, your private network is isolated with no path out.
4. **Set router's external gateway** to the public network — this is what actually enables internet access (`--external-gateway public`).
5. **Key Pair** — SSH public key uploaded to OpenStack so it can be injected into the VM at boot (via cloud-init).
6. **Security Group + rules** — acts as a **virtual firewall**; by default all inbound traffic is denied, so you must explicitly open ports (e.g., TCP 22 for SSH).
7. **Only now: create the instance** — referencing image, flavor, network, key pair, and security group together.
8. **Floating IP** (optional) — a public IP from the external pool, associated to the instance's private IP via NAT, needed only if the VM must be reachable from outside the cloud.

---

### Step-by-Step: Deploying via Horizon (Dashboard)

<details>
<summary><strong>View Contents</strong></summary>

**1. Create Router**
- Network → Routers → Create Router
- Give it a name, leave default checkboxes checked
- External Network → `public`
- Create Router

**2. Create Network**
- Network → Networks → Create Network
- Give it a name, leave default checkboxes checked
- Go to the **Subnet** tab within the same wizard:
  - Name the subnet
  - Network Address → a private CIDR range (`10.*` or `192.168.*`)
  - Set the Gateway IP
- Leave Subnet Details (DHCP, DNS, etc.) as default for now

**3. Plug Network into Router**
- Network → Routers → click your new router
- Go to the **Interfaces** tab → **Add Interface**
- Select your new network → Submit

**4. (Bonus) Set up a Security Group for SSH**
- Network → Security Groups → Create Security Group
- Add Rule:
  - Rule type → SSH (auto-fills TCP port 22)
  - Remote → CIDR
  - CIDR → `0.0.0.0/0` (allow from anywhere) or restrict to a known IP range
- Attach it to your instance: Compute → Instances → (▾ arrow) → Edit Security Groups → move your group into "Allocated" → Save

**5. Create the Instance**
- Compute → Instances → **Launch Instance**
- **Details** tab → give it a name
- **Source** tab → select your Image
- **Flavor** tab → select a resource size (vCPU/RAM/disk)
- **Networks** tab → select your new network
- **Security Groups** tab → select the group you created
- **Key Pair** tab → select/add your SSH key pair
- Skip the remaining tabs (Configuration, Metadata, etc.) for a basic deploy
- Click **Launch Instance**

**6. Associate a Floating IP (for external/public access)**
- Compute → Instances → (▾ arrow next to your instance) → **Associate Floating IP**
- Click **+** → select the IP Pool → Allocate IP
- Click **Associate**

**7. SSH into the instance** using the floating IP and your private key.

</details>

---

### Step-by-Step: Deploying via CLI

> [Detail Lab](../week-1/openstack/provision/README.md)

<details>
<summary><strong>View Contents</strong></summary>

**0. Authenticate first** (source credentials → obtains a Keystone token for all following commands):
```bash
cd /opt/stack/devstack
source openrc admin admin
```

**1. Create Router**
```bash
openstack router create ROUTER_NAME
```

**2. Create Network + Subnet**
```bash
openstack network create NETWORK_NAME

openstack subnet create --network NETWORK_NAME \
  --subnet-range 192.168.10.0/24 \
  SUBNET_NAME
```

**3. Plug subnet into router + enable internet**
```bash
openstack router add subnet ROUTER_NAME SUBNET_NAME
openstack router set --external-gateway public ROUTER_NAME

# verify
openstack router show ROUTER_NAME
```

**4. Import an SSH key pair**
```bash
openstack keypair create --public-key ~/.ssh/[public_key_file] KEY_NAME

# verify
openstack keypair list
```

**5. Create a Security Group + SSH rule**
```bash
openstack security group create SECURITY_GROUP_NAME \
  --description "GROUP_DESCRIPTION"

openstack security group rule create SECURITY_GROUP_NAME \
  --protocol tcp --dst-port 22:22 --remote-ip 0.0.0.0/0

# list rules to confirm
openstack security group rule list SECURITY_GROUP_NAME
```
*(useful cleanup commands: `openstack security group delete NAME`, `openstack security group rule delete RULE_ID`)*

**6. Gather IDs needed for instance creation**
```bash
openstack flavor list          # or create one: openstack flavor create --ram 512 --disk 1 --vcpus 1 m1.tiny
openstack image list
openstack security group list
openstack keypair list
openstack network list
```

**7. Launch the instance**
```bash
openstack server create --flavor FLAVOR_ID --image IMAGE_ID \
  --key-name KEY_NAME --user-data USER_DATA_FILE \
  --security-group SEC_GROUP_NAME --network NETWORK_NAME \
  INSTANCE_NAME

# confirm it's up
openstack server list
```

**8. Allocate + associate a Floating IP**
```bash
openstack floating ip create public
openstack server add floating ip INSTANCE_NAME FLOATING_IP

# list all floating IPs
openstack floating ip list
```

**9. SSH into the instance** using the floating IP and your private key.

</details>

---

### Key concepts to be sharp on:

| Concept | What it is | Why it matters |
|---|---|---|
| **Router** | L3 virtual device connecting internal networks to external ones | Without it, private networks have no route out — mirrors a physical router's job |
| **Flavor** | A named resource template (vCPU/RAM/disk) — e.g., `m1.tiny` | VMs don't take raw CPU/RAM numbers directly; you select or define a flavor first |
| **Security Group** | Stateful virtual firewall attached to instances (not networks) | Default = deny all inbound; must explicitly whitelist (e.g., SSH on port 22 from `0.0.0.0/0`) |
| **Key Pair** | SSH keypair registered with OpenStack (only the **public** key is uploaded) | Injected into the VM's `authorized_keys` at boot via cloud-init — enables passwordless SSH access |
| **Floating IP** | A public IP from an external pool, NAT'd to a private instance IP | Private IP alone isn't reachable from outside the cloud — floating IP is what makes a VM "public-facing" |

**CLI vs Horizon — same steps, different form:**
- Horizon walks you through the same order via wizards (Network → Router → Instance → Floating IP) with checkboxes doing sensible defaults for you.
- CLI requires you to explicitly gather **IDs first** (`openstack flavor list`, `image list`, `keypair list`, `network list`, `security group list`) before `openstack server create` — good to mention this as the "CLI forces you to understand dependencies explicitly" contrast.
- Auth note: CLI requires **sourcing credentials** first (`source openrc admin admin` or `demo demo`) — this is literally obtaining a Keystone token for subsequent commands, tying back to the auth flow from Section 1.

**Persistent Storage (Cinder) — the important conceptual point**
- A Cinder **volume's lifecycle is independent of the instance's lifecycle**. This is the single most interview-worthy fact from this whole workflow.
- Proof-of-concept ("destruction test"): attach a volume → write data → **delete the VM entirely** → volume still shows `status: available` → attach it to a **brand-new** instance → data is still there.
- Practical steps: create volume (`openstack volume create`) → attach (`openstack server add volume`) → inside the VM, it shows up as a raw block device (`/dev/vdb`, visible via `lsblk`) → must be **formatted** (`mkfs.ext4`) and **mounted** before use, just like a fresh physical disk.
- Why this matters conceptually: this is the OpenStack analog of a Kubernetes **PersistentVolume** outliving a Pod — same pattern (compute is disposable, storage is durable), different layer of the stack.

**Step-by-Step: Creating & Attaching a Volume**

<details>
<summary><strong>View Contents</strong></summary>

```bash
# create a volume
openstack volume create --size SIZE_GB VOLUME_NAME

# attach it to a running instance
openstack server add volume INSTANCE_NAME VOLUME_NAME
```
- SSH into the instance, confirm the raw device appears: `lsblk` → you'll see `vdb` (unmounted, no filesystem yet)
- Format and mount it like a brand-new disk:
```bash
sudo -i
mkfs.ext4 /dev/vdb              # format the empty disk
mkdir -p /mnt/FOLDER_NAME
mount /dev/vdb /mnt/FOLDER_NAME # mount it
```
- Create a test file inside `/mnt/FOLDER_NAME/` and verify with `ls -l`.

**Step-by-Step: The "Destruction Test" (proving volume persistence)**
```bash
# 1. Cleanly unmount before deleting the instance
umount /mnt/FOLDER_NAME
exit   # exit root
exit   # exit SSH session (avoids terminal hangs)

# 2. Re-auth and delete the instance
source /opt/stack/devstack/openrc demo demo
openstack server delete INSTANCE_NAME

# 3. Confirm the volume survived (status should show "available")
openstack volume list

# 4. Create a brand-new instance
openstack server create --flavor FLAVOR_ID --image IMAGE_ID \
  --key-name KEY_NAME --user-data USER_DATA_FILE \
  --security-group SEC_GROUP_NAME --network NETWORK_NAME \
  INSTANCE_NAME

# 5. Attach the OLD volume to the NEW instance
openstack volume list
openstack server add volume INSTANCE_NAME OLD_VOLUME

# 6. Re-associate a floating IP (old one, or allocate a new one)
openstack floating ip list
openstack server add floating ip INSTANCE_NAME FLOATING_IP
```
- SSH in again (you'll likely get an SSH host-key warning since it's technically a "new" host reusing an old IP — clear the stale entry first):
```bash
ssh-keygen -f '/opt/stack/.ssh/known_hosts' -R '<floating_ip>'
```
- Mount the drive **without reformatting** — this is the critical step; `mkfs` would wipe your data:
```bash
sudo -i
mkdir -p /mnt/FOLDER_NAME
mount /dev/vdb /mnt/FOLDER_NAME
```
- Check the folder — your file from the deleted instance should still be there. ✅ Proof that Cinder volume lifecycle is fully decoupled from instance lifecycle.

</details>

### Interview Questions
- *"Walk me through everything that needs to exist before you can SSH into a new OpenStack VM."*
  → Router → private network/subnet → subnet plugged into router → router's external gateway set → security group with SSH rule → key pair uploaded → instance launched referencing all of the above → floating IP associated.
- *"Why is a security group needed if the instance already has a private network?"*
  → Networking (routing/reachability) and security (what traffic is *permitted*) are separate concerns — by default all inbound traffic is denied regardless of network connectivity; the security group is the explicit allow-list.
- *"What happens to a Cinder volume if you delete the instance it's attached to?"*
  → Nothing — the volume is a separate resource with independent lifecycle; it persists in `available` state and can be reattached to any new instance, data intact.
- *"Why do you have to run `mkfs` on a newly attached volume?"*
  → A new Cinder volume is raw, unformatted block storage — just like a brand-new physical hard disk, it needs a filesystem written before the OS can mount and use it.
- *"What's the difference between a private IP and a floating IP?"*
  → Private IP = internal to the tenant network, not reachable from outside. Floating IP = a public IP from an external pool, NAT'd to the instance — required for any inbound access from outside the cloud.

</details>

## 5. Architecture Analysis

<details>
<summary><strong>View Contents</strong></summary>

<details>
<summary><strong>1. Architecture Analysis: Current OpenStack Installation via DevStack</strong></summary>

Perform a comprehensive analysis of the existing OpenStack environment deployed using DevStack to accurately identify its architecture, components, services, and network topology — before creating a technical architecture diagram.

**How to use this document:** Every section below has the commands to run and an **example output block** showing the kind of data you're looking for. Run each command against your real environment and replace the example output with your actual results. The "What this tells you" note under each block explains what to actually pull out of the output and put on the diagram.

---

## 1. Node Topology — Single-Node vs Multi-Node

```bash
# From the controller — list all registered hypervisors (one per compute node)
source openrc admin admin
openstack hypervisor list

# Cross-check against Nova's service registry
openstack compute service list

# Check DevStack's own multi-node config, if present
cat /opt/stack/devstack/local.conf | grep -iE "HOST_IP|SERVICE_HOST|MULTI_HOST"
```

**Example output:**

```
$ openstack hypervisor list
+----+---------------------+-----------------+--------------+-------+
| ID | Hypervisor Hostname | Hypervisor Type | Host IP      | State |
+----+---------------------+-----------------+--------------+-------+
| 1  | my-server           | QEMU            | 10.0.0.114   | up    |
+----+---------------------+-----------------+--------------+-------+

$ openstack compute service list
+----+----------------+------------+----------+---------+-------+
| ID | Binary         | Host       | Zone     | Status  | State |
+----+----------------+------------+----------+---------+-------+
| 1  | nova-scheduler | my-server  | interna  | enabled | up    |
| 2  | nova-conductor | my-server  | internal | enabled | up    |
| 3  | nova-conductor | my-server  | internal | enabled | up    |
| 4  | nova-compute   | my-server  | nova     | enabled | up    |
+----+----------------+------------+----------+---------+-------+
```

**What this tells you:** Only one hostname (`my-server`) appears across all services → this is a **single-node, all-in-one deployment**, meaning controller, compute, and network agents all run on the same host. If a second hostname shows up in `hypervisor list` (e.g., `my-server01`), it's multi-node, and that host becomes a separate box on your diagram.

---

## 2. Operating System and DevStack Version

```bash
# OS release
cat /etc/os-release

# Kernel
uname -a

# DevStack version / git commit
cd /opt/stack/devstack && git log -1 --oneline
cd /opt/stack/devstack && git branch -a --contains HEAD

# OpenStack release codename (from installed packages)
openstack --version
pip show nova | grep -i version
```

**Example output:**

```
$ cat /etc/os-release
NAME="Ubuntu"
VERSION="24.04.4 LTS (Noble Numbat)"

$ cd /opt/stack/devstack && git log -1 --oneline
a1b2c3d (HEAD, origin/stable/2024.1) Update local.conf sample

$ pip show nova | grep -i version
Version: 29.0.0
```

**What this tells you:** `stable/2024.1` + Nova `29.0.0` maps to the **2024.1 (Caracal)** OpenStack release running on **Ubuntu 22.04**. Put this in the diagram's title block/legend — release version affects which services and default drivers are even available (e.g., Placement being separate vs. embedded).

---

## 3. Enabled OpenStack Services

```bash
source openrc admin admin
openstack service list
openstack endpoint list --sort-column Service

# Cross-check against DevStack's enabled-service flags
grep -iE "^enable_service|ENABLED_SERVICES" /opt/stack/devstack/local.conf
```

**Example output:**

```
$ openstack service list
+----------------------------------+------------+----------------+
| ID                               | Name       | Type           |
+----------------------------------+------------+----------------+
| 0a1b2c3d4e5f                     | keystone   | identity       |
| 1a2b3c4d5e6f                     | nova       | compute        |
| 2a3b4c5d6e7f                     | neutron    | network        |
| 3a4b5c6d7e8f                     | glance     | image          |
| 4a5b6c7d8e9f                     | cinderv3   | volumev3       |
| 5a6b7c8d9e0f                     | placement  | placement      |
| 6a7b8c9d0e1f                     | heat       | orchestration  |
+----------------------------------+------------+----------------+

$ grep -iE "^enable_service" /opt/stack/devstack/local.conf
enable_service heat h-api h-api-cfn h-eng
enable_service n-cauth
disable_service tempest
```

**What this tells you:** This is your **service inventory box** for the diagram — every row here becomes a labeled service block. Note anything conspicuously *absent* (e.g., no `swift`/`object-store` row means object storage isn't deployed) — don't draw services that aren't actually enabled, even if they're common in "typical" OpenStack diagrams.

---

## 4. Service Placement — Controller vs Compute vs Storage Node

```bash
# Every systemd unit DevStack manages, grouped by host (run this ON each node)
systemctl list-units --type=service --state=running | grep -E "devstack@"

# Or, using the screen/tmux session DevStack itself uses
sudo systemctl list-units 'devstack@*' --no-legend | awk '{print $1}'
```

**Example output:**

```
$ systemctl list-units --type=service --state=running | grep devstack@
devstack@keystone.service                       loaded active running
devstack@n-api.service                          loaded active running
devstack@n-cpu.service                          loaded active running
devstack@n-sch.service                          loaded active running
devstack@n-cond.service                         loaded active running
devstack@neutron-api.service                    loaded active running
devstack@neutron-ovn-maintenance-worker.service loaded active running
devstack@neutron-periodic-workers.service       loaded active running
devstack@neutron-rpc-server.service             loaded active running
devstack@placement-api.service                  loaded active running
devstack@q-ovn-agent.service                    loaded active running
devstack@q-ovn-metadata-agent.service           loaded active running
devstack@g-api.service                          loaded active running
devstack@c-api.service                          loaded active running
devstack@c-vol.service                          loaded active running
devstack@c-sch.service                          loaded active running
devstack@placement-api.service                  loaded active running
devstack@horizon.service                        loaded active running
```

**What this tells you:** Run this command on *every* node. In a single-node lab, everything lands in one box. In a multi-node lab, you'd run this on the compute node separately and typically only see `n-cpu` (nova-compute) and the Neutron L2 agent (`q-agt`) there — confirming compute nodes are "dumb" execution hosts while control-plane logic (`n-api`, `n-sch`, `q-svc`, `keystone`) stays on the controller. This distinction is the core of the diagram's node boxes.

---

## 5. Network Architecture

### 5a. Management Network

```bash
ip -4 addr show
hostname -I
cat /opt/stack/devstack/local.conf | grep -iE "HOST_IP|MANAGEMENT"
```

**Example output:**

```text
$ ip -4 addr show
2: enp3s0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 8750 qdisc fq_codel state UP group default qlen 1000
    inet 10.0.0.112/25 metric 100 brd 10.0.0.127 scope global dynamic enp3s0
```

**What this tells you:** `10.0.0.112/25` on `enp3s0` is the **management/API network** interface. This network carries the **OpenStack control plane** traffic, including `openstack` CLI/API requests and inter-service communication (such as RPC between OpenStack services). In your network diagram, this corresponds to the **control plane (management/API) network segment**. The `/25` subnet indicates the host belongs to the `10.0.0.0/25` network, with a broadcast address of `10.0.0.127`.

### 5b. Provider / External Network

```bash
source openrc admin admin
openstack network list --external
openstack subnet show <external-subnet-id>
```

**Example output:**

```
$ openstack network list --external
+--------------------------------------+--------+---------+
| ID                                   | Name   | Subnets |
+--------------------------------------+--------+---------+
| 7f8e9d0c-1234-4a5b-8c6d-abcdef012345 | public | ...     |
+--------------------------------------+--------+---------+

$ openstack subnet show 9a8b7c6d-...
+-------------------+---------------------------------+
| Field             | Value                           |
+-------------------+---------------------------------+
| cidr              | 172.24.4.0/24                   |
| gateway_ip        | 172.24.4.1                      |
| allocation_pools  | 172.24.4.2-172.24.4.254         |
+-------------------+---------------------------------+
```

**What this tells you:** `172.24.4.0/24` is the **provider/external network** — the pool floating IPs are allocated from (Lab 12 in the hands-on guide). On the diagram this sits *outside* the tenant boundary, connected via the router's external gateway.

### 5c. Tenant Networks (VXLAN/GRE/VLAN)

```bash
source openrc admin admin
openstack network list
openstack network show <tenant-network-id>

# Confirm the tenant network's encapsulation type directly from ML2 config
grep -iE "tenant_network_types|type_drivers" /etc/neutron/plugins/ml2/ml2_conf.ini
```

**Example output:**

```bash
$ openstack network show private -f json
{
  "provider:network_type": "geneve",
  "provider:segmentation_id": 52911,
  "router:external": false
}

$ grep -iE "tenant_network_types|type_drivers" /etc/neutron/plugins/ml2/ml2_conf.ini
type_drivers = local,flat,vlan,geneve
tenant_network_types = geneve
```

**What this tells you:** Tenant traffic is encapsulated using **Geneve**, not VXLAN. The tenant network `private` uses a **Geneve overlay tunnel** with segmentation ID `52911` (the Geneve VNI equivalent). Draw tenant networks as **overlay tunnels between compute nodes** in your diagram, with Geneve encapsulation over the physical management/data network.

The flow is:

```
VM on Compute Node A
        |
        |  Tenant network traffic
        |
        v
Geneve tunnel (UDP 6081)
        |
        |
        v
VM on Compute Node B
```

Your ML2 configuration confirms:

```ini
tenant_network_types = geneve
```

so all new tenant networks will use Geneve overlays.

This is different from a VLAN-based deployment:

```ini
tenant_network_types = vlan
```

where tenant isolation happens using **802.1Q VLAN tags** on physical switches. A VLAN deployment would require trunked switch ports and would not use overlay tunnels.

### 5d. Floating IP Configuration

```bash
openstack floating ip list
openstack router list
openstack router show router1 -f json | grep -i external_gateway
```

**Example output:**

```
$ openstack floating ip list
+--------------------------------------+---------------------+------------------+
| ID                                   | Floating IP Address | Fixed IP Address |
+--------------------------------------+---------------------+------------------+
| 3c4d5e6f-...                         | 172.24.4.25          | 192.168.10.5    |
+--------------------------------------+---------------------+------------------+
```

**What this tells you:** This is the concrete NAT mapping (`172.24.4.25` ↔ `192.168.10.5`) you'd draw as an arrow crossing the router boundary on the diagram — the exact 1:1 translation happening at `router1`'s external gateway.

### 5e. Bridge Mappings (br-int, br-ex, br-tun)

```bash
# List OVS bridges and their ports
sudo ovs-vsctl show

# Confirm the physical interface mapped to br-ex
grep -iE "bridge_mappings|OVS_PHYSICAL_BRIDGE|PUBLIC_INTERFACE" /opt/stack/devstack/local.conf
cat /etc/neutron/plugins/ml2/openvswitch_agent.ini | grep -A2 "\[ovs\]"
```

**Example output:**

```bash
$ sudo ovs-vsctl show
148ee294-53c8-4740-ae29-c62d3fee32c0
    Manager "ptcp:6640:127.0.0.1"
        is_connected: true
    Bridge br-ex
        Port patch-provnet-698b7ff9-be41-4b03-937a-ae76350018fe-to-br-int
            Interface patch-provnet-698b7ff9-be41-4b03-937a-ae76350018fe-to-br-int
                type: patch
                options: {peer=patch-br-int-to-provnet-698b7ff9-be41-4b03-937a-ae76350018fe}
        Port br-ex
            Interface br-ex
                type: internal
    Bridge br-int
        fail_mode: secure
        datapath_type: system
        Port br-int
            Interface br-int
            type: internal
        Port patch-br-int-to-provnet-698b7ff9-be41-4b03-937a-ae76350018fe
            Interface patch-br-int-to-provnet-698b7ff9-be41-4b03-937a-ae76350018fe
            type: patch
            options: {peer=patch-provnet-698b7ff9-be41-4b03-937a-ae76350018fe}

$ grep bridge_mappings /etc/neutron/plugins/ml2/openvswitch_agent.ini
grep: /etc/neutron/plugins/ml2/openvswitch_agent.ini: No such file or directory
```

**What this tells you:** The hypervisor networking uses two main Open vSwitch bridges:

* **`br-int` (integration bridge)** — the internal switching point where VM logical interfaces connect.
* **`br-ex` (external bridge)** — the external/provider network bridge.

The connection between `br-int` and `br-ex` is provided by an **OVS patch port pair**:

```
br-int
  |
  | patch-br-int-to-provnet-698b7ff9-be41-4b03-937a-ae76350018fe
  |
  | patch-provnet-698b7ff9-be41-4b03-937a-ae76350018fe
  |
br-ex
```

The absence of:

```text
br-tun
vxlan-*
geneve-*
```

means this output does **not show any local tunnel interface** on this node.

For the diagram, represent this node as:

```
              Provider / External Network
                       |
                       |
                    br-ex
                       |
                 OVS patch port
                       |
                    br-int
                       |
              VM logical interfaces
```

No VXLAN tunnel endpoint or tunnel bridge is shown in this output.

---

## 6. Neutron Backend and Networking Mechanism

```bash
# ML2 mechanism drivers in use
grep -iE "mechanism_drivers" /etc/neutron/plugins/ml2/ml2_conf.ini

# Confirm OVS vs Linux Bridge agent is actually running
systemctl list-units | grep -iE "neutron-openvswitch-agent|neutron-linuxbridge-agent"

# List all Neutron agents and their live status
source openrc admin admin
openstack network agent list
```

**Example output:**

```bash id="7t5x1c"
$ grep mechanism_drivers /etc/neutron/plugins/ml2/ml2_conf.ini
mechanism_drivers = ovn

$ openstack network agent list
+--------------------------------------+------------------------------+-----------+-------+-------+-------------------+
| ID                                   | Agent Type                   | Host      | Alive | State | Binary            |
+--------------------------------------+------------------------------+-----------+-------+-------+-------------------+
| f5e71a47-ed0c-4df4-9dd0-af464490b847 | OVN Controller Gateway agent | my-server | :-)   | UP    | ovn-controller    |
| 59c8f31e-696d-54d6-8479-8459ae226ad7 | OVN Neutron agent            | my-server | :-)   | UP    | neutron-ovn-agent |
+--------------------------------------+------------------------------+-----------+-------+-------+-------------------+
```

**What this tells you:** The Neutron mechanism driver is **OVN**, not Open vSwitch ML2. The node is running the OVN networking components:

* **OVN Controller Gateway agent (`ovn-controller`)** — handles OVN logical network flows, datapath programming, and gateway functionality on this host.
* **OVN Neutron agent (`neutron-ovn-agent`)** — provides Neutron integration with OVN services.

Both agents are running on the same host:

```text
Host: my-server
State: UP
Alive: :-)
```

---

## 7. Compute Configuration

```bash
source openrc admin admin
openstack hypervisor list --long
openstack hypervisor show <hypervisor-id>

# Confirm the actual virtualization driver/type in use
grep -iE "^virt_type|^compute_driver" /etc/nova/nova.conf

# Confirm hardware virtualization is available (KVM vs QEMU-only fallback)
kvm-ok 2>/dev/null || egrep -c '(vmx|svm)' /proc/cpuinfo
```

**Example output:**

```bash id="5qv6vh"
$ openstack hypervisor show <hypervisor_id> -f json
{
  "hypervisor_type": "QEMU",
  "host_ip": "10.0.0.112",
  "hypervisor_hostname": "my-server",
  "state": "up",
  "status": "enabled"
}

$ openstack hypervisor stats show -f json
{
  "vcpus": 4,
  "vcpus_used": 0,
  "memory_mb": 15987,
  "memory_mb_used": 512,
  "local_gb": 60,
  "local_gb_used": 0
}

$ grep virt_type /etc/nova/nova.conf
virt_type = qemu
```

**What this tells you:** This compute node is using **QEMU virtualization** (`virt_type = qemu`), meaning Nova is configured to use QEMU rather than hardware-accelerated KVM. The hypervisor backend is reported as **QEMU**, running on host `my-server` with IP `10.0.0.112`.

The available compute resources are:

* **vCPU capacity:** 4 vCPUs
* **vCPU allocated:** 0 vCPUs
* **Memory capacity:** 15,987 MB (~16 GB)
* **Memory allocated:** 512 MB
* **Local disk capacity:** 60 GB
* **Local disk allocated:** 0 GB

For the architecture diagram, annotate the compute node box as:

```text
Compute Node: my-server
IP: 10.0.0.112

Hypervisor:
  QEMU
  virt_type = qemu

Resources:
  CPU: 4 vCPU (0 used)
  RAM: 15.9 GB (512 MB used)
  Disk: 60 GB (0 GB used)
```

This indicates the compute node is currently lightly utilized and has available capacity for launching additional instances.

---

## 8. Storage Architecture

### 8a. Image Storage (Glance)

```bash
source openrc admin admin
openstack image list --long
cat /etc/glance/glance-api.conf | grep -A5 "\[glance_store\]"
```

**Example output:**

```
$ grep -A5 "\[glance_store\]" /etc/glance/glance-api.conf
[glance_store]
filesystem_store_datadir = /opt/stack/data/glance/images/
```

**What this tells you:** Images are stored on **local filesystem** (`/opt/stack/data/glance/images/`), not a backend like Swift or Ceph — typical for DevStack. Draw Glance's storage as local disk, not an object store cluster.

### 8b. Block Storage (Cinder)

```bash
source openrc admin admin
openstack volume service list
cat /etc/cinder/cinder.conf | grep -iE "^volume_driver|^enabled_backends"
sudo lvs 2>/dev/null   # if LVM backend
```

**Example output:**

```
$ grep -iE "^volume_driver|^enabled_backends" /etc/cinder/cinder.conf
enabled_backends = lvmdriver-1
volume_driver = cinder.volume.drivers.lvm.LVMVolumeDriver

$ sudo lvs
  LV                                          VG          Attr
  volume-3c4d5e6f-1234-4a5b-8c6d-abcdef012345 stack-volumes-lvmdriver-1 -wi-ao----
```

**What this tells you:** Cinder uses the **LVM driver** backed by a volume group named `stack-volumes-lvmdriver-1` — meaning volumes are actually thin/thick LVM logical volumes carved out of a loopback or real disk on the controller. Diagram this as local LVM storage, not a SAN/NFS/Ceph backend.

### 8c. Object Storage (Swift), if enabled

```bash
source openrc admin admin
openstack object store account show 2>&1 | head -5
swift stat 2>&1 | head -5
```

**Example output:**

```
$ openstack object store account show
Service Unavailable (HTTP 503)
```

**What this tells you:** A 503/connection error here (with `swift` absent from `openstack service list` in section 3) confirms Swift simply **isn't deployed** — don't draw an object storage component at all.

---

## 9. Identity and Authentication (Keystone)

```bash
source openrc admin admin
openstack endpoint list
openstack project list
openstack user list
openstack role assignment list --names
```

**Example output:**

```bash id="1q7svh"
$ openstack endpoint list -c Service -c Interface -c URL
+-----------+-----------+-------------------------------------+
| Service   | Interface | URL                                 |
+-----------+-----------+-------------------------------------+
| keystone  | public    | http://10.0.0.112/identity          |
| nova      | public    | http://10.0.0.112/compute/v2.1      |
| neutron   | public    | http://10.0.0.112/networking        |
+-----------+-----------+-------------------------------------+

$ openstack project list
+----------------------------------+---------+
| ID                               | Name    |
+----------------------------------+---------+
| a1b2c3...                        | admin   |
| d4e5f6...                        | demo    |
| g7h8i9...                        | service |
+----------------------------------+---------+
```

**What this tells you:** All OpenStack public service endpoints are hosted on the same controller IP: **`10.0.0.112`**. This confirms a **centralized service endpoint layout** where Keystone (identity), Nova (compute API), and Neutron (network API) are exposed from the same node. There is no separate identity or API node shown in this deployment.

Service endpoint mapping:

```text
Controller Node: 10.0.0.112

Keystone:
  http://10.0.0.112/identity

Nova API:
  http://10.0.0.112/compute/v2.1

Neutron API:
  http://10.0.0.112/networking
```

The OpenStack projects include:

```text
admin
demo
service
```

The `service` project is used for OpenStack service accounts (such as Nova and Neutron services) to authenticate with Keystone.

---

## 10. Service Dependencies and Communication Paths

```bash
# Confirm the message bus every service RPCs over
grep -iE "^transport_url" /etc/nova/nova.conf /etc/neutron/neutron.conf /etc/cinder/cinder.conf

# Confirm the database backend every service persists state to
grep -iE "^connection" /etc/nova/nova.conf /etc/neutron/neutron.conf /etc/keystone/keystone.conf

# Live view of RPC traffic (management UI, optional)
sudo rabbitmqctl list_queues name messages consumers | head -20
```

**Example output:**

```bash id="j1f2mz"
$ grep -iE "^transport_url" /etc/nova/nova.conf /etc/neutron/neutron.conf
/etc/nova/nova.conf:transport_url = rabbit://stackrabbit:secret@10.0.0.112:5672/
/etc/neutron/neutron.conf:transport_url = rabbit://stackrabbit:secret@10.0.0.112:5672/

$ grep -iE "^connection" /etc/nova/nova.conf /etc/keystone/keystone.conf
/etc/nova/nova.conf:connection = mysql+pymysql://root:secret@127.0.0.1/nova_cell0
/etc/nova/nova.conf:connection = mysql+pymysql://root:secret@127.0.0.1/nova_api
/etc/keystone/keystone.conf:connection = mysql+pymysql://root:secret@127.0.0.1/keystone
```

**What this tells you:** OpenStack services are using a shared **RabbitMQ message broker** and a local **MariaDB/MySQL database backend** on the controller node.

Message queue dependency:

```text id="y3j7bx"
RabbitMQ
Address:
10.0.0.112:5672

Used by:
- Nova
- Neutron
- Other OpenStack services using oslo.messaging
```

Database dependency:

```text id="8j7p0c"
MariaDB/MySQL
Address:
127.0.0.1 (local controller)

Databases:
- nova_cell0
- nova_api
- keystone
```

---

## 11. IP Addresses, Hostnames, and Interfaces per Service

```bash
# Consolidated view: hostname, IPs, and interfaces in one shot
hostnamectl
ip -4 -brief addr show
cat /etc/hosts

# Cross-reference against Keystone's endpoint list (section 9) and OVS bridges (section 5e)
```

**Example output:**

Updated according to your output:

```bash id="1v8n6p"
$ hostnamectl
Static hostname: my-server

$ ip -4 -brief addr show
lo               UNKNOWN        127.0.0.1/8
enp3s0           UP             10.0.0.112/25
br-ex            UP             172.24.4.1/24
```

**What this tells you:** This provides the reference network identity for the OpenStack node. The host is named **`my-server`** and uses two important network interfaces:

| Purpose                          | Interface | Address         |
| -------------------------------- | --------- | --------------- |
| Management/API network           | `enp3s0`  | `10.0.0.112/25` |
| External/provider network bridge | `br-ex`   | `172.24.4.1/24` |

---

## 12. Custom Configurations, Plugins, or DevStack-Specific Modifications

```bash
# The actual local.conf used to deploy — this is the single source of truth for "what's custom"
cat /opt/stack/devstack/local.conf

# Any extra DevStack plugins pulled in beyond the base install
grep -iE "^enable_plugin" /opt/stack/devstack/local.conf

# Diff against a stock sample, if you have one, to spot deviations quickly
```

**Example output:**

```
$ grep -iE "^enable_plugin" /opt/stack/devstack/local.conf
enable_plugin heat https://opendev.org/openstack/heat stable/2024.1
enable_plugin neutron-dynamic-routing https://opendev.org/openstack/neutron-dynamic-routing
```

**What this tells you:** Two non-default plugins are pulled in (Heat orchestration and Neutron dynamic routing) — these explain any services in section 3 that aren't part of the "core six." Call these out explicitly on the diagram as "DevStack plugin additions" so it's clear they're not stock behavior.

---

## 13. High Availability Check

```bash
# Count instances of each control-plane service — >1 per role would indicate HA
openstack compute service list
openstack network agent list
systemctl list-units 'devstack@*' | wc -l

# Confirm there's no load balancer / VIP in front of the API
sudo ss -tulpn | grep -E ':5000|:8774|:9696'   # keystone, nova, neutron ports
```

**Example output:**

```
$ sudo ss -tulpn | grep -E ':5000|:8774|:9696'
tcp   LISTEN  0  128  10.0.0.10:5000   ...  keystone
tcp   LISTEN  0  128  10.0.0.10:8774   ...  nova-api
tcp   LISTEN  0  128  10.0.0.10:9696   ...  nova-api
```

**What this tells you:** Every API is bound directly to `10.0.0.10` with no HAProxy/VIP in front of it, and there's exactly one instance of each service (confirmed in sections 1 and 4). **This deployment has no HA** — single point of failure at every layer (API, DB, message bus, network agents). State this explicitly as a labeled caveat on the diagram; it's an important architectural fact, not just a footnote, since it's the single biggest difference from a production OpenStack design.

---

## 14. Configuration Files to Collect for Validation

Pull copies of these before finalizing the diagram — they're the ground truth behind everything above:

```bash
mkdir -p ~/openstack-arch-review
cp /opt/stack/devstack/local.conf         ~/openstack-arch-review/
cp /opt/stack/devstack/stackrc            ~/openstack-arch-review/
cp /etc/nova/nova.conf                    ~/openstack-arch-review/
cp /etc/neutron/neutron.conf              ~/openstack-arch-review/
cp /etc/neutron/plugins/ml2/ml2_conf.ini  ~/openstack-arch-review/
cp /etc/cinder/cinder.conf                ~/openstack-arch-review/
cp /etc/glance/glance-api.conf            ~/openstack-arch-review/
cp /etc/keystone/keystone.conf            ~/openstack-arch-review/

# Package it up to attach alongside the diagram as supporting evidence
tar czf openstack-arch-review.tar.gz ~/openstack-arch-review
```

**What this tells you:** Nothing new to observe here — this is your audit trail. Anyone reviewing the finished architecture diagram should be able to trace every box and arrow back to a specific line in one of these files.

---

## Expected Deliverable (Recap)

Once every section above is filled in with real output, you'll have concrete, sourced answers for the diagram's required elements:

| Diagram Element | Sourced From Section |
|---|---|
| Infrastructure nodes | 1, 2 |
| OpenStack services and placement | 3, 4 |
| Network topology | 5 |
| Storage components | 8 |
| Service interactions | 10 |
| External connectivity | 5b, 5d |
| Management and data traffic flow | 5a, 5e, 11 |
| HA posture / caveats | 13 |

Once you've swapped in your real command output across all 14 sections, this document is ready to hand off as the factual basis for the architecture diagram — every box, arrow, and IP on the diagram should trace back to a specific command's output above.

</details>

</details>


## Resources

### OpenStack

- [Learning OpenStack: the hard way](https://abayard.com/learning-openstack-the-hard-way/)
- [Why OpenStack](https://nws.netways.de/en/why-openstack/)
- [How to Build a Private Cloud Using OpenStack](https://openmetal.io/docs/edu/openstack/how-to-build-private-cloud-using-openstack/)
- [OpenStack Reference Architecture For 100, 300 and 500 Nodes](https://fuel-ccp.readthedocs.io/en/latest/design/ref_arch_100_nodes.html)
- [OpenStack Installation: A Practitioner’s Guide](https://platform9.com/media/openstack-installation-a-practitioner-s-guide.pdf)
- [Picking the Right Hardware for Your OpenStack Cloud in 2026](https://cloudification.io/cloud-blog/a-practical-guide-to-choose-the-right-hardware-for-your-openstack-cloud-in-2026/)

### Kubernetes

- [Kubernetes Documentation](https://kubernetes.io/docs/home/)
- [Kubernetes Browser based Lab - Killercode](https://killercoda.com/kubernetes)
- [Practice Kubernetes With Hands-on Labs](https://kodekloud.com/studio/labs/kubernetes)
- [DevOps with Kubernetes](https://devopswithkubernetes.com/)
- [Kubernetes Basic Learning](https://github.com/knrt10/kubernetes-basicLearning)
- [Kubernetes The Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way)
- [Kubernetes the (Very) Hard Way](https://labs.iximiuz.com/courses/kubernetes-the-very-hard-way-0cbfd997) (see my [summary](k8s/kubernetes-the-hard-way.md))
- [Kubernetes For You](https://github.com/ajeetraina/kubernetes101)
- [Learn Kubernetes from Scratch](https://morioh.com/a/970dae68e5dc/learn-kubernetes-from-scratch-or-kubernetes-cheat-sheet)
- [Kubernetes Learning Resources](https://github.com/kubernauts/Kubernetes-Learning-Resources/blob/master/pages/resources.md)
- [Awesome Kubernetes](https://github.com/ramitsurana/awesome-kubernetes)

### Terraform

- [Terraform Basics - Nana](https://www.youtube.com/watch?v=l5k1ai_GBDE)
- [Terraform Best Practices - Nana](https://www.youtube.com/watch?v=gxPykhPxRW0)
- [Complete HashiCorp Learn Terraform tutorials](https://learn.hashicorp.com/terraform)
- [Terraform Best Practices](https://www.terraform-best-practices.com)
- [Practice Terraform With Hands-on Labs](https://kodekloud.com/studio/labs/terraform)

### Ansible

- [Ansible for the Beginners - KodeKloud](https://www.youtube.com/playlist?list=PL2We04F3Y_42_PN52bT_U5o_lt6uPQqqq)
- [What is Ansible - Nana](https://www.youtube.com/watch?v=1id6ERvfozo)
- [Getting Started with Ansible](https://www.youtube.com/playlist?list=PLT98CRl2KxKEUHie1m24-wkyHpEsa4Y70)