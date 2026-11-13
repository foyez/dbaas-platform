Yes. With **ZITADEL**, I would change the architecture from the custom `/auth/login` + `/auth/refresh` design I described earlier.

For a Vue SPA, ZITADEL's recommended approach is **OIDC Authorization Code + PKCE**, and ZITADEL's Vue example uses `vue-oidc-context` on top of `oidc-client-ts`. ZITADEL also supports `offline_access` to obtain a refresh token. ([ZITADEL][1])

The important distinction is:

> **Your Vue app should not implement the OAuth login/refresh protocol itself. Let `oidc-client-ts` manage it.**

Your PaaS API then only needs to validate the ZITADEL access token.

---

# 1. Final architecture

I recommend this:

```text
                         ┌─────────────────┐
                         │    ZITADEL      │
                         │                 │
                         │ OIDC + PKCE     │
                         └────────┬────────┘
                                  │
                           Authorization
                              Code + PKCE
                                  │
                                  ▼
┌─────────────────────────────────────────────────┐
│                  Vue SPA                        │
│                                                 │
│  vue-oidc-context                               │
│       │                                         │
│       ├── access token                          │
│       ├── refresh token                         │
│       ├── token expiry                          │
│       └── authentication state                  │
│                                                 │
│             │                                   │
│             │ Authorization: Bearer ...         │
│             ▼                                   │
│          apiFetch()                             │
└──────────────┬──────────────────────────────────┘
               │
               ▼
       ┌─────────────────┐
       │   PaaS API      │
       │                 │
       │ /instances      │
       │ /instances/:id  │
       └────────┬────────┘
                │
                │ validate JWT
                ▼
            ZITADEL
```

Your frontend does **not** need:

```text
❌ /auth/login
❌ /auth/refresh
❌ /auth/logout backend endpoints
❌ custom refresh-token database
❌ refresh-token rotation implementation
❌ refresh-token hashing
❌ custom OAuth protocol implementation
```

ZITADEL handles those OAuth concerns.

ZITADEL's token endpoint supports the refresh-token grant and returns a new access token and refresh token; the `offline_access` scope is what requests a refresh token. ([ZITADEL][2])

---

# 2. Install the OIDC libraries

For your Vue 3 + Vite application:

```bash
npm install oidc-client-ts @dlukt/vue-oidc-context
```

This is also the library combination used in ZITADEL's official Vue example. ([ZITADEL][1])

You can continue using:

```text
Vue
TypeScript
Pinia
Vue Router
Tailwind
shadcn-vue
Zod
TanStack Vue Query
Playwright
Vitest
```

But authentication itself should be delegated to:

```text
vue-oidc-context
        ↓
oidc-client-ts
        ↓
ZITADEL
```

---

# 3. Configure ZITADEL

## Create a ZITADEL Project

In the ZITADEL Console, create something like:

```
Project: dbaas
```

Put both applications in this project.

This is useful because applications in the same project share roles and role assignments.

## Create the Vue.js application

Go to:

```
Project
  → Applications
    → New
```

Choose:

```
Application type: User Agent
Authentication method: PKCE
```

Do not use a client secret in Vue.js.

Your Vue application runs in the user's browser, so anything embedded in the JavaScript bundle must be considered public. ZITADEL specifically recommends Authorization Code + PKCE for SPA/User Agent applications.

Redirect URI

For development:

```
http://localhost:5173/callback
```

Post logout redirect URI

For example:

```
http://localhost:5173/
```

You can later add your production URLs, e.g.:

```
https://app.example.com/callback
https://app.example.com/
```

ZITADEL allows multiple redirect URIs.

You'll get a Client ID for the Vue application.

Conceptually:

```
ZITADEL
└── Project: dbaas
    └── Application: dbaas Frontend
        Type: User Agent
        Auth: PKCE
        Client ID: xxx
```

## Create the Go API application

This is the part that's easy to misunderstand.

Your Go API is not a second login application.

Create:

```
Project
  → Applications
    → New
```

and choose:

Application type: API

The API application represents your resource server/audience. ZITADEL describes API applications as applications accessed by other applications/services, and they can be used as the target of access tokens.

For example:

```
ZITADEL
└── Project: dbaas
    ├── dbaas Frontend
    │   Type: User Agent
    │   Auth: PKCE
    │
    └── dbaas API
        Type: API
```

The important thing is that your Vue app obtains an access token intended for your Go API, and your Go API validates that token.

---

# 4. Configure your scopes

For your PaaS I would request:

```text
openid
profile
email
offline_access
```

