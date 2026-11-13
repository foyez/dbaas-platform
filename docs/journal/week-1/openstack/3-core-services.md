# OpenStack Core Services

OpenStack is made up of several services, each responsible for a specific task. Together, these services provide a complete cloud platform.

---

# OpenStack Core Services Overview

| Service       | Purpose                              |
| ------------- | ------------------------------------ |
| **Keystone**  | Identity and Authentication          |
| **Nova**      | Virtual Machine (Compute) Management |
| **Neutron**   | Networking                           |
| **Glance**    | Image Management                     |
| **Cinder**    | Block Storage                        |
| **Swift**     | Object Storage                       |
| **Horizon**   | Web Dashboard                        |
| **Placement** | Resource Tracking                    |

---

# 1. Keystone (Identity Service)

**Purpose:** Authentication and authorization.

Keystone verifies **who the user is** and **what they are allowed to do**.

### Responsibilities

* User authentication
* User authorization
* Manage users
* Manage projects (tenants)
* Manage roles and permissions
* Provide service catalog

### Example

A user logs into OpenStack.

Keystone:

* Verifies username and password.
* Issues an authentication token.
* Allows access only to authorized services.

### Real-life Analogy

🏢 **Security Guard**

* Checks your ID.
* Gives you an access badge.
* Controls which rooms you can enter.

---

# 2. Nova (Compute Service)

**Purpose:** Manage Virtual Machines (VMs).

Nova creates and manages the lifecycle of instances.

### Responsibilities

* Launch VMs
* Stop VMs
* Restart VMs
* Delete VMs
* Schedule VMs on compute nodes
* Manage VM lifecycle

### Example

User requests:

* Ubuntu VM
* 4 vCPUs
* 8 GB RAM

Nova:

* Finds an available compute node.
* Creates the VM.
* Starts it.

### Real-life Analogy

🏗 **Construction Manager**

* Assigns work to available workers (compute nodes).
* Oversees the creation of new "buildings" (VMs).

---

# 3. Neutron (Networking Service)

**Purpose:** Manage networking.

Neutron connects VMs to each other and to external networks.

### Responsibilities

* Create networks
* Create subnets
* Manage routers
* Assign IP addresses
* Floating IPs
* Security groups
* DHCP

### Example

When a VM starts:

* Neutron creates the network.
* Assigns an IP address.
* Connects the VM to the Internet (if required).

### Real-life Analogy

🛣 **Road and Traffic Department**

* Builds roads.
* Assigns addresses.
* Controls traffic flow.

---

# 4. Glance (Image Service)

**Purpose:** Store and manage VM images.

An image is a template used to create a VM.

Examples:

* Ubuntu
* CentOS
* Windows Server

### Responsibilities

* Upload images
* Store images
* Retrieve images
* Manage image metadata

### Example

Before creating a VM:

Nova requests an Ubuntu image from Glance.

Glance sends the image to the compute node.

### Real-life Analogy

📚 **Library**

* Stores books (images).
* Users borrow a copy when needed.

---

# 5. Cinder (Block Storage)

**Purpose:** Provide persistent block storage.

Block storage behaves like a physical hard disk.

It remains available even if the VM is stopped or deleted (unless the volume is deleted).

### Responsibilities

* Create volumes
* Attach volumes to VMs
* Detach volumes
* Resize volumes
* Create snapshots

### Example

A database server needs extra storage.

Cinder:

* Creates a 100 GB volume.
* Attaches it to the VM.

### Real-life Analogy

💽 **External Hard Drive**

* Plug it into your computer.
* Store data permanently.

---

# 6. Swift (Object Storage)

**Purpose:** Store unstructured data as objects.

Examples:

* Photos
* Videos
* Documents
* Backups
* Logs

Unlike Cinder, Swift does not attach storage to a VM.

### Responsibilities

* Store objects
* Retrieve objects
* Replicate data
* Provide scalable storage

### Example

A website stores millions of images.

Swift stores the image files and serves them when requested.

### Real-life Analogy

☁ **Cloud File Storage**

* Similar to Google Drive or Amazon S3.
* Files are accessed directly, not mounted as disks.

---

# Cinder vs Swift

