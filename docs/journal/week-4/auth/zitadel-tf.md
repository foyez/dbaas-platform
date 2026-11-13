# Provisioning Zitadel using Terraform

**The Go API should be a separate `zitadel_application_api`, but you don't use its client secret in the Vue app.** The Vue app requests an audience scope for the API's **project**, and the Go API validates the resulting access token. ZITADEL explicitly documents this frontend → backend pattern. ([Terraform Registry][1])

For your setup, I'd use the following.

## 1. Terraform layout

```text
terraform/
├── versions.tf
├── provider.tf
├── variables.tf
├── project.tf
├── roles.tf
├── applications.tf
├── admin.tf
├── registration.tf
└── outputs.tf
```

---

## 2. `versions.tf`

```hcl
terraform {
  required_version = ">= 1.6.0"

  required_providers {
    zitadel = {
      source  = "zitadel/zitadel"
      version = "~> 3.3"
    }
  }
}
```

The current provider release is 3.3.0. ([Terraform Registry][1])

---

## 3. `provider.tf`

```hcl
provider "zitadel" {
  domain = var.zitadel_domain

  jwt_profile_file = var.zitadel_jwt_profile_file
}
```

I'd use a dedicated Terraform/service-account credential rather than your personal ZITADEL credentials.

---

## 4. `variables.tf`

```hcl
variable "zitadel_domain" {
  type        = string
  description = "ZITADEL instance domain"
}

variable "zitadel_org_id" {
  type        = string
  description = "ZITADEL organization ID"
}

variable "zitadel_jwt_profile_file" {
  type        = string
  sensitive   = true
  description = "Terraform service-account JWT profile"
}

variable "admin_user_id" {
  type        = string
  description = "ZITADEL user ID of the initial administrator"
}

variable "frontend_url" {
  type        = string
  description = "Public URL of the Vue frontend"
}

variable "dev_mode" {
  type        = bool
  default     = false
}
```

For development:

```hcl
frontend_url = "http://localhost:5173"
dev_mode     = true
```

Production:

```hcl
frontend_url = "https://app.example.com"
dev_mode     = false
```

---

# 5. Project

### `project.tf`

```hcl
resource "zitadel_project" "paas" {
  org_id = var.zitadel_org_id

  name = "My PaaS"

  project_role_assertion = true
  project_role_check     = true
}
```

`project_role_assertion = true` means ZITADEL can assert the project's roles in tokens. The current provider supports both `project_role_assertion` and `project_role_check`. ([Terraform Registry][2])

---

# 6. Roles

### `roles.tf`

```hcl
resource "zitadel_project_role" "admin" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id

  role_key     = "paas.admin"
  display_name = "Administrator"
  group        = "Application"
}

resource "zitadel_project_role" "user" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id

  role_key     = "paas.user"
  display_name = "User"
  group        = "Application"
}
```

Those attributes match the current provider schema. ([Terraform Registry][3])

---

# 7. Vue application

### `applications.tf`

```hcl
resource "zitadel_application_oidc" "frontend" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id

  name = "my-paas-web"

  # Browser SPA
  app_type         = "OIDC_APP_TYPE_USER_AGENT"
  auth_method_type = "OIDC_AUTH_METHOD_TYPE_NONE"

  # Authorization Code + PKCE
  response_types = [
    "OIDC_RESPONSE_TYPE_CODE"
  ]

  grant_types = [
    "OIDC_GRANT_TYPE_AUTHORIZATION_CODE",
    "OIDC_GRANT_TYPE_REFRESH_TOKEN"
  ]

  redirect_uris = [
    "${var.frontend_url}/auth/callback"
  ]

  post_logout_redirect_uris = [
    var.frontend_url
  ]

  access_token_type = "OIDC_TOKEN_TYPE_JWT"

  # Put project roles into the access token.
  access_token_role_assertion = true

  # Useful if your Vue app needs roles from the ID token.
  id_token_role_assertion = true

  dev_mode = var.dev_mode
}
```

