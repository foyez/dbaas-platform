Yes. This is the point where I would establish the authentication architecture **before adding more pages**.

For your Vue SPA, I recommend:

* **Access token:** short-lived, kept **only in memory**
* **Refresh token:** long-lived, stored in an **`HttpOnly` + `Secure` cookie**
* **Refresh token rotation:** every refresh invalidates the old refresh token
* **No access/refresh tokens in `localStorage` or `sessionStorage`**
* **Router guards:** protected pages require authentication
* **Auth pages:** `/login`, `/register`, etc. redirect authenticated users away
* **API client:** automatically refreshes once on `401`, then retries the original request
* **Logout:** server revokes the refresh-token session and clears the cookie
* **CSRF protection:** required because the refresh token is cookie-based

This aligns with current OAuth security guidance: refresh tokens for browser/public clients should use rotation or sender-constraining, and access tokens should have restricted scope/audience. ([datatracker.ietf.org][1]) OWASP specifically recommends **not putting authentication tokens in localStorage/sessionStorage** because JavaScript can access them if an XSS vulnerability exists. ([OWASP Cheat Sheet Series][2])

---

# 1. The architecture I recommend

Your application should look like this:

```text
                         Browser
                            │
             ┌──────────────┴──────────────┐
             │                             │
       Access Token                 Refresh Token
          memory                         cookie
             │                     HttpOnly + Secure
             │                             │
             ▼                             ▼
       Authorization                  POST /auth/refresh
       Bearer token                         │
             │                              │
             └──────────────┬───────────────┘
                            │
                            ▼
                         Backend
```

The important part:

```text
localStorage
    ❌ access token
    ❌ refresh token

sessionStorage
    ❌ access token
    ❌ refresh token

JavaScript memory
    ✅ access token

HttpOnly cookie
    ✅ refresh token
```

OWASP explicitly warns against storing authentication tokens, JWTs, refresh tokens, or other credentials in browser storage APIs. ([OWASP Cheat Sheet Series][3])

---

# 2. Recommended token lifetimes

For your PaaS, I'd start with:

### Access token

```text
10–15 minutes
```

I'd choose:

```text
15 minutes
```

The access token is deliberately short-lived.

If somebody steals it, the useful lifetime is limited.

### Refresh token

Don't think of this as simply:

```text
refresh token = 30 days
```

I'd implement a **refresh-token session**.

For example:

```text
Absolute lifetime: 30 days
Idle lifetime:      7 days
Rotation:           every refresh
```

So:

```text
Login
  │
  ▼
Refresh session created
  │
  ├── idle timeout: 7 days
  └── absolute timeout: 30 days
```

Every successful refresh:

```text
old refresh token
       │
       ▼
     invalid
       │
       ▼
new refresh token
```

RFC 9700 recommends refresh-token rotation for public clients and recommends refresh tokens expire after a period of inactivity. ([datatracker.ietf.org][1])

You can later adjust these values based on your threat model.

---

# 3. Don't store the refresh token as a database plaintext value

This is an important backend detail.

Suppose the browser has:

```text
refresh_token = abc123...
```

Your database should **not** contain:

```text
refresh_token = abc123...
```

Instead:

```text
refresh_token_hash = SHA256(refresh_token)
```

or preferably a keyed/appropriate server-side hashing strategy depending on your token design.

Your DB record might look like:

```text
refresh_sessions

id
user_id
token_hash
family_id
expires_at
last_used_at
revoked_at
created_at
replaced_by
```

For example:

```text
id:            rt_123
user_id:       user_456
family_id:     family_789
token_hash:    ...
expires_at:    2026-09-10
last_used_at:  2026-08-10
revoked_at:    null
replaced_by:   rt_124
```

The `family_id` becomes useful for detecting refresh-token reuse.

---

# 4. Refresh-token rotation

This is one of the most important parts.

Imagine:

```text
Login

RT1
 │
 ▼
POST /auth/refresh
 │
 ├── RT1 invalidated
 │
 └── RT2 issued
```

