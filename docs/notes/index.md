# Learning notes

What I learned building this platform, organised by topic instead of by when I learned it. Each page connects the concepts to the code that uses them, records what went wrong and why, and ends with review questions.

## How the project grew

1. **Run a cloud.** A single-node OpenStack (DevStack) install, then VMs created by hand and with Terraform.
2. **Run Kubernetes on it.** Single-node and two-node clusters on those VMs, installed with Ansible and `kubeadm`.
3. **Move to managed Kubernetes.** A STACKIT SKE cluster from Terraform, with Argo CD for delivery.
4. **Build the product.** A `PostgreSQL` operator on top of CloudNativePG, a Go API, ZITADEL login and a Vue UI.
5. **Run it properly.** TLS, external database access, autoscaling, metrics, logs and an audit trail.

## Topics

| Layer | Page | In one line |
|---|---|---|
| Linux | [Linux fundamentals](linux-fundamentals.md) | Kernel, distributions, the shell and the filesystem tree |
| | [Users and permissions](linux-users-permissions.md) | UIDs, sudo, rwx, special bits, capabilities |
| | [Processes and services](linux-processes-services.md) | Signals, cgroups, systemd, packages, cron |
| | [Networking and hardening](linux-networking-security.md) | ip, ss, netfilter/iptables, ufw, SSH |
| | [Logs and troubleshooting](linux-logs-troubleshooting.md) | journald, grep/awk, dmesg, a method, backups |
| Infrastructure | [OpenStack](openstack.md) | IaaS services and how a VM gets created |
| | [Terraform](terraform.md) | Declarative provisioning, state and its pitfalls |
| | [Kubernetes on VMs](kubernetes-on-vms.md) | Ansible + kubeadm, and what breaks on fresh nodes |
| Kubernetes | [Kubernetes concepts and architecture](kubernetes-concepts.md) | The component map, the API server as hub, control loops |
| | [The control plane](kubernetes-control-plane.md) | etcd, kube-apiserver, scheduler, controller-manager, CCM |
| | [Worker nodes](kubernetes-nodes.md) | kubelet, CRI, containerd, runc, static Pods, TLS bootstrapping |
| | [Cluster networking](kubernetes-cluster-networking.md) | CNI, overlays, Services and kube-proxy, CoreDNS |
| | [Workloads and scheduling](kubernetes-workloads.md) | Pods, controllers, requests, affinity, taints |
| | [Storage](kubernetes-storage.md) | Volumes, PV, PVC, StorageClass, CSI |
| | [Custom resources and operators](operators.md) | CRDs, reconcile loops and this project's operator |
| | [Building an operator with Kubebuilder](kubernetes-kubebuilder.md) | Scaffolding, markers, generated manifests, the manager and envtest |
| | [Running PostgreSQL on Kubernetes](kubernetes-postgres.md) | StatefulSet → operator → CloudNativePG, scaling, failover, backups |
| API | [API design in Go](api-design.md) | REST, methods, status codes, errors, versioning, pagination, layering in brief |
| | [Hardening the API](api-hardening.md) | Idempotency, validation, injection, SQL console limits, CORS, rate limiting, caching |
| | [Connection pooling](connection-pooling.md) | Pools, PgBouncer and the SQL console's limits |
| | [Health checks and scaling](health-and-scaling.md) | Probes, HPA and load testing with k6 |
| | [Building a Go SDK](go-sdk.md) | Client design, typed errors, testable examples, fakes |
| Code design | [Clean code principles](code-design-principles.md) | SOLID, DRY, KISS and YAGNI, with examples from the API and operator |
| | [Clean architecture and project structure](code-design-clean-architecture.md) | Layers, the dependency rule, the composition root and the repo layout |
| Networking | [TCP, DNS, TLS](networking.md) | How connections, names and encryption work |
| | [Exposing services](exposing-services.md) | Ingress, Gateway API, TCP passthrough, cert-manager |
| Security | [Kubernetes security](kubernetes-security.md) | RBAC, Secrets and hardened containers |
| | [Authentication](authentication.md) | OAuth 2.0, OIDC, PKCE and ZITADEL |
| Frontend | [Vue 3 SPA](frontend.md) | Vue for React developers, polling, serving a SPA |
| Operations | [GitOps and CI/CD](gitops.md) | Argo CD, Kustomize overlays and image tags |
| | [Observability](observability.md) | Metrics, logs, cardinality and SLOs |

## Using these notes for review

Every page ends with **review questions**. Try to answer each one out loud before opening it. If you can't, reread the page's "Key ideas" section, then follow its links into the code.
