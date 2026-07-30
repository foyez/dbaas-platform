# PostgreSQL DBASS Platform

## Connectivity Demonstration

The PostgreSQL connectivity component demonstrates how a client workload inside Kubernetes connects to and uses the managed PostgreSQL database provisioned by the PaaS Operator.

The database server runs **PostgreSQL 16**, while the client workload uses the lightweight **`alpine/psql:16.3`** image. The client container provides PostgreSQL command-line utilities such as `psql` and `pg_isready` for connecting to the database and validating its availability.

The connectivity architecture is:

```text
                 Kubernetes

+---------------------------+
| psql-client Deployment    |
| alpine/psql:16.3          |
|                           |
| psql                      |
| pg_isready                |
+-------------+-------------+
              |
              | PostgreSQL protocol
              |
              v

+---------------------------+
| Managed PostgreSQL 16     |
| Created by Operator       |
| customer-db               |
+---------------------------+
```

The `psql-client` Deployment does not run a database server. It acts only as a PostgreSQL client and connects to the managed PostgreSQL instance created and maintained by the Operator.

Database connection details are provided through a Kubernetes Secret containing:

* Database host
* PostgreSQL port
* Username
* Password
* Database name

These values are injected into the client container using PostgreSQL environment variables:

| Variable     | Purpose                    |
| ------------ | -------------------------- |
| `PGHOST`     | PostgreSQL server endpoint |
| `PGPORT`     | PostgreSQL service port    |
| `PGUSER`     | Database username          |
| `PGPASSWORD` | Database password          |
| `PGDATABASE` | Target database name       |

The `pg_isready` utility is used as a readiness probe to verify that the PostgreSQL server is reachable and accepting connections. It checks that the database endpoint is available and that the client can establish a connection.

---

### Connecting to PostgreSQL

Deploy the client workload:

```bash
kubectl apply -f psql-client.yaml
```

Verify that the client pod is running:

```bash
kubectl get pods -n postgres-demo
```

Access the client container:

```bash
kubectl exec -it deployment/psql-client -n postgres-demo -- sh
```

Connect to the managed database:

```bash
psql
```

The `psql` client automatically uses the configured PostgreSQL environment variables to establish the connection.

---

### Verifying the PostgreSQL Server

After connecting with `psql`, the PostgreSQL server information can be verified.

#### 1. Check PostgreSQL version

Run:

```sql
SELECT version();
```

Example output:

```text
PostgreSQL 16.4 (Ubuntu 16.4-1.pgdg22.04+1)
```

or:

```sql
SHOW server_version;
```

Example output:

```text
16.4
```

This confirms that the connected database server is running PostgreSQL 16.

---

#### 2. Check the current database connection

Run:

```sql
SELECT current_database();
```

Example:

```text
 current_database
------------------
 customer-db
```

The `psql` shortcut can also be used:

```text
\conninfo
```

Example:

```text
You are connected to database "customer-db" as user "app-user"
on host "managed-postgres.default.svc.cluster.local" port "5432".
SSL connection (protocol: TLSv1.3)
```

This shows:

* Database name
* Connected user
* Database host
* PostgreSQL port
* SSL connection status

---

#### 3. Identify the exact PostgreSQL instance

When multiple PostgreSQL instances exist, the current server identity can be checked using:

```sql
SELECT
    inet_server_addr(),
    inet_server_port(),
    pg_backend_pid();
```

Example output:

```text
 inet_server_addr | inet_server_port | pg_backend_pid
------------------+------------------+---------------
 10.244.1.25      | 5432             | 12345
```

Meaning:

* `inet_server_addr()` → IP address of the PostgreSQL server receiving the connection
* `inet_server_port()` → PostgreSQL port
* `pg_backend_pid()` → PostgreSQL process handling the current client session

---

### Verifying Kubernetes Service Connectivity

The Kubernetes Service used by the client can be checked with:

```bash
kubectl get svc -n postgres-demo
```

Example:

```text
NAME          TYPE        CLUSTER-IP
customer-db   ClusterIP   10.96.10.20
```

The service endpoints can be checked with:

```bash
kubectl get endpoints -n postgres-demo
```

Example:

```text
NAME          ENDPOINTS
customer-db   10.244.1.25:5432
```

The endpoint IP should match the value returned by:

```sql
SELECT inet_server_addr();
```

This confirms that the `psql-client` workload is communicating with the expected PostgreSQL instance.

---

### Checking Read/Write Instance

If the Operator supports PostgreSQL replication with primary and replica instances, the connected instance role can be checked:

```sql
SELECT pg_is_in_recovery();
```

Results:

```text
false
```

means:

* The client is connected to the primary instance.
* Read and write operations are allowed.

```text
true
```

means:

* The client is connected to a standby/replica instance.
* The instance is normally read-only.

---

### Verifying the Connection Target from Kubernetes

The client Deployment receives the database endpoint from the Kubernetes Secret:

```yaml
PGHOST:
  valueFrom:
    secretKeyRef:
      name: managed-postgres-app
      key: host
```

The configured host can be inspected using:

```bash
kubectl get secret managed-postgres-app \
-n postgres-demo \
-o jsonpath="{.data.host}" | base64 -d
```

Example output:

```text
customer-db-postgres.postgres-demo.svc.cluster.local
```

This hostname represents the PostgreSQL service endpoint used by the `psql-client` Deployment.

---

### Connectivity Verification Summary

For a PaaS demonstration, the following commands provide evidence that the managed PostgreSQL product is working correctly:

```sql
SELECT version();

\conninfo

SELECT pg_is_in_recovery();
```

They verify:

1. The PostgreSQL server version is correct (**PostgreSQL 16**).
2. The client is connected to the expected database service and instance.
3. The connection is established to the correct read/write PostgreSQL role.
