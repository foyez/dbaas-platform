# Design decisions

Each page records one decision: what was chosen, why, and what it costs. They're written so the trade-offs stay visible when the platform changes.

| # | Decision |
|---|---|
| 1 | [The Kubernetes API Is the Only State Store](kubernetes-as-state-store.md) |
| 2 | [A Narrow CRD of My Own Wraps CloudNativePG](own-crd-over-cnpg.md) |
| 3 | [One Shared LoadBalancer Exposes Databases Through Allocated TCP Ports](shared-loadbalancer.md) |
| 4 | [Creation Is Idempotent Through a Deterministic Resource Name](idempotent-creation.md) |
| 5 | [The SQL Console Goes Through the API](sql-console-through-api.md) |
| 6 | [Authentication Is Delegated to ZITADEL](auth-with-zitadel.md) |
| 7 | [Audit Events Are Structured Logs in Loki](audit-in-loki.md) |
| 8 | [Delivery Is GitOps, and CI Writes Image Tags Back to Git](gitops-delivery.md) |

## Known limitations and next steps

- **Backups:** no scheduled backups or point-in-time recovery yet. CNPG supports both through object storage, and they're the first thing to add.
- **Tenant isolation:** all instances share one namespace, with no per-tenant `ResourceQuota` or `NetworkPolicy`. A namespace per tenant is the planned direction.
- **Abuse limits:** there are no per-user limits on instance count or storage, and storage size isn't validated.
- **Error reporting:** failed reconciles, such as an unsupported major-version change, aren't reported as a `Failed` phase yet.
