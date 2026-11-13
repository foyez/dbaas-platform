# Install K8s in VM

## Issues needed to be fixed in VM

<details>
<summary><strong>View Contents</strong></summary>

### 1. Fix the hostname resolution warning

Check your hostname:

```bash
hostname
```

It is probably:

```
ubuntu-vm
```

Now check `/etc/hosts`:

```bash
cat /etc/hosts
```

You should have an entry like:

```
127.0.0.1 localhost
127.0.1.1 ubuntu-vm
```

If the second line is missing, add it:

```bash
sudo vi /etc/hosts
```

Add:

```
127.0.1.1 ubuntu-vm
```

Save and exit.

### 2. Check DNS configuration

Run:

```bash
cat /etc/resolv.conf
```

You should see nameservers, for example:

```
nameserver 127.0.0.53
```

or:

```
nameserver 8.8.8.8
nameserver 1.1.1.1
```

If it is empty or incorrect, temporarily fix it:

```bash
sudo nano /etc/resolv.conf
```

Add:

```
nameserver 8.8.8.8
nameserver 1.1.1.1
```

Then test:

```bash
ping -c 3 archive.ubuntu.com
```

If it works, DNS is fixed.

### 3. If `/etc/resolv.conf` keeps reverting

Ubuntu 22.04 usually uses `systemd-resolved`. Restart it:

```bash
sudo systemctl restart systemd-resolved
```

Check status:

```bash
systemctl status systemd-resolved
```

Then:

```bash
resolvectl status
```

You should see DNS servers assigned.

---

### 4. Test network vs DNS

First test raw connectivity:

```bash
ping -c 3 8.8.8.8
```

* Works → network is OK, DNS is broken.
* Fails → VM networking issue (NAT/bridge adapter, DHCP, etc.).

Then test DNS:

```bash
ping -c 3 google.com
```

</details>

## Step 1 - Install containerd (Container Runtime)

```sh
# Update packages
sudo apt-get update && sudo apt-get upgrade -y

# Install containerd
sudo apt-get install -y containerd

# Create default config
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml

# Critical: set SystemdCgroup = true
# kubeadm requires the kubelet and container runtime to use the same cgroup driver.
# Ubuntu uses systemd, so both must use SystemdCgroup.
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml

# Verify the change
grep SystemdCgroup /etc/containerd/config.toml
# Should print: SystemdCgroup = true

# Restart and enable containerd
sudo systemctl restart containerd
sudo systemctl enable containerd
sudo systemctl status containerd

# Verify
sudo ctr version
# Should show containerd version
```

## Step 2 – Install kubeadm, kubelet, kubectl

```sh
# Install dependencies
sudo apt-get install -y apt-transport-https ca-certificates curl gpg

# Add Kubernetes apt repository (v1.35)
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.35/deb/Release.key \
  | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

sudo tee /etc/apt/sources.list.d/kubernetes.list > /dev/null <<EOF
deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.35/deb/ /
EOF

sudo apt-get update

# Install exact version — pin it to avoid unintended upgrades
sudo apt-get install -y \
  kubelet=1.35.0-1.1 \
  kubeadm=1.35.0-1.1 \
  kubectl=1.35.0-1.1

# Pin the versions so apt upgrade doesn't break your cluster
sudo apt-mark hold kubelet kubeadm kubectl

# Verify
kubeadm version
kubelet --version
kubectl version --client
```

## Step 3 – Initialize the Control Plane

### Issues needed to be fixed

<details>
<summary><strong>View Contents</strong></summary>

`kubeadm` could fail because Linux IP forwarding can be disabled. Kubernetes requires the node to forward packets between pods, services, and interfaces.

Enable IPv4 forwarding:

### 1. Temporarily enable it (until reboot)

```bash
sudo sysctl -w net.ipv4.ip_forward=1
```

Verify:

```bash
cat /proc/sys/net/ipv4/ip_forward
```

Expected output:

```text
1
```

---

### 2. Make it persistent

Create a Kubernetes sysctl config:

```bash
sudo tee /etc/sysctl.d/kubernetes.conf <<EOF
net.ipv4.ip_forward = 1
EOF
```

Apply it:

```bash
sudo sysctl --system
```

Verify again:

```bash
sysctl net.ipv4.ip_forward
```

Expected:

```text
net.ipv4.ip_forward = 1
```

---

### 3. Also enable required Kubernetes networking settings

Load required kernel modules:

```bash
sudo modprobe overlay
sudo modprobe br_netfilter
```

Make them load after reboot:

```bash
sudo tee /etc/modules-load.d/kubernetes.conf <<EOF
overlay
br_netfilter
EOF
```

Set bridge networking parameters:

```bash
sudo tee /etc/sysctl.d/kubernetes-network.conf <<EOF
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
```

Apply:

```bash
sudo sysctl --system
```

---

### 4. Restart container runtime

For containerd:

```bash
sudo systemctl restart containerd
```

Check:

```bash
systemctl status containerd
```

---

### 5. Run kubeadm again

```bash
sudo kubeadm init \
  --pod-network-cidr=10.244.0.0/16 \
  --apiserver-advertise-address=$PRIVATE_IP \
  --kubernetes-version=1.35.0
```

The preflight error should be gone. If the next failure mentions **containerd cgroup driver**, **swap**, or **CRI**, those are common follow-up fixes for a fresh Kubernetes node.

</details>

