# Production Health Endpoints: Best Practices

## 1. Purpose

Health endpoints provide an operational interface for determining whether an application instance is:

* **Alive** — the process is running and able to respond.
* **Ready** — the instance is capable of serving application traffic.
* **Starting** — the application is still initializing.

The key production principle is:

> **Health endpoints should be simple, cheap, reliable, and expose as little information as possible. They should be reachable by the systems that need them, but should not unnecessarily be exposed to the public Internet.**

Health endpoints are not inherently security-sensitive. The main risks come from:

* exposing internal infrastructure information;
* performing expensive dependency checks;
* exposing administrative/debug information;
* allowing the public Internet to probe internal application state.

---

# 2. Health Endpoint Types

A production application can use three types of health checks.

## Liveness

**Question:**

> Is the application process fundamentally alive?

Example:

```text
GET /health/live
```

Response:

```http
HTTP/1.1 200 OK

{"status":"ok"}
```

The liveness check should be extremely cheap.

It should generally **not** check:

* PostgreSQL
* Redis
* Kafka
* external APIs
* DNS
* other applications

A temporary database outage should normally **not cause the application process to restart**.

A useful rule:

> Liveness should answer whether restarting this process might fix the problem.

---

## Readiness

**Question:**

> Can this application instance safely receive traffic?

Example:

```text
GET /health/ready
```

Possible responses:

```http
HTTP/1.1 200 OK

{"status":"ready"}
```

or:

```http
HTTP/1.1 503 Service Unavailable

{"status":"not_ready"}
```

Readiness can perform checks for dependencies that are genuinely required to serve requests.

For example:

```text
Application
   │
   ├── Database ── healthy
   ├── Required cache ── healthy
   └── Required service ── healthy
             │
             ▼
           READY
```

If the application cannot serve requests correctly, readiness should fail.

Importantly:

> **Readiness failure should normally stop traffic from being sent to the instance; it should not restart the process.**

This distinction is especially important with load balancers and Kubernetes.

---

## Startup

**Question:**

> Has the application completed its initialization?

Example:

```text
GET /health/startup
```

Startup checks are most useful with orchestration systems such as Kubernetes when an application has a slow or complex initialization phase. Kubernetes won't run liveness/readiness until startup succeeds.

For example:

```text
Process starts
      │
      ▼
Initialization
      │
      ├── migrations
      ├── configuration loading
      ├── cache initialization
      └── dependency initialization
      │
      ▼
Application ready
```

Without a startup concept, a slow-starting application can potentially be considered unhealthy before it has had enough time to initialize.

For non-Kubernetes deployments, a separate startup endpoint is often unnecessary. The deployment/runtime can usually handle startup and restart behavior itself.

---

# 3. Recommended Endpoint Design

A good general design is:

```text
/health/live
/health/ready
/health/startup
```

However, **do not create all three merely for the sake of having three endpoints**.

A practical recommendation is:

### Kubernetes

```text
/health/live
/health/ready
/health/startup
```

Use all three when the application's lifecycle benefits from the distinction.

### Non-Kubernetes

Start with:

```text
/health/live
/health/ready
```

Add `/health/startup` only if the deployment platform has a meaningful use for it.

For a very simple application, even one health endpoint may be sufficient.

---

# 4. Keep Health Responses Minimal

A health endpoint should not become a diagnostic or information-disclosure endpoint.

Good:

```json
{
  "status": "ok"
}
```

Potentially useful internally:

```json
{
  "status": "ready"
}
```

Avoid returning information such as:

```json
{
  "database": "postgres-prod.internal:5432",
  "redis": "redis-prod.internal",
  "kafka": "kafka-01.internal",
  "hostname": "prod-app-17",
  "version": "1.8.23",
  "environment": "production"
}
```

This information can help an attacker understand the internal architecture.

If detailed diagnostics are required, expose them through a **separate authenticated and authorized administrative/diagnostic interface**.

For example:

```text
Health endpoints
    │
    ├── /health/live
    └── /health/ready

Authenticated diagnostics
    │
    └── /admin/diagnostics
```

Do not turn `/health/ready` into a general diagnostic endpoint.

---

# 5. Health Endpoints Should Be Cheap