And importantly, request the audience/scope needed for your API.

ZITADEL documents `offline_access` as the scope that requests a refresh token. ([ZITADEL][4])

Your scope configuration might therefore look conceptually like:

```ts
scope:
  "openid profile email offline_access ..."
```

The final API audience/scope depends on how you configure your ZITADEL project/API.

---

# 5. Environment variables

Create:

```text
.env.development
```

```env
VITE_ZITADEL_AUTHORITY=https://your-instance.zitadel.cloud
VITE_ZITADEL_CLIENT_ID=your-client-id
VITE_ZITADEL_API_URL=http://localhost:8080/api/v1
```

And:

```text
.env.production
```

```env
VITE_ZITADEL_AUTHORITY=https://auth.example.com
VITE_ZITADEL_CLIENT_ID=your-production-client-id
VITE_ZITADEL_API_URL=https://api.example.com/api/v1
```

**Never put a ZITADEL client secret here.**

A browser SPA is a public client.

PKCE means the frontend does not need a client secret. ZITADEL's documentation specifically describes PKCE for applications running entirely in the browser. ([ZITADEL][1])

---

# 6. Create the OIDC plugin

Create:

```text
src/auth/oidc.ts
```

```ts
import {
  createOidcAuth,
} from "@dlukt/vue-oidc-context"

export const oidcAuth =
  createOidcAuth({
    authority:
      import.meta.env
        .VITE_ZITADEL_AUTHORITY,

    client_id:
      import.meta.env
        .VITE_ZITADEL_CLIENT_ID,

    redirect_uri:
      `${window.location.origin}/auth/callback`,

    post_logout_redirect_uri:
      window.location.origin,

    response_type: "code",

    scope:
      "openid profile email offline_access",

    automaticSilentRenew: true,
  })
```

The exact initialization API can vary slightly with the installed `vue-oidc-context` version, so use the current package API when wiring this file. The important configuration is:

```text
Authorization Code
+
PKCE
+
offline_access
+
automatic renewal
```

ZITADEL's official Vue integration uses this OIDC approach rather than a custom token implementation. ([ZITADEL][1])

---

# 7. Register it in `main.ts`

Your `main.ts` becomes roughly:

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
  oidcAuth,
} from "@/auth/oidc"

import "./assets/index.css"

const app =
  createApp(App)

const pinia =
  createPinia()

app.use(pinia)

app.use(
  oidcAuth,
)

app.use(router)

app.mount("#app")
```

---

# 8. Don't create your own auth store

This is another change from my previous recommendation.

Previously I suggested:

```text
stores/
└── auth.ts
```

With ZITADEL + `vue-oidc-context`, **don't duplicate authentication state in Pinia**.

You don't want:

```text
ZITADEL
   │
   ▼
vue-oidc-context
   │
   ▼
Pinia auth store
   │
   ▼
another copy of tokens
```

That creates synchronization problems.

Instead:

```text
ZITADEL
   │
   ▼
oidc-client-ts
   │
   ▼
vue-oidc-context
   │
   ├── authenticated
   ├── user
   ├── access token
   └── token lifecycle
```

Pinia can still be used for other genuinely global application state.

---

# 9. Token storage

There's an important nuance here.

ZITADEL's official Vue integration uses `oidc-client-ts`, which manages the OIDC session and token lifecycle. ZITADEL's documentation describes browser-side token storage for SPAs. ([ZITADEL][1])

For your application, I would configure the OIDC client with a storage strategy appropriate to your threat model, and **avoid putting tokens into application-managed `localStorage` yourself**.

In particular, don't write:

```ts
localStorage.setItem(
  "access_token",
  accessToken,
)
```

and don't create your own:

```ts
localStorage.setItem(
  "refresh_token",
  refreshToken,
)
```

The OIDC library should own the token lifecycle.

---

# 10. Your API client becomes much simpler

This is the biggest benefit.

Your PaaS API client does not need to know how ZITADEL refreshes tokens.

It just asks the OIDC client for the current access token.

Create:

```text
src/api/client.ts
```

```ts
import { oidcAuth } from "@/auth/oidc"

const API_URL =
  import.meta.env.VITE_ZITADEL_API_URL

