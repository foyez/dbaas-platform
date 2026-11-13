# Week 0 - Before You Start: Prerequisites for the Cloud-Native Platform Course

> **Who this is for:** complete beginners aiming to become a cloud or DevOps engineer, about to start the "Building a Cloud-Native Platform: From IaaS to PaaS" track (OpenStack → Kubernetes → SKE → PaaS product).
>
> **How this guide is different from a command cheatsheet:** every section answers three questions, in order — **what is this**, **why does it exist** (what problem does it solve, what would go wrong without it), and **how does it actually work**. Commands come last, and only as much as you need to recognize the shape of things — this is the map before the territory. Two companion deep-dive references exist for hands-on practice once the concepts click: a Linux Administration guide and a Networking & Virtualization guide, both structured the same way this one is.
>
> **Why "engineering mind" matters here:** the course will hand you failures — a VM that won't boot, an instance that can't reach the network, an Operator stuck in a reconcile loop. If you only memorized commands, you'll be stuck guessing. If you understand *why* each layer exists and *what it's responsible for*, you can reason your way to the broken layer instead of randomly trying things. That's the actual skill this guide is trying to build.

## Table of Contents

1. [How to Use This Guide](#1-how-to-use-this-guide)
2. [Foundation: The Command Line & Linux Basics](#2-foundation-the-command-line--linux-basics)
3. [Linux Administration](#3-linux-administration)
4. [Networking Fundamentals](#4-networking-fundamentals)
5. [Virtualization Fundamentals](#5-virtualization-fundamentals)
6. [OpenStack: What It Is and Why](#6-openstack-what-it-is-and-why)
7. [Version Control with Git](#7-version-control-with-git)
8. [Infrastructure as Code & Terraform](#8-infrastructure-as-code--terraform)
9. [Containers & Docker](#9-containers--docker)
10. [Kubernetes Fundamentals](#10-kubernetes-fundamentals)
11. [Kubernetes Operators & CRDs](#11-kubernetes-operators--crds)
12. [HTTP & REST API Fundamentals](#12-http--rest-api-fundamentals)
13. [Programming Fundamentals](#13-programming-fundamentals)
14. [CI/CD & GitOps Concepts](#14-cicd--gitops-concepts)
15. [Frontend Basics & Vue.js](#15-frontend-basics--vuejs)
16. [Security Basics: Auth, JWT, TLS](#16-security-basics-auth-jwt-tls)
17. [Observability Fundamentals](#17-observability-fundamentals)
18. [Suggested Learning Order & Timeline](#18-suggested-learning-order--timeline)
19. [Self-Check: Are You Ready?](#19-self-check-are-you-ready)

---

## 1. How to Use This Guide

The course is five weeks, each building on the last: raw infrastructure (week 1) → a Kubernetes-based product (week 2) → an API for that product (week 3) → a UI and secure public access (week 4) → observability and audit (week 5). That structure isn't arbitrary — it mirrors how real cloud platforms are actually layered, bottom to top. **You cannot skip layers and understand the ones above them**, because each week's tooling is built entirely on top of the previous week's concepts working correctly.

This guide follows the same bottom-up order. Don't jump to Kubernetes because it sounds more exciting than filesystem permissions — if you don't understand why a process runs as a specific user, you won't understand why a Kubernetes Pod's `securityContext` matters, three layers up.

---

## 2. Foundation: The Command Line & Linux Basics

### What it is
Nearly every tool in cloud engineering — Terraform, kubectl, Docker, the OpenStack CLI — is a program you run from a terminal, not a GUI. The terminal gives you a shell (commonly `bash` or `zsh`): a program whose entire job is reading text you type, interpreting it as a command, running it, and showing you the result.

### Why it exists
GUIs are built for humans clicking one thing at a time. Cloud infrastructure needs to be **automatable** — the same steps run identically a thousand times, in scripts, in CI pipelines, on servers with no screen attached at all. The command line is text in, text out — which means it's scriptable, and scriptable means automatable. This is the actual reason "learn the CLI" isn't optional for this field: it's not tradition, it's that automation requires a text interface.

### How it works — the concepts to actually own before week 1

**The filesystem is one tree, starting at `/`.** Unlike Windows' `C:\`, `D:\`, everything — disks, USB drives, network mounts — attaches somewhere under a single root. Know these locations by function, not just name:
- `/etc` — configuration files (almost everything you'll edit as an admin lives here)
- `/var/log` — logs
- `/home` — user directories
- `/usr/bin`, `/usr/local/bin` — where installed programs actually live
- `/tmp` — scratch space, cleared on reboot

**Every file has an owner, a group, and permissions** — this is the single most important Linux concept for a beginner to internalize before touching cloud tooling, because it resurfaces at every layer above (container users, Kubernetes security contexts, SSH key permissions, cloud IAM). A file's permission string like `-rwxr-xr--` reads as three groups of three: owner / group / everyone-else, each with read/write/execute. `chmod 644 file` and `chown user:group file` are the two commands you'll use constantly to fix "permission denied" errors — and you'll hit permission-denied errors *constantly* in this course, at every layer.

**Processes and the shell.** A running program is a process, identified by a PID. Programs you start block your terminal until they finish, unless you background them (`&`) or you're running a long-lived service — which is exactly why systemd exists (§3). Understand `ps aux`, `kill <pid>`, and piping (`|`, sending one command's output into another's input) before moving forward — pipes are the fundamental composability idea of Unix, and Terraform/kubectl/Docker workflows lean on shell composition constantly (`kubectl get pods | grep Running`, etc.).

**A text editor you're not afraid of.** You will edit config files over SSH, on a machine with no GUI, more times than you can count in this course. Learn `vim` or `nano` basics now — insert mode, save-and-quit — rather than mid-crisis during week 1.

### Why beginners should not skip this
Every subsequent section in this guide assumes you can navigate a filesystem, read a permission error and know what it means, and edit a file over SSH without panicking. Skipping this to "get to the interesting stuff" is the single most common reason beginners stall out in week 1 of courses like this.

---

## 3. Linux Administration

### What it is
The layer of Linux knowledge specifically about *operating* a server: managing who can log in, what services run and how they're supervised, how the system logs what happened, and how network traffic in and out is controlled.

### Why it exists
An OpenStack node, a Kubernetes worker node, a Docker host — every one of these is, underneath, an ordinary Linux server that needs to be administered like any other: users need controlled access, services need to start reliably and restart on crash, and you need to be able to answer "what happened, and when" after something breaks. None of the cloud-native tooling in weeks 2–5 replaces this layer — it all runs *on top of* a correctly administered Linux host.

### How it works — the core pieces
- **Users & groups**: every process runs as a specific user; permissions checks are entirely UID/GID-based, not name-based. Cloud services (OpenStack's `nova`, `neutron`; Kubernetes' containerd) each run as their own dedicated system user for isolation — a compromised or buggy service shouldn't be able to touch files owned by another.
- **systemd**: the process that supervises every long-running service — starts them at boot in the right order, restarts them if they crash, and gives you one consistent interface (`systemctl`) to manage all of them. Without a supervisor like this, "my service crashed at 3am" means "my service is down until a human notices," which is unacceptable for anything running in production.
- **SSH**: how you remotely and securely administer a server with no physical access — which describes every cloud server you will ever touch in your career. Key-based authentication (not passwords) is the standard because it's both more secure and scriptable.
- **journald / rsyslog**: where "what happened" lives. When something breaks at 2am with no one watching, the logs are the only record. journald is the modern, structured, systemd-native log store; rsyslog is the older but still-common way logs get shipped to a central place for long-term storage or analysis.
- **firewalld / nftables**: the layer that decides which network traffic is allowed to reach this machine at all, independent of whatever a cloud platform's own network rules (like OpenStack security groups) say. Two independent layers of "is this traffic allowed" is a deliberate defense-in-depth pattern you'll see again and again in cloud security.
- **Package management**: how software gets installed, updated, and removed in a controlled, dependency-aware way (`apt`/`dnf`), instead of manually downloading and compiling everything — which doesn't scale and doesn't give you a way to know what's installed or roll it back.

### Practice tools
`systemctl`, `journalctl`, `ss` (what's listening on which port), `ip` (interfaces and routing), `nmcli` (network connections), `tcpdump` (see actual packets on the wire) — these become the tools you reach for first when *anything* in this course misbehaves at the infrastructure layer, before you ever open an OpenStack dashboard.

*A full command reference for this section — unit files, journalctl filtering, firewalld zones, nftables rules — is available as a standalone deep-dive if you want hands-on drilling before week 1.*

---

## 4. Networking Fundamentals

### What it is
How machines find each other and exchange data — from a single cable between two computers up to "a VM in OpenStack reaches the public internet."

### Why it exists
Nothing in this course runs on one isolated machine. OpenStack manages virtual networks for tenants; Kubernetes gives every Pod an IP and expects them to reach each other; your PaaS product's API needs to be reachable from a browser over the internet, securely. If you don't understand what's actually happening when "the instance can't reach the internet," you're stuck clicking around a dashboard hoping something fixes itself.

### How it works — build this up in order, don't skip steps

**IP addresses, subnets, and ports (start here if these aren't second nature yet).** An IP address identifies a device on a network; a subnet (`10.0.1.0/24`) defines which addresses are "local" to each other without needing a router. A port (like `:22` for SSH, `:443` for HTTPS) lets one IP address run many independent services simultaneously — without ports, a machine could only run one network service at a time.

**TCP vs UDP.** TCP guarantees ordered, reliable delivery (used for almost everything: SSH, HTTP, databases) at the cost of some overhead; UDP is fire-and-forget, no delivery guarantee, used where speed matters more than reliability (DNS queries, video streaming, VXLAN's own encapsulation). Knowing which one a service uses tells you what kind of failure to expect — TCP failures usually mean "blocked" or "nothing listening"; UDP failures can silently vanish with no error at all.

**DNS.** How a name (`api.mycompany.com`) becomes an IP address. Every "Ingress" and "public URL" concept in weeks 3–4 ultimately rests on DNS resolution working correctly.

**L2 vs L3 — the conceptual foundation everything else builds on.** L2 (data link layer) is how devices on the *same* physical or virtual segment talk directly via MAC addresses — no routing involved. L3 (network layer) is how packets cross *between* segments, via IP addresses and routing decisions made by a router. This distinction is why VLANs (an L2 concept) and routing (an L3 concept) solve different problems, and it's the single most load-bearing concept in the whole networking section — nearly everything below is "a specific technique for doing L2 or L3 better."

**Routing.** Every device — not just dedicated routers — has a routing table it consults for every outbound packet: "what's the best-matching route to this destination, and which interface does it leave from?" Without correct routing, IP connectivity between two networks simply doesn't happen, no matter how correctly everything else is configured.

**NAT (Network Address Translation).** How many private-IP machines share one public IP for outbound traffic (this is why your home devices all "look like" one IP address to the internet), and the reverse — how a public IP gets mapped to a specific private machine for inbound traffic. This second pattern is *exactly* what an OpenStack "floating IP" is.

**VLANs.** A way to carve one physical network into multiple isolated broadcast domains using a tag in the Ethernet frame — so two tenants sharing physical switches can't see each other's traffic at all, without needing separate physical cabling for each.

**VXLAN.** VLANs run out of IDs (4094 max) and can't cross routed networks. VXLAN solves both: it wraps an entire Ethernet frame inside a regular UDP packet, letting you build a virtual L2 network that spans an arbitrary IP network, with millions of possible segment IDs. This is the default technology behind OpenStack's tenant network isolation, and understanding it is what makes "why is my tenant network slow/broken" debuggable instead of mysterious.

**Linux Bridges & Open vSwitch (OVS).** A bridge is a software switch — anything plugged into it (a physical NIC, a VM's virtual network interface) behaves as if connected to the same physical switch. OVS is a more powerful, programmable version of this idea, with built-in VXLAN support and fine-grained flow-based rules — and it's specifically what OpenStack's networking component (Neutron) uses under the hood on every compute node. You cannot meaningfully debug an OpenStack networking problem without understanding what OVS is doing.

### Practice tools
`ip link`, `ip route`, `bridge`, `ovs-vsctl` — and building a real topology by hand (VMs + a Linux bridge + an OVS bridge + a NAT gateway) before you ever let OpenStack automate it for you. Doing it manually once is what makes the automated version legible later, instead of magic.

*A full deep-dive with a hands-on 3-VM lab build exists as a standalone reference for this section.*

---

## 5. Virtualization Fundamentals

### What it is
How one physical machine runs multiple independent, isolated "virtual machines" (VMs), each believing it has its own dedicated hardware.

### Why it exists
Before virtualization, "I need a server" meant buying and racking a physical machine — slow, expensive, and wasteful if that server only uses 10% of its capacity. Virtualization lets one physical machine be safely divided among many workloads, which is the literal foundation of "cloud computing" as a business model — OpenStack itself exists to automate exactly this division at scale.

### How it works
- **Hypervisor**: the layer of software that creates and manages VMs. **KVM** turns the Linux kernel itself into a hypervisor, using the CPU's own hardware virtualization support (Intel VT-x / AMD-V) so guest code runs at near-native speed instead of being fully emulated in software.
- **QEMU**: the companion program that emulates everything else a VM needs — virtual disks, network cards, BIOS — while handing the actual CPU/memory virtualization off to KVM for speed. You'll almost always see them paired as "QEMU/KVM."
- **libvirt**: a management layer on top of QEMU/KVM that gives you one consistent, declarative way (XML definitions, plus a CLI called `virsh`) to define, start, stop, and inspect VMs — instead of hand-managing long QEMU command lines. This is exactly the layer OpenStack's Nova talks to on every compute node, which is why understanding libvirt directly makes Nova's behavior legible instead of magic.
- **Disk images — qcow2 vs raw**: a VM's virtual hard drive is just a file. `raw` is the simplest possible format (fastest, but no snapshots, no thin provisioning). `qcow2` adds snapshots and thin provisioning (only uses disk space actually written) and lets new images derive cheaply from a shared base image — this is how a cloud provider creates thousands of VMs from one "golden image" without copying the whole disk every time.
- **cloud-init**: how a freshly booted VM gets customized on first boot — hostname, SSH keys, network config, startup scripts — *without* needing a different disk image for every possible configuration. This is what lets one base image become a thousand differently-configured instances, and it's the mechanism behind every "user data" field you'll ever pass to Nova.

### Practice tools
`virsh`, `virt-install`, `virt-manager` — build a VM by hand from a base cloud image with a cloud-init seed before letting Nova automate the exact same process for you.

*A full deep-dive covering install steps and worked examples exists as a standalone reference for this section.*

---

## 6. OpenStack: What It Is and Why

### What it is
An open-source suite of services that together implement Infrastructure-as-a-Service (IaaS) — the ability to request "give me a VM with this much CPU/RAM, on this network, with this disk" through an API, and have it actually provisioned, entirely without a human racking a physical server.

### Why it exists
Public clouds (AWS, GCP, Azure) proved that on-demand, API-driven infrastructure was transformative — but running your own datacenter on someone else's public cloud isn't always the right answer (cost at scale, data sovereignty, regulatory requirements). OpenStack is what lets an organization run "their own AWS" on their own hardware, with the same API-driven, self-service model.

### How it works — the core services, and why each exists as a *separate* service instead of one monolith
- **Keystone** — identity and authentication. Every other service asks Keystone "is this request actually authorized?" before doing anything. Centralizing auth in one place, instead of every service reinventing it, is a deliberate and important design choice you'll see repeated in Kubernetes (RBAC) later.
- **Nova** — compute. Owns the lifecycle of virtual machines: scheduling them onto a compute node, and talking to libvirt on that node to actually boot them.
- **Neutron** — networking. Owns virtual networks, subnets, routers, security groups, and floating IPs — implemented, under the hood, using exactly the Linux networking primitives from §4 (OVS, network namespaces, NAT rules), just automated per-tenant.
- **Cinder** — block storage. Persistent virtual disks that can be attached to a VM independent of that VM's own lifecycle — so you can delete a VM without losing its data disk, or move a disk between VMs.
- **Glance** — image storage. Where the qcow2/raw base images from §5 live, ready for Nova to boot new instances from.
- **Horizon** — the web dashboard. A UI over all the above APIs — genuinely optional to use day-to-day once you're comfortable with the CLI/API, but useful for building intuition early.

**Why understanding this matters more than memorizing it:** when `nova boot` fails, the actual flow is Keystone (auth check) → Nova API → Nova scheduler (which compute node?) → Nova compute (talks to libvirt) → Neutron (attach a network port) → Cinder (attach a volume, if any) → Glance (fetch the image). A failure could be at any one of these — and each maps to something you *already* understand from §§3–5 underneath it (a systemd service down, a network namespace misconfigured, a permissions error on a disk file). This is the whole point of learning the layers below OpenStack first: OpenStack's own errors stop being mysterious once you can name which underlying layer is actually failing.

### Practice
**DevStack** is how you'll get hands-on with all of this — a shell-script installer that builds a full single-node OpenStack from source, intended for development/learning rather than production. Its logs and service structure have their own quirks (screen/systemd naming, `/opt/stack` layout) distinct from a packaged production install, worth knowing going in so you're not confused by DevStack-specific behavior.

---

## 7. Version Control with Git

### What it is
A system for tracking every change ever made to a set of files, who made it, and why — and for multiple people to work on the same codebase without overwriting each other.

### Why it exists
Every piece of this course — Terraform code, Kubernetes manifests, application source, CI/CD pipelines — is text files that change over time and need a history, a way to collaborate, and a way to roll back a bad change. Without version control, "who changed this and why" and "how do I undo this" have no good answer. GitOps (§14) — an entire methodology this course uses — is built entirely on the premise that Git is the source of truth for infrastructure state, so Git fluency isn't optional supporting knowledge here, it's load-bearing.

### How it works — the core mental model
A Git repository tracks **commits** — snapshots of your files at a point in time, each with a message explaining the change and a link back to the previous commit, forming a history. **Branches** let you develop a change in isolation from the main line of work; **merging** (or a pull request, the collaborative review process around a merge) brings that change back in once it's reviewed and ready.

```bash
git init                       # start tracking a new project
git clone <url>                 # copy an existing repo
git add <file>                  # stage a change
git commit -m "message"         # record a snapshot
git push / git pull             # sync with a remote repo (e.g. GitHub)
git branch feature-x && git checkout feature-x   # work in isolation
git log --oneline               # history
git diff                        # see uncommitted changes
```

### Why beginners underestimate this
It's tempting to treat Git as "just how I save my code to GitHub" and move on. But GitOps in week 1's bonus and throughout weeks 2–4 means your actual *infrastructure state* — not just application code — lives in Git, and a CI/CD pipeline reacts to Git changes automatically. If you're shaky on branches, merges, and resolving conflicts, that shakiness becomes infrastructure incidents once automation is watching your repo.

---

## 8. Infrastructure as Code & Terraform

### What it is
Describing infrastructure (VMs, networks, Kubernetes clusters) as text files, instead of clicking through a dashboard or running one-off CLI commands — and using a tool that reads that description and makes reality match it.

### Why it exists
Manually clicking "create VM" in a dashboard doesn't scale, isn't repeatable, and leaves no record of *why* something exists. If a human has to remember every manual step to rebuild an environment, that environment isn't actually recoverable in a crisis. Infrastructure as Code (IaC) makes infrastructure **versionable** (it's just text, so Git works), **repeatable** (run the same file, get the same result), and **reviewable** (a proposed infrastructure change can be read and approved like any code change, before it touches anything real).

### How it works — Terraform's core concepts
- **Providers** — plugins that know how to talk to a specific platform's API (an OpenStack provider, an AWS provider, the STACKIT provider you'll use in week 2). Terraform's core doesn't know anything about OpenStack or Kubernetes itself — providers are what teach it.
- **Resources** — the actual things you want to exist: a VM, a network, a Kubernetes cluster. You declare *what* you want, not the steps to get there.
- **State** — Terraform's record of what it believes currently exists, so it can compute the difference between "what you've declared" and "what's actually there" and make only the necessary changes, rather than recreating everything every run. Losing or corrupting state is one of the most common real-world Terraform incidents — understanding *why* state exists (it's Terraform's only memory of reality) makes that risk make sense instead of feeling arbitrary.
- **The core workflow**:
```bash
terraform init      # download providers, initialize state
terraform plan       # show what WOULD change, without changing anything
terraform apply      # actually make it happen
terraform destroy    # tear it down
```
`plan` before `apply` is the single most important habit to build early — it's how you catch "this is about to delete my database" before it happens instead of after.

### Why this matters for the course specifically
Week 1 has you provisioning an OpenStack VM with Terraform; week 2 has you provisioning an entire Kubernetes cluster (SKE) with it. Same tool, same three concepts, different provider — which is exactly the point: learn Terraform's model once, and it transfers everywhere.

---

## 9. Containers & Docker

### What it is
A way to package an application together with everything it needs to run (libraries, dependencies, config) into one portable unit that runs identically anywhere — a laptop, a test server, a production cluster.

### Why it exists
"It works on my machine" is a real and constant problem — differences in installed library versions, OS configuration, or environment variables between where you built something and where you run it. Containers solve this by packaging the *entire* runtime environment, not just the code.

### How it works — containers vs VMs (a distinction beginners often blur)
A VM virtualizes hardware — each VM has its own full OS kernel, which is why VMs are heavier and slower to start. A container virtualizes at the OS level — all containers on a host share the *same* kernel, isolated from each other using Linux kernel features (namespaces for isolation, cgroups for resource limits) rather than full hardware emulation. This is why containers start in milliseconds instead of the tens-of-seconds a VM takes, and why you can run far more containers than VMs on the same hardware. Understanding this distinction is also what makes "why does Kubernetes run containers, and why did we bother with VMs at all in week 1" make sense — they solve different problems at different layers, and a real cloud platform uses both.

- **Image** — the packaged, immutable template (built from a `Dockerfile`, a text file describing how to build it layer by layer).
- **Container** — a running instance of an image, analogous to how a process is a running instance of a program.
- **Registry** — where built images are stored and pulled from (Docker Hub publicly, or a private one like the STACKIT Container Registry you'll use in week 3).

```bash
docker build -t myapi:latest .
docker run -p 8080:8080 myapi:latest
docker push myregistry.example.com/myapi:latest
```

### Why this matters for the course
Every service you deploy to Kubernetes in weeks 2–5 starts as a Docker image. If you don't understand what's actually inside that image and why it was built the way it was, debugging "why won't my Pod start" in Kubernetes is guesswork.

---

## 10. Kubernetes Fundamentals

### What it is
A system for running many containers across many machines, automatically — deciding which machine runs which container, restarting failed containers, and giving you one API to manage all of it as a cluster instead of managing individual servers by hand.

### Why it exists
Once you have more than a handful of containers across more than one server, manually deciding "which container goes on which machine" and "what happens when a machine dies" stops being feasible by hand. Kubernetes automates exactly this — you declare *what* you want running (and how many copies, with what resources), and it continuously works to make reality match that declaration, recovering automatically from failures. This "declare desired state, let a controller continuously reconcile reality toward it" pattern is the single most important idea in Kubernetes, and it reappears directly in §11 (Operators) as the *same* pattern applied to your own custom product.

### How it works — the core pieces
- **Control plane** — the "brain": decides scheduling, watches for desired-vs-actual drift, exposes the API you interact with (`kubectl`).
- **Worker nodes** — the machines that actually run your containers, each running a container runtime and an agent (`kubelet`) that takes instructions from the control plane.
- **Pod** — the smallest deployable unit; usually one container (sometimes a small group that must run together), with its own IP.
- **Deployment** — declares "I want N copies of this Pod running, always" — Kubernetes handles creating, replacing, and rescheduling Pods to keep that promise true.
- **Service** — a stable network identity/IP for a set of Pods, since individual Pods come and go and their IPs aren't stable — Services solve "how do other things reliably reach this constantly-changing set of Pods."

```bash
kubectl get pods
kubectl get nodes
kubectl apply -f deployment.yaml
kubectl logs <pod>
kubectl describe pod <pod>       # events + why something isn't scheduling/starting
```

### Why this matters for the course
Weeks 2–5 are *entirely* built on Kubernetes — the PaaS product itself runs as Kubernetes workloads, the API and UI deploy as Kubernetes Deployments, Ingress (week 4) is a Kubernetes concept, and the Operator pattern (§11) is Kubernetes-native. This is the single highest-leverage topic to be genuinely solid on before week 2, more than any other item in this guide.

---

## 11. Kubernetes Operators & CRDs

### What it is
A way to teach Kubernetes about *your own* custom kind of resource — not just built-in things like Pods and Deployments, but something like "a Database instance" or "a PaaS product instance" — and have Kubernetes manage its lifecycle the same automated way it manages Pods.

### Why it exists
Kubernetes ships knowing about generic building blocks (Pods, Services, Deployments). It has no built-in concept of "a managed PostgreSQL database" or whatever your PaaS product actually is. A **Custom Resource Definition (CRD)** teaches the Kubernetes API about a new resource type; an **Operator** is the program that watches for those custom resources and does the actual work to make reality match what's declared — provisioning the underlying Deployment, Service, storage, etc. This is precisely how you build "a product" on Kubernetes instead of just running containers on it — it's the mechanism behind week 2's entire deliverable.

### How it works — the reconciler pattern
An Operator runs a **reconcile loop**: continuously watch for custom resources of your type, compare desired state (what the resource says should exist) against actual state (what's really running), and take whatever action closes the gap — create missing pieces, fix drifted ones, clean up deleted ones. This loop is expected to be **idempotent** — running it repeatedly with no actual drift should be a no-op, because it *will* run repeatedly, constantly, by design.

```yaml
apiVersion: mycompany.com/v1
kind: ManagedDatabase
metadata:
  name: my-db
spec:
  size: 10Gi
  version: "14"
```
This is a Custom Resource. Your Operator watches for `ManagedDatabase` objects and reconciles: sees a new one, creates the underlying Deployment/PVC/Service; sees one deleted, cleans them up.

### Why beginners find this hard
It's an abstract pattern layered on top of an abstract system (Kubernetes itself). Don't try to learn this from theory alone — build the *smallest possible* toy Operator (kubebuilder or Operator SDK, both Go-based) that manages something trivial before week 2, so the reconcile loop is a concrete thing you've seen work, not a diagram you've read.

---

## 12. HTTP & REST API Fundamentals

### What it is
HTTP is the protocol nearly the entire web (and nearly every cloud API, including OpenStack's and Kubernetes' own) is built on. REST is a set of conventions for designing APIs *using* HTTP in a consistent, predictable way.

### Why it exists
Every service in this course — OpenStack, Kubernetes, your own PaaS API in week 3 — exposes an HTTP API. Understanding HTTP isn't optional background; it's the actual protocol every `kubectl`, `terraform`, and `curl` command you run is speaking underneath.

### How it works
- **Methods** map to actions: `GET` (read), `POST` (create), `PUT`/`PATCH` (update), `DELETE` (remove).
- **Status codes** communicate outcome: `2xx` success, `4xx` client error (you asked for something invalid — `404` not found, `401` unauthorized), `5xx` server error (something broke on the server's side). Reading status codes correctly is the fastest first diagnostic step for *any* API failure in this course.
- **REST** conventions: resources are nouns identified by URLs (`/instances/123`), and you act on them with the HTTP method, not by encoding the verb in the URL (`/getInstance?id=123` is not RESTful; `GET /instances/123` is).
- **OpenAPI** — a standard, machine-readable format for documenting exactly what an API's endpoints, request bodies, and responses look like. This is what week 3 requires you to write, and it's also how tools can auto-generate client code or interactive docs from your API definition.

```
GET  /instances          → list all instances
POST /instances           → create a new instance
GET  /instances/{id}      → get one instance
DELETE /instances/{id}    → delete one instance
```

### Why this matters for the course
Week 3's entire deliverable is a REST API for your PaaS product. If HTTP semantics aren't second nature, you'll design an API that's technically functional but violates conventions every experienced engineer expects — which matters, because APIs are a contract other people (or your own UI in week 4) depend on.

---

## 13. Programming Fundamentals

### What it is
The ability to write code — variables, functions, loops, conditionals, error handling — well enough to build a real service, not just follow a tutorial.

### Why it exists
This course isn't only configuration and infrastructure — week 3 has you writing an actual API server with unit tests, and week 5's bonus has you writing a Go SDK. At minimum you need one language you're comfortable writing real, tested code in.

### How it works — what to actually prioritize
You don't need to be a software engineering expert, but you do need genuine comfort with:
- Basic syntax and control flow in at least one language
- Functions, and structuring code into reusable pieces instead of one long script
- Error handling — in cloud/infra code specifically, *most* of what you write is handling the ways a network call or external system can fail, not the happy path
- Writing and running unit tests
- Making HTTP calls from code (every cloud SDK, every API client, is built on this)

**Language choice for this course:** Go is a strong default — it's what Kubernetes itself, most Kubernetes Operators (kubebuilder/Operator SDK), and the week 5 bonus SDK are written in, so investing there pays off across multiple weeks rather than being single-purpose. Python is a reasonable alternative if you're coming from a more scripting-heavy background and want a gentler on-ramp, though you'll eventually want at least reading fluency in Go given how much of the Kubernetes ecosystem is written in it.

---

## 14. CI/CD & GitOps Concepts

### What it is
**CI (Continuous Integration)** — automatically building and testing code every time it changes. **CD (Continuous Delivery/Deployment)** — automatically deploying that code once it passes. **GitOps** is a specific flavor of CD where Git is the single source of truth for desired infrastructure/application state, and an automated agent continuously reconciles the real system to match what's in Git — rather than a human or a script pushing changes imperatively.

### Why it exists
Manually running tests and manually deploying doesn't scale past a handful of changes, and it's error-prone — a tired human forgetting a step is a common cause of real outages. Automating build/test/deploy removes that human error surface and makes "what's actually running" traceable back to an exact, reviewed Git commit — genuinely important for both reliability and auditability.

### How it works
A typical pipeline: code pushed to Git → CI system (GitHub Actions, GitLab CI) automatically builds and runs tests → on success, either deploys directly (CD) or updates a Git repo that a GitOps agent (like **ArgoCD**, mentioned across this course's bonus sections) is watching → the GitOps agent notices the change and reconciles the live cluster to match. Notice this is the *exact same* reconcile-loop pattern from §11's Operators, just applied to "keep the cluster's state matching Git" instead of "keep a custom resource's dependents correct" — recognizing that this is one recurring pattern, not several unrelated tools, is the real payoff of learning it properly.

### Why this matters for the course
Nearly every week's bonus section is "now automate the thing you just did manually via GitOps/CI." Doing the manual version first (which the core course scope has you do) and understanding *why* each step exists is what makes automating it meaningful instead of cargo-culting a pipeline YAML file you don't understand.

---

## 15. Frontend Basics & Vue.js

### What it is
The UI layer — what runs in a user's browser, calling your backend API and rendering the result as an actual interface a human can click through.

### Why it exists
An API alone (week 3) is only usable by other programs or people comfortable with `curl`. A real product needs a UI non-technical users can actually operate — this is what turns "a backend service" into "a product."

### How it works
Before Vue.js specifically, you need baseline **HTML** (structure), **CSS** (styling), and **JavaScript** (behavior/interactivity) — Vue is a framework built on top of all three, not a replacement for knowing them. Vue's core idea (shared by React, Angular, and most modern frontend frameworks) is **reactive, component-based UI**: you break the interface into reusable components, each with its own state, and the framework automatically re-renders whatever changed when that state updates — instead of you manually writing code to update the DOM by hand every time data changes.

### Why beginners with a backend/infra background underestimate this
If your prior experience is entirely backend or infrastructure, frontend concepts (the DOM, component lifecycles, reactive state) are a genuinely different mental model, not just "more programming." Budget real, dedicated time for this before week 4 rather than assuming general programming competence transfers automatically — it partially does, but the gaps are real.

---

## 16. Security Basics: Auth, JWT, TLS

### What it is
How you verify *who* is making a request (authentication), what they're *allowed* to do (authorization), and how you keep data private in transit between a browser/client and your servers (encryption).

### Why it exists
An API and UI exposed to the public internet (week 4) is immediately a target — without auth, anyone can call your API as if they were any user; without encryption, anyone on the network path can read (or alter) the traffic. This isn't optional hardening for later — it's required for the product to be safe to expose publicly at all.

### How it works
- **JWT (JSON Web Token)** — a compact, signed token a server issues after successful login, which the client then sends with every subsequent request to prove who they are, without the server needing to look up a session on every single call. The signature is what makes it trustworthy — the server can verify the token hasn't been tampered with, without a database round-trip.
- **TLS/SSL** — the encryption layer behind `https://`. A **certificate** proves a server's identity and enables encrypting traffic between client and server so it can't be read or modified in transit. **cert-manager** (a Kubernetes tool) automates obtaining and renewing these certificates (commonly via Let's Encrypt) instead of manually managing them — manual certificate renewal is a classic, entirely preventable cause of production outages when someone forgets.

### Why this matters for the course
Week 4 explicitly requires JWT-based auth between UI and API, plus SSL/TLS on the public-facing endpoint — this isn't background knowledge, it's a literal deliverable.

---

## 17. Observability Fundamentals

### What it is
The ability to answer "what is my system actually doing right now, and why did it just fail" — through three main signal types: **metrics** (numeric measurements over time — CPU%, request count, latency), **logs** (discrete event records — "user X created instance Y at time Z"), and eventually traces (following one request across multiple services, beyond this course's scope but worth knowing exists).

### Why it exists
"It's down" is not actionable on its own. Observability is what turns an outage from "guess and restart things" into "look at the dashboard, see exactly which component's error rate spiked, look at its logs for the exact error, fix the actual cause." At any real operational scale, you cannot SSH into every machine and manually check — you need the system to be telling you what's happening continuously, ahead of time, not only when you go looking.

### How it works
- **Prometheus** — collects and stores metrics, scraping numeric data from every service on a schedule, with a query language (PromQL) to analyze trends and set alerts.
- **Grafana** — visualizes metrics (and logs) as dashboards, turning raw numbers into something a human can actually read at a glance during an incident.
- **Loki** — aggregates logs from every service into one searchable place, so "what happened across my whole system in the last 5 minutes" doesn't mean SSHing into every individual machine.
- **Audit logging** — distinct from general application logs: a deliberate, permanent record specifically of *user actions* (who created/modified/deleted what, and when) — required for accountability and compliance, not just debugging.

### Why this matters for the course
Week 5 requires wiring up all of the above for both the platform itself (internal monitoring) and, separately, giving your *product's users* visibility into their own instances' logs — recognize these as two different audiences with two different needs, not one monitoring task.

---

## 18. Suggested Learning Order & Timeline

Don't try to learn everything simultaneously — this is designed to be worked through roughly top-to-bottom, since each layer genuinely depends on the one before it.

| Phase | Topics | Why this grouping |
|---|---|---|
| **[Phase 1 — Foundation](./phase-1.md)** | §2 Command line, §3 Linux admin, §7 Git | Nothing else is possible without these; also the fastest to get comfortable with |
| **[Phase 2 — Systems layer](./phase-2.md)** | §4 Networking, §5 Virtualization | Build the mental model before OpenStack automates it for you |
| **Phase 3 — OpenStack + IaC** | §6 OpenStack, §8 Terraform | Directly prepares for week 1 |
| **Phase 4 — Cloud-native core** | §9 Docker, §10 Kubernetes, §11 Operators/CRDs | Directly prepares for week 2 — the single most important phase to be solid on |
| **Phase 5 — Product layer** | §12 HTTP/REST, §13 Programming | Directly prepares for week 3 |
| **Phase 6 — Delivery layer** | §14 CI/CD & GitOps, §15 Frontend, §16 Security | Directly prepares for week 4, plus the bonus automation sections throughout |
| **Phase 7 — Operations layer** | §17 Observability | Directly prepares for week 5 |

If you're starting from zero with limited time before the course begins, Phases 1–4 are non-negotiable — they're the actual prerequisite for OpenStack and Kubernetes both working, and both weeks 1 and 2 will be genuinely confusing without them. Phases 5–7 can realistically be picked up just-in-time, one week ahead of when the course needs them, *if* you're already comfortable programming and have used an HTTP API before.

---

## 19. Self-Check: Are You Ready?

Before starting week 1, you should be able to honestly say yes to each of these — if not, that's exactly where to spend more prep time, not a reason to feel behind:

- [ ] I can navigate a Linux filesystem, understand file permissions, and edit a file over SSH without hesitation
- [ ] I know what a systemd service is and can check whether one is running and why it isn't
- [ ] I can explain, in my own words, the difference between L2 and L3 networking
- [ ] I understand what NAT does and why a VM with a private IP can still reach the internet
- [ ] I can explain what a hypervisor is and why KVM/QEMU/libvirt are three different, cooperating things
- [ ] I can name what Nova, Neutron, Cinder, Glance, and Keystone are each responsible for
- [ ] I'm comfortable with Git branches, commits, and resolving a merge conflict
- [ ] I understand Terraform's plan/apply workflow and why `state` exists
- [ ] I know why a container starts faster than a VM
- [ ] I can explain what a Kubernetes Deployment and Service each do, and why Services exist at all
- [ ] I understand the "reconcile loop" idea well enough to explain it without looking it up
- [ ] I know the difference between `PUT` and `PATCH`, and what a `404` vs a `500` tells me
- [ ] I can write a small, tested program in at least one language
- [ ] I understand, at a conceptual level, what a JWT proves and what TLS protects
- [ ] I can explain the difference between a metric and a log, and when I'd reach for each

This list is intentionally about **understanding**, not command memorization — you can always look up a flag; you can't look up "why is this the right layer to be debugging right now" in the middle of an incident. That instinct is what this whole guide is trying to build.