Health checks can be called very frequently.

For example, with multiple instances:

```text
Load balancer
     │
     ├── instance 1 → health check
     ├── instance 2 → health check
     ├── instance 3 → health check
     └── instance 4 → health check
```

Therefore health endpoints should:

* execute quickly;
* use minimal CPU;
* avoid expensive database queries;
* avoid large responses;
* avoid external API calls unless absolutely necessary;
* avoid creating unnecessary logs;
* avoid causing side effects.

A particularly important rule:

> **Never make the health endpoint itself capable of causing significant load during an outage.**

---

# 6. Liveness and Dependency Checks

A common mistake is:

```text
GET /health/live
    │
    ├── PostgreSQL
    ├── Redis
    ├── Kafka
    ├── external API
    └── DNS
```

This can create a cascading failure.

For example:

```text
PostgreSQL fails
      │
      ▼
Liveness fails
      │
      ▼
All application instances restart
      │
      ▼
All instances reconnect simultaneously
      │
      ▼
Database is overloaded further
      │
      ▼
More failures
```

Instead:

```text
/health/live
    │
    └── process is functioning → 200


/health/ready
    │
    ├── required DB → healthy
    ├── required dependency → healthy
    └── application state → healthy
```

This allows an application to remain alive while temporarily removing itself from traffic.

---

# 7. Kubernetes Architecture

In Kubernetes, health endpoints integrate naturally with probes.

Typical configuration:

```yaml
containers:
  - name: app
    image: my-app:1.0.0

    livenessProbe:
      httpGet:
        path: /health/live
        port: health

    readinessProbe:
      httpGet:
        path: /health/ready
        port: health

    startupProbe:
      httpGet:
        path: /health/startup
        port: health
```

The semantics are:

```text
startupProbe
      │
      ▼
Application initialized?
      │
      ▼
livenessProbe ──► Should Kubernetes restart it?
      │
      ▼
readinessProbe ─► Should it receive traffic?
```

Kubernetes treats these differently:

| Probe     | Purpose                                 | Failure effect                                     |
| --------- | --------------------------------------- | -------------------------------------------------- |
| Startup   | Application has finished initialization | Prevents liveness/readiness from running too early |
| Liveness  | Process is fundamentally healthy        | Container may be restarted                         |
| Readiness | Instance can serve traffic              | Pod is removed from Service endpoints              |

---

# 8. Kubernetes: Do Not Expose Health Through the Public Ingress

Ideally:

```text
                         Internet
                            │
                            ▼
                     ┌─────────────┐
                     │   Ingress   │
                     └──────┬──────┘
                            │
                            ▼
                     Application API
                            │
                            ▼
                          :8080


Kubernetes kubelet
        │
        ▼
      :8081
        │
        ▼
   Health endpoints
```

The public Ingress should expose the application API, not the operational health interface.

For example:

```text
Public:
https://example.com/api/...

Private:
http://pod:8081/health/live
http://pod:8081/health/ready
```

This is preferable to relying on authentication for basic health probes.

---

# 9. Kubernetes: Use a Dedicated Health Port When Appropriate

A strong design is to expose application traffic and health traffic on separate ports.

```text
Pod
├── :8080 → application
│
└── :8081 → health
```

For example:

```yaml
containers:
  - name: app
    ports:
      - name: http
        containerPort: 8080

      - name: health
        containerPort: 8081

    livenessProbe:
      httpGet:
        path: /health/live
        port: health

    readinessProbe:
      httpGet:
        path: /health/ready
        port: health
```

The application can listen on both ports within the same process.

The public Service can expose only the application port:

```text
Internet
   │
   ▼
Ingress
   │
   ▼
Service :8080
   │
   ▼
Pod :8080
```

The health port does not need to be exposed through the public Service.

This provides a clean separation between:

```text
Application interface
        vs.
Operational interface
```

---

# 10. NetworkPolicy Considerations in Kubernetes

Kubernetes `NetworkPolicy` operates primarily at the **network layer (IP/port)**.

It cannot normally express:

```text
allow /health/live
deny /admin
```

because those are HTTP paths.

Therefore, if you want network-level isolation, a dedicated health port is useful:

```text
:8080 → application traffic
:8081 → health traffic
```