export class ApiError
  extends Error {
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
export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const accessToken =
    await oidcAuth.userManager
      .getUser()
      .then(user =>
        user?.access_token,
      )

  const headers =
    new Headers(options.headers)

  headers.set(
    "Accept",
    "application/json",
  )

  if (
    options.body &&
    !headers.has("Content-Type")
  ) {
    headers.set(
      "Content-Type",
      "application/json",
    )
  }

  if (accessToken) {
    headers.set(
      "Authorization",
      `Bearer ${accessToken}`,
    )
  }

  const response =
    await fetch(
      `${API_URL}${path}`,
      {
        ...options,
        headers,
      },
    )

  if (!response.ok) {
    throw await parseApiError(
      response,
    )
  }

  if (
    response.status === 204
  ) {
    return undefined as T
  }

  return response.json()
}
```

But I'd make one further improvement.

---

# 11. Let the OIDC library handle token renewal

Don't implement:

```text
401
 ↓
POST /auth/refresh
```

yourself.

That's exactly the OAuth protocol work that `oidc-client-ts` is there to handle.

ZITADEL's own Vue example states that `oidc-client-ts` handles authentication state and token refresh with offline access. ([ZITADEL][1])

Your frontend should therefore be:

```text
Vue
 │
 ▼
oidc-client-ts
 │
 ├── access token
 ├── refresh token
 └── renewal
      │
      ▼
   ZITADEL
```

Then:

```text
apiFetch()
   │
   ▼
get valid access token
   │
   ▼
Authorization: Bearer ...
```

---

# 12. Create an API token helper

I'd make a small abstraction:

```text
src/auth/access-token.ts
```

```ts
import { oidcAuth } from "./oidc"

export async function getAccessToken() {
  const user =
    await oidcAuth.userManager
      .getUser()

  if (!user) {
    return null
  }

  return user.access_token
}
```

Then your API client becomes:

```ts
import { getAccessToken } from "@/auth/access-token"

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token =
    await getAccessToken()

  const headers =
    new Headers(options.headers)

  headers.set(
    "Accept",
    "application/json",
  )

  if (token) {
    headers.set(
      "Authorization",
      `Bearer ${token}`,
    )
  }

  const response =
    await fetch(
      `${API_URL}${path}`,
      {
        ...options,
        headers,
      },
    )

  if (!response.ok) {
    throw await parseApiError(
      response,
    )
  }

  return response.json()
}
```

---

# 13. Your router becomes very clean

You still need route protection.

I'd keep the metadata approach.

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
const routes = [
  {
    path: "/login",
    name: "login",
    component: LoginPage,
    meta: {
      guestOnly: true,
    },
  },

  {
    path: "/auth/callback",
    name: "auth-callback",
    component: AuthCallbackPage,
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
]
```

---

# 14. Router guard with ZITADEL

Instead of asking Pinia:

```ts
auth.isAuthenticated
```

ask the OIDC context.

Conceptually:

```ts
router.beforeEach(
  async (to) => {
    const isAuthenticated =
      oidcAuth.isAuthenticated

    if (
      to.meta.requiresAuth &&
      !isAuthenticated
    ) {
      await oidcAuth.signIn()

      return false
    }

    if (
      to.meta.guestOnly &&
      isAuthenticated
    ) {
      return {
        name: "dashboard",
      }
    }

    return true
  },
)
```

However, I actually prefer an even cleaner UX for ZITADEL.

---

# 15. Protected route → redirect directly to ZITADEL

You don't necessarily need to make `/login` a real login form.

Because ZITADEL is your identity provider.

Your:

```text
/login
```

can simply contain:

```text
┌─────────────────────────────┐
│                             │
│       Your PaaS             │
│                             │
│   [ Continue with login ]   │
│                             │
└─────────────────────────────┘
```

Clicking it:

```ts
await oidcAuth.signIn()
```

takes the user to:

```text
ZITADEL
```

They authenticate there.

Then:

```text
ZITADEL
   │
   ▼
/auth/callback
   │
   ▼
Vue
   │
   ▼
/
```

ZITADEL explicitly describes this flow: the app redirects the user to the hosted ZITADEL login, then receives an authorization code and exchanges it for tokens. ([ZITADEL][5])

---

# 16. Your callback page

Create:

```text
src/pages/auth/AuthCallbackPage.vue
```

```vue
<script setup lang="ts">
import {
  onMounted,
} from "vue"

import {
  useRouter,
} from "vue-router"

import { oidcAuth } from "@/auth/oidc"

const router =
  useRouter()

onMounted(
  async () => {
    try {
      await oidcAuth.signInCallback()

      await router.replace({
        name: "dashboard",
      })
    } catch (error) {
      console.error(
        "Authentication failed",
        error,
      )

      await router.replace({
        name: "login",
      })
    }
  },
)
</script>

<template>
  <div
    class="flex min-h-screen items-center justify-center"
  >
    <div class="text-center">
      <div
        class="text-lg font-medium"
      >
        Signing you in…
      </div>

      <div
        class="text-muted-foreground mt-2 text-sm"
      >
        Please wait.
      </div>
    </div>
  </div>
</template>
```