Then:

```text
RT2
 │
 ▼
POST /auth/refresh
 │
 ├── RT2 invalidated
 │
 └── RT3 issued
```

But suppose an attacker stole `RT1`.

They try:

```text
attacker
   │
   ▼
POST /auth/refresh
RT1
```

Your backend sees:

```text
RT1 = already used
```

That is a **refresh-token reuse event**.

At that point, revoke the entire token family:

```text
RT1 ❌
RT2 ❌
RT3 ❌
```

and require a fresh login.

RFC 9700 specifically describes this rotation/reuse-detection approach. ([datatracker.ietf.org][1])

---

# 5. Your backend endpoints

I would establish these endpoints:

```text
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

Eventually:

```text
POST /api/v1/auth/register
POST /api/v1/auth/forgot-password
POST /api/v1/auth/reset-password
```

But for the SPA architecture, the first four are the important ones.

---

# 6. Login response

Don't return the refresh token to Vue.

Your login response should look something like:

```http
POST /api/v1/auth/login
```

Response:

```json
{
  "accessToken": "eyJ...",
  "expiresIn": 900,
  "user": {
    "id": "usr-123",
    "email": "user@example.com",
    "name": "John"
  }
}
```

And simultaneously:

```http
Set-Cookie: __Host-refresh=...;
  HttpOnly;
  Secure;
  SameSite=Strict;
  Path=/;
```

The browser stores the refresh token automatically.

Vue never sees it.

`HttpOnly` prevents JavaScript from reading the cookie, while `Secure` ensures it is sent only over HTTPS. OWASP recommends these cookie protections for session credentials. ([OWASP Cheat Sheet Series][3])

---

# 7. Use a `__Host-` cookie

For your production deployment, I'd use:

```text
__Host-refresh
```

rather than:

```text
refresh_token
```

For example:

```http
Set-Cookie: __Host-refresh=abc...;
  Path=/;
  HttpOnly;
  Secure;
  SameSite=Strict
```

The `__Host-` prefix gives you additional browser-enforced restrictions: it must use `Secure`, `Path=/`, and cannot specify a `Domain`. OWASP recommends this pattern for session identifiers. ([OWASP Cheat Sheet Series][3])

---

# 8. What about localhost?

During development you can use:

```text
http://localhost:5173
```

and:

```text
http://localhost:8080
```

You need to configure CORS correctly.

For example:

```text
Vue
http://localhost:5173

        │
        │ credentials
        ▼

API
http://localhost:8080
```

Your fetch requests need:

```ts
credentials: "include"
```

And the API must explicitly allow:

```text
http://localhost:5173
```

rather than:

```text
Access-Control-Allow-Origin: *
```

when credentials are involved.

OWASP recommends explicit trusted origins rather than blindly allowing all origins. ([OWASP Cheat Sheet Series][2])

---

# 9. Create an auth type

Create:

```text
src/types/auth.ts
```

```ts
export interface User {
  id: string
  email: string
  name: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  accessToken: string
  expiresIn: number
  user: User
}

export interface RefreshResponse {
  accessToken: string
  expiresIn: number
}

export interface MeResponse {
  user: User
}
```

---

# 10. Create an auth API module

Create:

```text
src/api/auth.ts
```

```ts
import { apiFetch } from "./client"

import type {
  LoginRequest,
  LoginResponse,
  RefreshResponse,
  MeResponse,
} from "@/types/auth"

export function login(
  request: LoginRequest,
): Promise<LoginResponse> {
  return apiFetch<LoginResponse>(
    "/auth/login",
    {
      method: "POST",
      body: JSON.stringify(request),
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    },
  )
}

export function refresh(): Promise<RefreshResponse> {
  return apiFetch<RefreshResponse>(
    "/auth/refresh",
    {
      method: "POST",
      credentials: "include",
    },
  )
}

