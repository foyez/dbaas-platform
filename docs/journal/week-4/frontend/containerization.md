# Production-Grade Containerization of a Vue Application

## 1. Overview

A Vue application is a **frontend application**. After it is built, it does not require Node.js, npm, Vite, or the Vue development server to run in production.

A typical production architecture is:

```text
                    Internet
                       │
                    HTTPS :443
                       │
                       ▼
            ┌──────────────────────┐
            │ Kubernetes Ingress / │
            │ Gateway / Load Bal.  │
            │                      │
            │ TLS termination      │
            └──────────┬───────────┘
                       │
                       │ HTTP
                       ▼
            ┌──────────────────────┐
            │ Kubernetes Service   │
            └──────────┬───────────┘
                       │
                       ▼
            ┌──────────────────────┐
            │ Nginx container      │
            │                      │
            │ non-root user        │
            │ port 8080            │
            │                      │
            │ Vue dist/ files      │
            └──────────────────────┘
```

The responsibilities are intentionally separated:

- **Vue/Vite** builds the frontend.
- **Nginx** serves the generated static files.
- **Kubernetes** manages the container lifecycle and scaling.
- **Ingress/Gateway/Load Balancer** handles external networking and normally TLS termination.

The production container therefore contains only Nginx and the compiled Vue application.

---

# 2. Multi-Stage Docker Build

The recommended Dockerfile uses two stages:

1. **Builder** — installs Node dependencies and builds the Vue application.
2. **Runtime** — contains only Nginx and the generated `dist/` directory.

```dockerfile
# syntax=docker/dockerfile:1

# ============================================================
# Build stage
# ============================================================
FROM node:22-alpine AS builder

WORKDIR /app

# Copy dependency manifests first.
# This allows Docker to cache npm ci when source code changes.
COPY package.json package-lock.json ./

RUN npm ci

# Copy application source.
COPY . .

# Generate the production Vue application.
RUN npm run build


# ============================================================
# Runtime stage
# ============================================================
FROM nginxinc/nginx-unprivileged:alpine AS runtime

# Replace the default nginx configuration.
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy only the compiled Vue application.
COPY --from=builder /app/dist /usr/share/nginx/html

# nginx-unprivileged listens on 8080.
EXPOSE 8080

# Defense in depth: explicitly run as the nginx user.
USER nginx

ENTRYPOINT ["nginx"]
CMD ["-g", "daemon off;"]
```

## Why two stages?

The builder needs Node.js and the complete dependency tree:

```text
node
npm
vite
vue
typescript
other devDependencies
application source
```

None of these are required after the application has been built.

`npm run build` produces something similar to:

```text
dist/
├── index.html
└── assets/
    ├── index-a1b2c3.js
    ├── index-d4e5f6.css
    └── ...
```

The runtime image copies only this `dist/` directory.

Therefore the final image does **not** contain:

- Node.js
- npm
- Vite
- TypeScript
- development dependencies
- Vue source code
- build tooling

This reduces the runtime image and its attack surface.

---

# 3. Why Use `nginx-unprivileged`?

The runtime uses:

```dockerfile
FROM nginxinc/nginx-unprivileged:alpine
```

rather than:

```dockerfile
FROM nginx:alpine
```

The unprivileged image is designed to run Nginx without requiring a root process.

The container listens on:

```text
8080
```

instead of the traditional privileged HTTP port:

```text
80
```

This is not a problem in Kubernetes.

For example:

```text
Internet
   │
   │ HTTPS :443
   ▼
Ingress
   │
   │ :80
   ▼
Service
   │
   │ :8080
   ▼
Nginx container
```

The public port and container port do not have to be the same.

Running the application as non-root is an important container-hardening measure. The Dockerfile also explicitly specifies:

```dockerfile
USER nginx
```

even though the base image is already intended for unprivileged operation.

This provides defense in depth and makes the intended security posture obvious.

---

# 4. Nginx Configuration

The `nginx.conf` used by the container is:

```nginx
server {
    listen 8080;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    # Vue Router / SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Vite-generated static assets
    location /assets/ {
        try_files $uri =404;

        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Kubernetes health endpoint
    location = /healthz {
        access_log off;
        default_type text/plain;
        return 200 "ok\n";
    }
}
```

Each section has a specific purpose.

---

## 5. `listen 8080`

```nginx
listen 8080;
```

Nginx listens on port `8080`.

This is intentional because the container runs as a non-root user.

A process running without the required privileges should not need to bind to a privileged port such as `80`.

Kubernetes can still expose the application on whatever port is appropriate.

For example:

```text
Service port:       80
Container port:     8080
```

---

# 6. `root /usr/share/nginx/html`

```nginx
root /usr/share/nginx/html;
```

This tells Nginx where the Vue application's compiled files are located.

The Dockerfile puts the build output here:

```dockerfile
COPY --from=builder /app/dist /usr/share/nginx/html
```

So the resulting container looks approximately like:

```text
/usr/share/nginx/html/
├── index.html
└── assets/
    ├── index-a1b2c3.js
    ├── index-d4e5f6.css
    └── ...
```

Nginx simply serves these files.

There is no Vue server running in the production container.

---

# 7. Vue Router and `try_files`

This is one of the most important parts:

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

A Vue SPA commonly uses client-side routing.

For example:

```text
/
 /login
 /dashboard
 /users
 /settings
```

The browser expects Vue Router to handle these routes.

However, `/dashboard` is not necessarily a physical file:

```text
/usr/share/nginx/html/dashboard
```

does not exist.

Without the fallback, Nginx could return:

```text
404 Not Found
```

Instead:

```nginx
try_files $uri $uri/ /index.html;
```

means:

