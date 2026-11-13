# GitOps Pipeline: Forgejo → Docker Hub → ArgoCD

Picking up from where you already are: SKE cluster provisioned and ArgoCD installed, both via `.forgejo/workflows/platform-bootstrap.yaml`. This guide covers everything after that — repo layout, building and pushing the operator image, and wiring ArgoCD to deploy from your Forgejo repo.

## Table of Contents

1. [The End-to-End Picture](#1-the-end-to-end-picture)
2. [Step 1 — Repo Layout](#step-1--repo-layout)
3. [Step 2 — Docker Hub Access Token & Forgejo Secrets](#step-2--docker-hub-access-token--forgejo-secrets)
4. [Step 3 — The GitOps Overlay for Your Operator](#step-3--the-gitops-overlay-for-your-operator)
5. [Step 4 — GitOps Manifests for CNPG and Your PostgreSQL CRs](#step-4--gitops-manifests-for-cnpg-and-your-postgresql-crs)
6. [Step 5 — Give ArgoCD Access to Your Private Forgejo Repo](#step-5--give-argocd-access-to-your-private-forgejo-repo)
7. [Step 6 — ArgoCD Application Manifests](#step-6--argocd-application-manifests)
8. [Step 7 — Bootstrap the Applications](#step-7--bootstrap-the-applications)
9. [Step 8 — The CI Workflow: Build, Push, Bump Tag](#step-8--the-ci-workflow-build-push-bump-tag)
10. [Step 9 — Run It End to End](#step-9--run-it-end-to-end)
11. [Optional Refinement: App-of-Apps](#optional-refinement-app-of-apps)

---

## 1. The End-to-End Picture

```
 you push code to operator/
        │
        ▼
 Forgejo Actions (operator-ci.yaml)
   ├── go build / go vet
   ├── podman build → push image to Docker Hub
   └── bump image tag in gitops/postgres-platform/kustomization.yaml
        │
        │  (a normal git commit, pushed back to the same repo)
        ▼
 gitops/  directory now has a new image tag committed
        │
        ▼
 ArgoCD (watching gitops/ on a timer + webhook)
   detects the diff → kustomize build → kubectl apply
        │
        ▼
 New operator Pod running in the cluster
```

The important idea underneath all of this: **your CI pipeline never talks to the cluster directly.** It only ever writes an intention into Git (a new image tag). ArgoCD is the only thing that actually applies anything to Kubernetes, and it only ever applies what's committed. This is what "Git is the source of truth" means in practice, not just as a slogan — if you ever wonder what's running in production, the answer is always "whatever the `gitops/` directory says," full stop.

---

## Step 1 — Repo Layout

You already have `terraform/` and `.forgejo/workflows/` at the repo root. Add two more top-level directories:

```
.
├── .forgejo/workflows/
│   ├── platform-bootstrap.yaml        # existing — SKE + ArgoCD install
│   └── operator-ci.yaml               # new — Step 8
├── terraform/                          # existing
├── operator/                           # your Kubebuilder project, unchanged in structure
│   ├── api/v1/
│   ├── internal/controller/
│   ├── cmd/main.go
│   ├── Dockerfile
│   ├── Makefile
│   └── config/
│       ├── crd/
│       ├── rbac/
│       ├── manager/
│       │   └── kustomization.yaml      # image name/tag transformer lives here (Kubebuilder scaffolds this)
│       └── default/
│           └── kustomization.yaml      # ties crd+rbac+manager together
└── gitops/                             # everything ArgoCD watches — Steps 3–4
    ├── cnpg/
    │   └── kustomization.yaml
    ├── postgres-platform/
    │   └── kustomization.yaml          # CI bumps the image tag here — NOT inside operator/
    ├── applications/
    │   └── customer-db.yaml
    └── argocd/
        ├── app-cnpg.yaml
        ├── app-postgres-platform.yaml
        └── app-databases.yaml
```

**Why the image tag lives in `gitops/postgres-platform/`, not in `operator/config/manager/`:** if CI patched a file under `operator/`, and your CI workflow also *triggers* on changes under `operator/`, the bot's own commit would re-queue itself. Keeping the CI-writable file physically outside the path the workflow triggers on is a small structural habit that avoids that whole class of problem — worth doing even though (as you'll see in Step 8) Forgejo's built-in Actions token also refuses to self-trigger, which is a second, independent safety net.

---

## Step 2 — Docker Hub Access Token & Forgejo Secrets

Rather than your Docker Hub password, generate a scoped **access token**:

1. Docker Hub → **Account Settings → Security → New Access Token**
2. Give it a description like `forgejo-ci-postgres-platform`, permission **Read & Write**.
3. Copy the token now — Docker Hub won't show it again.

In your Forgejo repo: **Settings → Actions → Secrets**, add:

| Secret name | Value |
|---|---|
| `DOCKERHUB_USERNAME` | your Docker Hub username |
| `DOCKERHUB_TOKEN` | the access token from step 3 above |

These sit alongside the `AWS_ACCESS_KEY_ID`, `STACKIT_PROJECT_ID`, etc. you already have configured for `platform-bootstrap.yaml` — same mechanism, different job.

---

## Step 3 — The GitOps Overlay for Your Operator

Kubebuilder already scaffolds an image transformer in `operator/config/manager/kustomization.yaml` (something like `images: [{name: controller, newName: controller, newTag: latest}]`). Rather than editing that file directly, create a **separate overlay** in `gitops/` that references it and overrides the image — this is the file CI will actually touch.

**File: `gitops/postgres-platform/kustomization.yaml`**

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../operator/config/default

images:
  - name: controller
    newName: docker.io/youruser/postgres-platform-operator
    newTag: init   # CI overwrites this line — the value here is only a placeholder
```

Replace `youruser` with your actual Docker Hub username. `../../operator/config/default` is Kubebuilder's own bundled kustomization (CRD + RBAC + Deployment) — you're not duplicating any of that, just pointing at it and layering an image override on top.

**✅ Verify this resolves correctly before wiring up ArgoCD:**

```bash
kustomize build gitops/postgres-platform | grep -A2 "image:"
```

You should see the full `docker.io/youruser/postgres-platform-operator:init` reference in the rendered Deployment — confirming the overlay is actually reaching the base and overriding the image, before you ever hand this to ArgoCD.

---

## Step 4 — GitOps Manifests for CNPG and Your PostgreSQL CRs

**File: `gitops/cnpg/kustomization.yaml`** — the CNPG install becomes GitOps-managed instead of a one-off `kubectl apply`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/main/releases/cnpg-1.24.0.yaml
```

> ArgoCD's repo-server needs outbound network access to `raw.githubusercontent.com` to resolve this at sync time — worth a quick check if this Application later shows an error instead of just deploying.

**File: `gitops/applications/customer-db.yaml`** — your actual product instances, the same CR from the earlier build guide:

```yaml
apiVersion: database.example.com/v1
kind: PostgreSQL
metadata:
  name: customer-db
  namespace: default
spec:
  version: "16"
  replicas: 3
  storage: 10Gi
  database: customer
  username: customer-admin
  connectionPooler:
    enabled: true
    poolSize: 20
```

You can add more `PostgreSQL` CRs to this same directory as separate files as you provision more databases — each is picked up automatically by the Application that watches `gitops/applications/` (Step 6).

---

## Step 5 — Give ArgoCD Access to Your Private Forgejo Repo

ArgoCD needs credentials to clone your repo before it can sync anything from it. **Do not commit this credential to Git** — create it directly against the cluster.

1. In Forgejo: **Settings → Applications → Generate New Token**, scope: `repo` (read access is enough — ArgoCD never needs to push).
2. Create the secret directly, without saving it as a file first:

```bash
kubectl create secret generic postgres-platform-repo \
  -n argocd \
  --from-literal=type=git \
  --from-literal=url=https://<your-forgejo-host>/<org>/<repo>.git \
  --from-literal=username=<your-forgejo-username> \
  --from-literal=password=<the-token-from-step-1>

kubectl label secret postgres-platform-repo -n argocd \
  argocd.argoproj.io/secret-type=repository
```

Replace `<your-forgejo-host>/<org>/<repo>.git` with your repo's actual HTTPS clone URL (visible under the repo's **Clone** button in Forgejo).

**✅ Verify:**

```bash
kubectl get secret postgres-platform-repo -n argocd -o jsonpath='{.metadata.labels}'
# should show: {"argocd.argoproj.io/secret-type":"repository"}
```

If you're on the `argocd` CLI rather than raw `kubectl`, the equivalent is `argocd repo add https://<host>/<org>/<repo>.git --username <user> --password <token>` — same end result, ArgoCD just wraps the same Secret creation for you.

> **On production hygiene:** the imperative `kubectl create secret` above is fine for now, but it means the credential only exists in the cluster, not in Git — which is correct (you never want a plaintext token in a repo), but does mean it's not disaster-recovery-covered by GitOps alone. If this matters to you later, look at **Sealed Secrets** or **SOPS** — both let you commit an *encrypted* version of a secret to Git safely. Not needed to get this working today.

---

## Step 6 — ArgoCD Application Manifests

Three separate `Application` objects — one per concern, each pointing at a different `gitops/` subdirectory.

<details>
<summary><code>gitops/argocd/app-cnpg.yaml</code></summary>

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: cnpg-operator
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://<your-forgejo-host>/<org>/<repo>.git
    targetRevision: main
    path: gitops/cnpg
  destination:
    server: https://kubernetes.default.svc
    namespace: cnpg-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```
</details>

<details>
<summary><code>gitops/argocd/app-postgres-platform.yaml</code></summary>

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: postgres-platform-operator
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://<your-forgejo-host>/<org>/<repo>.git
    targetRevision: main
    path: gitops/postgres-platform
  destination:
    server: https://kubernetes.default.svc
    namespace: postgres-platform-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```
</details>

<details>
<summary><code>gitops/argocd/app-databases.yaml</code></summary>

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: postgres-databases
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://<your-forgejo-host>/<org>/<repo>.git
    targetRevision: main
    path: gitops/applications
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```
</details>

**Why three, not one:** each has a genuinely different failure mode and update cadence. CNPG changes rarely (a version bump you do deliberately). The operator changes on every code push. Database CRs change whenever a customer wants a new database or a config tweak. Bundling them into one Application means a bad CNPG sync and a routine database creation would show up in ArgoCD as the same blob of status — separating them keeps the failure signal meaningful.

`selfHeal: true` on all three is what makes this GitOps rather than "apply once": if someone hand-edits a live resource (or your own operator's reconciler creates a child object that then drifts — shouldn't happen given your idempotency work, but this is the safety net if it ever does), ArgoCD reverts it back to match Git.

---

## Step 7 — Bootstrap the Applications

These three YAML files live in Git, but ArgoCD doesn't know about *itself* yet — apply them once, directly:

```bash
kubectl apply -f gitops/argocd/app-cnpg.yaml
kubectl apply -f gitops/argocd/app-postgres-platform.yaml
kubectl apply -f gitops/argocd/app-databases.yaml
```

**✅ Verify:**

```bash
kubectl get applications -n argocd
```

```
NAME                          SYNC STATUS   HEALTH STATUS
cnpg-operator                 Synced        Healthy
postgres-platform-operator    Synced        Healthy
postgres-databases            Synced        Healthy
```

If `postgres-platform-operator` shows `OutOfSync` at this point, that's expected the very first time — it just means Step 3's overlay hasn't had a real image pushed to it yet (Step 8 does that). Everything else should be `Synced`/`Healthy` already.

---

## Step 8 — The CI Workflow: Build, Push, Bump Tag

**File: `.forgejo/workflows/operator-ci.yaml`** (new file, alongside your existing `platform-bootstrap.yaml`):

```yaml
name: Operator CI
on:
  push:
    branches:
      - main
    paths:
      - 'operator/**'

permissions:
  contents: write   # needed to push the image-tag bump commit back

jobs:
  build-and-release:
    name: Build, Push, Update GitOps
    runs-on: stackit-ubuntu-22
    defaults:
      run:
        working-directory: operator
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Confirm generated code is committed
        run: |
          make manifests generate
          git diff --exit-code -- . || \
            (echo "::error::CRD/deepcopy code is out of date — run 'make manifests generate' locally and commit the result." && exit 1)

      - name: Build and vet
        run: |
          go build ./...
          go vet ./...

      - name: Install Podman
        run: sudo apt-get update && sudo apt-get install -y podman

      - name: Log in to Docker Hub
        run: |
          echo "${{ secrets.DOCKERHUB_TOKEN }}" | \
            podman login docker.io -u "${{ secrets.DOCKERHUB_USERNAME }}" --password-stdin

      - name: Compute image tag
        id: vars
        run: echo "tag=$(git rev-parse --short HEAD)" >> "$GITHUB_OUTPUT"

      - name: Build and push operator image
        env:
          CONTAINER_TOOL: podman
          IMG: docker.io/${{ secrets.DOCKERHUB_USERNAME }}/postgres-platform-operator:${{ steps.vars.outputs.tag }}
        run: |
          make docker-build
          make docker-push

      - name: Bump image tag in gitops overlay
        run: |
          sed -i "s#newTag: .*#newTag: ${{ steps.vars.outputs.tag }}#" \
            ../gitops/postgres-platform/kustomization.yaml

      - name: Commit and push updated image tag
        run: |
          git config user.name "forgejo-actions"
          git config user.email "actions@noreply.local"
          git add ../gitops/postgres-platform/kustomization.yaml
          git commit -m "chore: bump postgres-platform-operator image to ${{ steps.vars.outputs.tag }}" \
            || (echo "No changes to commit" && exit 0)
          git push origin HEAD:main
```

A few things worth understanding rather than just copying:

- **`paths: ['operator/**']`** — this workflow only fires when operator source actually changes, not on every push to `main` (which would otherwise include the bot's own tag-bump commits, Terraform changes, README edits...).
- **The push at the end just works without extra token wiring** because `actions/checkout@v4` persists Forgejo's built-in Actions credentials by default, and you granted `permissions: contents: write` at the top of the workflow. No separate PAT needed for this part.
- **Commits made with this built-in token never re-trigger workflows** — this is a deliberate Forgejo Actions guarantee, specifically to prevent exactly the infinite-loop scenario a naive CI-writes-to-Git setup risks. Between that guarantee and the `paths` filter above, you have two independent reasons this won't loop.
- **The `git diff --exit-code` check** on generated code is a cheap correctness gate: if someone edited `api/v1/postgresql_types.go` (Part 3/Step 12 territory from the build guide) and forgot to run `make manifests generate` before committing, this fails the build loudly instead of silently shipping a stale CRD.

> **One thing to verify on your specific runner, not assume:** whether `stackit-ubuntu-22` supports rootless Podman builds out of the box. Some containerized CI runners need `--storage-driver=vfs` or extra privileges for nested container builds. Push a trivial change first and watch the "Build and push operator image" step — if it fails on storage/overlay errors rather than anything in your code, that's a runner-configuration problem, not a workflow-logic one, and worth isolating before debugging anything else.

---

## Step 9 — Run It End to End

```bash
# Trivial, verifiable change to the operator
echo "// ci trigger test" >> operator/internal/controller/postgresql_controller.go
git add operator/internal/controller/postgresql_controller.go
git commit -m "test: trigger operator CI"
git push origin main
```

**Watch it happen, in order:**

```bash
# 1. Forgejo Actions run — watch in the Forgejo UI, Actions tab
#    (build → push to Docker Hub → commit tag bump)

# 2. Confirm the bot commit landed
git log --oneline -3
# should show your commit, then "chore: bump postgres-platform-operator image to <sha>"

# 3. Confirm the image tag actually changed in Git
grep newTag gitops/postgres-platform/kustomization.yaml

# 4. Watch ArgoCD pick it up
kubectl get application postgres-platform-operator -n argocd -w
# SYNC STATUS flips OutOfSync → Synced

# 5. Confirm the running Pod is on the new image
kubectl get deployment -n postgres-platform-system \
  postgres-platform-controller-manager \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

That last command's output should end in the same short SHA you saw in step 2/3 above — the full loop, from a code push to a running Pod, with no manual `kubectl apply` anywhere in between.

---

## Optional Refinement: App-of-Apps

Right now the three `Application` objects in `gitops/argocd/` were applied manually once (Step 7) and aren't themselves managed by ArgoCD — if you add a fourth Application later, you'd `kubectl apply` it by hand again. The standard fix is a **root Application** that watches `gitops/argocd/` itself:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: app-of-apps
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://<your-forgejo-host>/<org>/<repo>.git
    targetRevision: main
    path: gitops/argocd
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

Apply this one manually, once — from then on, adding a new `Application` YAML file to `gitops/argocd/` and pushing it is enough; the root Application syncs it in automatically. Not necessary for the three you have now, but worth doing the moment you're adding Applications often enough that manual `kubectl apply` starts feeling like a chore rather than a one-time bootstrap step.