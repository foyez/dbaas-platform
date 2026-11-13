
TLS setup with stackit webhook:

```
Let's Encrypt
        │
ACME DNS-01
        │
        ▼
┌───────────────┐
│  cert-manager │
└───────┬───────┘
        │
STACKIT webhook
        │
        ▼
STACKIT DNS API
        │
        ▼
_acme-challenge...
        │
        ▼
Certificate
        │
        ▼
Kubernetes Secret
        │
        ▼
NGINX Ingress
        │
HTTPS :443
```



Yes. Since you've **already created the DNS zone/records**, let's do this specifically for your STACKIT SKE + NGINX setup.

I checked the current STACKIT documentation: their recommended Kubernetes integration is **cert-manager + STACKIT cert-manager webhook + DNS-01**, and it supports wildcard certificates. ([Docs][1])

We'll create one certificate for:

```text
*.foyez.runs.onstackit.cloud
```

and use it for both:

```text
app.foyez.runs.onstackit.cloud
api.foyez.runs.onstackit.cloud
```

---

# Step 1 — Create a STACKIT service account

You need a dedicated STACKIT service account for cert-manager.

In the STACKIT Portal:

**IAM & Management → Service accounts → Create service account**

For example:

```text
dbaas-cert-manager
```

Then create a **service account key** for it and download the JSON.

STACKIT's documentation says the service account needs **DNS Admin** permissions because cert-manager must create/delete temporary TXT records for the DNS-01 challenge. ([Docs][1])

You should end up with something like:

```text
stackit-cert-manager-key.json
```

**Do not commit this file to Git.**

---

# Step 2 — Install cert-manager

You currently don't have it:

```bash
kubectl get pods -A | grep cert-manager
```

so install it.

```bash
kubectl create namespace cert-manager
```

Add the Helm repository:

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update
```

Install:

```bash
helm install cert-manager \
  --namespace cert-manager \
  jetstack/cert-manager \
  --set webhook.timeoutSeconds=15 \
  --set crds.enabled=true
```

This follows STACKIT's current installation guidance. ([Docs][1])

Check:

```bash
kubectl get pods -n cert-manager
```

You should get:

```text
cert-manager-xxxxx             1/1 Running
cert-manager-cainjector-xxxxx  1/1 Running
cert-manager-webhook-xxxxx     1/1 Running
```

Don't continue until these are running.

---

# Step 3 — Install the STACKIT cert-manager webhook

Add the STACKIT Helm repository:

```bash
helm repo add stackit-cert-manager-webhook \
  https://stackitcloud.github.io/stackit-cert-manager-webhook

helm repo update
```

Install it:

```bash
helm install stackit-cert-manager-webhook \
  --namespace cert-manager \
  stackit-cert-manager-webhook/stackit-cert-manager-webhook
```

Check:

```bash
kubectl get pods -n cert-manager
```

You should now have four relevant pods:

```text
cert-manager-xxxxx
cert-manager-cainjector-xxxxx
cert-manager-webhook-xxxxx
stackit-cert-manager-webhook-xxxxx
```

all `Running`.

This is the STACKIT-provided DNS-01 webhook. ([Docs][1])

---

# Step 4 — Create the STACKIT credential Secret

Now we need to make the service-account key available to the webhook.

The STACKIT documentation uses a Secret containing `sa.json`. ([Docs][1])

Since the official example uses a namespace-specific `Issuer`, we'll follow that approach.

You have two application namespaces:

```text
dbaas-frontend
dbaas-api
```

I'd actually create the certificate in **`dbaas-frontend`** and use that TLS Secret from the frontend Ingress.

However, Kubernetes Secrets are namespace-scoped, so you **cannot directly reference that Secret from `dbaas-api`**.

Therefore, for simplicity, I recommend creating the wildcard certificate in each namespace.

First:

```bash
kubectl create secret generic stackit-sa-authentication \
  -n dbaas-frontend \
  --from-file=sa.json=./stackit-cert-manager-key.json
```

Then:

```bash
kubectl create secret generic stackit-sa-authentication \
  -n dbaas-api \
  --from-file=sa.json=./stackit-cert-manager-key.json
```

Verify:

```bash
kubectl get secret stackit-sa-authentication -n dbaas-frontend
kubectl get secret stackit-sa-authentication -n dbaas-api
```

You can then securely delete the local key if you have it stored safely elsewhere.

---

# Step 5 — Enable STACKIT authentication in the webhook

Now upgrade the webhook:

```bash
helm upgrade stackit-cert-manager-webhook \
  --namespace cert-manager \
  stackit-cert-manager-webhook/stackit-cert-manager-webhook \
  --set stackitSaAuthentication.enabled=true
