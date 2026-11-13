# Kubernetes the (Very) Hard Way

Before automating cluster installs with `kubeadm` and Ansible, I worked through [Kubernetes the (Very) Hard Way](https://labs.iximiuz.com/courses/kubernetes-the-very-hard-way-0cbfd997) by Márk Sági-Kazár on iximiuz Labs. It builds a cluster by hand, one component at a time, and tests each one before moving on. It's based on Kelsey Hightower's [Kubernetes the Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way).

This page is a summary of what the course covers and how it fed into the project. The lessons themselves are on iximiuz Labs.

## What the course builds

| Module | Components | What it shows |
|---|---|---|
| Worker node | containerd, kubelet | A node is a container runtime plus an agent. The kubelet can run static pods from a directory with no control plane at all. |
| Control plane | etcd, kube-apiserver, kube-scheduler, kube-controller-manager | All cluster state lives in etcd and the API server is its only client. The scheduler and the controllers are just API clients that watch and write objects. |
| Cluster | joining nodes, pod networking, kube-proxy, CoreDNS | Nodes register through the API server. Pod-to-pod traffic and Service virtual IPs are separate problems, solved by the network plugin and kube-proxy respectively. |

Every component runs as a systemd service with its own certificates and flags, which makes the trust relationships between them visible in a way `kubeadm init` hides.

## Where the same ideas show up in this project

- **Installing Kubernetes on OpenStack.** The [single-node](install-k8s.md) and [two-node](install-two-node.md) installs, and their [Ansible automation](../ansible/automate-provisioning-vm-k8s-installation.md), use `kubeadm` to generate the certificates, static pods and kubeconfigs the course creates by hand.
- **The operator.** The platform's [operator](../../../components/operator.md) follows the same watch-and-reconcile loop as the built-in controllers in kube-controller-manager.
- **Cluster DNS.** The local `make dev` setup routes `*.dbaas.localhost` inside the cluster by adding a rewrite rule to CoreDNS, the component the cluster module installs.
