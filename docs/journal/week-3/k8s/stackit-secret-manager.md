Good. Since you already have a **STACKIT Secret Manager** with a user and stored secrets, the next step is to connect Kubernetes to it so your `stackit-registry-secret` is **created automatically** instead of manually running `kubectl create secret`.

For a GitOps setup with Argo CD, the usual pattern is:

```text
STACKIT Secret Manager
          |
          |
          v
 External Secrets Operator
          |
          |
          v
 Kubernetes Secret
 (stackit-registry-secret)
          |
          |
          v
 Deployment imagePullSecrets
```

You should install and configure **External Secrets Operator (ESO)**.

---

## 1. Install External Secrets Operator

Create:

```text
gitops/external-secrets/base/kustomization.yaml
```

Example:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - https://github.com/external-secrets/external-secrets/releases/download/v0.10.5/external-secrets.yaml
```

Then create an Argo CD Application:

```text
gitops/argocd/applications/external-secrets.yaml
```

similar to your CNPG application.

---

## 2. Create namespace

ESO usually runs in:

```text
external-secrets-system
```

Your structure:

```text
gitops/
└── external-secrets/
    └── base/
        ├── kustomization.yaml
        └── namespace.yaml
```

---

## 3. Configure STACKIT Secret Manager access

You need a Kubernetes Secret containing the STACKIT credentials.

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: stackit-secret-manager-auth
  namespace: external-secrets-system

type: Opaque

stringData:
  username: "<stackit-user>"
  password: "<stackit-password>"
```

Do **not** commit this plain secret to Git.

For production, store this using:

* Sealed Secrets
* SOPS
* Argo CD Vault Plugin

---

## 4. Create a SecretStore

Example:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore

metadata:
  name: stackit-secret-manager
  namespace: dbaas-api

spec:
  provider:
    stackit:
      auth:
        secretRef:
          username:
            name: stackit-secret-manager-auth
            key: username
          password:
            name: stackit-secret-manager-auth
            key: password
```

This tells ESO:

"Use STACKIT Secret Manager as the source."

---

## 5. Create ExternalSecret for Docker registry

Example:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret

metadata:
  name: stackit-registry-secret
  namespace: dbaas-api

spec:
  refreshInterval: 1h

  secretStoreRef:
    name: stackit-secret-manager
    kind: SecretStore

  target:
    name: stackit-registry-secret
    creationPolicy: Owner
    template:
      type: kubernetes.io/dockerconfigjson

  data:
    - secretKey: .dockerconfigjson
      remoteRef:
        key: docker-registry-config
```

Where:

```text
docker-registry-config
```

is the secret name/key you stored in STACKIT Secret Manager.

---

## 6. Your Deployment stays unchanged

Keep:

```yaml
spec:
  template:
    spec:
      imagePullSecrets:
        - name: stackit-registry-secret
```

The difference:

Before:

```text
You
 |
kubectl create secret
 |
Kubernetes
```

After:

```text
STACKIT Secret Manager
 |
External Secrets Operator
 |
Kubernetes Secret
 |
Pod image pull
```

---

## 7. Verify

After Argo CD sync:

Check ESO:

```bash
kubectl get externalsecret -A
```

Expected:

```text
NAMESPACE   NAME
dbaas-api   stackit-registry-secret
```

Check the generated secret:

```bash
kubectl get secret -n dbaas-api
```

Expected:

```text
stackit-registry-secret
```

Then restart API pods:

```bash
kubectl rollout restart deployment dbaas-deploy -n dbaas-api
```

---

For your platform architecture, I would add External Secrets as a **cluster-level platform component**:

```text
gitops/
├── cnpg/
├── external-secrets/
├── ingress-nginx/
├── operator/
├── api/
└── database/
```

and make it another child Application under your Argo CD App-of-Apps. This way a new SKE cluster can bootstrap itself without any manual secret creation.














