# Install K8s on Two VMs (Control Plane + Worker)

## 0. Node plan

| Hostname | Role | Private IP |
|---|---|---|
| `control-plane` | control-plane | `192.168.10.10` |
| `worker-01` | worker | `192.168.10.11` |

Run **Part A** on both VMs. Run **Part B** only on `control-plane`. Run **Part C** only on `worker-01`.

---

## Security group / firewall reminder (do this before Part A)

Whether you provisioned these VMs via Terraform or by hand, confirm these are open **between the two private IPs** (not just from your laptop):

| Port | Protocol | Purpose |
|---|---|---|
| 22 | TCP | SSH |
| 6443 | TCP | kube-apiserver |
| 10250 | TCP | kubelet |
| 8472 | UDP | Flannel VXLAN |
| 30000-32767 | TCP | NodePort services (needed for the Step 5 test) |

If you skip the 8472/UDP rule, pods can be scheduled and `Ready` but cross-node pod traffic (and therefore CoreDNS) will silently fail with `NXDOMAIN`-style symptoms — this is the exact failure mode diagnosed earlier when Flannel traffic was blocked.

---

## Part A — Run on BOTH nodes

### A1. Set hostname and `/etc/hosts`

On `control-plane`:
```bash
sudo hostnamectl set-hostname control-plane
```
On `worker-01`:
```bash
sudo hostnamectl set-hostname worker-01
```

On **both** nodes, edit `/etc/hosts` to include both machines:
```bash
sudo tee -a /etc/hosts <<EOF
192.168.10.10 control-plane
192.168.10.11 worker-01
EOF
```
Also make sure the `127.0.1.1` line matches this node's own hostname (`control-plane` or `worker-01`, not the generic `ubuntu-vm` from the original doc).

### A2. Check DNS

```bash
cat /etc/resolv.conf
ping -c 3 archive.ubuntu.com
```
If this fails and you're on a managed OpenStack network with `dns_nameservers` set on the subnet, don't hand-edit `/etc/resolv.conf` — it'll be pushed via DHCP already, and a manual edit will just get reverted or fight `systemd-resolved`. Only fall back to manual edits if you're not using DHCP-provided DNS at all:
```bash
sudo systemctl restart systemd-resolved
resolvectl status
```

### A3. Disable swap (new step — required)

```bash
sudo swapoff -a
sudo sed -i '/ swap / s/^/#/' /etc/fstab

# Confirm
free -h   # "Swap" row should show 0
```

### A4. Install containerd

```bash
sudo apt-get update && sudo apt-get upgrade -y
sudo apt-get install -y containerd

sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
grep SystemdCgroup /etc/containerd/config.toml   # should print: SystemdCgroup = true

sudo systemctl restart containerd
sudo systemctl enable containerd
sudo ctr version
```

### A5. Kernel modules and sysctl

```bash
sudo modprobe overlay
sudo modprobe br_netfilter

sudo tee /etc/modules-load.d/kubernetes.conf <<EOF
overlay
br_netfilter
EOF

sudo tee /etc/sysctl.d/kubernetes.conf <<EOF
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

sudo sysctl --system
sysctl net.ipv4.ip_forward   # expect: net.ipv4.ip_forward = 1
```

### A6. Install kubeadm, kubelet, kubectl

```bash
sudo apt-get install -y apt-transport-https ca-certificates curl gpg

curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.35/deb/Release.key \
  | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

sudo tee /etc/apt/sources.list.d/kubernetes.list > /dev/null <<EOF
deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.35/deb/ /
EOF

sudo apt-get update

sudo apt-get install -y \
  kubelet=1.35.0-1.1 \
  kubeadm=1.35.0-1.1 \
  kubectl=1.35.0-1.1

sudo apt-mark hold kubelet kubeadm kubectl

kubeadm version
kubelet --version
kubectl version --client
```

**A1-A6 are now done on both nodes.** Everything below is role-specific.

---

## Part B — Control plane only (`control-plane`)

### B1. Initialize the control plane

```bash
sudo kubeadm init \
  --pod-network-cidr=10.244.0.0/16 \
  --apiserver-advertise-address=192.168.10.10 \
  --apiserver-bind-port=6443 \
  --kubernetes-version=1.35.0
```