Depending on the version/configuration of `vue-oidc-context`, the callback handling may instead be exposed through the composable/provider. Follow the installed package's API for that exact call.

---

# 17. Login page

Your login page becomes extremely simple.

```vue
<script setup lang="ts">
import {
  useRouter,
} from "vue-router"

import {
  Button,
} from "@/components/ui/button"

import { oidcAuth } from "@/auth/oidc"

const router =
  useRouter()

async function login() {
  await oidcAuth.signIn()
}
</script>

<template>
  <div
    class="flex min-h-screen items-center justify-center bg-muted/40 px-4"
  >
    <div
      class="w-full max-w-md rounded-xl border bg-background p-8 shadow-sm"
    >
      <div class="mb-8">
        <h1
          class="text-2xl font-semibold"
        >
          Welcome back
        </h1>

        <p
          class="mt-2 text-sm text-muted-foreground"
        >
          Sign in to manage your PostgreSQL
          instances.
        </p>
      </div>

      <Button
        class="w-full"
        size="lg"
        @click="login"
      >
        Continue with ZITADEL
      </Button>
    </div>
  </div>
</template>
```

No password input.

No email input.

No password validation.

No login API.

ZITADEL owns all of that.

---

# 18. Logout

Similarly, don't create:

```text
POST /api/v1/auth/logout
```

unless you have some separate application-specific reason.

Use the OIDC logout flow.

For example:

```ts
async function logout() {
  await oidcAuth.signOut()
}
```

The user gets logged out of your application and the ZITADEL session can be terminated according to the configured OIDC logout flow.

---

# 19. Your instance API doesn't change much

This is nice.

Your existing:

```text
src/api/instances.ts
```

can remain:

```ts
import { apiFetch } from "./client"

import type {
  Instance,
  CreateInstanceRequest,
} from "@/types/instance"

export function getInstances() {
  return apiFetch<Instance[]>(
    "/instances",
  )
}

export function getInstance(
  id: string,
) {
  return apiFetch<Instance>(
    `/instances/${id}`,
  )
}

export function createInstance(
  request: CreateInstanceRequest,
) {
  return apiFetch<Instance>(
    "/instances",
    {
      method: "POST",
      body: JSON.stringify(request),
    },
  )
}
```

Authentication is completely transparent.

---

# 20. Your component doesn't care about auth

For example:

```ts
const {
  data: instances,
  isLoading,
} = useInstances()
```

No:

```ts
const token = ...
```

No:

```ts
refreshToken()
```

No:

```ts
localStorage
```

No:

```ts
Authorization
```

Your components shouldn't know about any of that.

---

# 21. What your application now looks like

I'd structure it as:

```text
src/
│
├── api/
│   ├── client.ts
│   ├── errors.ts
│   └── instances.ts
│
├── auth/
│   ├── oidc.ts
│   └── access-token.ts
│
├── components/
│   ├── auth/
│   │   └── LoginButton.vue
│   │
│   ├── dashboard/
│   │
│   ├── instances/
│   │
│   ├── layout/
│   │
│   └── ui/
│
├── composables/
│   ├── useInstances.ts
│   ├── useInstance.ts
│   └── useCreateInstance.ts
│
├── pages/
│   ├── auth/
│   │   ├── LoginPage.vue
│   │   └── AuthCallbackPage.vue
│   │
│   ├── DashboardPage.vue
│   ├── InstancesPage.vue
│   ├── CreateInstancePage.vue
│   └── InstanceDetailPage.vue
│
├── router/
│   └── index.ts
│
├── types/
│   └── instance.ts
│
├── lib/
│   └── ...
│
├── App.vue
└── main.ts
```

Notice:

```text
stores/
```

doesn't contain authentication anymore.

---

# 22. One important thing for your PaaS API

Your API is:

```text
Vue
  │
  │ Bearer access token
  ▼
Go/API?
  │
  ▼
CNPG/Kubernetes
```

Your backend should **not trust the Vue application**.

It should validate the ZITADEL access token.

For every request:

```http
GET /api/v1/instances

Authorization: Bearer eyJ...
```

Your API verifies:

```text
issuer
audience
signature
expiration
scope
roles
```

Then:

```text
valid
  ↓
request.user
  ↓
handler
```