Absolutely. Let's start from where you are **right now** and do this step by step. I’ll keep it concrete and avoid introducing SOPS, Sealed Secrets, etc. for now.

Your goal is:

> You already have a secret in STACKIT Secret Manager → Kubernetes should automatically create `stackit-registry-secret` → your Deployment uses it.

### Step 1 — Install External Secrets Operator

First, install **External Secrets Operator (ESO)** into your Kubernetes cluster.

Create this file:

```text
gitops/external-secrets/base/kustomization.yaml
```

Put this in it:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - https://github.com/external-secrets/external-secrets/releases/download/v0.10.5/external-secrets.yaml
```

Then commit it to Git.

**Do not create any Secret yet.**

---

### Step 2 — Deploy ESO with Argo CD

Create an Argo CD Application for ESO.

For example:

```text
gitops/argocd/applications/external-secrets.yaml
```

Put something like:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: external-secrets
  namespace: argocd
spec:
  project: default

  source:
    repoURL: <YOUR-GIT-REPOSITORY>
    targetRevision: HEAD
    path: gitops/external-secrets/base

  destination:
    server: https://kubernetes.default.svc
    namespace: external-secrets-system

  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

Adjust `<YOUR-GIT-REPOSITORY>` to match your existing Argo CD setup.

Then commit and push.

Check:

```bash
kubectl get pods -n external-secrets-system
```

You should see ESO pods running.

**Stop here and verify this works before continuing.**

---

## Step 3 — Give ESO access to STACKIT Secret Manager

This is the part that was confusing.

You told me:

> I already have a STACKIT Secret Manager with a user and stored secrets.

Good.

Now ESO needs to log in to STACKIT Secret Manager **using that user**.

You therefore need the STACKIT user's credentials available to Kubernetes.

For now, because you're learning, **don't worry about SOPS or Sealed Secrets yet**.

We can create this Kubernetes Secret manually once:

```bash
kubectl create secret generic stackit-secret-manager-auth \
  -n external-secrets-system \
  --from-literal=username='<YOUR-STACKIT-USERNAME>' \
  --from-literal=password='<YOUR-STACKIT-PASSWORD>'
```

Replace:

```text
<YOUR-STACKIT-USERNAME>
<YOUR-STACKIT-PASSWORD>
```

with the credentials of your STACKIT Secret Manager user.

This is different from your Docker registry secret.

You are creating:

```text
stackit-secret-manager-auth
```

Its purpose is:

> "ESO, here are the credentials you can use to access STACKIT Secret Manager."

---

## Step 4 — Check that Kubernetes has the authentication Secret

Run:

```bash
kubectl get secret stackit-secret-manager-auth \
  -n external-secrets-system
```

You should get something like:

```text
NAME                             TYPE     DATA   AGE
stackit-secret-manager-auth     Opaque   2      10s
```

**Do not put this Secret into Git yet.**

For now, create it manually with `kubectl` so you can understand how everything works.

---

## Step 5 — Create the SecretStore

Now tell ESO:

> "Use those credentials to access STACKIT Secret Manager."

Create:

```text
gitops/external-secrets/base/secret-store.yaml
```

Put:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: stackit-secret-manager
  namespace: dbaas-api

spec:
  provider:
    stackit:
      auth:
        secretRef:
          username:
            name: stackit-secret-manager-auth
            key: username
          password:
            name: stackit-secret-manager-auth
            key: password
```

Add it to your `kustomization.yaml`:

```yaml
resources:
  - https://github.com/external-secrets/external-secrets/releases/download/v0.10.5/external-secrets.yaml
  - secret-store.yaml
```

Commit and push.

---

## Step 6 — Check the SecretStore

Run:

```bash
kubectl get secretstore -n dbaas-api
```

You should see:

```text
NAME                     STATUS
stackit-secret-manager   Valid
```

If it says `Valid`, excellent.

That means:

```text
ESO
 |
 | username/password
 v
STACKIT Secret Manager
```