A NetworkPolicy can then control access to port `8081`.

However, be careful when restricting health ports with NetworkPolicy because the source address observed for kubelet probes depends on the Kubernetes networking implementation.

Do not assume that every cluster will see the kubelet as a particular Pod or node IP without verifying the behavior of your CNI/networking configuration.

A good approach is:

1. Do not expose the health port through the public Ingress.
2. Keep the health port separate where practical.
3. Add NetworkPolicy when there is a concrete access-control requirement.
4. Test the policy with the actual cluster networking implementation.

---

# 11. Authentication for Health Endpoints

Health endpoints generally do **not** need application authentication if they are properly network-restricted.

For example, avoid requiring:

```http
Authorization: Bearer <token>
```

for a Kubernetes liveness probe unless there is a specific reason to do so.

Why?

Because the health mechanism itself should remain extremely reliable.

If the health check requires:

```text
Application
   │
   ▼
Authentication service
   │
   ▼
Health check
```

then an authentication outage can make the health check fail even though the application itself is functioning.

Prefer:

```text
Kubernetes/network
       │
       ▼
private health endpoint
       │
       ▼
minimal response
```

Security should primarily come from **network exposure and minimal information**, rather than making the basic health mechanism dependent on authentication.

---

# 12. Non-Kubernetes Deployments

The same principles apply when deploying containers without Kubernetes.

For example:

```text
Docker
Docker Compose
systemd + Docker
VM + Docker
ECS
Nomad
custom deployment platform
```

You can still run one application process listening on multiple ports.

For example:

```text
Container
├── :8080 → application
└── :8081 → health
```

The Go application could conceptually do:

```go
go http.ListenAndServe(":8080", apiHandler)
http.ListenAndServe(":8081", healthHandler)
```

The important point is:

> **Kubernetes is not required for an application to expose multiple ports.**

One process can listen on multiple TCP ports.

---

# 13. Docker Example

Suppose the application listens on:

```text
:8080 → API
:8081 → health
```

Run it with:

```bash
docker run \
  -p 443:8080 \
  -p 127.0.0.1:8081:8081 \
  my-app:1.0.0
```

The result is:

```text
Internet
   │
   ▼
Host :443
   │
   ▼
Container :8080
   │
   ▼
Application


Host localhost :8081
   │
   ▼
Container :8081
   │
   ▼
Health
```

The key distinction is:

```text
-p 443:8080
```

publishes the application externally.

Whereas:

```text
-p 127.0.0.1:8081:8081
```

publishes the health port only on the host's loopback interface.

Therefore an external client cannot directly connect to port `8081`.

---

# 14. Docker Compose Example

A similar design can be used with Docker Compose:

```yaml
services:
  app:
    image: my-app:1.0.0

    ports:
      - "443:8080"

    expose:
      - "8081"

    healthcheck:
      test:
        - CMD
        - curl
        - -f
        - http://localhost:8081/health/live
      interval: 30s
      timeout: 5s
      retries: 3
```

The important distinction is:

```yaml
ports:
  - "443:8080"
```

publishes a port on the host.

Whereas:

```yaml
expose:
  - "8081"
```

makes the port available to the Docker network without publishing it externally.

---

# 15. Non-Kubernetes + Reverse Proxy

Another common architecture is:

```text
Internet
    │
    ▼
Nginx / HAProxy / Load Balancer
    │
    ▼
Application container :8080
```

The application may have:

```text
:8080
├── /api/...
└── /health/...
```

The reverse proxy can expose only the application routes and prevent public access to the health endpoints.

For example:

```text
Internet
   │
   ├── /api/*      → application
   │
   └── /health/*   → blocked
```

The monitoring system can access the health endpoint through the private network instead.

---

# 16. Non-Kubernetes + VM

If the application runs directly on a VM or in a container on a VM, another simple option is to bind the health interface only to localhost at the host level.

For example:

```text
Application
    │
    ├── public API :8080
    │
    └── health :8081
             │
             ▼
       localhost only
```

Monitoring can then run:

```bash
curl http://127.0.0.1:8081/health/ready
```

while the Internet cannot directly reach that port.

A host firewall can provide another layer of protection.

---

# 17. Do Health Endpoints Have to Be Private?