```sh
# Get the private IP of this VM (will be used as the advertise address)
PRIVATE_IP=$(hostname -I | awk '{print $1}')
echo "Private IP: $PRIVATE_IP"

# Initialize the control plane
# --pod-network-cidr must match the CNI plugin (Flannel uses 10.244.0.0/16)
# --apiserver-advertise-address: the IP the API server binds to and advertises
sudo kubeadm init \
  --pod-network-cidr=10.244.0.0/16 \
  --apiserver-advertise-address=$PRIVATE_IP \
  --kubernetes-version=1.35.0

# This will take 2-5 minutes

# IMPORTANT: Save the full output — it contains the join command for workers
# It looks like:
# kubeadm join 10.0.10.x:6443 --token <token> --discovery-token-ca-cert-hash sha256:<hash>
```

**Expected output:**

```txt
[init] Using Kubernetes version: v1.31.x
[preflight] Running pre-flight checks
[kubelet-start] Writing kubelet environment file
[certs] Generating certificates
[kubeconfig] Writing "admin.conf" file
[control-plane] Using manifest folder "/etc/kubernetes/manifests"
[etcd] Creating static Pod manifest for local etcd
[wait-control-plane] Waiting for the kubelet to boot up the control plane
[apiclient] All control plane components are healthy
[upload-config] Storing the configuration used in ConfigMap
[mark-control-plane] Marking the node as control-plane
[bootstrap-token] Configuring bootstrap tokens
...

Your Kubernetes control-plane has initialized successfully!

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each:

kubeadm join 10.20.0.100:6443 --token <token> \
    --discovery-token-ca-cert-hash sha256:<hash>
```

After `kubeadm init` succeeds, set up `kubectl` access:

```sh
# Set up kubeconfig for the ubuntu user
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# Verify cluster
kubectl cluster-info

# Kubernetes control plane is running at https://10.0.10.28:6443
# CoreDNS is running at https://10.0.10.28:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

# Verify — control plane node will show NotReady until CNI is installed
kubectl get nodes
kubectl get pods -n kube-system
```

## Step 4 – Install a CNI Plugin (Flannel)

**CNI Plugins:**
- **Calico:** https://docs.tigera.io/calico/latest/getting-started/kubernetes/quickstart
- **Flannel:** https://github.com/flannel-io/flannel#deploying-flannel-manually

Kubernetes needs a **Container Network Interface (CNI)** plugin for pod networking. Without a CNI plugin, pods cannot communicate across nodes and all pods remain in `Pending`. Flannel is simple and well-suited for learning.

```sh
# Install Flannel — the CIDR must match what you passed to kubeadm init
kubectl apply -f https://github.com/flannel-io/flannel/releases/download/v0.26.4/kube-flannel.yml

# Watch Flannel pods come up (~30 seconds)
kubectl get pods -n kube-flannel -w

# After Flannel is running, the node should become Ready
kubectl get nodes
# NAME                 STATUS   ROLES           AGE   VERSION
# k8s-control-plane    Ready    control-plane   3m    v1.35.0
```

> **What Flannel does:** It creates a virtual network overlay — each node gets a subnet (e.g., `10.244.0.0/24`) and Flannel routes pod traffic between nodes using VXLAN encapsulation. The CNI plugin is what makes `10.244.x.x` pod IPs routable across nodes.

**What Flannel does:**

* Assigns IP addresses to pods (via overlay network)
* Routes traffic between pods (using VXLAN or host-gw)
* Provides pod-to-pod communication across nodes
* Simplifies networking setup without advanced policy enforcement

## Step 5 – Validate the Cluster

```bash
# Node status
kubectl get nodes -o wide

# Confirms the API server and DNS are responding
kubectl cluster-info

# All system pods should be Running
kubectl get pods -n kube-system -o wide
kubectl get pods -n kube-flannel -o wide

# Check component health
kubectl get componentstatuses   # deprecated but still useful
kubectl cluster-info

# Deploy a test workload to verify the cluster works end-to-end
kubectl create deployment nginx-test --image=nginx --replicas=1
kubectl get pods -w   # wait for Running

# Expose it as a NodePort to test connectivity
kubectl expose deployment nginx-test --port=80 --type=NodePort
kubectl get svc nginx-test
# Note the NodePort (e.g., 32xxx)

# Test from your local machine (using floating IP):
FLOATING_IP=$(cd /path/to/terraform && terraform output -raw control_plane_floating_ip)
NODE_PORT=$(kubectl get svc nginx-test -o jsonpath='{.spec.ports[0].nodePort}')
curl http://$FLOATING_IP:$NODE_PORT
# Should return the nginx welcome page HTML

# Clean up test resources
kubectl delete deployment nginx-test
kubectl delete svc nginx-test
```

### Copy kubeconfig to Your Local Machine

```sh
# Run this on your LOCAL machine
FLOATING_IP=$(cd terraform && terraform output -raw control_plane_floating_ip)

# Copy the kubeconfig
scp -i ~/.ssh/id_rsa ubuntu@$FLOATING_IP:~/.kube/config ~/.kube/config-openstack

# Replace the internal IP with the floating IP so kubectl works from outside
sed -i "s|https://10\.0\.10\.[0-9]*:6443|https://$FLOATING_IP:6443|g" ~/.kube/config-openstack

# Use it
export KUBECONFIG=~/.kube/config-openstack
kubectl get nodes
kubectl get nodes --insecure-skip-tls-verify # Bypass the Cert
```