is working.

---

# Step 7 — Tell ESO which STACKIT secret you want

Now we get to your existing secret.

You said you already have secrets stored in STACKIT Secret Manager.

Let's say your Docker registry secret is called:

```text
docker-registry-config
```

You need to know the **exact name/key** of the secret in STACKIT.

For example:

```text
STACKIT Secret Manager

docker-registry-config
        |
        +-- contains Docker registry credentials
```

Now create:

```text
gitops/external-secrets/base/registry-external-secret.yaml
```

Put:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret

metadata:
  name: stackit-registry-secret
  namespace: dbaas-api

spec:
  refreshInterval: 1h

  secretStoreRef:
    name: stackit-secret-manager
    kind: SecretStore

  target:
    name: stackit-registry-secret
    creationPolicy: Owner
    template:
      type: kubernetes.io/dockerconfigjson

  data:
    - secretKey: .dockerconfigjson
      remoteRef:
        key: docker-registry-config
```

Then add it to your Kustomization:

```yaml
resources:
  - https://github.com/external-secrets/external-secrets/releases/download/v0.10.5/external-secrets.yaml
  - secret-store.yaml
  - registry-external-secret.yaml
```

Commit and push.

---

# Step 8 — ESO creates the Kubernetes Secret

Now ESO should do this automatically:

```text
STACKIT Secret Manager

docker-registry-config
          |
          | ESO reads it
          v
External Secrets Operator
          |
          | creates
          v
Kubernetes

dbaas-api/
    stackit-registry-secret
```

Check:

```bash
kubectl get externalsecret -n dbaas-api
```

You want:

```text
NAME                     STATUS
stackit-registry-secret  SecretSynced
```

Then:

```bash
kubectl get secret stackit-registry-secret -n dbaas-api
```

You should see:

```text
NAME                     TYPE                             DATA
stackit-registry-secret  kubernetes.io/dockerconfigjson   1
```

**That is the result we're trying to achieve.**

---

# Step 9 — Your Deployment uses it

Your Deployment should have:

```yaml
spec:
  template:
    spec:
      imagePullSecrets:
        - name: stackit-registry-secret
```

You don't manually create that Secret anymore.

ESO creates it.

So your complete flow is now:

```text
                    STACKIT
                Secret Manager
                     |
                     |
             docker credentials
                     |
                     v
          External Secrets Operator
                     |
                     |
                     v
              Kubernetes Secret
          stackit-registry-secret
                     |
                     |
                     v
                 Deployment
                     |
                     |
                     v
              Pull image from
             STACKIT registry
```

---

# Step 10 — Don't worry about SOPS yet

This is the important part.

For **your first working setup**, I suggest you do **not** introduce:

* SOPS
* Sealed Secrets
* Argo CD Vault Plugin

yet.

First get this working:

```text
STACKIT Secret Manager
        ↓
       ESO
        ↓
Kubernetes Secret
        ↓
Deployment
```

Once that works, we can make the **ESO authentication Secret** GitOps-friendly.

Right now, the only thing you create manually is:

```bash
kubectl create secret generic stackit-secret-manager-auth ...
```

That Secret contains the credentials ESO needs to log into STACKIT.

Your actual registry secret:

```text
stackit-registry-secret
```

will be created **automatically** by ESO.

---

## What you should do right now

Don't do everything at once.

Do these **three things first**:

**1. Install ESO through your existing Argo CD setup.**

**2. Verify:**

```bash
kubectl get pods -n external-secrets-system
```

**3. Create the ESO authentication Secret:**

```bash
kubectl create secret generic stackit-secret-manager-auth \
  -n external-secrets-system \
  --from-literal=username='<YOUR-STACKIT-USERNAME>' \
  --from-literal=password='<YOUR-STACKIT-PASSWORD>'
```

Then stop.

Once those three steps work, the next step is creating the `SecretStore`. We can do that one step at a time.
