# Creating Service Account and Assign roles

STACKIT's current documentation confirms that a service account has **no resource access by default**; you must assign it a role. For the cert-manager DNS-01 integration, STACKIT specifically requires the service account to have **DNS Admin** permissions because cert-manager needs to create and delete the temporary TXT record. ([Docs][1])

## 1. Open your STACKIT project

Go to the [STACKIT Portal](https://portal.stackit.cloud/?utm_source=chatgpt.com) and select the project containing:

```text
foyez.runs.onstackit.cloud
```

Make sure it is project:

```text
e7990770-226d-4f75-b9a7-e2269a8d2194
```

This is important because your Issuer currently tells the webhook to use exactly that project.

---

## 2. Create the service account

In the STACKIT Portal:

**IAM and Management → Service accounts**

Then:

1. Click **Create service account**
2. Give it a name, for example:

```text
cert-manager-dns
```

3. Click **Create**

STACKIT's current procedure is to create service accounts from **IAM and Management → Service accounts** within the selected project. ([Docs][2])

You should end up with something similar to:

```text
cert-manager-dns-xxxxxxxx@...
```

Save that service-account identity.

---

## 3. Assign DNS Admin to the service account

This is the part you're currently missing.

Stay in the **same project**.

Go to:

**IAM and Management → Access**

Then:

1. Click **Grant Access**
2. In **Subject**, search for the service account you just created.
3. Select it and click **+ Add**
4. Under **Role**, select:

```text
DNS Admin
```

5. Click **Save**

STACKIT documents exactly this workflow: **IAM and Management → Access → Grant Access → Subject → Add → Role → Save**. ([Docs][3])

Your final access should conceptually look like:

```text
Project: e7990770-226d-4f75-b9a7-e2269a8d2194

Subject:
  cert-manager-dns-xxxxxxxx@...

Role:
  DNS Admin
```

**This role should be assigned at the project containing the DNS zone.**

---

## 4. Create a service-account key

The service account also needs credentials that the cert-manager STACKIT webhook can use.

Go to:

**IAM and Management → Service accounts**

Select your new service account.

Then create a **service account key**.

STACKIT documents service-account keys separately from the service account itself; the key is what allows an application to authenticate as that service account. ([Docs][4])

Download/save the resulting key securely.

**Do not paste the key into this chat.**

---

## 5. Make sure cert-manager's STACKIT webhook uses that key

This is important in your case.

The STACKIT cert-manager integration documentation uses a Kubernetes Secret named:

```text
stackit-sa-authentication
```

containing the service-account key, and the webhook needs service-account authentication enabled. ([Docs][1])

For example, the documentation uses:

```bash
kubectl create secret generic stackit-sa-authentication \
  -n <namespace> \
  --from-literal=sa.json='<YOUR_SERVICE_ACCOUNT_KEY>'
```

The exact namespace/config depends on how you installed the STACKIT webhook.

So **don't blindly run that command yet**. First we should inspect your existing webhook installation, because you may already have a Secret and just need to replace its credentials.

---

## 6. Verify the role assignment

Back in STACKIT:

**IAM and Management → Access**

You should now see your service account listed with:

```text
DNS Admin
```

If it only shows the service account without a role, the 403 will continue.

STACKIT explicitly states that service accounts need assigned roles before they can be used. ([Docs][3])

---

## 7. Verify the DNS zone is in the same project

Your Issuer says:

```yaml
config:
  projectId: e7990770-226d-4f75-b9a7-e2269a8d2194
```

The STACKIT documentation also notes that the `projectId` in the Issuer needs to correspond to the project where the DNS zone was created. ([Docs][1])

So verify:

```text
STACKIT project
    e7990770-226d-4f75-b9a7-e2269a8d2194
          │
          └── DNS zone
              foyez.runs.onstackit.cloud
```

If the zone is in another project, change the `projectId` in your Issuer instead.

---

# For your specific problem

Your current state is:

```text
Service account
    ?
    │
    ├── DNS Admin?       ← need to verify
    │
    └── key configured?  ← need to verify
            │
            ▼
STACKIT cert-manager webhook
            │
            ▼
STACKIT DNS API
            │
            └── 403 Unauthorized ❌
```

I would **not change your Issuer** yet. The Issuer itself looks reasonable.

### Do these two things first

**In STACKIT:**

```text
IAM and Management
  → Service accounts
  → select your SA
  → create key

IAM and Management
  → Access
  → Grant Access
  → select your SA
  → DNS Admin
  → Save
```

Then, on Kubernetes, run:

```bash
kubectl get pods -A | grep -Ei 'stackit|webhook'
```

and:

```bash
kubectl get secret -A | grep -i stackit
```

**Paste those two outputs here.** Don't paste any Secret contents. From that I can give you the exact next commands to connect the new STACKIT service-account key to your existing webhook.

[1]: https://docs.stackit.cloud/products/network/core-networking/dns/tutorials/use-stackit-dns-for-dns01-to-act-as-a-dns01-acme-issuer-with-cert-manager/?utm_source=chatgpt.com "Use STACKIT DNS for DNS01 to act as a DNS01 ACME Issuer with Cert-Manager | Docs"
[2]: https://docs.stackit.cloud/platform/access-and-identity/service-accounts/how-tos/manage-service-accounts/?utm_source=chatgpt.com "Create, manage and delete service accounts | Docs"
[3]: https://docs.stackit.cloud/platform/access-and-identity/roles-permissions/assign-roles-to-account/?utm_source=chatgpt.com "Assign roles to a user or service account | Docs"
[4]: https://docs.stackit.cloud/platform/access-and-identity/service-accounts/how-tos/manage-service-account-keys/?utm_source=chatgpt.com "Create, manage and delete service account keys | Docs"