export function logout(): Promise<void> {
  return apiFetch<void>(
    "/auth/logout",
    {
      method: "POST",
      credentials: "include",
    },
  )
}

export function getCurrentUser(): Promise<MeResponse> {
  return apiFetch<MeResponse>(
    "/auth/me",
    {
      method: "GET",
      credentials: "include",
    },
  )
}
```

---

# 11. Keep the access token in memory

Create:

```text
src/stores/auth.ts
```

You can use Pinia here.

Unlike the instance pages, **I actually recommend Pinia for authentication**.

Authentication state is truly global.

Install:

```bash
npm install pinia
```

Then:

```ts
import { defineStore } from "pinia"

import {
  getCurrentUser,
  login as loginApi,
  logout as logoutApi,
  refresh as refreshApi,
} from "@/api/auth"

import type {
  LoginRequest,
  User,
} from "@/types/auth"

interface AuthState {
  accessToken: string | null
  user: User | null
  isInitialized: boolean
  isLoading: boolean
}
```

Then:

```ts
export const useAuthStore = defineStore(
  "auth",
  {
    state: (): AuthState => ({
      accessToken: null,
      user: null,
      isInitialized: false,
      isLoading: false,
    }),

    getters: {
      isAuthenticated: state =>
        state.accessToken !== null &&
        state.user !== null,
    },

    actions: {
      async login(
        credentials: LoginRequest,
      ) {
        this.isLoading = true

        try {
          const response =
            await loginApi(credentials)

          this.accessToken =
            response.accessToken

          this.user =
            response.user
        } finally {
          this.isLoading = false
        }
      },

      async initialize() {
        if (this.isInitialized) {
          return
        }

        try {
          const response =
            await refreshApi()

          this.accessToken =
            response.accessToken

          const me =
            await getCurrentUser()

          this.user = me.user
        } catch {
          this.accessToken = null
          this.user = null
        } finally {
          this.isInitialized = true
        }
      },

      async refreshAccessToken() {
        const response =
          await refreshApi()

        this.accessToken =
          response.accessToken

        return response.accessToken
      },

      async logout() {
        try {
          await logoutApi()
        } finally {
          this.accessToken = null
          this.user = null
        }
      },

      clearAuth() {
        this.accessToken = null
        this.user = null
      },
    },
  },
)
```

### Important

Do **not** add:

```ts
persist: true
```

to this store.

You don't want:

```text
Pinia
  ↓
localStorage
  ↓
access token
```

The access token should disappear when the JavaScript application context disappears.

---

# 12. Why does refresh work after browser reload?

This is the nice part.

Imagine:

```text
User logs in

access token
    ↓
memory

refresh token
    ↓
HttpOnly cookie
```

Then the user presses:

```text
F5
```

The access token disappears.

That's okay.

Your Vue app starts:

```text
App startup
     │
     ▼
auth.initialize()
     │
     ▼
POST /auth/refresh
     │
     │ browser automatically sends
     │ HttpOnly refresh cookie
     ▼
Backend
     │
     ▼
new access token
     │
     ▼
memory
```

The user remains logged in.

This is exactly why the refresh token exists.

---

# 13. Create the API client properly

Your current `apiFetch()` should become the central authentication mechanism.

Create/update:

```text
src/api/client.ts
```

Conceptually:

```ts
import { useAuthStore } from "@/stores/auth"

const API_URL =
  import.meta.env.VITE_API_URL

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message)
    this.name = "ApiError"
  }
}
```

Then:

```ts
interface ApiFetchOptions
  extends RequestInit {
  skipAuth?: boolean
}
```

And the first request:

```ts
export async function apiFetch<T>(
  path: string,
  options: ApiFetchOptions = {},
): Promise<T> {
  const auth =
    useAuthStore()

  const headers =
    new Headers(options.headers)

  headers.set(
    "Content-Type",
    "application/json",
  )

  if (
    auth.accessToken &&
    !options.skipAuth
  ) {
    headers.set(
      "Authorization",
      `Bearer ${auth.accessToken}`,
    )
  }

  const response =
    await fetch(
      `${API_URL}${path}`,
      {
        ...options,
        headers,
        credentials: "include",
      },
    )

  // ...
}
```

The critical part is:

```ts
Authorization: Bearer <access-token>
```

Your regular API endpoints should **not** use the refresh cookie as their authentication mechanism.

---

# 14. Automatically refresh after 401

This is the behavior you want:

```text
GET /instances
       │
       ▼