ZITADEL's frontend/backend integration documentation describes this architecture: the frontend sends the access token to the backend, and the backend validates/authorizes it. ([ZITADEL][6])

---

# 23. Make the API audience-specific

This is particularly important for your application.

Your frontend shouldn't simply request:

```text
openid profile email
```

and then throw that token at every backend.

Configure your ZITADEL project/API so the access token is intended for your PaaS API.

ZITADEL's documentation specifically calls out requesting the API project's audience scope so the backend can validate the token for that API. ([ZITADEL][6])

Conceptually:

```text
Scopes:

openid
profile
email
offline_access
<your-api-audience>
```

Then your backend checks the expected audience.

---

# 24. Roles are also useful for your PaaS

You're likely going to eventually have:

```text
owner
admin
developer
viewer
```

Don't implement this purely in Vue.

Vue can hide buttons:

```vue
<Button v-if="canCreateInstance">
  Create Instance
</Button>
```

but your API must enforce authorization.

For example:

```text
POST /instances
```

might require:

```text
instance:create
```

while:

```text
GET /instances
```

might require:

```text
instance:read
```

ZITADEL supports project roles and can assert role claims using scopes/configuration. ([ZITADEL][4])

---

# 25. The final authentication flow

This is what I'd use for your project:

```text
                     ┌──────────────┐
                     │   ZITADEL    │
                     └──────┬───────┘
                            │
                    OIDC Authorization
                    Code + PKCE
                            │
                            ▼
┌─────────────────────────────────────────────┐
│                  Vue SPA                    │
│                                             │
│        vue-oidc-context                     │
│              │                              │
│              ▼                              │
│         oidc-client-ts                      │
│              │                              │
│       ┌──────┴───────┐                      │
│       │              │                      │
│ access token    refresh token               │
│       │              │                      │
│       └──────┬───────┘                      │
│              │                              │
│              ▼                              │
│          apiFetch()                         │
│              │                              │
│              │ Authorization: Bearer        │
└──────────────┼──────────────────────────────┘
               │
               ▼
        ┌───────────────┐
        │   PaaS API    │
        │               │
        │ validate JWT  │
        │ check scopes  │
        │ check roles   │
        └───────┬───────┘
                │
                ▼
             CNPG
```

And your frontend route flow:

```text
                    Vue
                     │
                     ▼
                Vue Router
                     │
             ┌───────┴────────┐
             │                │
         requiresAuth      guestOnly
             │                │
             ▼                ▼
      authenticated?       logged in?
        │       │            │
       yes      no           yes
        │       │            │
        ▼       ▼            ▼
       Page    ZITADEL      Dashboard
                │
                ▼
          /auth/callback
                │
                ▼
            Dashboard
```

### One key correction from my previous answer

Because you're using ZITADEL, **do not implement the custom refresh-token architecture I previously described**. In particular, don't create your own refresh-token cookie, refresh endpoint, refresh-token database, or refresh-token rotation mechanism.

Use ZITADEL's OIDC implementation and `oidc-client-ts` to manage the OAuth lifecycle. ZITADEL's current Vue documentation explicitly recommends this integration, including PKCE and token refresh with offline access. ([ZITADEL][1])

Your application should focus on just three authentication responsibilities:

1. **Route protection** — Vue Router.
2. **Getting/maintaining the ZITADEL access token** — `vue-oidc-context` / `oidc-client-ts`.
3. **Sending the access token to your API** — your centralized `apiFetch()`.

Everything else belongs to ZITADEL.

[1]: https://zitadel.com/docs/sdk-examples/vue?utm_source=chatgpt.com "Vue.js | ZITADEL Docs"
[2]: https://zitadel.com/docs/apis/openidoauth/endpoints?utm_source=chatgpt.com "OpenID Connect Endpoints in ZITADEL | ZITADEL Docs"
[3]: https://zitadel.com/docs/guides/manage/console/applications-overview?utm_source=chatgpt.com "Applications | ZITADEL Docs"
[4]: https://zitadel.com/docs/apis/openidoauth/scopes?utm_source=chatgpt.com "Scopes in ZITADEL | ZITADEL Docs"
[5]: https://zitadel.com/docs/guides/start/quickstart?utm_source=chatgpt.com "The ZITADEL Quick Start Guide | ZITADEL Docs"
[6]: https://zitadel.com/docs/guides/solution-scenarios/frontend-calling-backend-API?utm_source=chatgpt.com "Frontend and Back-end API Communication in ZITADEL | ZITADEL Docs"
