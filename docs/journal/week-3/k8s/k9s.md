# K9s

## 1. What K9s is

**K9s = terminal UI for Kubernetes.**

Use it to quickly:

* 👀 View resources
* 📋 Inspect resources
* 📜 Read logs
* 🔍 Troubleshoot
* 🐚 Enter containers
* ✏️ Edit resources
* 🗑️ Delete resources
* 📊 Check CPU/memory
* 🔄 Switch contexts/namespaces

> **K9s is a Kubernetes tool, so understanding Kubernetes matters more than memorizing K9s shortcuts.**

---

# 2. The Kubernetes structure you MUST understand

```text
Namespace
   │
   ├── Deployment
   │      │
   │      └── ReplicaSet
   │              │
   │              └── Pod
   │                    └── Container
   │
   ├── Service ────────> Pods
   │
   ├── ConfigMap
   ├── Secret
   └── Ingress
```

### Most important relationships

```text
Deployment → manages Pods

Service → sends traffic to Pods

Ingress → exposes/routes external traffic

ConfigMap/Secret → provides configuration

Node → runs Pods
```

---

# 3. Essential navigation

| Key / Command | Purpose                |
| ------------- | ---------------------- |
| `:`           | Command/resource mode  |
| `?`           | Help                   |
| `Esc`         | Back                   |
| `Ctrl-C`      | Exit                   |
| `:ctx`        | Kubernetes contexts    |
| `:ns`         | Namespaces             |
| `:po`         | Pods                   |
| `:deploy`     | Deployments            |
| `:svc`        | Services               |
| `:ing`        | Ingresses              |
| `:node`       | Nodes                  |
| `:cm`         | ConfigMaps             |
| `:sec`        | Secrets                |
| `:pvc`        | PersistentVolumeClaims |
| `:events`     | Events                 |

### Resource navigation

```text
:po
:deploy
:svc
:node
```

You can also use `?` whenever you forget a shortcut.

---

# 4. The BIG 5 shortcuts

These are the ones I'd memorize first:

```text
l     → Logs
d     → Describe
s     → Shell
e     → Edit
Ctrl-D → Delete
```

### Remember:

```text
l = What is the application saying?

d = What does Kubernetes know?

s = Let me inspect inside the container.

e = Change the resource.

Ctrl-D = Delete it.
```

⚠️ `e` and `Ctrl-D` can make real changes. Be especially careful in production.

---

# 5. Filtering

Don't scroll through hundreds of resources.

```text
/my-app
```

Filter the current view.

Useful when you have:

```text
100+ pods
multiple applications
multiple replicas
large namespaces
```

---

# 6. The most important troubleshooting workflow

When someone says:

> **"My application is broken."**

Do this:

```text
1. Check context
       ↓
2. Check namespace
       ↓
3. Check Deployment
       ↓
4. Check Pods
       ↓
5. Check logs
       ↓
6. Describe Pod
       ↓
7. Check Events
       ↓
8. Check Node
       ↓
9. Check Service
       ↓
10. Check configuration
```

### In K9s:

```text
:ctx
:ns
:deploy
:po
l
d
:events
:node
:svc
```

---

# 7. Pod status — know these

### 🟢 `Running`

Container is running.

**Does NOT automatically mean application is healthy.**

### 🟡 `Pending`

Pod hasn't started.

Common causes:

* insufficient CPU/memory
* scheduling problem
* volume problem
* node constraints

### 🔴 `CrashLoopBackOff`

Container starts → crashes → Kubernetes restarts it.

Check:

```text
l
d
:events
```

Especially check **Last State** and **Restart Count**.

### 🔴 `ImagePullBackOff`

Kubernetes can't pull the container image.

Check:

* image name/tag
* registry access
* credentials
* network

### 🔴 `Error`

Container/application exited with an error.

Check logs and describe.

---

# 8. `Running` ≠ healthy

Very important:

```text
Pod Running
     ≠
Application healthy
```

Check:

```text
Readiness probe → Should it receive traffic?

Liveness probe → Is it still alive?

Startup probe → Has it finished starting?
```

A pod can be running but **not Ready**.