The current provider supports `OIDC_APP_TYPE_USER_AGENT`, `OIDC_AUTH_METHOD_TYPE_NONE`, authorization-code/refresh-token grants, JWT access tokens, and role assertions. ([Terraform Registry][4])

---

# 8. Go API application

Now the important piece.

```hcl
resource "zitadel_application_api" "backend" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id

  name = "my-paas-api"

  auth_method_type = "API_AUTH_METHOD_TYPE_PRIVATE_KEY_JWT"
}
```

The current `zitadel_application_api` resource supports `BASIC` and `PRIVATE_KEY_JWT` authentication. ([Terraform Registry][5])

### What is this application for?

It represents:

```text
my-paas-api
```

as a protected API/resource in ZITADEL.

Your Go API should validate that incoming access tokens are intended for it.

**But the API's client secret is not something you put into Vue.**

---

# 9. The audience

This is the part that's easy to miss.

Your Vue application authenticates the user, but the resulting token needs to say:

```text
"I'm intended for the PaaS API."
```

ZITADEL does this using an audience scope:

```text
urn:zitadel:iam:org:project:id:<API_PROJECT_ID>:aud
```

ZITADEL explicitly documents this for frontend → backend communication. ([ZITADEL][6])

In **your current design**, the Vue app and Go API are in the **same project**, so you can use:

```text
urn:zitadel:iam:org:project:id:<PAAS_PROJECT_ID>:aud
```

The frontend therefore requests:

```text
openid
profile
email
urn:zitadel:iam:org:project:id:<PAAS_PROJECT_ID>:aud
```

---

# 10. Output the IDs Terraform needs

### `outputs.tf`

```hcl
output "paas_project_id" {
  value = zitadel_project.paas.id
}

output "frontend_client_id" {
  value     = zitadel_application_oidc.frontend.client_id
  sensitive = true
}

output "api_client_id" {
  value     = zitadel_application_api.backend.client_id
  sensitive = true
}

output "api_audience_scope" {
  value = "urn:zitadel:iam:org:project:id:${zitadel_project.paas.id}:aud"
}
```

After:

```bash
terraform apply
```

you'll get something conceptually like:

```text
paas_project_id     = "123456789..."
frontend_client_id  = "234567890..."
api_client_id       = "345678901..."

api_audience_scope  =
urn:zitadel:iam:org:project:id:123456789...:aud
```

---

# 11. Give yourself admin

### `admin.tf`

```hcl
resource "zitadel_user_grant" "admin" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id
  user_id    = var.admin_user_id

  role_keys = [
    zitadel_project_role.admin.role_key
  ]
}
```

The current `zitadel_user_grant` resource supports `project_id`, `user_id`, and `role_keys`. ([Terraform Registry][7])

So:

```text
YOU
 │
 └── paas.admin
```

---

# 12.Authentication & Registration for `paas.user`

### `identity_providers.tf`

Add Google authentication:

```hcl
resource "zitadel_org_idp_google" "google" {
  org_id = var.zitadel_org_id

  name          = "Google"
  client_id     = var.google_client_id
  client_secret = var.google_client_secret

  scopes = [
    "openid",
    "profile",
    "email",
  ]

  is_creation_allowed = true
  is_linking_allowed  = true
  is_auto_creation    = true
  is_auto_update      = true
}
```

This allows users to register/login through Google.

### `login_policy.tf`

Add the login and registration configuration:

```hcl
resource "zitadel_login_policy" "paas" {
  org_id = var.zitadel_org_id

  user_login         = true
  allow_register     = true
  allow_external_idp = true

  idps = [
    zitadel_org_idp_google.google.id
  ]

  force_mfa            = false
  force_mfa_local_only = false

  passwordless_type = "PASSWORDLESS_TYPE_ALLOWED"

  hide_password_reset = false

  password_check_lifetime       = "240h0m0s"
  external_login_check_lifetime = "240h0m0s"
  multi_factor_check_lifetime   = "24h0m0s"
  mfa_init_skip_lifetime        = "720h0m0s"

  ignore_unknown_usernames = true

  default_redirect_uri = var.frontend_url
}
```

