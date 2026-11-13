# Kubernetes security: RBAC, Secrets and containers

!!! abstract "Where it's used in this project"
    The API's ServiceAccount and Role ([`gitops/api/base/rbac.yaml`](https://github.com/foyez/dbaas-platform/blob/main/gitops/api/base/rbac.yaml)), the operator's generated RBAC, the distroless API image ([`platform/Containerfile`](https://github.com/foyez/dbaas-platform/blob/main/platform/Containerfile)) and how secrets reach the cluster.

## Key ideas

### ServiceAccounts and RBAC

Pods call the Kubernetes API as a **ServiceAccount**. What it may do is granted by RBAC:

| Resource | Scope |
|---|---|
| `Role` + `RoleBinding` | Permissions inside one namespace |
| `ClusterRole` + `ClusterRoleBinding` | Permissions across the cluster, or for cluster-scoped objects |

A rule names API groups, resources and verbs. Grant the least that works, and check with `kubectl auth can-i list postgresqls --as=system:serviceaccount:dbaas-api:dbaas-api -n database-system`.

The same idea exists outside Kubernetes: a STACKIT service account has no access until it's given a role, such as DNS Admin for cert-manager's DNS-01 challenges.

### Secrets

A Kubernetes Secret is only base64-encoded, not encrypted, and must never be committed to Git. Options for GitOps:

| Approach | How |
|---|---|
| Create outside Git | `kubectl create secret` during bootstrap; simple, but manual |
| **External Secrets Operator** | A `SecretStore` points at a vault (STACKIT Secrets Manager); an `ExternalSecret` says which keys to sync into a Kubernetes Secret |
| Sealed Secrets / SOPS | Commit an encrypted manifest that only the cluster can decrypt |

### Hardened containers

- **Distroless images** contain the app and its runtime libraries, nothing else: no shell or package manager, so less to exploit and fewer CVEs.
- **Run as non-root** (`distroless/static:nonroot`, UID 65532).
- **Security context**: `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true`, drop all Linux capabilities.

## How this project uses it

- **The API's Role lives where the data is.** The API runs in `dbaas-api`, but its Role and RoleBinding are in `database-system`, because that's where the `PostgreSQL` resources and their Secrets are. It can manage `postgresqls` and only `get` Secrets.
- **The operator's RBAC is generated** from `+kubebuilder:rbac` markers, except for the port-allocator ConfigMap and the ingress-nginx objects, which are hand-written to avoid a kustomize namespace collision.
- **The API image is distroless and non-root**, with a read-only root filesystem.
- **Secrets are created outside Git.** External Secrets Operator is installed, but the `SecretStore` and `ExternalSecret` resources aren't in the repository yet.

## What I learned building it

- **A `403 Forbidden` on a custom resource is almost always RBAC placement**: a Role in the wrong namespace (the caller's instead of the resource's), a missing RoleBinding, or a wrong API group or plural resource name.
- **Debugging distroless pods needs a helper pod.** There's no shell to `exec` into, so run a throwaway pod (`kubectl run curl-test --rm -it --image=curlimages/curl -- sh`) in the same namespace to test connectivity.
- **Cloud service accounts start with no permissions.** cert-manager's DNS-01 challenge failed until the STACKIT service account got the DNS Admin role, in the same project as the DNS zone.

## Review questions

??? question "The API runs in namespace `dbaas-api` and manages resources in `database-system`. Where do its Role and RoleBinding go?"
    In `database-system`, where the resources are. The RoleBinding's subject is the ServiceAccount from `dbaas-api`. A Role only grants access within its own namespace.

??? question "Is a Kubernetes Secret encrypted?"
    Not by default. The values are base64-encoded, and etcd stores them in plain form unless encryption at rest is configured. Access control comes from RBAC.

??? question "How does External Secrets Operator keep secrets out of Git?"
    Git holds only an `ExternalSecret` that references a key in an external vault. The operator fetches the value and writes it into a Kubernetes Secret inside the cluster.

??? question "Why use a distroless, non-root image?"
    Fewer packages means fewer vulnerabilities and nothing for an attacker to use (no shell, no package manager). Running as non-root limits what a compromised process can do.