Expect near the end:
```
[init] Using Kubernetes version: v1.35.0
...
Your Kubernetes control-plane has initialized successfully!

Then you can join any number of worker nodes by running the following on each:

kubeadm join 192.168.10.10:6443 --token <token> \
    --discovery-token-ca-cert-hash sha256:<hash>
```
**Copy that `kubeadm join ...` line — you'll need it in Part C.** If you lose it, regenerate anytime with:
```bash
kubeadm token create --print-join-command
```

### B2. Set up kubectl access

```bash
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

kubectl cluster-info
kubectl get nodes          # will show NotReady until Flannel is installed
```

### B3. Install Flannel (pinned version)

```bash
kubectl apply -f https://github.com/flannel-io/flannel/releases/download/v0.26.4/kube-flannel.yml

kubectl get pods -n kube-flannel -w   # wait for Running, ~30s
kubectl get nodes                      # control-plane should now show Ready
```

---

## Part C — Worker only (`worker-01`)

### C1. Join the cluster

Run the exact command copied from B1 (or regenerated via `kubeadm token create --print-join-command` on the control plane):

```bash
sudo kubeadm join 192.168.10.10:6443 \
  --token <token> \
  --discovery-token-ca-cert-hash sha256:<hash>
```

Expect:
```
This node has joined the cluster:
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.
```

---

## Part D — Validate the two-node cluster

Run these from `control-plane`:

```bash
# Both nodes should now be listed
kubectl get nodes -o wide
# NAME             STATUS   ROLES           VERSION
# control-plane    Ready    control-plane   v1.35.0
# worker-01        Ready    <none>          v1.35.0

kubectl cluster-info
kubectl get pods -n kube-system -o wide
kubectl get pods -n kube-flannel -o wide

# Cross-node pod networking test
kubectl run test-a --image=busybox --restart=Never -- sleep 3600
kubectl run test-b --image=busybox --restart=Never -- sleep 3600
kubectl wait --for=condition=Ready pod/test-a pod/test-b --timeout=60s
kubectl exec test-a -- ping -c 3 $(kubectl get pod test-b -o jsonpath='{.status.podIP}')
kubectl delete pod test-a test-b

# DNS test
kubectl run dns-test --image=busybox --restart=Never -- nslookup kubernetes.default
kubectl logs dns-test
kubectl delete pod dns-test
```

### End-to-end workload test

```bash
kubectl create deployment nginx-test --image=nginx --replicas=2
kubectl get pods -o wide -w   # confirm one pod lands on each node

kubectl expose deployment nginx-test --port=80 --type=NodePort
NODE_PORT=$(kubectl get svc nginx-test -o jsonpath='{.spec.ports[0].nodePort}')

# From your local machine, via the control-plane's floating IP
FLOATING_IP=$(cd /path/to/terraform && terraform output -raw control_plane_floating_ip)
curl http://$FLOATING_IP:$NODE_PORT
# Should return the nginx welcome page, whichever node actually served it

kubectl delete deployment nginx-test
kubectl delete svc nginx-test
```

If the NodePort curl fails but pods show `Running`, it's almost always the `30000-32767` security-group rule from the top of this doc not being open.

---

## Copy kubeconfig to your local machine

```bash
# Run on your LOCAL machine
FLOATING_IP=$(cd terraform && terraform output -raw control_plane_floating_ip)

scp -i ~/.ssh/id_ed25519 ubuntu@$FLOATING_IP:~/.kube/config ~/.kube/config-openstack

sed -i "s|https://192\.168\.10\.10:6443|https://$FLOATING_IP:6443|g" ~/.kube/config-openstack

export KUBECONFIG=~/.kube/config-openstack
kubectl get nodes
```

`--insecure-skip-tls-verify` is fine for a quick local check, but avoid making it a habit — it silently accepts any cert, which defeats the point of TLS on the API server.

## Adding a third node later

Repeat Part A on the new VM (with its own hostname/IP added to `/etc/hosts` on **every** existing node too), then run Part C on it using a fresh join command from `kubeadm token create --print-join-command`. No changes needed on the control plane or other workers.