The important settings are:

```hcl
allow_register     = true
allow_external_idp = true
idps               = [zitadel_org_idp_google.google.id]
```

So users can either:

```text
Email/password registration
        OR
Google registration
```

### `registration.tf`

Keep the default-role Action from the previous tutorial, but make sure it is configured for **Post Creation**:

```hcl
resource "zitadel_action" "assign_default_user_role" {
  org_id = var.zitadel_org_id

  name            = "assign-default-paas-user-role"
  timeout         = "10s"
  allowed_to_fail = false

  script = <<-EOT
    function addDefaultUserRole(ctx, api) {
      api.userGrants.push({
        projectID: "${zitadel_project.paas.id}",
        roles: ["${zitadel_project_role.user.role_key}"]
      });
    }
  EOT
}

resource "zitadel_trigger_actions" "user_registration" {
  org_id = var.zitadel_org_id

  flow_type    = "FLOW_TYPE_INTERNAL_AUTHENTICATION"
  trigger_type = "TRIGGER_TYPE_POST_CREATION"

  action_ids = [
    zitadel_action.assign_default_user_role.id
  ]
}
```

The resulting flow is:

```text
Normal registration
       │
       ▼
    ZITADEL
       │
       ▼
 User created
       │
       ▼
 paas.user assigned
```

For Google:

```text
Continue with Google
       │
       ▼
 Google authentication
       │
       ▼
    ZITADEL
       │
       ▼
 User created
       │
       ▼
 paas.user assigned
```

### Update `variables.tf`

**Create a Google Cloud project**

