# DevStack OpenStack Installation Config Notes

# Overview

This configuration installs a complete OpenStack cloud using **DevStack**.

The goal is to create a private cloud environment capable of:

* Creating virtual machines
* Managing users and authentication
* Uploading VM images
* Creating virtual networks
* Providing virtual routers and IP addresses
* Attaching storage volumes
* Managing everything through a web dashboard

After creating `local.conf`, install OpenStack with:

```bash
./stack.sh
```

---

# Recommended DevStack Configuration (Modern OVN Based)

```ini
[[local|localrc]]

# ============================================================
# Host Configuration
# ============================================================

# IP address of the DevStack machine

HOST_IP=10.0.0.140

SERVICE_HOST=$HOST_IP


# ============================================================
# Authentication and Database Passwords
# ============================================================

ADMIN_PASSWORD=secret

DATABASE_PASSWORD=$ADMIN_PASSWORD

RABBIT_PASSWORD=$ADMIN_PASSWORD

SERVICE_PASSWORD=$ADMIN_PASSWORD


# ============================================================
# Core Infrastructure Services
# ============================================================

# Keystone authentication
# MySQL database
# RabbitMQ message queue

enable_service key mysql rabbit


# ============================================================
# Nova Compute Service
# ============================================================

# API service
# Database conductor
# Scheduler
# Compute service
# Browser console

enable_service n-api
enable_service n-cond
enable_service n-sch
enable_service n-cpu
enable_service n-novnc


# ============================================================
# Glance Image Service
# ============================================================

enable_service g-api


# ============================================================
# Neutron Networking (OVN)
# ============================================================

# Disable old Nova networking

disable_service n-net


# Enable Neutron API

enable_service neutron
enable_service q-svc


# Enable modern OVN networking backend

enable_service ovn-northd
enable_service ovn-controller
enable_service q-ovn-metadata-agent


# ============================================================
# Cinder Block Storage
# ============================================================

enable_service c-api
enable_service c-sch
enable_service c-vol


# ============================================================
# Placement Resource Tracking
# ============================================================

enable_service placement-api


# ============================================================
# Horizon Dashboard
# ============================================================

enable_service horizon


# ============================================================
# Neutron OVN Configuration
# ============================================================

Q_AGENT=openvswitch

Q_ML2_PLUGIN_MECHANISM_DRIVERS=ovn

Q_ML2_TENANT_NETWORK_TYPE=vxlan


# Disable old Neutron agents
# OVN replaces these functions

disable_service q-agt
disable_service q-dhcp
disable_service q-l3


# ============================================================
# System Settings
# ============================================================

USE_SYSTEMD=True

LOGFILE=$DEST/logs/stack.sh.log

VERBOSE=True
```

---

# OpenStack Architecture Overview

OpenStack is not one application.

It is a collection of independent services working together.

The main services are:

| Service   | Purpose                                |
| --------- | -------------------------------------- |
| Keystone  | Authentication and identity management |
| Nova      | Virtual machine lifecycle management   |
| Neutron   | Virtual networking                     |
| Glance    | VM image management                    |
| Cinder    | Persistent block storage               |
| Swift     | Object storage                         |
| Horizon   | Web dashboard                          |
| Placement | Resource tracking                      |

---

# Service Explanation

# Keystone - Identity Service

Keystone manages:

* Users
* Projects
* Roles
* Permissions
* Authentication tokens

Example:

```
User Login

      |
      v

Keystone

      |
      v

Authentication Token

      |
      v

Access OpenStack APIs
```

Every OpenStack service uses Keystone.

---

# Nova - Compute Service

Nova creates and manages virtual machines.

Nova components:

## n-api

Receives VM requests.

Example:

```
Create VM
Delete VM
Restart VM
```

---

## n-sch

Nova Scheduler.

Decides:

"Which compute node should run this VM?"

Example:

```
Compute Node A
CPU available

Compute Node B
CPU full


Scheduler chooses A
```

---

## n-cond

Nova Conductor.

Handles:

* Database operations
* Communication between Nova services

---

## n-cpu

The actual VM creator.

It communicates with:

* KVM/QEMU
* Neutron
* Glance
* Cinder

Flow:

```
Nova Compute

       |

      KVM

       |

 Virtual Machine
```

---

## n-novnc

Provides browser console access.

Example:

```
Horizon

   |

NoVNC

   |

VM Console
```