1. If the requested file exists, serve it.
2. If the requested directory exists, use it.
3. Otherwise, return `index.html`.

The browser receives `index.html`, loads the Vue application, and Vue Router handles `/dashboard`.

This is what allows direct navigation to SPA routes to work correctly.

---

# 8. Static Asset Caching

Vite normally generates hashed asset filenames:

```text
index-a1b2c3.js
index-d4e5f6.css
```

When the contents change, the filename changes.

That makes these files excellent candidates for long-term caching:

```nginx
location /assets/ {
    try_files $uri =404;

    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

The browser can therefore cache:

```text
index-a1b2c3.js
```

for a long time.

When a new frontend version is deployed, Vite generates a different filename:

```text
index-z9y8x7.js
```

so the browser downloads the new asset.

This gives good performance without requiring aggressive cache invalidation of every deployment.

---

# 9. `/healthz`

The Vue application itself does **not** provide `/healthz`.

Vue is not running as a server in production.

The endpoint is provided by Nginx:

```nginx
location = /healthz {
    access_log off;
    default_type text/plain;
    return 200 "ok\n";
}
```

A request:

```text
GET /healthz
```

receives:

```text
HTTP/1.1 200 OK

ok
```

This gives Kubernetes a simple way to verify that the frontend container is able to serve HTTP requests.

For example:

```yaml
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080

livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
```

The important distinction is that this endpoint verifies the **frontend container/Nginx**, not the health of your backend API or database.

Backend health should be handled by the backend's own health endpoints.

---

# 10. TLS/HTTPS

TLS should normally **not** be configured inside this Vue container when deploying to managed Kubernetes.

Instead:

```text
                    HTTPS
Internet ─────────────────────────► Ingress/Gateway
                                      │
                                      │ TLS termination
                                      ▼
                                 HTTP :80
                                      │
                                      ▼
                                  Service
                                      │
                                      ▼
                               Nginx :8080
```

The certificate can be managed by the cloud provider, ingress controller, cert-manager, or another platform-specific mechanism.

This keeps the frontend container simple.

The Nginx container's responsibility is:

```text
serve static Vue files
handle SPA routing
provide health endpoint
set caching headers
```

The Kubernetes/networking layer handles:

```text
TLS
load balancing
external networking
certificates
WAF/routing where applicable
```

End-to-end TLS from the ingress to the pod can still be used if organizational security requirements require it, but it is not normally necessary merely because the public application uses HTTPS.

---

# 11. `.dockerignore`

A production build should also have a `.dockerignore`:

```text
node_modules
dist
.git
.gitignore
.env
.env.*
*.log
Dockerfile*
docker-compose*.yml
README.md
```

This prevents unnecessary files from being sent to the Docker build context.

In particular, `.env` files and existing `node_modules` should not accidentally become part of the build context.

Secrets should never be baked into the Docker image.

---

# 12. Resulting Production Image

The build process is:

```text
                     BUILD TIME
                         │
                         ▼
              ┌─────────────────────┐
              │ node:22-alpine      │
              │                     │
              │ npm ci              │
              │ npm run build       │
              └──────────┬──────────┘
                         │
                         │ dist/
                         ▼
              ┌─────────────────────┐
              │ nginx-unprivileged  │
              │                     │
              │ USER nginx          │
              │                     │
              │ index.html          │
              │ assets/*            │
              │ nginx.conf          │
              └─────────────────────┘
                         │
                         ▼
                   PRODUCTION
```

The final container has a very small responsibility: **serve the already-built frontend**.

---

# 13. Production Security Characteristics

This approach provides several useful security properties:

|Concern|Approach|
|---|---|
|Root process|No|
|Node in runtime|No|
|npm in runtime|No|
|Build tools in runtime|No|
|Source code in runtime|No|
|TLS certificates in image|No|
|Secrets in image|No|
|Privileged port|No|
|SPA routing|Nginx|
|Health endpoint|Nginx|
|Static asset caching|Nginx|
|TLS|Kubernetes/cloud ingress|
|Scaling|Kubernetes|
|Process supervision|Kubernetes|

At the Kubernetes level, the container should additionally be deployed with security settings such as:

```yaml
securityContext:
  runAsNonRoot: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  seccompProfile:
    type: RuntimeDefault
```

Resource requests/limits and appropriate readiness/liveness probes should also be configured.

---

# 14. Recommended Production Model

For a Vue SPA running in a managed Kubernetes cluster, the clean separation of responsibilities is:

```text
┌──────────────────────────────────────────────────────┐
│                    Managed Kubernetes                │
│                                                      │
│  ┌──────────────┐                                    │
│  │ Ingress /    │                                    │
│  │ Gateway      │                                    │
│  │              │                                    │
│  │ TLS :443     │                                    │
│  └──────┬───────┘                                    │
│         │                                            │
│         ▼                                            │
│  ┌──────────────┐                                    │
│  │ Service      │                                    │
│  └──────┬───────┘                                    │
│         │                                            │
│         ▼                                            │
│  ┌─────────────────────────────┐                     │
│  │ Vue frontend pod            │                     │
│  │                             │                     │
│  │ nginx-unprivileged          │                     │
│  │ USER nginx                  │                     │
│  │ :8080                       │                     │
│  │                             │                     │
│  │ Vue static files            │                     │
│  │ SPA routing                 │                     │
│  │ /healthz                    │                     │
│  └─────────────────────────────┘                     │
│                                                      │
└──────────────────────────────────────────────────────┘
```

This is preferable to running the Vue development server, Node.js, or TLS directly inside the production frontend container.

The resulting production image is deliberately boring: **build with Node, serve with unprivileged Nginx, let Kubernetes handle infrastructure concerns.**