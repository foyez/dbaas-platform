# DevStack Installation Guide (Clean Linux Installation)

> **Recommended OS:** Ubuntu 24.04 LTS (Noble)
> DevStack officially supports the latest two Ubuntu LTS releases, Rocky Linux 9, and openEuler. Ubuntu 24.04 is the most tested and generally provides the smoothest installation experience.

---

# 1. Create the `stack` User (Optional)

> **Note:** DevStack should be run as a **non-root user** with **sudo** privileges.
>
> If you are using a cloud image (such as Ubuntu cloud images), the default user (`ubuntu`, `cloud-user`, etc.) is usually sufficient.

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

---

# 2. Download DevStack

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

---

# 3. Create `local.conf`

Create a file named **`local.conf`** in the root of the `devstack` directory.

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

### Sample Configuration

A sample configuration file is available at:

```
samples/local.conf
```

### Password Requirements

> **Important:** Use **only alphanumeric characters** in passwords.
>
> Some OpenStack services may fail when special characters are used.

---

# 4. Start the Installation

Run the installation script:

```sh
./stack.sh
```

The installation typically takes **15–30 minutes**, depending on:

* Internet speed
* System performance
* Package download time

During installation, DevStack will download and configure numerous packages and Git repositories.

---

# 5. Installation Complete

After the installation finishes successfully, your DevStack environment will include:

* Keystone (Identity)
* Glance (Image Service)
* Nova (Compute)
* Placement
* Cinder (Block Storage)
* Neutron (Networking)
* Horizon (Web Dashboard)

Additional features include:

* Floating IP support
* External network access for virtual machines
* OpenStack CLI tools
* Tempest testing environment

---

# 6. Access Horizon Dashboard

Open the Horizon dashboard in your browser using the URL displayed at the end of the installation.

```text
http://YOUR_VM_IP/dashboard
```

Login:

Username:

```text
admin
```

Password:

```text
secret
```

From the dashboard, you can:

* Manage virtual machines
* Create networks
* Upload images
* Manage storage volumes
* Configure projects and users

---

# 7. Use the OpenStack CLI

Load the OpenStack environment variables:

```sh
source openrc
```

Verify:

```bash
openstack token issue
```

If you receive a token, OpenStack is working.

You can now use the OpenStack CLI:

```sh
openstack server list
```

---

# 8. Create and Access Virtual Machines

After sourcing `openrc`, you can:

* Create virtual machines
* Assign floating IPs
* SSH into instances
* Manage networking and storage

---

# 9. Run Tempest Tests

Navigate to the Tempest directory:

```sh
cd /opt/stack/tempest
```

Run the configured Tempest test suite:

```sh
tempest run
```

---

# 10. Development Workflow

DevStack is intended for OpenStack development.

After installation you can:

* Modify OpenStack source code
* Restart affected services
* Test changes locally
* Validate functionality using Tempest

---

# Step 11: Test the Installation

List services:

```bash
openstack service list
```

List images:

```bash
openstack image list
```

List networks:

```bash
openstack network list
```

List flavors:

```bash
openstack flavor list
```

List compute services:

```bash
openstack compute service list
```

---

# Step 12: Launch Your First VM

From the Horizon dashboard:

1. Go to **Project → Compute → Instances**.
2. Click **Launch Instance**.
3. Enter:

   * Name
   * Image (e.g., CirrOS or Ubuntu)
   * Flavor
   * Network
4. Click **Launch**.

Or use the CLI:

```bash
openstack server create \
  --flavor m1.tiny \
  --image cirros \
  --network private \
  test-vm
```

---

# Useful Commands

List running VMs:

```bash
openstack server list
```

List networks:

```bash
openstack network list
```

List routers:

```bash
openstack router list
```

List volumes:

```bash
openstack volume list
```

Check services:

```bash
systemctl status devstack@*
```

---

# DevStack Directory Structure

```text
/opt/stack
│
├── devstack
├── logs
├── data
└── requirements
```

---

# What Gets Installed?

A standard DevStack installation typically includes:

* **Keystone** – Identity
* **Nova** – Compute
* **Neutron** – Networking
* **Glance** – Image service
* **Cinder** – Block storage
* **Horizon** – Web dashboard
* **Placement** – Resource tracking
* **RabbitMQ** – Message queue
* **MariaDB** – Database

---

# Learning Exercises

Once DevStack is running, try these tasks:

1. Create a new network.
2. Create a subnet.
3. Create a router.
4. Upload an Ubuntu image.
5. Launch a VM.
6. Assign a floating IP.
7. Attach a Cinder volume.
8. Create a snapshot.
9. Delete the VM.
10. Launch another VM from the snapshot.

---

## Official Resources

For the latest installation instructions and requirements, see:

* [DevStack documentation](https://docs.openstack.org/devstack/latest/?utm_source=chatgpt.com)
* [DevStack source repository](https://opendev.org/openstack/devstack?utm_source=chatgpt.com)

These are the authoritative references, as DevStack installation steps and supported operating systems can change between releases.

---

# Summary

The complete installation workflow is:

```sh
sudo useradd -s /bin/bash -d /opt/stack -m stack
sudo chmod +x /opt/stack
echo "stack ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/stack
sudo -u stack -i

git clone https://opendev.org/openstack/devstack
cd devstack

# Create local.conf (see example above)

./stack.sh
```

Once installation completes, you can:

```sh
source openrc
openstack server list
cd /opt/stack/tempest
tempest run
```

You now have a fully functional DevStack environment suitable for learning, development, testing, and contributing to OpenStack.