Not necessarily.

A public endpoint such as:

```text
GET /health/live

200 OK
{"status":"ok"}
```

is not inherently a vulnerability.

Many production systems expose simple health endpoints publicly.

The important question is:

> **Does exposing this endpoint provide useful sensitive information or create an operational/security risk?**

A minimal endpoint:

```json
{"status":"ok"}
```

usually reveals very little.

A detailed endpoint containing:

```text
database host
Redis host
Kafka brokers
internal hostname
environment
dependency versions
stack traces
configuration
```

is much more problematic.

Therefore:

> **Prefer private health endpoints when practical, but don't treat public reachability of a minimal health endpoint as automatically insecure.**

---

# 18. Separate Health From Diagnostics

Health is not the same as diagnostics.

### Health

```text
/health/live
/health/ready
```

Should be:

* fast;
* minimal;
* stable;
* unauthenticated when network-restricted;
* safe to call frequently.

### Diagnostics

```text
/admin/diagnostics
/admin/debug
```

May contain:

* dependency status;
* connection pool information;
* configuration information;
* detailed metrics;
* internal state.

Diagnostics should generally require:

* authentication;
* authorization;
* restricted network access;
* auditing where appropriate.

Do not combine the two.

---

# 19. HTTP Status Codes

Use HTTP status codes consistently.

### Healthy

```http
200 OK
```

### Not ready

```http
503 Service Unavailable
```

For example:

```json
{
  "status": "not_ready"
}
```

The body is primarily useful for humans and monitoring systems. The status code should carry the operational meaning.

Avoid returning:

```http
200 OK
{"status":"database_down"}
```

for a readiness check if the instance genuinely cannot serve traffic.

A `200` response generally means:

> This check succeeded.

---

# 20. Graceful Shutdown

Health checks become especially important during application shutdown.

A good lifecycle is:

```text
Running
   │
   │ shutdown signal
   ▼
Mark NOT READY
   │
   ▼
Stop receiving new traffic
   │
   ▼
Finish in-flight requests
   │
   ▼
Close dependencies
   │
   ▼
Exit
```

This prevents an instance that is shutting down from continuing to receive new requests.

This is particularly important behind a load balancer and in Kubernetes.

---

# 21. Avoid Health Check Side Effects

Health endpoints should be safe to call repeatedly.

Avoid:

```text
GET /health/ready
    │
    ├── writes database record
    ├── sends Kafka message
    ├── modifies cache
    └── triggers external API
```

Health checks should be effectively read-only.

A useful rule:

> **Calling a health endpoint 1,000 times should not change application state and should not create significant load.**

---

# 22. Logging

Because health checks can be called frequently, logging every successful health request at `INFO` level can create unnecessary noise.

Instead:

```text
Successful health checks → debug/trace/no log
Failed health checks     → appropriate warning/error/metric
```

The exact policy depends on the application and observability stack.

Also make sure monitoring systems do not generate an alert for every individual transient probe failure.

Use appropriate thresholds and aggregation.

---

# 23. Monitoring vs Health Checks

Health endpoints and metrics have different purposes.

### Health endpoint

Answers:

> Can this instance operate?

Example:

```text
/health/ready → 200
```

### Metrics

Answer:

> What is happening inside the system?

Examples:

```text
http_requests_total
http_request_duration_seconds
database_connections
queue_depth
error_rate
```

### Logs

Answer:

> What happened?

Do not try to turn `/health/ready` into a complete monitoring system.

---

# 24. Recommended Architecture

## Kubernetes

A strong production design is:

```text
                         Internet
                            │
                            ▼
                     ┌─────────────┐
                     │   Ingress   │
                     └──────┬──────┘
                            │
                            ▼
                     Service :8080
                            │
                            ▼
                    ┌──────────────┐
                    │     Pod      │
                    │              │
                    │ :8080 API    │
                    │              │
                    │ :8081 health │◄──── kubelet
                    └──────────────┘
```

Endpoints:

```text
/health/live
/health/ready
/health/startup
```

Use Kubernetes probes appropriately.

Do not expose the health port through the public Ingress.

---

## Non-Kubernetes

A strong production design is:

```text
                         Internet
                            │
                            ▼
                     Reverse Proxy
                            │
                            ▼
                    :8080 Application
                            │
                    ┌───────┴───────┐
                    │               │
                public          private
                                  │
                                :8081
                                  │
                            Health endpoints
                                  ▲
                                  │
                         Monitoring/deployer
```

Endpoints:

```text
/health/live
/health/ready
```

Usually no `/health/startup` unless the deployment platform actually needs it.

---

# 25. Interview Answer

A concise interview answer could be:

> **I separate liveness, readiness, and startup semantics. Liveness answers whether the process is alive and should be very cheap, while readiness answers whether the instance can actually serve traffic and may check required dependencies. Startup is useful for slow initialization, especially with Kubernetes.**
>
> **In Kubernetes, I configure these as liveness, readiness, and startup probes. I prefer keeping health endpoints off the public Ingress, and when practical I expose them on a separate health port so application traffic and operational traffic are separated.**
>
> **Without Kubernetes, I use the same principles but adapt them to the deployment platform. A simple container can listen on two ports—for example, `8080` for the public API and `8081` for private health checks. Docker can publish `8080` publicly while binding `8081` only to localhost.**
>
> **I keep health responses minimal, avoid exposing infrastructure details, avoid expensive dependency checks in liveness, use `503` for failed readiness, and make health checks read-only, fast, and safe to execute frequently. Detailed diagnostics belong behind a separate authenticated and authorized interface.**
>
> **Finally, during graceful shutdown I mark the instance unready first, allow traffic to drain, and then terminate the process.**

---

# 26. Practical Decision Guide

| Situation                        | Recommended approach                                          |
| -------------------------------- | ------------------------------------------------------------- |
| Simple application               | `/health/live` may be enough                                  |
| Application behind load balancer | `/health/live` + `/health/ready`                              |
| Kubernetes                       | Liveness + readiness; startup when needed                     |
| Slow application startup         | Add startup semantics                                         |
| Public API                       | Prefer health endpoint on private interface/port              |
| Single port only                 | Keep health minimal and block it at reverse proxy if desired  |
| Multiple ports possible          | Public API port + private health port                         |
| Detailed system diagnostics      | Separate authenticated admin/diagnostic endpoint              |
| Database temporarily unavailable | Usually readiness failure, not liveness failure               |
| Application fundamentally dead   | Liveness failure                                              |
| Application shutting down        | Mark not ready before termination                             |
| Health endpoint output           | Minimal JSON/status                                           |
| Health endpoint logging          | Avoid noisy successful-request logs                           |
| Network isolation in Kubernetes  | Separate health port + appropriate NetworkPolicy where needed |
| Docker/VM deployment             | Bind health port to localhost/private network                 |

---

# 27. Golden Rules

1. **Liveness means "should this process be restarted?"**
2. **Readiness means "should this instance receive traffic?"**
3. **Startup means "has initialization finished?"**
4. **Don't put expensive dependency checks into liveness.**
5. **Don't expose infrastructure details through health responses.**
6. **Keep health endpoints fast, deterministic, and side-effect free.**
7. **Use `503` when readiness fails.**
8. **Prefer a separate private health port when practical.**
9. **Don't expose health endpoints through a public Ingress unnecessarily.**
10. **Don't require authentication for basic health checks when network restriction already provides the required boundary.**
11. **Keep detailed diagnostics separate and protected.**
12. **During graceful shutdown, become unready before terminating.**
13. **Kubernetes does not define the concept of health endpoints—the application does; Kubernetes simply consumes them through probes.**
14. **Without Kubernetes, use the same health semantics only where they provide value; don't create endpoints merely to imitate Kubernetes.**

The overarching design is:

```text
                 PUBLIC
                   │
                   ▼
            Application API
                   │
                   │
              ┌────┴────┐
              │  App    │
              │ Process │
              └────┬────┘
                   │
             PRIVATE HEALTH
                   │
             ┌─────┴─────┐
             │           │
          /live        /ready
             │           │
             └─────┬─────┘
                   │
             Monitoring /
             Orchestrator
```

**The goal is not simply to have health endpoints. The goal is to give the deployment platform a reliable, minimal signal about application lifecycle while keeping operational information and network exposure under control.**