---

# 9. Logs vs Describe vs Events

### Logs

```text
l
```

Use when you want to know:

> **What is my application doing/saying?**

Look for:

```text
ERROR
Exception
timeout
connection refused
authentication failed
OOM
```

### Describe

```text
d
```

Use when you want to know:

> **What does Kubernetes know about this resource?**

Look at:

```text
State
Last State
Restart Count
Conditions
Volumes
Image
Events
```

### Events

```text
:events
```

Useful for:

```text
FailedScheduling
FailedMount
FailedPull
Unhealthy
BackOff
Insufficient CPU
Insufficient memory
```

---

# 10. Service troubleshooting

If:

```text
Pod = Running
Logs = OK
```

but users can't reach the application:

```text
Check Service
      ↓
Check selector
      ↓
Check Pod labels
      ↓
Check endpoints
      ↓
Check Ingress
```

Remember:

```text
Service
   ↓
selector
   ↓
matching Pod labels
   ↓
traffic
```

A Service can exist while pointing to **zero pods** because of a wrong selector.

---

# 11. Context & Namespace — GOLDEN RULE

Before changing anything:

```text
Which cluster?
Which namespace?
Which resource?
```

Check:

```text
:ctx
:ns
```

### Production habit

For investigation:

```bash
k9s --readonly
```

Use normal write access only when you actually need to change something.

---

# 12. Labels & Selectors

Labels identify resources:

```yaml
app: backend
environment: production
```

Selectors find matching resources:

```text
app=backend
```

Critical relationship:

```text
Service selector
       ↓
Pod labels
```

If they don't match → Service won't route traffic to the Pods.

---

# 13. Node problems

If **many unrelated Pods** suddenly have problems:

```text
Check Pods
    ↓
Check Nodes
```

Use:

```text
:node
```

Look at:

```text
CPU
Memory
Disk
Conditions
Taints
```

Think:

> **One pod broken → investigate application.**

> **Many pods broken → investigate cluster/node.**

---

# 14. ConfigMap & Secret

```text
:cm
:sec
```

Applications commonly depend on:

```text
environment variables
configuration
database URLs
credentials
certificates
```

If the application suddenly can't connect to something:

```text
Logs
  ↓
ConfigMap / Secret
  ↓
Check configuration
```

⚠️ Never casually expose passwords, tokens, API keys, or certificates.

---

# 15. Shell

```text
s
```

Use it when you need to inspect the container itself.

Examples:

```bash
env
ls
ps
curl localhost:8080
```

Useful for checking:

* environment
* files
* processes
* connectivity
* application health

---

# 16. Production safety rules 🚨

### ALWAYS verify:

```text
Context
Namespace
Resource
```

### Before changing:

```text
e
```

Ask:

> Do I understand what I'm changing?

### Before deleting:

```text
Ctrl-D
```

Ask:

> Who owns this resource?

Example:

```text
Deployment
   ↓
ReplicaSet
   ↓
Pod
```

Deleting a Pod is very different from deleting a Deployment.

---

# 17. Don't memorize everything

You only need to be fluent with:

```text
:ctx
:ns
:po
:deploy
:svc
:node

l
d
s
e
Ctrl-D
/
?
Esc
```

And understand:

```text
Deployment
Pod
Service
Ingress
Node
ConfigMap
Secret
Labels
Selectors
Probes
Events
```

---

# 🧠 One-page mental model

```text
                 Kubernetes
                      │
                 Namespace
                      │
              ┌───────┴───────┐
              │               │
         Deployment        Service
              │               │
          ReplicaSet          │
              │               │
             Pod <────────────┘
              │
          Container
              │
        ┌─────┴─────┐
       Logs       Config
        │        CM/Secret
        │
      Events
        │
       Node
```

### If something breaks:

```text
Context?
   ↓
Namespace?
   ↓
Deployment?
   ↓
Pod?
   ↓
Running / Ready?
   ↓
Logs?
   ↓
Describe?
   ↓
Events?
   ↓
Node?
   ↓
Service?
   ↓
Ingress / Config?
```

**If you know this page well, you know the core K9s workflow.**