401 Unauthorized
       │
       ▼
POST /auth/refresh
       │
       ├── success
       │      │
       │      ▼
       │   new access token
       │      │
       │      ▼
       │   retry GET /instances
       │
       └── failure
              │
              ▼
           logout
              │
              ▼
           /login
```

But there's an important concurrency problem.

Imagine your dashboard loads:

```text
GET /instances
GET /auth/me
GET /something
```

at the same time.

All three receive:

```text
401
```

You don't want:

```text
POST /auth/refresh
POST /auth/refresh
POST /auth/refresh
```

because refresh-token rotation can cause those requests to race.

Instead, use a **single refresh promise**.

---

# 15. Implement refresh locking

In `client.ts`:

```ts
let refreshPromise:
  Promise<string> | null = null
```

Then:

```ts
async function refreshAccessToken(): Promise<string> {
  if (!refreshPromise) {
    const auth =
      useAuthStore()

    refreshPromise =
      auth.refreshAccessToken()
        .finally(() => {
          refreshPromise = null
        })
  }

  return refreshPromise
}
```

Now:

```text
Request A ──┐
Request B ──┼──► refreshPromise
Request C ──┘
```

Only one refresh request reaches the backend.

---

# 16. Retry exactly once

Your `apiFetch()` should roughly work like this:

```ts
export async function apiFetch<T>(
  path: string,
  options: ApiFetchOptions = {},
): Promise<T> {
  const auth =
    useAuthStore()

  const response =
    await request(
      path,
      options,
      auth.accessToken,
    )

  if (
    response.status !== 401 ||
    options.skipAuth
  ) {
    return handleResponse<T>(
      response,
    )
  }

  try {
    const accessToken =
      await refreshAccessToken()

    const retryResponse =
      await request(
        path,
        options,
        accessToken,
      )

    return handleResponse<T>(
      retryResponse,
    )
  } catch {
    auth.clearAuth()

    throw new ApiError(
      401,
      "UNAUTHORIZED",
      "Your session has expired.",
    )
  }
}
```

The important rule is:

> **Never retry infinitely.**

One refresh attempt.

One original-request retry.

Then fail.

---

# 17. Don't let `/auth/refresh` recursively refresh

This is why `skipAuth` exists.

For example:

```ts
refresh()
```

must not trigger:

```text
401
 ↓
refresh
 ↓
401
 ↓
refresh
 ↓
...
```

So:

```ts
export function refresh() {
  return apiFetch<RefreshResponse>(
    "/auth/refresh",
    {
      method: "POST",
      skipAuth: true,
    },
  )
}
```

Same for login.

---

# 18. Initialize auth before routing

This is crucial for your protected pages.

In:

```text
src/main.ts
```

you should initialize authentication **before mounting the router application**.

For example:

```ts
import {
  createApp,
} from "vue"

import {
  createPinia,
} from "pinia"

import App from "./App.vue"
import router from "./router"

import {
  useAuthStore,
} from "@/stores/auth"

import "./assets/index.css"

async function bootstrap() {
  const app =
    createApp(App)

  const pinia =
    createPinia()

  app.use(pinia)

  const auth =
    useAuthStore(pinia)

  await auth.initialize()

  app.use(router)

  await router.isReady()

  app.mount("#app")
}

bootstrap()
```

This avoids the common problem where:

```text
Browser loads
     │
     ▼
Router sees user as logged out
     │
     ▼
redirect /login
     │
     ▼
refresh completes
     │
     ▼
