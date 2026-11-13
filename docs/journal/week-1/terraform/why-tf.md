# Why Terraform

## 1. The Big Picture: Software Delivery
"Done" doesn't mean tests pass — it means the code is **actually running for the user**. Software delivery = deploying, scaling, securing, and maintaining code in production.

## 2. Rise of DevOps
- **Old world**: Dev team builds → "throws it over the wall" → Ops team manually deploys.
- **Problem**: Manual work → slow, error-prone releases → snowflake servers (servers that all drift slightly different from each other = "configuration drift") → more bugs → longer release cycles → silos → company slows down.
- **New world**: Cloud (AWS/Azure/GCP) + automation tools (Terraform, Chef, Docker, Kubernetes) blur the line between Dev and Ops → **DevOps**.
- **DevOps goal**: make software delivery dramatically more efficient.
- **4 core values (CAMS)**: Culture, Automation, Measurement, Sharing.
- Real results: Nordstrom (+100% features/month, -50% defects), HP LaserJet (dev time on new features 5%→40%), Etsy (25–50 deploys/day).

## 3. Infrastructure as Code (IaC)
Writing code to define/deploy/update/destroy infrastructure — instead of clicking around a UI or running one-off commands.

### 5 Categories of IaC Tools
| Category | Tools | Purpose |
|---|---|---|
| Ad hoc scripts | Bash, Python, Ruby | Quick one-off automation |
| Configuration management | Chef, Puppet, Ansible | Install/manage software on existing servers |
| Server templating | Docker, Packer, Vagrant | Bake a full "image" (snapshot) of a server |
| Orchestration | Kubernetes, ECS, Nomad, Mesos | Manage fleets of containers/VMs at scale |
| Provisioning | Terraform, CloudFormation, Pulumi, OpenStack Heat | Create the actual infrastructure (servers, DBs, networks, etc.) |

### Ad Hoc Scripts — Pros/Cons
- ✅ Flexible, use any language.
- ❌ No standard structure, gets messy fast (spaghetti code), not idempotent by default.

### Configuration Management (Chef/Puppet/Ansible)
Key advantages over raw scripts:
- **Coding conventions** — consistent structure everyone follows.
- **💡 Idempotence (important, often overlooked concept)**: code that gives the *same end result no matter how many times you run it*. E.g., "install Apache if not already installed" instead of blindly installing every time. Ad hoc scripts need manual if-checks for this; tools like Ansible build it in.
- **Distribution** — one config applied across many servers in parallel or in rolling batches.

### Server Templating (Docker/Packer/Vagrant)
Instead of configuring servers live, you bake a ready-made "image" (OS + software + files) and deploy that everywhere.
- **VMs** (via Packer/Vagrant/hypervisors like VirtualBox): emulate entire hardware+OS → strong isolation, but heavy (slow boot, more CPU/memory).
- **Containers** (via Docker): emulate only the OS *user space*, share host kernel → boot in milliseconds, near-zero overhead, but weaker isolation than VMs.
- 🧠 **Kernel space vs user space** (a concept most people skip): Kernel space = direct hardware access, no safety net, reserved for OS core. User space = sandboxed, must go through OS APIs — this is why containers (user-space only) are lighter but less isolated than VMs.
- This is the foundation of **immutable infrastructure** (see below).

### Orchestration (Kubernetes etc.)
Handles: deploying containers efficiently, rolling/blue-green/canary updates, auto-healing (replacing crashed instances), auto-scaling, load balancing, service discovery — all via declarative YAML.

### Provisioning (Terraform etc.)
Creates the actual cloud resources: servers, DBs, load balancers, queues, firewalls, DNS, SSL certs, subnets — everything.
- Provisioning + server templating is a common combo: Terraform deploys a server using an AMI (image) built by Packer.

