# Operator

The operator owns the `PostgreSQL` custom resource (`database.example.com/v1`). It's built with [Kubebuilder](https://book.kubebuilder.io/) and hands the hard database work (replication, failover, secrets) to [CloudNativePG](https://cloudnative-pg.io/). The operator itself only adds the product layer on top: a small spec, a normalised status and external connectivity.

## The resource

```yaml
apiVersion: database.example.com/v1
kind: PostgreSQL
metadata:
  name: orders-db
  namespace: database-system
spec:
  version: "16"      # PostgreSQL major version
  instances: 1       # replicas, default 1
  storage: 2Gi
  database: orders
  username: orders-admin
status:
  phase: Ready       # Pending | Creating | Ready | Failed | Deleting
  readyInstances: 1
  externalPort: 30101
  conditions: [...]
```

Five spec fields are the whole contract the API and UI depend on. [Design decision 2](../design-decisions/own-crd-over-cnpg.md) explains why the platform wraps CNPG in its own CRD instead of exposing CNPG directly.

## Reconcile loop

Each reconcile runs the same steps, and every step is safe to repeat:

1. **Deletion.** If the resource is being deleted, release its external port and remove its ingress-nginx entries, then drop the finalizer.
2. **Finalizer.** Otherwise make sure the `tcp-expose` finalizer is set, so that cleanup always gets a chance to run.
3. **CNPG cluster.** Create or update the CloudNativePG `Cluster` that matches the spec.
4. **External exposure.** Allocate a port from the shared pool (30100–30999) and wire it into ingress-nginx's `tcp-services` ConfigMap and Service.
5. **Status.** Copy readiness from the CNPG cluster into `phase`, `readyInstances` and the conditions.

## External connectivity

CNPG clusters are only reachable inside the cluster (`ClusterIP`). To give each instance a real connection string without a LoadBalancer per tenant, the operator:

- picks the lowest free port from a ConfigMap-backed allocator, using optimistic locking so concurrent reconciles can't hand out the same port
- adds `port → database-system/<instance>-rw:5432` to ingress-nginx's `tcp-services` ConfigMap
- opens the same port on the ingress-nginx Service

Exposure is best effort. If ingress-nginx isn't installed, the CR still becomes Ready with only internal connectivity. [Design decision 3](../design-decisions/shared-loadbalancer.md) covers why ports are used instead of hostnames.

!!! note "PgBouncer"
    The operator originally also managed a PgBouncer pooler directly, to practise working with Deployments, Services and PVCs. That code is still in `pgbouncer.go` but disabled. CNPG's own pooler is the better fit if pooling is needed.

## Develop and test

```sh
cd operator
make test            # unit + envtest
make manifests       # regenerate CRDs and RBAC after changing types
```

In the local platform the operator is built and deployed by `make dev`. After changing its code, re-run `make dev` to rebuild and roll it out.
