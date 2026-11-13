# Core OpenStack Architecture

An OpenStack cloud is typically built using three main types of nodes (servers):

1. **Controller Node**
2. **Compute Node**
3. **Network Node**

Each node has a specific responsibility.

---

# OpenStack Architecture Diagram

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

---

# 1. Controller Node

The **Controller Node** is the **brain** of the OpenStack cloud.

It manages and coordinates all OpenStack services.

## Responsibilities

* User authentication
* Receives API requests
* Schedules virtual machines
* Stores cloud information in the database
* Manages images
* Controls networking
* Runs the web dashboard

Think of it as the **manager** that tells other nodes what to do.

### Common Services on the Controller Node

| Service        | Purpose                                |
| -------------- | -------------------------------------- |
| Keystone       | Authentication and identity management |
| Nova API       | Handles VM requests                    |
| Nova Scheduler | Decides where VMs should run           |
| Glance         | Stores VM images                       |
| Neutron Server | Manages networking                     |
| Horizon        | Web dashboard                          |
| RabbitMQ       | Message queue for communication        |
| MariaDB/MySQL  | Stores cloud data                      |

---

# 2. Compute Node

The **Compute Node** is where the **Virtual Machines (VMs)** actually run.

It provides the CPU and memory resources for user instances.

## Responsibilities

* Launch VMs
* Stop VMs
* Restart VMs
* Delete VMs
* Monitor VM status
* Connect VMs to virtual networks

### Main Components

| Component    | Purpose               |
| ------------ | --------------------- |
| Nova Compute | Controls VM lifecycle |
| Hypervisor   | Runs virtual machines |

### Hypervisors

A hypervisor creates and manages virtual machines.

Common hypervisors include:

* KVM (most common)
* VMware ESXi
* Hyper-V
* Xen

---

# Example

Suppose a user requests a VM with:

* 4 vCPUs
* 8 GB RAM
* Ubuntu image

The Compute Node:

* Creates the VM
* Allocates CPU and RAM
* Boots the operating system
* Connects it to the network

---

# 3. Network Node

The **Network Node** manages networking for the cloud.

It allows VMs to communicate with each other and with external networks.

## Responsibilities

* Create virtual networks
* Manage routers
* Assign IP addresses
* Provide DHCP services
* Manage Floating IPs
* Connect VMs to the Internet

### Common Services

| Service                     | Purpose                    |
| --------------------------- | -------------------------- |
| Neutron L3 Agent            | Routing                    |
| DHCP Agent                  | Assigns IP addresses       |
| Metadata Agent              | Provides instance metadata |
| Linux Bridge / Open vSwitch | Virtual switching          |

---

# Example

When a VM is created:

1. Network Node creates a virtual network.
2. DHCP assigns an IP address.
3. Router connects the VM to the Internet.
4. Floating IP allows public access.

---

# How All Three Nodes Work Together

Suppose a user wants to create a VM.

### Step 1: User Request

The user clicks **"Launch Instance"** in Horizon or uses the CLI.

↓

### Step 2: Controller Node

The Controller Node:

* Authenticates the user (Keystone)
* Finds an available Compute Node (Nova Scheduler)
* Retrieves the VM image (Glance)
* Configures networking (Neutron)

↓

### Step 3: Compute Node

The selected Compute Node:

* Downloads the image
* Creates the VM
* Starts the VM using the hypervisor

↓

### Step 4: Network Node

The Network Node:

* Creates networking
* Assigns an IP address
* Connects the VM to the network
* Optionally assigns a Floating IP for Internet access

↓

### Step 5: VM Ready

The virtual machine is now running and accessible.

---

# Communication Between Nodes

```text
User
 │
 ▼
Controller Node
 │
 ├── Authenticate user
 ├── Schedule VM
 ├── Store information
 └── Configure networking
       │
       ▼
Compute Node
 │
 ├── Create VM
 ├── Allocate CPU
 ├── Allocate RAM
 └── Start VM
       │
       ▼
Network Node
 │
 ├── Assign IP
 ├── Configure router
 ├── Enable DHCP
 └── Internet access
```

---

# Easy Analogy

Imagine a **hotel**:

### 🧑‍💼 Controller Node = Hotel Reception

* Receives guests
* Checks reservations
* Assigns rooms
* Coordinates hotel services

### 🛏️ Compute Node = Hotel Rooms

* Where guests stay
* Provides the actual accommodation (VMs)

### 🚪 Network Node = Roads and Utilities

* Connects rooms to the outside world
* Provides communication and access (networking)

---

# Summary Table

| Node                | Main Role                                  | Key Services                                                                       |
| ------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------- |
| **Controller Node** | Manages the cloud and coordinates services | Keystone, Nova API, Scheduler, Glance, Neutron Server, Horizon, RabbitMQ, Database |
| **Compute Node**    | Runs Virtual Machines                      | Nova Compute, Hypervisor (KVM, ESXi, etc.)                                         |
| **Network Node**    | Provides networking for VMs                | Neutron L3 Agent, DHCP Agent, Metadata Agent, Open vSwitch/Linux Bridge            |

---

# Quick Revision

* **Controller Node:** The **brain** of OpenStack. Manages authentication, scheduling, APIs, images, databases, and networking services.
* **Compute Node:** Runs the **virtual machines** using a hypervisor and the Nova Compute service.
* **Network Node:** Provides **network connectivity**, IP addresses, routing, DHCP, and Internet access for VMs.

> **Memory tip:** **Controller = Control**, **Compute = Compute (VMs)**, **Network = Networking**.