## 4. Benefits of IaC (why bother learning a new tool)
- **Self-service** — no single "only sysadmin who knows the magic commands" bottleneck.
- **Speed & safety** — automated = faster and more consistent than manual.
- **Documentation** — the code itself documents your infra (protects against "bus factor" — the number of people you can lose before the business can't function; you never want a bus factor of 1).
- **Version control** — full history of infra changes; can revert to a known-good state.
- **Validation** — code review + automated tests + static analysis before changes go live.
- **Reuse** — package infra into reusable modules.
- **Happiness** — removes tedious manual ops work.
- Stat: DevOps/IaC orgs deploy 200x more often, recover from failure 24x faster, lead times 2,555x shorter (2016 State of DevOps Report).

## 5. How Terraform Actually Works (the "under the hood" part most people skip)
- 🔧 Terraform is **open source, written in Go**, and **compiles into a single binary** (one per OS) — literally just a file called `terraform`. No servers of its own needed.
- You run this binary from your laptop, CI server, anywhere.
- It works by making **API calls to providers** (AWS, Azure, GCP, DigitalOcean, etc.) on your behalf — it piggybacks on infrastructure and auth (API keys) that already exists. It doesn't need its own "control plane" server.
- You write **Terraform configuration files** (HCL language) describing desired infra → Terraform translates this into the right sequence of provider API calls → `terraform apply` executes them.
- Terraform can coordinate **multiple providers in one config** (e.g., spin up an AWS server, then create a Google Cloud DNS record pointing to it) — genuinely multi-cloud in one workflow.

### 🧠 Myth-busting: "Portability between clouds"
People assume Terraform lets you write once and deploy identically to AWS/Azure/GCP. **False.** Each cloud's services are fundamentally different (different features, configs, scaling models) — Terraform doesn't magically unify them. What it *does* give you is the **same language/workflow/tooling** across all providers, even though the actual resource code is provider-specific.

## 6. Terraform vs Other Tools — Key Trade-offs (the meaty comparison part)

### a) Configuration Management vs Provisioning
- Config mgmt (Chef/Puppet/Ansible) = install/manage software on servers that already exist.
- Provisioning (Terraform/CloudFormation/Pulumi) = create the servers/infra itself.
- Common combo: Terraform provisions, Ansible configures.

### b) Mutable vs Immutable Infrastructure
- **Mutable** (Chef/Puppet/Ansible default): updates happen *in place* on existing servers → risk of configuration drift over time (each server quietly becomes unique = hard-to-reproduce bugs).
- **Immutable** (Terraform + Packer/Docker style): never change a running server. To update, build a *new* image, deploy *new* servers, kill the old ones. Easier to reason about, fewer drift bugs — but rebuilding/redeploying images takes longer, and drift can still creep in once a server is running for a while.

### c) Declarative vs Procedural (⭐ one of the most important concepts)
- **Procedural** (Chef, Ansible): you write step-by-step instructions for *how* to reach a state. Problem: rerunning old scripts isn't safe — e.g., "add 10 servers" script run twice = 20 servers, not 10. You must always know the current state and write a *new* script for every change. Reading the code alone doesn't tell you what's actually deployed — you'd need the full history of every script ever run.
- **Declarative** (Terraform, CloudFormation, Puppet, Heat, Pulumi): you just describe the *end state you want* (e.g., "15 servers should exist"). Terraform figures out the diff itself and only creates/destroys what's needed. `terraform plan` shows you exactly what will change *before* you apply it. This is why Terraform code stays small, reusable, and always reflects the real current state.

### d) General-Purpose Language (GPL) vs Domain-Specific Language (DSL)
- Chef (Ruby) and Pulumi (JS/TS/Python/Go/C#/Java) = GPL — more powerful, more familiar if you already know the language, bigger tooling ecosystem, but more ways to write inconsistent code.
- Terraform (HCL), Puppet, Ansible/CloudFormation/Heat (YAML) = DSL — easier to learn, more concise, more uniform code, but less flexible for complex logic (loops/conditionals harder).

### e) Master vs Masterless
- Chef/Puppet traditionally need a **master server** storing state and pushing/pulling updates. Pros: central visibility, can auto-revert drift. Cons: extra infra to run/maintain/secure.
- Terraform, Ansible, CloudFormation, Heat, Pulumi = **masterless** — the "master" is really just the cloud provider's own API (which you don't have to run or maintain yourself).

### f) Agent vs Agentless
- Chef/Puppet need agent software installed & running on every server (bootstrapping problem: how do you install the agent in the first place? + ongoing maintenance + security surface from open ports).
- Terraform/Ansible/CloudFormation/Heat/Pulumi = **agentless** — Terraform just calls cloud provider APIs (the provider already manages its own "agents" on physical infra); Ansible just uses SSH (which servers already run anyway).

### g) Paid vs Free
- CloudFormation & Heat: fully free tools.
- Terraform, Chef, Puppet, Ansible: usable fully in the free/open-source version for production.
- **Pulumi is the outlier**: practically requires the paid "Pulumi Service" for real production use (team state locking, encryption, fault-tolerant checkpointing aren't in the free backend options).

### h) Community Size (as of ~2022 data in the book)
Ansible and Terraform have by far the largest and fastest-growing communities (GitHub stars, contributors, StackOverflow questions, module registries) compared to Chef, Puppet, Pulumi, CloudFormation, Heat.

### i) Maturity
Chef & Puppet = oldest/most mature. Terraform = medium maturity (newer but growing fast, reached 1.0 stability). Pulumi & Heat = least mature/newest.

## 7. Combining Tools (realistic real-world setups)
1. **Terraform + Ansible** — Terraform provisions infra, Ansible configures apps on top. Easy to start, but still procedural/mutable under the hood.
2. **Terraform + Packer** — Packer bakes VM images, Terraform deploys them + rest of infra. Immutable, but VM builds are slow and deployment strategies (like blue-green) are limited natively in Terraform.
3. **Terraform + Packer + Docker + Kubernetes** — Packer builds a base image with Docker+Kubernetes agents, Terraform stands up the cluster + infra, Kubernetes manages containers on top (auto-healing, auto-scaling, rolling/canary deploys). More powerful but more layers to learn/debug.

## 8. Summary Table (mental cheat-sheet)

| | Chef | Puppet | Ansible | Pulumi | CloudFormation | Heat | Terraform |
|---|---|---|---|---|---|---|---|
| Type | Config Mgmt | Config Mgmt | Config Mgmt | Provisioning | Provisioning | Provisioning | Provisioning |
| Infra style | Mutable | Mutable | Mutable | Immutable | Immutable | Immutable | Immutable |
| Paradigm | Procedural | Declarative | Procedural | Declarative | Declarative | Declarative | Declarative |
| Language | GPL (Ruby) | DSL | DSL (YAML) | GPL | DSL | DSL | DSL (HCL) |
| Master server? | Yes | Yes | No | No | No | No | No |
| Agent needed? | Yes | Yes | No | No | No | No | No |
| Paid required? | No | No | No | Practically yes | N/A | N/A | No |

**Bottom line (why the book's author picked Terraform):** open source, cloud-agnostic, provisioning-focused, immutable, declarative, masterless, agentless, free-usable, large community, decently mature.

## 9. Quick "Did You Know" Facts (the deep-cut details)
- Terraform's binary is compiled from **Go**, giving it a single-file, cross-platform install with zero runtime dependencies.
- Terraform has **no agents and no master server of its own** — it rides entirely on cloud providers' existing APIs and your existing API keys/SSH keys.
- **Idempotence** is the hidden reason config management tools feel "safer" than raw scripts — same input, same result, every time, regardless of current state.
- **"Bus factor"** = literally how many key people you can lose before the company can't operate. IaC increases this because infra logic lives in code, not one person's head.
- Kernel space vs user space is *the* technical reason containers are lightweight (share host kernel) while VMs are heavy (each runs a full separate OS/kernel).
- `terraform plan` is essentially a **dry-run diff engine** — it's what makes the declarative model trustworthy, since you see exactly what will be added/changed/destroyed before committing.