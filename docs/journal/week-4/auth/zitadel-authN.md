# ZITADEL Authentication: Vue.js → Go API

This document describes the complete authentication setup for a Vue.js SPA and Go API using ZITADEL.

The application uses:

* **Vue.js** — frontend / SPA
* **ZITADEL** — identity provider and authentication server
* **Go + Gin** — backend API
* **OIDC Authorization Code Flow with PKCE** — user authentication
* **ZITADEL project roles** — application authorization

The authorization model is intentionally simple:

```text
admin → administrator
user  → normal authenticated user
```

The frontend never stores or handles user passwords. ZITADEL hosts the login and registration UI.

---

## 1. Create a ZITADEL account

Go to the ZITADEL Cloud website and create a ZITADEL account.

During onboarding:

1. Create your ZITADEL account.
2. Verify your email.
3. Create your ZITADEL instance.
4. Choose the instance name and region.
5. Create the initial **Instance Administrator**.

The Instance Administrator has administrative control over the ZITADEL instance. This is **not** the same thing as your application's `admin` role.

ZITADEL recommends separate instances for environments such as development, test, and production.

---

## 2. Create the ZITADEL Instance

A ZITADEL **Instance** is the isolated identity environment containing your organizations, users, projects, applications, policies, and authentication configuration.

For development, create something such as:

```text
Instance:
myapp-dev
```

ZITADEL will provide an issuer/domain similar to:

```text
https://myapp-dev-xxxxx.zitadel.cloud
```

This becomes the issuer used by both the frontend and backend.

Example:

```env
VITE_ZITADEL_DOMAIN=https://myapp-dev-xxxxx.zitadel.cloud
```

---

## 3. Create the Project

Inside the ZITADEL Management Console:

```text
Projects → Create Project
```

For example:

```text
Project name:
MyApp

Framework:
Vue
```

A project groups applications and their roles. This is useful because your Vue frontend and API can belong to the same project and share the application's role definitions and assignments.

Conceptually:

```text
ZITADEL Instance
└── MyApp Project
    ├── Vue SPA
    ├── Go API
    └── Roles
        ├── admin
        └── user
```

---

# 4. Create the Vue.js Application

Your Vue application is a browser-based SPA, so create a:

```text
Application type:
User Agent
```

Use:

```text
Authentication:
Authorization Code + PKCE
```

PKCE is the recommended authentication flow for browser-based applications because the frontend cannot safely keep a client secret. ZITADEL specifically recommends Authorization Code with PKCE for SPAs such as Vue applications.

In the ZITADEL project:

```text
Project
└── Applications
    └── New Application
```

Select:

```text
User Agent
PKCE
```

---

## 5. Configure Vue redirect URIs

For local development, assume Vue runs on:

```text
http://localhost:5173
```

Configure the application with:

```text
Redirect URI:
http://localhost:5173/auth/callback

Post Logout Redirect URI:
http://localhost:5173
```

The redirect URI is where ZITADEL sends the browser after successful authentication. ZITADEL validates that the redirect URI used by the application matches a registered URI.

For production, use HTTPS:

```text
https://app.example.com/auth/callback
https://app.example.com
```

Do not use HTTP in production.

For local development, ZITADEL's development mode can be used when testing HTTP URLs.

---

# 6. Save the Vue Client ID

After creating the application, copy its:

```text
Client ID
```

The Vue application needs:

```env
VITE_ZITADEL_DOMAIN=https://your-instance.zitadel.cloud
VITE_ZITADEL_CLIENT_ID=your-client-id
VITE_ZITADEL_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ZITADEL_POST_LOGOUT_URI=http://localhost:5173
```

Do **not** put a client secret in the Vue application.

A SPA is a public client. PKCE is used specifically so the browser does not need a client secret.

---

# 7. Enable user registration

If users should be able to create their own ZITADEL accounts, enable:

```text
Login Behavior and Security
    ↓
Register allowed
```

`Register allowed` enables the registration option on the ZITADEL login UI. It can allow local username/password registration and configured external identity providers.

For a simple local registration flow, make sure:

```text
Username Password allowed: Enabled
Register allowed:          Enabled
```

You can leave external identity providers disabled if you only want email/password registration.

The important point is that **registration happens at ZITADEL**, not in your Vue application.

---

# 8. How registration works

The Vue application does not implement:

```text
email
password
confirm password
```

itself.

Instead:

```text
Vue
 │
 │ Login/Register
 ▼
ZITADEL
 │
 ├── Login
 │
 └── Register
      │
      ▼
   User account
```

ZITADEL hosts the registration form and manages the user's authentication credentials.

By default, self-registered users are added to the default organization. You can target another organization using the appropriate organization scope if your application needs multi-organization behavior.

---

# 9. Create the application roles

Inside the project:

```text
Project
└── Roles
    ├── admin
    └── user
```

Create:

| Key     | Display Name  | Group |
| ------- | ------------- | ----- |
| `admin` | Administrator | empty |
| `user`  | User          | empty |

The `Key` is what the Go API checks:

```go
auth.RequireRole("admin")
```

---

# 10. Assign yourself the admin role

Go to:

```text
Project
→ Role Assignments
→ Create Role Assignment
```

Select your ZITADEL account.

Then select:

```text
admin
```

Do not select `user` for your account.

The resulting configuration is:

```text
Your ZITADEL account
        ↓
      admin
```

Normal users should receive:

```text
Normal user
     ↓
   user
```

You do not need a separate authorization database for this two-role model.

---

# 11. Return roles during authentication

Creating and assigning roles is not enough. The application must receive the role information during authentication.

Configure the project/application so that project roles are **asserted/returned during authentication**.

ZITADEL supports returning role information through OIDC authentication, including the UserInfo endpoint and ID/Access Tokens depending on the application's configuration.

For your Go API, the important requirement is that the authentication/introspection result used by the API contains the project roles.

After changing role assignments or authentication settings, obtain a **new access token** before testing. Existing tokens may not contain the newly assigned role information.

---

# 12. Vue.js authentication

For Vue 3, use an OIDC library rather than implementing OAuth manually.

ZITADEL's Vue example uses `vue-oidc-context`, which wraps `oidc-client-ts` and handles the Authorization Code + PKCE flow and session state.

Install:

```bash
npm install vue-oidc-context oidc-client-ts
```

Configure the OIDC client with:

```env
VITE_ZITADEL_DOMAIN=https://your-instance.zitadel.cloud
VITE_ZITADEL_CLIENT_ID=your-client-id
VITE_ZITADEL_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ZITADEL_POST_LOGOUT_URI=http://localhost:5173
```

Conceptually, the Vue configuration is:

```text
authority:
    https://your-instance.zitadel.cloud

client_id:
    <ZITADEL Vue application client ID>

redirect_uri:
    http://localhost:5173/auth/callback

post_logout_redirect_uri:
    http://localhost:5173

response_type:
    code

scope:
    openid profile email
```

The exact configuration depends on the OIDC library version being used.

---

# 13. Vue login flow

When the user clicks **Login**:

```text
Vue
 │
 │ Authorization request
 │ + PKCE challenge
 ▼
ZITADEL
 │
 │ Login UI
 ▼
User authenticates
 │
 ▼
ZITADEL
 │
 │ authorization code
 ▼
Vue callback
 │
 │ code + PKCE verifier
 ▼
ZITADEL token endpoint
 │
 ├── access token
 ├── ID token
 └── refresh token (when configured/requested)
```

The access token is then used when calling the Go API.

ZITADEL documents this Authorization Code + PKCE flow for SPAs and recommends PKCE for browser applications.

---

# 14. Vue calls the Go API

After authentication, the Vue application gets an access token.

For an API request:

```http
GET /api/v1/instances
Authorization: Bearer <access_token>
```

For example:

```js
const response = await fetch(
  'http://localhost:8080/api/v1/instances',
  {
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
  },
)
```

The frontend does not send the user's password to the Go API.

It sends the ZITADEL access token.

---

# 15. Go API configuration

The Go API needs its ZITADEL configuration:

```env
ZITADEL_DOMAIN=your-instance.zitadel.cloud
ZITADEL_KEY_PATH=./zitadel-api-application-key.json # not sa key
```

The `ZITADEL_KEY_PATH` is for the backend's authentication when it communicates with ZITADEL for authorization/token introspection.

Keep this key on the backend.

Never expose it through:

```text
VITE_*
```

or put it in the Vue application.

---

# 16. Initialize ZITADEL in Go

Initialize the ZITADEL middleware during application startup:

```go
auth, err := authmw.New(ctx, cfg.Zitadel)
if err != nil {
    return fmt.Errorf("initialize auth: %w", err)
}
```

Do this in the application's composition/root layer, not inside `router.New()`.

The router receives the already-initialized middleware:

```go
r := router.New(instanceHandler, auth)
```

---

# 17. Protect the Go API

Normal authenticated endpoints:

```go
v1 := r.Group("/api/v1")

protected := v1.Group("")
protected.Use(authmw.Gin(auth.RequireAuth()))

protected.GET("/instances", h.ListInstances)
protected.GET("/instances/:id", h.GetInstance)
protected.POST("/instances", h.CreateInstance)
protected.PATCH("/instances/:id", h.UpdateInstance)
```

Admin endpoints:

```go
admin := protected.Group("")
admin.Use(authmw.Gin(auth.RequireRole("admin")))

admin.DELETE("/instances/:id", h.DeleteInstance)
```

The responsibilities are:

```text
RequireAuth()
    ↓
Is the access token valid?
    ↓
Who is the user?
```

and:

```text
RequireRole("admin")
    ↓
Does the authenticated user have
the admin project role?
```

---

# 18. Complete authentication flow

The complete system looks like this:

```text
                    ZITADEL
                       │
              ┌────────┴─────────┐
              │                  │
           Instance            Project
              │                  │
              │          ┌───────┴────────┐
              │          │                │
              │       Vue App           Roles
              │       User Agent       ├─ admin
              │          PKCE          └─ user
              │
              ▼
        User Registration
              │
              ▼
        ZITADEL Login
              │
              ▼
        Authorization Code
              │
              ▼
        Vue + PKCE
              │
              ▼
        Access Token
              │
              │ Authorization: Bearer ...
              ▼
          Go API
              │
              ▼
        RequireAuth()
              │
              ├── invalid → 401
              │
              ▼
        RequireRole()
              │
              ├── forbidden → 403
              │
              ▼
           Handler
```

---

# 19. What each component owns

Keep the responsibilities separated:

```text
ZITADEL
├── User accounts
├── Passwords / passkeys / MFA
├── Login
├── Registration
├── Identity
└── Application roles

Vue.js
├── Login button
├── Logout button
├── OIDC callback
├── Session state
└── Sending access tokens to API

Go API
├── Validate/authenticate access tokens
├── Enforce roles
├── Identify the authenticated user
└── Application/business authorization

Application Database
├── Instances
├── Domain data
└── Resource ownership
```

The frontend should **not** authenticate users directly against your Go API with passwords.

The Go API should **not** store ZITADEL passwords or implement its own login system.

ZITADEL handles identity; your API consumes the resulting access token. ZITADEL's frontend/backend integration guidance follows this model: the SPA authenticates with Authorization Code + PKCE and then sends the access token to the protected backend API.

---

# 20. Local development configuration

### Vue

```env
VITE_ZITADEL_DOMAIN=https://your-instance.zitadel.cloud
VITE_ZITADEL_CLIENT_ID=<vue-client-id>
VITE_ZITADEL_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ZITADEL_POST_LOGOUT_URI=http://localhost:5173
```

### Go

```env
APP_ENV=development
SERVER_ADDRESS=:8080

ZITADEL_DOMAIN=your-instance.zitadel.cloud
ZITADEL_KEY_PATH=./zitadel-key.json
```

The Vue app runs:

```text
http://localhost:5173
```

The Go API runs:

```text
http://localhost:8080
```

The browser flow is:

```text
Vue
http://localhost:5173
        │
        ▼
ZITADEL
        │
        ▼
Vue callback
http://localhost:5173/auth/callback
        │
        ▼
Access token
        │
        ▼
Go API
http://localhost:8080/api/v1/...
```

---

# 21. Production

For production:

```text
Vue:
https://app.example.com

API:
https://api.example.com

ZITADEL:
https://login.example.com
```

Register the production redirect URIs in the ZITADEL application:

```text
https://app.example.com/auth/callback
https://app.example.com
```

Use HTTPS and do not use development HTTP settings in production.

For production, keep the ZITADEL backend key outside the source repository and inject it through your deployment/secret-management system.

---

## Final checklist

Before testing the complete flow, verify:

```text
ZITADEL
[ ] Account created
[ ] Instance created
[ ] Instance administrator created
[ ] Project created

Vue application
[ ] User Agent application created
[ ] Authorization Code + PKCE selected
[ ] Redirect URI configured
[ ] Post logout URI configured
[ ] Client ID copied

User registration
[ ] Username/password allowed
[ ] Register allowed
[ ] Registration organization configured

Authorization
[ ] admin role created
[ ] user role created
[ ] admin assigned to your account
[ ] user assigned to normal users
[ ] Assert/return roles during authentication enabled

Vue
[ ] OIDC library configured
[ ] Login implemented
[ ] Callback implemented
[ ] Logout implemented
[ ] Access token sent to API

Go API
[ ] ZITADEL domain configured
[ ] Backend key configured
[ ] authmw initialized at startup
[ ] RequireAuth() applied
[ ] RequireRole("admin") applied to admin routes
[ ] 401 handled for unauthenticated requests
[ ] 403 handled for unauthorized requests
```

The final responsibility split is:

```text
             ZITADEL
                │
        Authentication
                │
                ▼
             Vue.js
                │
          Access Token
                │
                ▼
             Go API
                │
        ┌───────┴────────┐
        │                │
   RequireAuth()   RequireRole("admin")
        │                │
        ▼                ▼
    Identity          Authorization
```

This keeps authentication centralized in ZITADEL, keeps the Vue application as a public OIDC client using PKCE, and keeps the Go API responsible for validating the authenticated identity and enforcing API authorization.