| Feature         | Cinder              | Swift                         |
| --------------- | ------------------- | ----------------------------- |
| Storage Type    | Block Storage       | Object Storage                |
| Appears as Disk | ✅ Yes               | ❌ No                          |
| Attached to VM  | ✅ Yes               | ❌ No                          |
| Best For        | Databases, OS disks | Images, backups, videos, logs |

**Easy memory tip:**

* **Cinder = Disk Drive**
* **Swift = Cloud File Storage**

---

# 7. Horizon (Dashboard)

**Purpose:** Web interface for OpenStack.

Instead of using the command line, users can manage cloud resources through a browser.

### Responsibilities

* Create VMs
* Create networks
* Manage storage
* Manage users
* View cloud resources

### Example

A user:

* Logs into Horizon.
* Clicks **Launch Instance**.
* Creates a VM without using commands.

### Real-life Analogy

🖥 **Control Panel**

* Everything can be managed through a graphical interface.

---

# 8. Placement Service

**Purpose:** Track available resources in the cloud.

Nova uses Placement to determine where a VM can run.

### Tracks

* CPU
* RAM
* Disk space
* Resource availability

### Example

A user requests:

* 8 CPUs
* 16 GB RAM

Placement checks all compute nodes.

It reports which node has enough available resources.

Nova then schedules the VM there.

### Real-life Analogy

📋 **Resource Manager**

* Keeps track of which hotel rooms are available before assigning a guest.

---

# How Core Services Work Together

Suppose a user launches a VM.

### Step 1 – Login

**Keystone**

* Authenticates the user.
* Issues a token.

↓

### Step 2 – Request VM

**Nova**

* Receives the request.

↓

### Step 3 – Check Resources

**Placement**

* Finds a compute node with enough CPU, RAM, and disk.

↓

### Step 4 – Get Image

**Glance**

* Provides the selected VM image.

↓

### Step 5 – Configure Network

**Neutron**

* Creates networking.
* Assigns an IP address.
* Configures routing and security groups.

↓

### Step 6 – Attach Storage (Optional)

**Cinder**

* Attaches a persistent volume if requested.

↓

### Step 7 – Access

**Horizon**

* Displays the VM status through the web dashboard.

---

# Service Interaction Diagram

```text
                User
                  │
                  ▼
            Horizon (Dashboard)
                  │
                  ▼
         Keystone (Authentication)
                  │
                  ▼
          Nova (Create VM)
                  │
      ┌───────────┼───────────┐
      ▼           ▼           ▼
 Placement     Glance      Neutron
(Resource)     (Image)    (Network)
      │                       │
      └───────────┬───────────┘
                  ▼
           Compute Node
                  │
                  ▼
        Cinder (Optional Disk)
```

---

# Easy Analogy (Building an Apartment)

| Service       | Analogy                                      |
| ------------- | -------------------------------------------- |
| **Keystone**  | Security guard (checks identity)             |
| **Nova**      | Building manager (creates apartments/VMs)    |
| **Placement** | Room availability board (tracks free space)  |
| **Glance**    | Blueprint library (stores VM templates)      |
| **Neutron**   | Roads and utilities (networking)             |
| **Cinder**    | Extra storage room (block storage)           |
| **Swift**     | Warehouse (stores files and backups)         |
| **Horizon**   | Reception desk/control panel (web interface) |

---

# Quick Revision

| Service       | Remember It As                                        |
| ------------- | ----------------------------------------------------- |
| **Keystone**  | Identity and authentication                           |
| **Nova**      | Creates and manages VMs                               |
| **Neutron**   | Networking and IP management                          |
| **Glance**    | Stores VM images                                      |
| **Cinder**    | Block storage (virtual hard disks)                    |
| **Swift**     | Object storage (files, images, backups)               |
| **Horizon**   | Web dashboard                                         |
| **Placement** | Tracks CPU, RAM, and disk resources for VM scheduling |

## Memory Trick

**"KNNGCSHP"** (pronounced *"King's Chip"*):

* **K** – **Keystone** → Who can log in?
* **N** – **Nova** → Create and manage VMs.
* **N** – **Neutron** → Connect VMs with networking.
* **G** – **Glance** → VM image library.
* **C** – **Cinder** → Persistent virtual disks.
* **S** – **Swift** → Object/file storage.
* **H** – **Horizon** → Web dashboard.
* **P** – **Placement** → Finds resources for new VMs.
