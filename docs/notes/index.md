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
| Infrastructure | [OpenStack](openstack.md) | IaaS services and how a VM gets created |
| | [Terraform](terraform.md) | Declarative provisioning, state and its pitfalls |
| | [Kubernetes on VMs](kubernetes-on-vms.md) | Ansible + kubeadm, and what breaks on fresh nodes |
| Kubernetes | [Kubernetes concepts](kubernetes-concepts.md) | Components, resources and control loops |
| | [Custom resources and operators](operators.md) | CRDs, reconcile loops and this project's operator |
| API | [API design in Go](api-design.md) | Clean architecture, HTTP semantics, idempotent create |
| | [Connection pooling](connection-pooling.md) | Pools, PgBouncer and the SQL console's limits |
| | [Health checks and scaling](health-and-scaling.md) | Probes, HPA and load testing with k6 |
| Networking | [TCP, DNS, TLS](networking.md) | How connections, names and encryption work |
| | [Exposing services](exposing-services.md) | Ingress, Gateway API, TCP passthrough, cert-manager |
| Security | [Kubernetes security](kubernetes-security.md) | RBAC, Secrets and hardened containers |
| | [Authentication](authentication.md) | OAuth 2.0, OIDC, PKCE and ZITADEL |
| Frontend | [Vue 3 SPA](frontend.md) | Vue for React developers, polling, serving a SPA |
| Operations | [GitOps and CI/CD](gitops.md) | Argo CD, Kustomize overlays and image tags |
| | [Observability](observability.md) | Metrics, logs, cardinality and SLOs |

## Using these notes for review

Every page ends with **review questions**. Try to answer each one out loud before opening it. If you can't, reread the page's "Key ideas" section, then follow its links into the code.
