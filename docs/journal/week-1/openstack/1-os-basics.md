# OpenStack Basics

Before learning OpenStack, it's important to understand **cloud computing** and **Infrastructure as a Service (IaaS)**.

---

# What is Cloud Computing?

Cloud computing is the **delivery of computing resources over the internet** instead of owning and managing physical hardware yourself.

These resources include:

* Virtual Machines (VMs)
* Storage
* Networks
* Databases
* Applications

Instead of buying servers, you **rent resources when you need them**.

### Example

Suppose a company needs 20 servers for a new application.

**Traditional approach**

* Buy 20 physical servers.
* Install operating systems.
* Configure networking.
* Maintain hardware.
* Replace failed components.

**Cloud approach**

* Create 20 virtual servers in minutes.
* Pay only for the resources used.
* The cloud provider manages the physical hardware.

---

# What is OpenStack?

**OpenStack** is a **free, open-source cloud platform** used to build and manage **private and public cloud infrastructures**.

Think of OpenStack as software that turns a collection of physical servers into a cloud.

It allows users to:

* Create Virtual Machines (VMs)
* Create networks
* Allocate storage
* Manage users
* Deploy cloud applications
* Scale resources on demand

### Simple Definition

> **OpenStack is software that helps organizations build their own cloud, similar to AWS, Microsoft Azure, or Google Cloud.**

---

# Why Do Companies Use OpenStack?

Organizations use OpenStack because it is:

* Free and open source
* Highly scalable
* Flexible
* Supports multiple hypervisors
* Suitable for private clouds
* Vendor-independent (no vendor lock-in)

Large organizations such as telecom companies, research institutions, universities, and enterprises often use OpenStack for private cloud deployments.

---

# What Problems Does OpenStack Solve?

Without OpenStack:

* Manual server creation
* Manual network configuration
* Manual storage management
* Difficult scaling
* Time-consuming infrastructure management

With OpenStack:

* Self-service resource provisioning
* Centralized management
* Automation
* Better resource utilization
* Faster deployment

---

# What is IaaS (Infrastructure as a Service)?

**Infrastructure as a Service (IaaS)** is a cloud service model where the provider offers **virtualized infrastructure** over the internet.

The provider supplies:

* Virtual Machines
* Storage
* Networking
* CPU
* Memory

The customer manages:

* Operating system
* Applications
* Data
* Middleware

---

# Cloud Service Models

There are three main cloud service models:

| Model                            | What You Manage        | Example                                    |
| -------------------------------- | ---------------------- | ------------------------------------------ |
| **IaaS**                         | OS, applications, data | OpenStack, AWS EC2, Azure Virtual Machines |
| **PaaS (Platform as a Service)** | Applications and data  | Google App Engine, Heroku                  |
| **SaaS (Software as a Service)** | Only use the software  | Gmail, Microsoft 365, Salesforce           |

---

# How OpenStack Fits into IaaS

OpenStack is an **IaaS platform**.

It provides the infrastructure needed to run cloud environments, including:

* Compute (Virtual Machines)
* Networking
* Block storage
* Object storage
* Identity management
* Image management

Users can request these resources through a web dashboard, command-line interface (CLI), or APIs.

---

# Private Cloud vs Public Cloud

### Public Cloud

Infrastructure is owned by a cloud provider and shared among customers.

Examples:

* Amazon Web Services (AWS)
* Microsoft Azure
* Google Cloud Platform (GCP)

### Private Cloud

Infrastructure is dedicated to a single organization.

Advantages:

* More control
* Better security
* Greater customization
* Compliance with organizational policies

OpenStack is commonly used to build **private clouds**.

---

# Key Benefits of OpenStack

* **Open source** – No licensing fees.
* **Scalable** – Add more servers as demand grows.
* **Flexible** – Supports different hardware and hypervisors.
* **Self-service** – Users can create resources on demand.
* **Automated** – APIs and orchestration reduce manual work.
* **High Availability** – Can be designed for minimal downtime.

---

# Real-World Example

Imagine a university has 50 physical servers.

Without OpenStack:

* Administrators manually configure each server.
* Provisioning new servers can take days.

With OpenStack:

* The 50 servers become a private cloud.
* Students and researchers can launch VMs in minutes.
* Storage and networking are managed centrally.
* Resources can be allocated and reclaimed efficiently.

---

# Summary

* **Cloud computing** delivers computing resources over the internet.
* **OpenStack** is an open-source platform for building and managing cloud infrastructure.
* **IaaS** provides virtualized infrastructure such as VMs, storage, and networking.
* OpenStack is primarily an **IaaS platform**, often used to build **private clouds**.
* Its main advantages are scalability, automation, flexibility, and centralized management.

---

## Quick Revision

* **Cloud Computing:** Delivery of computing resources over the internet.
* **OpenStack:** Open-source software for building and managing cloud infrastructure.
* **IaaS:** Provides virtual machines, storage, and networking; users manage the OS and applications.
* **Private Cloud:** Dedicated to one organization.
* **Public Cloud:** Shared infrastructure operated by a cloud provider.
* **Main OpenStack Goal:** Transform physical servers into a cloud that users can manage easily.
