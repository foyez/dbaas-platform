# Validation of OpenStack Installation

After DevStack finishes installing, the next step is to verify that the OpenStack cloud is working correctly.

A successful installation should be checked from two different layers:

1. **OpenStack Service Layer**

   * Verify that OpenStack services are running.
   * Check that APIs, agents, and components are healthy.

2. **Operating System Layer**

   * Verify that the underlying Linux services are running.
   * Inspect logs to identify errors or service failures.

---

# 1. Verify Horizon Dashboard (Web Interface)

Horizon is the graphical interface for managing OpenStack.

Check that the dashboard is reachable:

```bash
curl -I http://localhost/dashboard
```

Expected result:

```text
HTTP/1.1 200 OK
```

or:

```text
HTTP/1.1 302 Found
```

A redirect (`302`) is normal because Horizon redirects users to the login page.

Open the dashboard:

```
http://<your-server-ip>/dashboard
```

Login credentials:

```
Username:
admin

Password:
The value configured in ADMIN_PASSWORD inside local.conf
```

Example:

```ini
ADMIN_PASSWORD=secret
```

Login:

```
Username: admin
Password: secret
```

---

# 2. Run DevStack Test Suite

DevStack includes tests to verify that the deployment is functioning.

Run:

```bash
./run_tests.sh
```

The test suite checks:

* OpenStack API availability
* Service communication
* Authentication
* Networking
* Compute functionality

If the installation is healthy, tests should complete successfully.

---

# 3. Authenticate with OpenStack CLI

Before checking OpenStack services, load administrator credentials.

Run:

```bash
source /opt/stack/devstack/openrc admin admin
```

This loads:

* Keystone authentication URL
* Admin username
* Admin project
* Service credentials

Verify authentication:

```bash
openstack token issue
```

Expected output:

```
+------------+--------------------------------------+
| Field      | Value                                |
+------------+--------------------------------------+
| id         | gAAAAAB...                           |
| expires    | 2026-...                             |
+------------+--------------------------------------+
```

A token confirms Keystone is working.

---

# 4. Check OpenStack Service Health

## 4.1 Check Nova Compute Services

Nova manages virtual machines.

Run:

```bash
openstack compute service list
```

Example output:

```
+----------------+----------+-------+---------+
| Binary         | Host     | State | Status  |
+----------------+----------+-------+---------+
| nova-scheduler | devstack | up    | enabled |
| nova-conductor | devstack | up    | enabled |
| nova-compute   | devstack | up    | enabled |
+----------------+----------+-------+---------+
```

Check:

| Field  | Expected |
| ------ | -------- |
| Status | enabled  |
| State  | up       |

Meaning:

* The service is allowed to receive requests.
* The service is running correctly.

---

## 4.2 Check Neutron Networking Services

Neutron provides virtual networking.

For OVN deployments:

```bash
openstack network agent list
```

You should see agents related to:

* OVN Controller
* Metadata Agent

Expected:

```
Alive: :-)
State: UP
```

Example:

```
+----------------------+-------+-------+
| Agent Type           | Alive | State |
+----------------------+-------+-------+
| OVN Controller       | :-)   | UP    |
| Metadata agent       | :-)   | UP    |
+----------------------+-------+-------+
```

---

### Traditional Neutron Agent Check

If using the older Open vSwitch agent architecture:

```bash
openstack network agent list
```

Expected agents:

```
Open vSwitch agent
DHCP agent
L3 agent
Metadata agent
```

Health:

```
Alive: :-)
State: UP
```

---

# 4.3 Check Cinder Block Storage

Cinder manages persistent volumes.

Run:

```bash
openstack volume service list
```

Example:

```
+------------------+----------+---------+
| Binary           | Host     | State   |
+------------------+----------+---------+
| cinder-scheduler | devstack | up      |
| cinder-volume    | devstack | up      |
+------------------+----------+---------+
```

Check:

| Field  | Expected |
| ------ | -------- |
| Status | enabled  |
| State  | up       |

---

# 5. Check System-Level Services and Logs

OpenStack services run on top of Linux.

If OpenStack commands fail, check the operating system layer.

DevStack services are managed using `systemd`.

---

# 5.1 View All DevStack Service Logs

To monitor every OpenStack service:

```bash
sudo journalctl -f -u devstack@*
```

Explanation:

```
journalctl
```

Reads Linux system logs.

```
-f
```

Follows logs live.

```
-u devstack@*
```

Shows all DevStack services.

Example output:

```
nova-compute started successfully
neutron service connected
cinder-volume ready
```

This is the best command when troubleshooting.

---

# 5.2 Nova Compute Logs

Nova Compute creates virtual machines.

Service:

```
devstack@n-cpu
```

Check logs:

```bash
sudo journalctl -u devstack@n-cpu -f
```

Look for:

Good:

```
Compute service started
```

Problems:

```
Unable to connect to libvirt
Failed to spawn instance
```

---

# 5.3 Neutron Logs

## OVN Based Deployment

Modern OpenStack uses:

```
ovn-controller
ovn-northd
q-ovn-metadata-agent
```

Check:

```bash
sudo journalctl -u devstack@q-ovn-metadata-agent -f
```

For OVN controller:

```bash
sudo journalctl -u devstack@ovn-controller -f
```

---

## Traditional Neutron Agent Deployment

If using:

* Open vSwitch Agent
* L3 Agent

Check:

### Layer 2 Agent

```bash
sudo journalctl -u devstack@q-agt -f
```

Responsible for:

* VM network connections
* Virtual switching
* VLAN/VXLAN configuration

### Layer 3 Agent

```bash
sudo journalctl -u devstack@q-l3 -f
```

Responsible for:

* Routing
* NAT
* Floating IPs

---

# 5.4 Cinder Storage Logs

Cinder creates and manages volumes.

Service:

```
devstack@c-vol
```

Check:

```bash
sudo journalctl -u devstack@c-vol -f
```

Look for:

Good:

```
Volume service started
```

Problems:

```
Unable to create volume
Backend unavailable
```

---

# 5.5 Keystone Authentication Logs

Keystone handles authentication.

DevStack normally runs Keystone through Apache.

Check Apache logs:

```bash
sudo tail -f /var/log/apache2/error.log
```

Look for:

Good:

```
Keystone application started
```

Problems:

```
Authentication failure
Database connection error
```

---

# 6. Complete Health Check Flow

A recommended validation order:

```
1. Check Dashboard
        |
        v
curl -I http://localhost/dashboard


2. Authenticate CLI
        |
        v
source openrc


3. Verify Keystone
        |
        v
openstack token issue


4. Verify Nova
        |
        v
openstack compute service list


5. Verify Neutron
        |
        v
openstack network agent list


6. Verify Cinder
        |
        v
openstack volume service list


7. Check Linux Logs
        |
        v
journalctl -f -u devstack@*
```

---

# What Does a Healthy OpenStack Installation Look Like?

A healthy DevStack deployment should have:

| Component | Expected State                        |
| --------- | ------------------------------------- |
| Horizon   | Dashboard loads                       |
| Keystone  | Token generation works                |
| Nova      | Services enabled and UP               |
| Neutron   | Agents alive and networking available |
| Glance    | Images can be listed/uploaded         |
| Cinder    | Volume services UP                    |
| Placement | Resource inventory available          |
| RabbitMQ  | Message queue running                 |
| MySQL     | Database accessible                   |

When all these checks pass, the OpenStack cloud is ready to create instances, networks, and storage resources.