---

# Glance - Image Service

Glance stores VM templates.

Examples:

```
ubuntu-22.04.qcow2

centos-stream.qcow2

debian.qcow2
```

When creating a VM:

```
Glance Image

      |

      v

Nova

      |

      v

Virtual Machine
```

---

# Neutron - Networking Service

Neutron provides cloud networking.

It manages:

* Virtual networks
* Subnets
* Routers
* Floating IPs
* Security groups
* Network isolation

There are two major Neutron architectures.

---

# Traditional Neutron Architecture

Older deployments use:

```
Neutron Server

      |

      +---- q-agt
      |
      +---- q-dhcp
      |
      +---- q-l3

```

Components:

## q-svc

Neutron API server.

Receives network commands.

---

## q-agt

Layer 2 agent.

Connects VMs to virtual switches.

---

## q-dhcp

Provides VM IP addresses.

---

## q-l3

Provides:

* Routing
* NAT
* Floating IPs

---

# Modern Neutron Architecture - OVN

New OpenStack deployments use:

```
Neutron

   |

   |

  OVN

   |

   |

Open vSwitch

   |

   |

Virtual Machines
```

OVN replaces multiple Neutron agents.

---

# What is OVN?

OVN means:

**Open Virtual Network**

It is the modern networking backend for OpenStack.

It provides:

* Virtual switching
* Routing
* DHCP
* NAT
* Security rules
* Network isolation

---

# OVN Components

## ovn-northd

The controller of OVN.

It converts Neutron requests into OVN network rules.

Example:

User creates:

```
Network A

Subnet:
10.0.0.0/24
```

Flow:

```
Neutron API

      |

      v

ovn-northd

      |

      v

OVN Database Rules
```

---

## ovn-controller

Runs on compute nodes.

Responsible for:

* Programming Open vSwitch
* Connecting VMs
* Applying network rules

---

## q-ovn-metadata-agent

Provides VM metadata.

Used for:

* SSH key injection
* cloud-init
* instance information

---

# Why OVN Does Not Need q-dhcp and q-l3

Traditional Neutron:

```
q-dhcp → DHCP

q-l3 → Router
```

OVN:

```
OVN

Handles:

DHCP
Routing
NAT
Switching
```

Therefore:

```
disable_service q-dhcp
disable_service q-l3
```

is correct.

---

# Cinder - Block Storage

Cinder provides persistent disks.

Without Cinder:

```
VM deleted

Disk deleted
```

With Cinder:

```
VM deleted

Disk remains

Attach to another VM
```

Components:

## c-api

Receives storage requests.

## c-sch

Chooses storage backend.

## c-vol

Creates and manages volumes.

---

# Placement Service

Placement tracks available resources.

Example:

```
Compute Node

CPU: 32 cores
RAM: 128 GB
Disk: 2 TB
```

Nova Scheduler asks:

```
Where can I place this VM?
```

Placement answers:

```
Compute Node 1
has enough resources
```

---

# Horizon Dashboard

Horizon provides the web interface.

Instead of:

```bash
openstack server create
```

You can use:

```
Browser

   |

Horizon

   |

OpenStack APIs
```

You can manage:

* Instances
* Networks
* Images
* Storage
* Users

---

# Swift Object Storage

Swift is different from Cinder.

Cinder:

```
Block Storage

Like a hard disk
```

Swift:

```
Object Storage

Like Amazon S3
```

Used for:

* Backups
* Large files
* Media storage

It is usually not enabled in DevStack because it requires additional storage configuration.

---

# Traditional Neutron vs OVN

| Feature             | Traditional Agents | OVN             |
| ------------------- | ------------------ | --------------- |
| Layer 2 Networking  | q-agt              | OVN Controller  |
| DHCP                | q-dhcp             | OVN             |
| Routing             | q-l3               | OVN             |
| Architecture        | Agent based        | Distributed     |
| Scalability         | Medium             | High            |
| Modern OpenStack    | Less common        | Recommended     |
| Learning simplicity | Easier             | Slightly harder |

---

# Final Recommendation

For a new DevStack installation:

Use:

```
Keystone
Nova
Glance
Neutron + OVN
Cinder
Placement
Horizon
```

with:

```
Open vSwitch
VXLAN
OVN networking
```

This represents the current OpenStack architecture and is closer to what you would see in modern production clouds.