```

Check:

```bash
kubectl get pods -n cert-manager
```

If the webhook doesn't start correctly, check:

```bash
kubectl logs \
  -n cert-manager \
  deployment/stackit-cert-manager-webhook
```

---

# Step 6 — Find your STACKIT project ID

You need the **project ID containing your DNS zone**.

It looks like:

```text
4ec710a8-fe7c-42d3-aca4-84d9ff69a14e
```

You can get it from the STACKIT Portal.

This is important: the `projectId` must correspond to the project containing the DNS zone. STACKIT explicitly calls this out in their documentation. ([Docs][1])

Let's assume:

```text
YOUR_PROJECT_ID
```

for now.

---

# Step 7 — Create the Issuer

Create:

```text
letsencrypt-prod.yaml
```

with:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-prod
  namespace: dbaas-frontend

spec:
  acme:
    email: YOUR_EMAIL@example.com

    server: https://acme-v02.api.letsencrypt.org/directory

    privateKeySecretRef:
      name: letsencrypt-prod

    solvers:
      - dns01:
          webhook:
            solverName: stackit
            groupName: acme.stackit.de
            config:
              projectId: YOUR_PROJECT_ID
```

Replace:

```text
YOUR_EMAIL@example.com
```

and:

```text
YOUR_PROJECT_ID
```

Then:

```bash
kubectl apply -f letsencrypt-prod.yaml
```

Check:

```bash
kubectl get issuer -n dbaas-frontend
```

You want:

```text
NAME               READY
letsencrypt-prod   True
```

---

# Step 8 — Create the wildcard certificate

Now create:

```text
dbaas-wildcard-certificate.yaml
```

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: dbaas-wildcard
  namespace: dbaas-frontend

spec:
  secretName: dbaas-wildcard-tls

  issuerRef:
    name: letsencrypt-prod
    kind: Issuer

  commonName: "*.foyez.runs.onstackit.cloud"

  dnsNames:
    - "*.foyez.runs.onstackit.cloud"
    - "foyez.runs.onstackit.cloud"
```

The second DNS name is optional for your current two subdomains, but it's useful if you later want the root domain too.

Apply:

```bash
kubectl apply -f dbaas-wildcard-certificate.yaml
```

Now:

```bash
kubectl get certificate -n dbaas-frontend
```

Initially:

```text
NAME              READY   SECRET               AGE
dbaas-wildcard    False   dbaas-wildcard-tls    10s
```

Wait a little.

Then:

```bash
kubectl get certificate -n dbaas-frontend
```

You want:

```text
NAME              READY   SECRET               AGE
dbaas-wildcard    True    dbaas-wildcard-tls    2m
```

---

# Step 9 — Watch the DNS-01 challenge

This is a very useful command while testing:

```bash
kubectl get challenges -n dbaas-frontend
```

You should see something like:

```text
NAME                           STATE     DOMAIN
dbaas-wildcard-xxxxx           pending   foyez.runs.onstackit.cloud
```

Eventually:

```text
NAME                           STATE      DOMAIN
dbaas-wildcard-xxxxx           valid      foyez.runs.onstackit.cloud
```

You can get more details with:

```bash
kubectl describe challenge -n dbaas-frontend
```

During the process, cert-manager + STACKIT webhook will create a temporary record roughly like:

```text
_acme-challenge.foyez.runs.onstackit.cloud
```

You **don't need to create this manually**. That's the whole point of DNS-01 automation. STACKIT's webhook creates and removes it through the DNS API. ([Docs][1])

---

# Step 10 — Make the certificate available to the API namespace

Here's an important Kubernetes detail.

Your certificate currently creates:

```text
dbaas-frontend/dbaas-wildcard-tls
```

But your API Ingress is in:

```text
dbaas-api
```

A Secret cannot be referenced across namespaces.

So for the API, you have two choices.

### Option A — Create another Certificate

This is the cleanest approach.

Create another Issuer in `dbaas-api` using the same configuration, then another Certificate.

Or:

### Option B — Copy the TLS Secret

For a simple setup, you can copy the resulting TLS Secret:

```bash
kubectl get secret dbaas-wildcard-tls \
  -n dbaas-frontend \
  -o yaml