Oops, user was actually logged in
```

Instead:

```text
Browser
   │
   ▼
initialize auth
   │
   ▼
refresh cookie
   │
   ▼
authenticated?
   │
   ▼
start router
```

---

# 19. Add route metadata

This is how I would define protected and auth-only routes.

```ts
declare module "vue-router" {
  interface RouteMeta {
    requiresAuth?: boolean
    guestOnly?: boolean
  }
}
```

Then:

```ts
const router = createRouter({
  history: createWebHistory(),

  routes: [
    {
      path: "/login",
      name: "login",
      component: LoginPage,
      meta: {
        guestOnly: true,
      },
    },

    {
      path: "/register",
      name: "register",
      component: RegisterPage,
      meta: {
        guestOnly: true,
      },
    },

    {
      path: "/",
      component: AppLayout,

      children: [
        {
          path: "",
          name: "dashboard",
          component: DashboardPage,
          meta: {
            requiresAuth: true,
          },
        },

        {
          path: "instances",
          name: "instances",
          component: InstancesPage,
          meta: {
            requiresAuth: true,
          },
        },

        {
          path: "instances/new",
          name: "instance-create",
          component: CreateInstancePage,
          meta: {
            requiresAuth: true,
          },
        },

        {
          path: "instances/:id",
          name: "instance-detail",
          component: InstanceDetailPage,
          meta: {
            requiresAuth: true,
          },
        },
      ],
    },
  ],
})
```

Now the intent is explicit.

---

# 20. Add the authentication guard

After creating the router:

```ts
import {
  useAuthStore,
} from "@/stores/auth"
```

Then:

```ts
router.beforeEach(
  (to) => {
    const auth =
      useAuthStore()

    if (
      to.meta.requiresAuth &&
      !auth.isAuthenticated
    ) {
      return {
        name: "login",

        query: {
          redirect:
            to.fullPath,
        },
      }
    }

    if (
      to.meta.guestOnly &&
      auth.isAuthenticated
    ) {
      return {
        name: "dashboard",
      }
    }

    return true
  },
)
```

Now:

```text
/                  🔒
/instances         🔒
/instances/new     🔒
/instances/:id     🔒
```

require authentication.

And:

```text
/login             👤 guest only
/register          👤 guest only
```

redirect authenticated users to:

```text
/
```

---

# 21. Preserve the original URL

This part is particularly nice:

```text
User visits:

/instances/inst-123
```

while logged out.

Router sends them to:

```text
/login?redirect=/instances/inst-123
```

After successful login:

```text
/login
    │
    ▼
authentication
    │
    ▼
redirect=/instances/inst-123
    │
    ▼
/instances/inst-123
```

So they don't always get dumped on the dashboard.

Your login page can do:

```ts
const redirect =
  route.query.redirect

router.replace(
  typeof redirect === "string"
    ? redirect
    : {
        name: "dashboard",
      },
)
```

---

# 22. Add `App.vue`

Your `App.vue` can now be extremely simple:

```vue
<script setup lang="ts">
</script>

<template>
  <RouterView />
</template>
```

That's it.

The router determines whether the user gets:

```text
Auth pages
```

or:

```text
AppLayout
```

---

# 23. Your auth page structure

I'd now organize your pages as:

```text
src/pages/
├── auth/
│   ├── LoginPage.vue
│   ├── RegisterPage.vue
│   ├── ForgotPasswordPage.vue
│   └── ResetPasswordPage.vue
│
├── DashboardPage.vue
├── InstancesPage.vue
├── CreateInstancePage.vue
└── InstanceDetailPage.vue
```

And components:

```text
src/components/
├── auth/
│   ├── LoginForm.vue
│   └── RegisterForm.vue
│
├── dashboard/
│
├── instances/
│
├── layout/
│
└── ui/
```

---

# 24. Authentication flow

Your complete login flow becomes:

```text
                 ┌───────────────┐
                 │    /login     │
                 └───────┬───────┘
                         │
                         │ email/password
                         ▼
                 POST /auth/login
                         │
              ┌──────────┴──────────┐
              │                     │
              ▼                     ▼
        Access Token          HttpOnly Cookie
          memory              Refresh Token
              │                     │
              └──────────┬──────────┘
                         │
                         ▼
                     Dashboard
