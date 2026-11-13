# Secret Management

### Option 1 — Create secret outside Git (simple)

```bash
kubectl create secret docker-registry registry-secret \
  --docker-server=registry.onstackit.cloud \
  --docker-username=$REGISTRY_USER \
  --docker-password=$REGISTRY_PASSWORD \
  -n <namespace>
```

### Option 2 — Use a secret management tool (recommended)

Examples:

* STACKIT Secret Manager (if available in your setup)
* External Secrets Operator
* HashiCorp Vault
* Cloud provider secret stores

### Option 3 — Store an encrypted Secret manifest

If you use GitOps (Argo CD/Flux), tools like:

* Sealed Secrets
* SOPS

allow you to keep an encrypted YAML file in Git.

For your current SKE setup, a clean structure would be:

```
k8s/
├── namespace.yaml
├── deployment.yaml
├── service.yaml
├── ingress.yaml
├── configmap.yaml
└── registry-secret.yaml   # generated/encrypted, not plain credentials
```

So yes, a `secret.yaml` is valid; I omitted it because registry credentials are usually treated differently from application configuration secrets. For a real deployment pipeline, you would usually automate its creation rather than keep the raw Secret YAML in the repository.


Since you created `stackit-registry-secret` **after** the pod was already failing, Kubernetes will not automatically recreate the failed pull immediately in every case. The clean approach is to restart the pods so kubelet retries the image pull using the new secret.

First verify the secret exists in the **same namespace** as your Deployment:

```bash
kubectl get secret -n dbaas-api
```

You should see:

```text
stackit-registry-secret
```

Then delete the failed pods:

```bash
kubectl delete pods -n dbaas-api -l app=dbaas
```

Because your Deployment has:

```yaml
replicas: 2
```

the ReplicaSet will create new pods automatically.

Watch:

```bash
kubectl get pods -n dbaas-api -w
```

You should see:

```text
dbaas-deploy-xxxxx   0/1   ContainerCreating
dbaas-deploy-xxxxx   1/1   Running
```

---

Also check the secret is a valid docker registry secret:

```bash
kubectl get secret stackit-registry-secret \
  -n dbaas-api \
  -o jsonpath='{.type}'
```

Expected:

```text
kubernetes.io/dockerconfigjson
```

If not, recreate it:

```bash
kubectl create secret docker-registry stackit-registry-secret \
  --namespace dbaas-api \
  --docker-server=registry.onstackit.cloud \
  --docker-username=<username> \
  --docker-password=<password>
```

---

## For GitOps best practice

Right now you manually created:

```text
dbaas-api/stackit-registry-secret
```

but Argo CD does not know about it.

You should eventually manage it through GitOps. Common options:

### Option 1: External Secrets Operator (recommended)

Store credentials in:

* STACKIT Secret Manager
* Vault
* AWS Secrets Manager
* etc.

and let Kubernetes create:

```yaml
kind: Secret
metadata:
  name: stackit-registry-secret
```

### Option 2: Sealed Secrets

Commit an encrypted secret:

```text
gitops/api/base/
├── deployment.yaml
├── service.yaml
└── registry-secret.yaml   # encrypted
```

### Option 3: Create it during cluster bootstrap

For a private registry used by all workloads, create it once in your infrastructure layer.

---

For your current setup, do:

```bash
kubectl delete pods -n dbaas-api -l app=dbaas
```

and the API should start pulling the image successfully.
