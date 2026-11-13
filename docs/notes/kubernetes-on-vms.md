# Kubernetes on VMs (Ansible + kubeadm)

!!! abstract "Where it's used in this project"
    Before moving to managed SKE, the project built its own clusters on OpenStack VMs: single-node and two-node, installed by Ansible with `kubeadm`. Code: [`infra/openstack/single-node`](https://github.com/foyez/dbaas-platform/tree/main/infra/openstack/single-node) and [`two-nodes`](https://github.com/foyez/dbaas-platform/tree/main/infra/openstack/two-nodes).

## Key ideas

**Ansible** is agentless configuration management: it connects over SSH and runs modules that describe a desired state, so re-running a playbook is safe. Its building blocks are the **inventory** (which hosts), **playbooks** and **roles** (what to do), **variables** (`group_vars`) and **handlers** (run only when a task changed something, such as restarting containerd).

**kubeadm** bootstraps a cluster: `kubeadm init` creates the control plane (certificates, etcd, API server, scheduler, controller-manager as static pods) and prints a `join` command that workers use to register.

Every node needs, before kubeadm:

1. swap disabled
2. the `overlay` and `br_netfilter` kernel modules
3. sysctls for IP forwarding and bridged traffic through iptables
4. a container runtime (containerd) with the systemd cgroup driver
5. `kubeadm`, `kubelet` and `kubectl` at matching versions

Then a **CNI plugin** (Flannel here) gives pods an IP network across nodes. Until it's running, nodes stay `NotReady`.

The project's `bootstrap.sh` chains it together: Terraform creates the VMs and writes the Ansible inventory → Ansible runs the `common`, `containerd`, `kubernetes`, `control-plane`, `worker` and `flannel` roles → the control plane generates a join command with `kubeadm token create --print-join-command`, and Ansible passes it to the workers through host variables, so nobody copies a token by hand.

## What I learned building it

- **Most `kubeadm init` failures are preflight prerequisites**: IP forwarding off, swap on, missing kernel modules or a cgroup driver mismatch. Automating them in the Ansible roles means a fresh VM works first time.
- **Nodes have to be able to reach each other on specific ports**, not just from your laptop: 6443 (API server), 10250 (kubelet), 8472/UDP (Flannel VXLAN) and 30000–32767 (NodePorts).
- **A blocked VXLAN port fails quietly.** Pods schedule and show `Running`, but cross-node traffic and therefore CoreDNS fail with `NXDOMAIN`-style errors. When DNS breaks only across nodes, check UDP 8472 first.
- **Hostname resolution matters.** Each node's hostname has to be in `/etc/hosts` on every node, and on a DHCP-managed network you fix DNS on the subnet rather than hand-editing `/etc/resolv.conf`, which gets overwritten.
- **Reach private nodes through a bastion.** Only the control plane had a floating IP, so workers were reached with `ssh -J` (ProxyJump) or a `~/.ssh/config` entry.
- **"kubeadm exited 0" isn't validation.** The real checks are: all nodes `Ready`, control-plane pods running, pod-to-pod traffic across nodes, and DNS resolution from inside a pod.
- **Doing it by hand first pays off.** Kubernetes the (Very) Hard Way builds every component manually, which makes it clear what `kubeadm` is generating and why each component exists.

## Review questions

??? question "Why does Kubernetes require IP forwarding and `br_netfilter`?"
    Nodes route packets between pod network interfaces, and Service traffic is handled by iptables or IPVS rules from kube-proxy. Bridged pod traffic has to pass through iptables for those rules to apply, which `br_netfilter` plus the bridge sysctls enable.

??? question "A node stays `NotReady` right after joining. What's the most likely cause?"
    The CNI plugin isn't running on it yet, or can't start. Check the Flannel DaemonSet pod on that node.

??? question "What does Ansible idempotency buy you here?"
    You can re-run the whole playbook after a partial failure, or to add a node, and it only changes what isn't already in the desired state.

??? question "Why use Flannel VXLAN, and what does it need from the network?"
    It wraps pod-to-pod packets in UDP so pods on different nodes can talk over a plain L3 network with no special routing. It needs UDP 8472 open between nodes.
