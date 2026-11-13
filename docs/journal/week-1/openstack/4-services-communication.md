# How OpenStack Services Communicate

OpenStack is made up of many independent services. These services work together using:

* **APIs**
* **Message Queue (RabbitMQ)**
* **Database (MariaDB/MySQL)**
* **Scheduler**
* **Hypervisor**

Think of OpenStack as a company where different departments communicate to complete a task.

---

# Communication Flow

```text
User
   │
   ▼
Horizon / CLI / API
   │
   ▼
Keystone (Authentication)
   │
   ▼
Nova API
   │
   ├──────────► Database (Store information)
   │
   ├──────────► Scheduler (Choose Compute Node)
   │
   ├──────────► RabbitMQ (Send commands)
   │
   ▼
Nova Compute
   │
   ▼
Hypervisor (KVM)
   │
   ▼
Virtual Machine
```

---

# 1. APIs (Application Programming Interfaces)

### What is an API?

An **API** is a way for software to communicate with another software.

Instead of directly accessing a service, users and applications send **API requests**.

### Example

A user clicks **Launch Instance** in Horizon.

Horizon sends an API request to Nova:

```
Create VM
Image = Ubuntu
CPU = 2
RAM = 4 GB
```

Nova receives the request and begins creating the VM.

### Why APIs?

They allow:

* Different OpenStack services to communicate.
* External applications to automate cloud management.
* Users to manage OpenStack via CLI, SDKs, or dashboards.

### Real-Life Analogy

📞 **API = Phone Call**

You don't walk into a restaurant's kitchen to order food. Instead, you tell the waiter (API), who passes your order to the kitchen.

---

# 2. Message Queue (RabbitMQ)

### What is RabbitMQ?

RabbitMQ is the **messaging system** used by OpenStack services.

Instead of services talking directly all the time, they send messages through RabbitMQ.

### Responsibilities

* Deliver commands between services.
* Handle asynchronous communication.
* Improve reliability.
* Prevent services from waiting on each other.

### Example

Nova API receives a request to create a VM.

Instead of contacting the Compute Node directly:

```
Nova API
      │
      ▼
RabbitMQ
      │
      ▼
Nova Compute
```

RabbitMQ delivers the message:

> "Create a VM with Ubuntu."

Nova Compute receives the message and starts creating the VM.

### Why RabbitMQ?

Without RabbitMQ:

* Services would need direct communication.
* Failures would affect many services.

With RabbitMQ:

* Messages are queued.
* Services can process them when ready.
* Communication is more reliable.

### Real-Life Analogy

📮 **RabbitMQ = Postal Service**

You send a letter to the post office. The post office delivers it to the correct recipient when possible.

---

# 3. Database (MariaDB/MySQL)

### Purpose

The database stores **configuration and state information** about the cloud.

It does **not** store virtual machines or image files.

### Stores

* Users
* Projects
* VM information
* Networks
* Volumes
* IP addresses
* Service configuration

### Example

When a VM is created:

Nova stores information like:

```
VM Name
Status
Owner
CPU
RAM
Compute Node
```

The actual VM runs on the Compute Node, but its metadata is stored in the database.

### Why a Database?

Without it, OpenStack would not know:

* Which VMs exist.
* Who owns them.
* Their status.
* Network configurations.

### Real-Life Analogy

📚 **Database = Company Records Office**

It keeps records of employees, departments, and assets.

---

# 4. Scheduler (Nova Scheduler)

### Purpose

The Scheduler decides **which Compute Node should run a VM**.

It does not create VMs.

It only chooses the best location.

### Factors Considered

* Available CPU
* Available RAM
* Available disk space
* Policies
* Availability zones

### Example

Cloud has:

| Compute Node | Free RAM |
| ------------ | -------- |
| Compute-1    | 4 GB     |
| Compute-2    | 32 GB    |
| Compute-3    | 8 GB     |

User requests:

```
16 GB RAM VM
```

Scheduler chooses:

✅ Compute-2

because it has enough resources.

### Real-Life Analogy

🏨 **Hotel Receptionist**

The receptionist checks which rooms are available before assigning a guest.

---

# 5. Hypervisor

### Purpose

The Hypervisor actually runs Virtual Machines.

Without a hypervisor, VMs cannot exist.

### Common Hypervisors

* KVM (most common)
* VMware ESXi
* Hyper-V
* Xen

### Responsibilities

* Create VMs
* Allocate CPU
* Allocate RAM
* Manage virtual disks
* Start/stop VMs

### Example

Nova Compute tells the hypervisor:

```
Create Ubuntu VM
CPU = 4
RAM = 8 GB
```

The hypervisor allocates resources and boots the VM.

### Real-Life Analogy

🏗 **Construction Worker**

Nova Compute gives the instructions, and the hypervisor does the actual construction.

---

# Complete VM Creation Process

Suppose a user launches a VM.

### Step 1 – User Login

**Keystone**

* Authenticates the user.
* Issues an authentication token.

↓

### Step 2 – API Request

User sends:

```
Launch Ubuntu VM
```

to **Nova API**.

↓

### Step 3 – Store Request

Nova records the request in the **Database**.

↓

### Step 4 – Choose Compute Node

The **Scheduler** checks resource availability and selects the best Compute Node.

↓

### Step 5 – Send Command

Nova API sends a message through **RabbitMQ** to the selected Nova Compute service.

↓

### Step 6 – Create VM

Nova Compute instructs the **Hypervisor** (such as KVM) to create and start the VM.

↓

### Step 7 – VM Running

The VM is now active and ready for use.

---

# Communication Diagram

```text
                 User
                   │
                   ▼
        Horizon / CLI / REST API
                   │
                   ▼
             Keystone (Login)
                   │
                   ▼
              Nova API
          ┌──────┼────────┐
          │      │        │
          ▼      ▼        ▼
      Database Scheduler RabbitMQ
                  │          │
                  ▼          ▼
             Compute Node (Nova Compute)
                        │
                        ▼
                  Hypervisor (KVM)
                        │
                        ▼
                 Virtual Machine
```

---

# Simple Analogy

Imagine ordering a taxi:

* **API** 🚖 = You place the booking.
* **Database** 📒 = Stores your booking details.
* **Scheduler** 🧑‍💼 = Chooses the nearest available driver.
* **RabbitMQ** 📡 = Sends the trip request to the driver.
* **Hypervisor** 🚗 = The driver who actually picks you up and completes the trip.

---

# Summary Table

| Component                    | Purpose                                              |
| ---------------------------- | ---------------------------------------------------- |
| **API**                      | Receives and processes user or application requests. |
| **RabbitMQ**                 | Delivers messages between OpenStack services.        |
| **Database (MariaDB/MySQL)** | Stores cloud configuration and state information.    |
| **Scheduler**                | Selects the best Compute Node for a VM.              |
| **Hypervisor**               | Creates and runs virtual machines on Compute Nodes.  |

---

# Quick Revision

* **API** → Entry point for requests from users, CLI, or applications.
* **RabbitMQ** → Messaging system connecting OpenStack services.
* **Database** → Stores metadata and configuration, not the actual VM data.
* **Scheduler** → Decides where a VM should run based on available resources.
* **Hypervisor** → The software that creates and runs virtual machines on physical hardware.

## Memory Tip

Remember the sequence:

**API → Database → Scheduler → RabbitMQ → Hypervisor → VM**

This is the typical flow when a user requests a new virtual machine in OpenStack.