```

---

# 25. Browser refresh flow

```text
F5
 │
 ▼
Vue application starts
 │
 ▼
auth.initialize()
 │
 ▼
POST /auth/refresh
 │
 │ HttpOnly cookie automatically sent
 ▼
Backend validates refresh token
 │
 ├── valid
 │     │
 │     ├── rotate refresh token
 │     └── return access token
 │
 └── invalid
       │
       ▼
    logged out
```

The user never needs to know this happened.

---

# 26. Access-token expiration flow

Suppose the user is on:

```text
/instances
```

and the access token expires.

Then:

```text
GET /instances
       │
       ▼
401
       │
       ▼
refresh
       │
       ▼
new access token
       │
       ▼
retry GET /instances
       │
       ▼
200
```

The user sees nothing.

---

# 27. Actual logout flow

Logout should be:

```text
POST /auth/logout
```

Backend:

```text
refresh session
      │
      ▼
revoke
      │
      ▼
Set-Cookie: __Host-refresh=;
           Max-Age=0
```

Frontend:

```ts
await auth.logout()

router.replace({
  name: "login",
})
```

Don't just do:

```ts
auth.clearAuth()
```

because that only clears the access token from memory.

The refresh session also needs to be revoked server-side.

---

# 28. CSRF is important with this design

There is one tradeoff with putting the refresh token in a cookie.

Cookies are automatically attached to requests.

Therefore:

```text
HttpOnly
```

protects the cookie from JavaScript reading it, but it **doesn't by itself prevent CSRF**. OWASP explicitly notes that HttpOnly protects confidentiality, not CSRF. ([OWASP Cheat Sheet Series][3])

I'd use:

```text
SameSite=Strict
```

where your deployment permits it.

For additional defense, use a CSRF token for state-changing requests.

For example:

```text
X-CSRF-Token: <token>
```

The server verifies it for:

```text
POST
PUT
PATCH
DELETE
```

Your `GET` endpoints don't need CSRF tokens if they truly don't mutate state.

OWASP recommends CSRF defenses in addition to SameSite in deployments where SameSite alone isn't sufficient. ([OWASP Cheat Sheet Series][4])

---

# 29. Don't put the CSRF token in HttpOnly

There is a deliberate difference:

```text
Refresh token
    ↓
HttpOnly cookie
    ↓
JavaScript CANNOT read it
```

versus:

```text
CSRF token
    ↓
regular cookie or bootstrap response
    ↓
JavaScript CAN read it
    ↓
X-CSRF-Token header
```

That's okay because the CSRF token isn't the authentication credential.

A common pattern is:

```text
__Host-refresh
HttpOnly
Secure
SameSite=Strict

