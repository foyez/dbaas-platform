# SSH to a Worker Node via a Jump Host (Bastion)

## Scenario

* **Control Plane (Jump Host)**

  * Public/Floating IP: `172.24.4.228`
  * Private IP: `192.168.10.10`

* **Worker Node**

  * Private IP: `192.168.10.11`

The worker node does **not** have a public IP, so it cannot be accessed directly from my local machine. I must first connect to the control plane (jump host), which has access to the private network.

---

## Option 1: SSH from Your Local Machine Using ProxyJump

If both nodes use the same SSH key, this is the most convenient method because it establishes the connection to the worker in a **single command**.

```bash
ssh -J ubuntu@172.24.4.228 ubuntu@192.168.10.11
```

Replace `ubuntu` with your actual VM username if it's different (for example `ec2-user`, `debian`, `cloud-user`, `rocky`, etc.).

If your key isn't the default:

```bash
ssh -i ~/.ssh/id_rsa \
    -J ubuntu@172.24.4.228 \
    ubuntu@192.168.10.11
```

### How it works

```
My Laptop
    │
    │ SSH
    ▼
Jump Host (172.24.4.228)
    │
    │ SSH
    ▼
Worker (192.168.10.11)
```

### Pros

* Single SSH command
* No need to manually log in to the control plane
* Ideal for automation and scripting

---

## Option 2: SSH to the Control Plane First

This is the traditional and easiest method to understand.

### Step 1

SSH into the control plane.

```bash
ssh ubuntu@172.24.4.228
```

### Step 2

From the control plane, SSH into the worker.

```bash
ssh ubuntu@192.168.10.11
```

### How it works

```
My Laptop
      │
      ▼
Control Plane
      │
      ▼
Worker
```

### Pros

* Simple and easy to troubleshoot
* Useful when you want to perform tasks on the control plane before accessing workers
* Good for learning how the network is laid out

---

## Option 3: Configure `~/.ssh/config`

Create an SSH configuration file if you frequently SSH into these machines or if different SSH keys are used for each host.

Edit:

```bash
nano ~/.ssh/config
```

Add:

```text
Host control-plane
    HostName 172.24.4.228
    User ubuntu
    IdentityFile ~/.ssh/id_rsa

Host worker
    HostName 192.168.10.11
    User ubuntu
    ProxyJump control-plane
    IdentityFile ~/.ssh/id_rsa
```

Save the file and ensure it has the correct permissions:

```bash
chmod 600 ~/.ssh/config

ls -l ~/.ssh/config
# if ~/.ssh/config is owned by another user, e.g., root,
# change the ownership of the config file to your user
sudo chown stack:stack ~/.ssh/config
```

Now you can connect using simple host aliases:

```bash
ssh control-plane
```

or directly to the worker:

```bash
ssh worker
```

SSH automatically connects through the control plane because of the `ProxyJump` configuration.

### Pros

* Simplest commands (`ssh worker`)
* Great for daily use
* Keeps SSH options in one place
* Easy to extend with additional worker nodes

---

## Network Topology

```
                 Internet
                     │
                     │
        172.24.4.228 (Floating IP)
                     │
              Control Plane / Jump Host
             192.168.10.10
                     │
      Private Network (192.168.10.0/24)
                     │
             Worker Node
             192.168.10.11
```

## Why This Works

* The control plane has a public (floating) IP and is reachable from the internet.
* The worker node only has a private IP.
* The control plane acts as a bridge (jump host/bastion) to the private network, allowing access to the worker.

## Summary

| Method                         | Command                                                   | Best For                                             |
| ------------------------------ | --------------------------------------------------------- | ---------------------------------------------------- |
| **ProxyJump**                  | `ssh -J ubuntu@172.24.4.228 ubuntu@192.168.10.11`         | Quick one-off access and automation                  |
| **SSH to Control Plane First** | `ssh ubuntu@172.24.4.228` then `ssh ubuntu@192.168.10.11` | Learning, troubleshooting, and manual administration |
| **SSH Config**                 | `ssh worker`                                              | Frequent day-to-day 