```

But **don't manually copy the YAML blindly**, because it contains namespace metadata and other fields.

For production, I'd use **two Certificate resources** and let cert-manager manage both.

---

# Step 11 — Create the API Issuer

Create:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-prod
  namespace: dbaas-api

spec:
  acme:
    email: YOUR_EMAIL@example.com
    server: https://acme-v02.api.letsencrypt.org/directory

    privateKeySecretRef:
      name: letsencrypt-prod

    solvers:
      - dns01:
          webhook:
            solverName: stackit
            groupName: acme.stackit.de
            config:
              projectId: YOUR_PROJECT_ID
```

Then:

```bash
kubectl apply -f api-issuer.yaml
```

---

# Step 12 — Create API certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: dbaas-api-tls
  namespace: dbaas-api

spec:
  secretName: dbaas-api-tls

  issuerRef:
    name: letsencrypt-prod
    kind: Issuer

  commonName: "api.foyez.runs.onstackit.cloud"

  dnsNames:
    - "api.foyez.runs.onstackit.cloud"
```

Apply:

```bash
kubectl apply -f api-certificate.yaml
```

Check:

```bash
kubectl get certificate -A
```

Eventually:

```text
NAMESPACE         NAME               READY
dbaas-frontend    dbaas-wildcard     True
dbaas-api         dbaas-api-tls      True
```

---

# Step 13 — Update your frontend Ingress

Now modify your existing Ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dbaas-frontend
  namespace: dbaas-frontend

spec:
  ingressClassName: nginx

  tls:
    - hosts:
        - app.foyez.runs.onstackit.cloud
      secretName: dbaas-wildcard-tls

  rules:
    - host: app.foyez.runs.onstackit.cloud
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: dbaas-frontend-svc
                port:
                  number: 3000
```

Notice that **we don't need a cert-manager annotation here** because we're explicitly creating the `Certificate` resource ourselves.

---

# Step 14 — Update API Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dbaas-api
  namespace: dbaas-api

spec:
  ingressClassName: nginx

  tls:
    - hosts:
        - api.foyez.runs.onstackit.cloud
      secretName: dbaas-api-tls

  rules:
    - host: api.foyez.runs.onstackit.cloud
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: dbaas-svc
                port:
                  number: 8080
```

Apply:

```bash
kubectl apply -f frontend-ingress.yaml
kubectl apply -f api-ingress.yaml
```

---

# Step 15 — Test HTTPS

Now:

```bash
curl -I https://app.foyez.runs.onstackit.cloud
```

and:

```bash
curl -I https://api.foyez.runs.onstackit.cloud
```

You should no longer get the `Crypto.subtle` problem in your browser because:

```text
https://app.foyez.runs.onstackit.cloud
```

is now a secure context.

---

# Step 16 — Fix your ZITADEL redirect URI

This is important for your Vue application.

In ZITADEL, configure the redirect URI as:

```text
https://app.foyez.runs.onstackit.cloud/auth/callback
```

Your frontend should therefore have:

```text
Frontend
https://app.foyez.runs.onstackit.cloud

API
https://api.foyez.runs.onstackit.cloud

ZITADEL callback
https://app.foyez.runs.onstackit.cloud/auth/callback
```

And your Vue OIDC configuration should **not** use `http://`.

---

## One correction to my earlier recommendation

Because you're on **STACKIT SKE and already have STACKIT DNS**, I would **not use the generic HTTP-01 Let's Encrypt setup I initially suggested**.

The current STACKIT documentation specifically describes:

```text
cert-manager
     +
STACKIT cert-manager webhook
     +
STACKIT DNS
     +
Let's Encrypt DNS-01
```

and specifically documents wildcard certificates. ([Docs][1])

That's the setup I'd use for your cluster.

### Before you start

You already have the DNS zone, so the remaining things you need are:

* [ ] STACKIT service account
* [ ] Service account key JSON
* [ ] DNS Admin permission for that service account
* [ ] cert-manager
* [ ] STACKIT cert-manager webhook
* [ ] `projectId`
* [ ] Issuer
* [ ] Certificate
* [ ] TLS Secret referenced by NGINX
* [ ] ZITADEL redirect URI changed to HTTPS

**Don't proceed past Step 1 by pasting your STACKIT private key here.** Keep that JSON local/secret.

If you give me your **STACKIT project ID** (the ID itself is not a secret) and tell me whether your DNS zone is exactly `foyez.runs.onstackit.cloud`, I can tailor the Issuer/Certificate YAML exactly to your setup.