XSRF-TOKEN
Secure
SameSite=Strict
NOT HttpOnly
```

Then Vue sends:

```http
X-CSRF-Token: <value>
```

for state-changing operations.

---

# 30. Your final project structure

I'd now change your project toward:

```text
src/
├── api/
│   ├── client.ts
│   ├── errors.ts
│   ├── auth.ts
│   └── instances.ts
│
├── components/
│   ├── auth/
│   │   ├── LoginForm.vue
│   │   └── RegisterForm.vue
│   │
│   ├── dashboard/
│   │   ├── DashboardStats.vue
│   │   ├── RecentInstances.vue
│   │   └── GettingStarted.vue
│   │
│   ├── instances/
│   │   ├── InstanceForm.vue
│   │   ├── InstanceStatus.vue
│   │   └── InstanceListSkeleton.vue
│   │
│   ├── layout/
│   │   ├── AppHeader.vue
│   │   ├── AppLayout.vue
│   │   └── AppSidebar.vue
│   │
│   └── ui/
│       └── ...
│
├── composables/
│   ├── useInstance.ts
│   ├── useInstances.ts
│   └── useCreateInstance.ts
│
├── lib/
│   ├── date.ts
│   └── instance-status.ts
│
├── pages/
│   ├── auth/
│   │   ├── LoginPage.vue
│   │   └── RegisterPage.vue
│   │
│   ├── DashboardPage.vue
│   ├── InstancesPage.vue
│   ├── CreateInstancePage.vue
│   └── InstanceDetailPage.vue
│
├── stores/
│   └── auth.ts
│
├── types/
│   ├── auth.ts
│   └── instance.ts
│
├── router/
│   └── index.ts
│
├── App.vue
└── main.ts
```

---

# 31. The most important security decisions

I'd write these down as project rules:

| Credential/data        | Where                        | Lifetime               |
| ---------------------- | ---------------------------- | ---------------------- |
| Access token           | JS memory / Pinia            | ~15 min                |
| Refresh token          | `HttpOnly` cookie            | 30d absolute / 7d idle |
| Refresh token DB value | hashed                       | session lifetime       |
| CSRF token             | non-HttpOnly cookie/header   | session                |
| User info              | Pinia memory                 | current session        |
| Password               | **never stored client-side** | —                      |

And:

```text
❌ localStorage access token
❌ localStorage refresh token
❌ sessionStorage access token
❌ sessionStorage refresh token
❌ refresh token in JSON response
❌ refresh token in URL
❌ refresh token in JavaScript
❌ long-lived access token
❌ infinite 401 retry
❌ multiple simultaneous refresh requests
```

RFC 9700 also recommends keeping access-token privileges restricted and audience-specific where possible, which is worth applying to your API tokens. ([datatracker.ietf.org][1])

---

# 32. One thing I'd change in your overall architecture

Earlier I recommended **not using Pinia yet** because your instance data didn't need global state.

I would make an exception for authentication:

```text
Instances
    → composables

Authentication
    → Pinia
```

That's because authentication is genuinely application-wide:

```text
Dashboard
Instances
Instance detail
Create instance
Header
Sidebar
Login redirect
Logout
API client
```

all need to know whether the user is authenticated.

So:

```text
Pinia
└── auth

Composables
├── useInstances
├── useInstance
└── useCreateInstance
```

is a clean boundary.

---

## The resulting request architecture

Ultimately, every protected API request follows:

```text
Vue component
      │
      ▼
composable
      │
      ▼
api function
      │
      ▼
apiFetch()
      │
      ├── access token available?
      │          │
      │          ▼
      │       Authorization
      │          │
      │          ▼
      │       API request
      │
      └── 401?
            │
            ▼
       /auth/refresh
            │
            ├── success → rotate refresh token
            │              ↓
            │         new access token
            │              ↓
            │         retry request
            │
            └── failure → clear auth
                           ↓
                         /login
```

That is the architecture I'd establish **now**, before adding more authenticated PaaS features.

For the backend, the most important pieces are **refresh-token rotation/reuse detection, hashed refresh-token storage, secure cookie configuration, CSRF protection, and revocation on logout/password change**. For the Vue side, the most important pieces are **in-memory access tokens, one centralized API client, a single-flight refresh mechanism, and router guards**. ([datatracker.ietf.org][1])

[1]: https://datatracker.ietf.org/doc/html/rfc9700?utm_source=chatgpt.com "RFC 9700 - Best Current Practice for OAuth 2.0 Security"
[2]: https://cheatsheetseries.owasp.org/cheatsheets/HTML5_Security_Cheat_Sheet.html?utm_source=chatgpt.com "HTML5 Security - OWASP Cheat Sheet Series"
[3]: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html?utm_source=chatgpt.com "Session Management - OWASP Cheat Sheet Series"
[4]: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html?utm_source=chatgpt.com "Cross-Site Request Forgery Prevention - OWASP Cheat Sheet Series"