Open the [Google Cloud Console](https://console.cloud.google.com/?utm_source=chatgpt.com) and create/select a project for your PaaS.

Then go to:

**Google Auth Platform → Clients → Create Client → Web application**. Google documents this flow here. ([Google for Developers][1])

Add:

```hcl
variable "google_client_id" {
  type      = string
  sensitive = true
}

variable "google_client_secret" {
  type      = string
  sensitive = true
}
```

## Create a Google Cloud project

Open [Google Cloud Console](https://console.cloud.google.com/?utm_source=chatgpt.com) and create/select a project for your PaaS.

Then go to:

**Google Auth Platform → Clients → Create Client → Web application**

---

### Get the ZITADEL callback URI

Before creating the Google OAuth client, get the callback URL from ZITADEL.

In ZITADEL:

```text
Organization
  → Settings
  → Identity Providers
  → Google
```

Open the **Google provider configuration/template**.

ZITADEL displays a **Callback URL** there. Copy it. ZITADEL's current documentation specifically instructs you to copy this URL and use it in Google Cloud. ([ZITADEL][1])

For your instance, it will look like:

```text
https://<your-zitadel-domain>/idps/callback
```

For example:

```text
https://dbass-k4ho6i.eu1.zitadel.cloud/idps/callback
```

> **Important:** This is the callback from **Google → ZITADEL**. It is **not** your Vue frontend callback.

---

### Add the callback URI to Google

In [Google Cloud Console](https://console.cloud.google.com/?utm_source=chatgpt.com):

```text
Google Cloud Project
  → Google Auth Platform
  → Clients
  → Create Client
  → Web application
```

Under **Authorized redirect URIs**, add the ZITADEL callback:

```text
https://dbass-k4ho6i.eu1.zitadel.cloud/idps/callback
```

Google must have this exact URI registered. ZITADEL's Google setup documentation specifies using the ZITADEL Callback URL as Google's authorized redirect URI.

So the configuration is:

```text
Google Cloud
└── OAuth Client
    └── Authorized redirect URI
        └── https://dbass-k4ho6i.eu1.zitadel.cloud/idps/callback
```

---

### Get Client ID and Client Secret

After creating the OAuth client, Google gives you:

```text
Client ID
Client Secret
```

Keep the secret out of Git.

Put them into your CI/CD secrets as:

```text
TF_VAR_google_client_id
TF_VAR_google_client_secret
```

The overall flow is:

```text
Your Vue app
     │
     ▼
  ZITADEL
     │
     ▼
  Google
     │
     │  redirect_uri
     ▼
ZITADEL /idps/callback
     │
     ▼
  ZITADEL
     │
     ▼
 Your Vue app
```


**Don't add your Vue URL (`https://app.foyez...`) as Google's redirect URI.** That URL belongs to your **ZITADEL application**, while `/idps/callback` belongs to the **Google → ZITADEL** integration. ZITADEL's hosted-login documentation confirms that the current login flow uses `${CUSTOM_DOMAIN}/idps/callback`. ([ZITADEL][2])

[1]: https://zitadel.com/docs/guides/integrate/identity-providers/google?utm_source=chatgpt.com "Configure Google as an Identity Provider in ZITADEL | ZITADEL Docs"
[2]: https://zitadel.com/docs/guides/integrate/login/hosted-login?utm_source=chatgpt.com "Hosted Login UI: Authenticate Your Users | ZITADEL Docs"

---

# 13. Final ZITADEL structure

After Terraform:

```text
My PaaS
│
├── Roles
│   ├── paas.admin
│   └── paas.user
│
├── Applications
│   │
│   ├── my-paas-web
│   │     └── OIDC / SPA / PKCE
│   │
│   └── my-paas-api
│         └── API
│
├── Users
│   │
│   ├── You
│   │    └── paas.admin
│   │
│   └── New registrations
│        └── paas.user
│
└── Actions
    └── Post Creation
         └── assign paas.user
```

---

# 14. Vue configuration

Your Vue OIDC configuration should request the API audience.

For example:

```ts
const oidcConfig = {
  authority: `https://${import.meta.env.VITE_ZITADEL_DOMAIN}`,

  client_id: import.meta.env.VITE_ZITADEL_CLIENT_ID,

  redirect_uri: `${window.location.origin}/auth/callback`,

  post_logout_redirect_uri: window.location.origin,

  response_type: "code",

  scope: [
    "openid",
    "profile",
    "email",

    // Go API audience
    `urn:zitadel:iam:org:project:id:${import.meta.env.VITE_ZITADEL_PROJECT_ID}:aud`,
  ].join(" "),
};
```

So your `.env` is:

```env
VITE_ZITADEL_DOMAIN=my-instance.zitadel.cloud
VITE_ZITADEL_CLIENT_ID=...
VITE_ZITADEL_PROJECT_ID=...
```

The audience scope is the critical part for allowing the token to be used by your Go backend. ([ZITADEL][11])

---

# 15. Go API

Your request flow becomes:

```text
Vue
 │
 │ Authorization: Bearer eyJ...
 ▼
Go API
 │
 ├── validate signature
 ├── validate issuer
 ├── validate expiry
 ├── validate audience
 │
 └── validate role
       │
       ├── paas.user  → normal endpoints
       │
       └── paas.admin → admin endpoints
```

ZITADEL's official Go API example uses its Go SDK authorization middleware and shows role-protected endpoints. It also notes that the API must be in the token's audience. ([ZITADEL][12])

You can use:

```bash
go get github.com/zitadel/zitadel-go/v3
```

and use ZITADEL's authorization middleware rather than implementing JWT verification from scratch. ([ZITADEL][12])

---

## 16. One security detail I'd strongly recommend

**Don't treat the audience as authorization.**

A token can request an audience scope, so your Go API should check **both**:

```text
audience == my-paas-api/project
            AND
role == paas.user or paas.admin
```

ZITADEL's current security guidance explicitly warns that audience checking alone is insufficient; role/grant authorization must also be enforced. ([ZITADEL Knowledge Base][13])

So:

```go
if !validAudience(token) {
    return http.StatusUnauthorized
}

if !hasRole(token, "paas.user") &&
   !hasRole(token, "paas.admin") {
    return http.StatusForbidden
}
```

And for admin endpoints:

```go
if !hasRole(token, "paas.admin") {
    return http.StatusForbidden
}
```

---

## 17. One architectural simplification

Because your **Vue and Go API are part of the same PaaS**, I'd actually keep them in the **same ZITADEL project** for now:

```text
My PaaS
│
├── my-paas-web
└── my-paas-api
```

That makes role management straightforward:

```text
paas.user
paas.admin
```

If later you split your system into multiple APIs/projects, you can introduce separate audiences/projects.

So the final architecture I'd use is:

```text
                         ZITADEL
                            │
             ┌──────────────┴──────────────┐
             │                             │
        my-paas-web                  my-paas-api
        OIDC SPA                    API resource
             │                             │
             │        access token         │
             └──────────────┬──────────────┘
                            │
                           Vue
                            │
                    Authorization: Bearer
                            │
                            ▼
                          Go API
                            │
                   ┌────────┴────────┐
                   │                 │
              paas.user        paas.admin
                   │                 │
                   └────────┬────────┘
                            │
                           CNPG
```

This gives you **registration, login, password management, roles, OIDC, API audience validation, and Terraform-managed IAM**, without putting authentication logic or passwords into your Go/CNPG application.

One caveat: if you're using **ZITADEL Cloud vs. self-hosted ZITADEL**, or if your current Terraform project already exists, I would adapt the organization/provider authentication pieces rather than blindly creating a new organization. The resources above are based on the current v3.3.0 provider schema. ([Terraform Registry][1])

[1]: https://registry.terraform.io/providers/zitadel/zitadel/latest/docs?utm_source=chatgpt.com "Docs overview | zitadel/zitadel | Terraform | Terraform Registry"
[2]: https://registry.terraform.io/providers/zitadel/zitadel/latest/docs/resources/project?utm_source=chatgpt.com "zitadel_project | Resources | zitadel/zitadel | Terraform | Terraform Registry"
[3]: https://registry.terraform.io/providers/zitadel/zitadel/latest/docs/resources/project_role?utm_source=chatgpt.com "zitadel_project_role | Resources | zitadel/zitadel | Terraform | Terraform Registry"
[4]: https://registry.terraform.io/providers/zitadel/zitadel/latest/docs/resources/application_oidc?utm_source=chatgpt.com "zitadel_application_oidc | Resources | zitadel/zitadel | Terraform | Terraform Registry"
[5]: https://registry.terraform.io/providers/zitadel/zitadel/latest/docs/resources/application_api?utm_source=chatgpt.com "zitadel_application_api | Resources | zitadel/zitadel | Terraform | Terraform Registry"
[6]: https://zitadel.com/docs/apis/openidoauth/scopes?utm_source=chatgpt.com "Scopes in ZITADEL | ZITADEL Docs"
[7]: https://registry.terraform.io/providers/zitadel/zitadel/latest/docs/resources/user_grant?utm_source=chatgpt.com "zitadel_user_grant | Resources | zitadel/zitadel | Terraform | Terraform Registry"
[8]: https://registry.terraform.io/providers/zitadel/zitadel/latest/docs/resources/action?utm_source=chatgpt.com "zitadel_action | Resources | zitadel/zitadel | Terraform | Terraform Registry"
[9]: https://zitadel.com/docs/apis/actions/internal-authentication?utm_source=chatgpt.com "Internal Authentication Flow | ZITADEL Docs"
[10]: https://registry.terraform.io/providers/zitadel/zitadel/latest/docs/resources/trigger_actions?utm_source=chatgpt.com "zitadel_trigger_actions | Resources | zitadel/zitadel | Terraform | Terraform Registry"
[11]: https://zitadel.com/docs/guides/solution-scenarios/frontend-calling-backend-API?utm_source=chatgpt.com "Frontend and Back-end API Communication in ZITADEL | ZITADEL Docs"
[12]: https://zitadel.com/docs/examples/secure-api/go?utm_source=chatgpt.com "Secure Go API Application with ZITADEL | ZITADEL Docs"
[13]: https://help.zitadel.com/security-best-practices-validating-audience-aud-claims-in-zitadel-access-tokens?utm_source=chatgpt.com "Security Best Practices: Validating Audience (aud) Claims in ZITADEL Access Tokens"
