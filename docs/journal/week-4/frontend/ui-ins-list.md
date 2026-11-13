Absolutely. For `GET /instances`, I'd add an **Instances list page** and keep the same architecture we used for the detail page:

```text
InstancesPage.vue
      │
      ▼
useInstances()
      │
      ▼
instancesApi.list()
      │
      ▼
GET /api/v1/instances
```

For a PaaS, I'd also design the list API so it can grow into pagination, filtering, and sorting without changing the frontend architecture.

I'll assume the response is currently:

```json
{
  "items": [
    {
      "id": "inst-7f3a2b",
      "name": "my-postgres-db",
      "version": 16,
      "storage": "10Gi",
      "status": "Running",
      "createdAt": "2026-08-04T09:12:00Z"
    }
  ]
}
```

If your backend currently returns a raw array, I'll show the small adjustment below.

---

# 1. Define the list response

You already have:

```text
src/types/instance.ts
```

I'd make the types explicit.

```ts
export const INSTANCE_STATUSES = [
  "Pending",
  "Provisioning",
  "Running",
  "Failed",
  "Deleting",
  "Deleted",
] as const

export type InstanceStatus =
  (typeof INSTANCE_STATUSES)[number]

export interface Instance {
  id: string
  name: string
  version: number
  storage: string
  status: InstanceStatus
  createdAt: string
}

export interface CreateInstanceRequest {
  name: string
  instances: number
  version: number
  storage: string
}

export interface InstanceListResponse {
  items: Instance[]
}
```

This gives you a nice separation between:

```text
Instance
```

and:

```text
InstanceListResponse
```

which will become important when you add pagination.

---

# 2. Add `listInstances()` to the API

Update:

```text
src/api/instances.ts
```

```ts
import { apiFetch } from "./client"

import type {
  CreateInstanceRequest,
  Instance,
  InstanceListResponse,
} from "@/types/instance"

export async function createInstance(
  request: CreateInstanceRequest,
  idempotencyKey: string,
): Promise<Instance> {
  return apiFetch<Instance>("/instances", {
    method: "POST",

    headers: {
      "Content-Type": "application/json",
      "Idempotency-Key": idempotencyKey,
    },

    body: JSON.stringify(request),
  })
}

export async function getInstance(
  id: string,
): Promise<Instance> {
  return apiFetch<Instance>(
    `/instances/${encodeURIComponent(id)}`,
  )
}

export async function listInstances(): Promise<InstanceListResponse> {
  return apiFetch<InstanceListResponse>(
    "/instances",
  )
}
```

The API layer remains extremely boring, which is exactly what you want.

---

# 3. Create `useInstances()`

Create:

```text
src/composables/useInstances.ts
```

```ts
import {
  ref,
  computed,
} from "vue"

import { listInstances } from "@/api/instances"
import { ApiError } from "@/api/errors"
import type { Instance } from "@/types/instance"

export function useInstances() {
  const instances = ref<Instance[]>([])
  const isLoading = ref(false)
  const error = ref<ApiError | null>(null)

  const isEmpty = computed(
    () =>
      !isLoading.value &&
      !error.value &&
      instances.value.length === 0,
  )

  async function fetchInstances() {
    isLoading.value = true
    error.value = null

    try {
      const response = await listInstances()

      instances.value = response.items

      return response
    } catch (err) {
      if (err instanceof ApiError) {
        error.value = err
      } else {
        error.value = new ApiError(
          0,
          "NETWORK_ERROR",
          "Unable to connect to the server.",
        )
      }

      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    instances,
    isLoading,
    error,
    isEmpty,
    fetchInstances,
  }
}
```

Notice that the page doesn't know how the HTTP request works.

It just says:

```ts
await fetchInstances()
```

That's the separation I'd keep throughout the application.

---

# 4. Create the list route

Your router now becomes:

```ts
import {
  createRouter,
  createWebHistory,
} from "vue-router"

import CreateInstancePage from "@/pages/CreateInstancePage.vue"
import InstanceDetailPage from "@/pages/InstanceDetailPage.vue"
import InstancesPage from "@/pages/InstancesPage.vue"

const router = createRouter({
  history: createWebHistory(),

  routes: [
    {
      path: "/instances",
      component: InstancesPage,
    },

    {
      path: "/instances/new",
      component: CreateInstancePage,
    },

    {
      path: "/instances/:id",
      component: InstanceDetailPage,
    },
  ],
})

export default router
```

So your resource navigation is now:

```text
/instances
    │
    ├── /instances/new
    │
    └── /instances/:id
```

---

# 5. Create the page

Create:

```text
src/pages/InstancesPage.vue
```

I'd keep the page fairly thin.

```vue
<script setup lang="ts">
import { onMounted } from "vue"
import { useRouter } from "vue-router"
import { Plus, RefreshCw } from "lucide-vue-next"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

import { Button } from "@/components/ui/button"

import InstanceStatus from "@/components/instances/InstanceStatus.vue"

import { useInstances } from "@/composables/useInstances"

import { formatDate } from "@/lib/date"

const router = useRouter()

const {
  instances,
  isLoading,
  error,
  isEmpty,
  fetchInstances,
} = useInstances()

function createInstance() {
  router.push("/instances/new")
}

function openInstance(id: string) {
  router.push(`/instances/${encodeURIComponent(id)}`)
}

async function load() {
  try {
    await fetchInstances()
  } catch {
    // Error is exposed by useInstances().
  }
}

onMounted(load)
</script>

<template>
  <main
    class="
      min-h-screen
      bg-muted/40
      px-4
      py-8
    "
  >
    <div class="mx-auto max-w-6xl space-y-6">

      <!-- Header -->

      <div
        class="
          flex
          flex-col
          gap-4
          sm:flex-row
          sm:items-center
          sm:justify-between
        "
      >
        <div>
          <h1
            class="
              text-2xl
              font-semibold
              tracking-tight
            "
          >
            Instances
          </h1>

          <p
            class="
              mt-1
              text-sm
              text-muted-foreground
            "
          >
            Manage your PostgreSQL instances.
          </p>
        </div>

        <Button @click="createInstance">
          <Plus class="mr-2 h-4 w-4" />

          Create instance
        </Button>
      </div>

      <!-- Loading -->

      <Card v-if="isLoading">
        <CardContent class="py-12">
          <div
            class="
              flex
              items-center
              justify-center
              text-sm
              text-muted-foreground
            "
          >
            Loading instances...
          </div>
        </CardContent>
      </Card>

      <!-- Error -->

      <Card v-else-if="error">
        <CardHeader>
          <CardTitle>
            Unable to load instances
          </CardTitle>

          <CardDescription>
            {{ error.message }}
          </CardDescription>
        </CardHeader>

        <CardContent>
          <Button
            variant="outline"
            @click="load"
          >
            <RefreshCw class="mr-2 h-4 w-4" />

            Try again
          </Button>
        </CardContent>
      </Card>

      <!-- Empty -->

      <Card v-else-if="isEmpty">
        <CardContent
          class="
            flex
            flex-col
            items-center
            justify-center
            py-16
            text-center
          "
        >
          <div
            class="
              flex
              h-12
              w-12
              items-center
              justify-center
              rounded-full
              bg-muted
            "
          >
            <Plus class="h-5 w-5" />
          </div>

          <h2
            class="
              mt-4
              font-semibold
            "
          >
            No instances yet
          </h2>

          <p
            class="
              mt-1
              max-w-sm
              text-sm
              text-muted-foreground
            "
          >
            Create your first PostgreSQL
            instance to get started.
          </p>

          <Button
            class="mt-6"
            @click="createInstance"
          >
            <Plus class="mr-2 h-4 w-4" />

            Create instance
          </Button>
        </CardContent>
      </Card>

      <!-- Instances -->

      <Card v-else>
        <CardHeader>
          <CardTitle>
            Your instances
          </CardTitle>

          <CardDescription>
            {{ instances.length }}
            {{ instances.length === 1
              ? "instance"
              : "instances"
            }}
          </CardDescription>
        </CardHeader>

        <CardContent class="p-0">
          <div class="overflow-x-auto">

            <table class="w-full">
              <thead>
                <tr
                  class="
                    border-b
                    text-left
                    text-sm
                    text-muted-foreground
                  "
                >
                  <th
                    class="
                      px-6
                      py-3
                      font-medium
                    "
                  >
                    Name
                  </th>

                  <th
                    class="
                      px-6
                      py-3
                      font-medium
                    "
                  >
                    Status
                  </th>

                  <th
                    class="
                      hidden
                      px-6
                      py-3
                      font-medium
                      sm:table-cell
                    "
                  >
                    PostgreSQL
                  </th>

                  <th
                    class="
                      hidden
                      px-6
                      py-3
                      font-medium
                      md:table-cell
                    "
                  >
                    Storage
                  </th>

                  <th
                    class="
                      hidden
                      px-6
                      py-3
                      font-medium
                      lg:table-cell
                    "
                  >
                    Created
                  </th>
                </tr>
              </thead>

              <tbody>
                <tr
                  v-for="instance in instances"
                  :key="instance.id"
                  class="
                    cursor-pointer
                    border-b
                    last:border-0
                    hover:bg-muted/50
                  "
                  @click="openInstance(instance.id)"
                >
                  <td class="px-6 py-4">
                    <div>
                      <div
                        class="
                          font-medium
                          text-foreground
                        "
                      >
                        {{ instance.name }}
                      </div>

                      <div
                        class="
                          mt-1
                          font-mono
                          text-xs
                          text-muted-foreground
                        "
                      >
                        {{ instance.id }}
                      </div>
                    </div>
                  </td>

                  <td class="px-6 py-4">
                    <InstanceStatus
                      :status="instance.status"
                    />
                  </td>

                  <td
                    class="
                      hidden
                      px-6
                      py-4
                      text-sm
                      sm:table-cell
                    "
                  >
                    PostgreSQL
                    {{ instance.version }}
                  </td>

                  <td
                    class="
                      hidden
                      px-6
                      py-4
                      text-sm
                      md:table-cell
                    "
                  >
                    {{ instance.storage }}
                  </td>

                  <td
                    class="
                      hidden
                      px-6
                      py-4
                      text-sm
                      text-muted-foreground
                      lg:table-cell
                    "
                  >
                    {{ formatDate(instance.createdAt) }}
                  </td>
                </tr>
              </tbody>
            </table>

          </div>
        </CardContent>
      </Card>

    </div>
  </main>
</template>
```

---

# 6. Make rows keyboard accessible

There's one thing I'd improve in the table above.

Don't make a `<tr>` clickable without considering keyboard accessibility.

A better approach is to make the **instance name a link**.

Instead of:

```vue
<tr
  class="cursor-pointer"
  @click="openInstance(instance.id)"
>
```

use:

```vue
<tr
  :key="instance.id"
  class="border-b last:border-0"
>
```

and change the name:

```vue
<td class="px-6 py-4">
  <RouterLink
    :to="{
      name: 'instance-detail',
      params: {
        id: instance.id,
      },
    }"
    class="
      font-medium
      text-foreground
      hover:underline
    "
  >
    {{ instance.name }}
  </RouterLink>

  <div
    class="
      mt-1
      font-mono
      text-xs
      text-muted-foreground
    "
  >
    {{ instance.id }}
  </div>
</td>
```

Then give your route a name:

```ts
{
  path: "/instances/:id",
  name: "instance-detail",
  component: InstanceDetailPage,
}
```

I'd recommend this version.

You can then remove:

```ts
openInstance()
```

from the page entirely.

---

# 7. Add loading skeletons

Once you have shadcn, I wouldn't show:

```text
Loading instances...
```

for a polished production UI.

Use skeleton rows.

Create:

```text
src/components/instances/InstanceListSkeleton.vue
```

```vue
<script setup lang="ts">
import { Skeleton } from "@/components/ui/skeleton"
</script>

<template>
  <div class="space-y-3">
    <div
      v-for="index in 5"
      :key="index"
      class="
        flex
        items-center
        justify-between
        rounded-lg
        border
        bg-background
        p-4
      "
    >
      <div class="space-y-2">
        <Skeleton class="h-4 w-40" />
        <Skeleton class="h-3 w-24" />
      </div>

      <Skeleton class="h-6 w-20" />
    </div>
  </div>
</template>
```

Then:

```ts
import InstanceListSkeleton
  from "@/components/instances/InstanceListSkeleton.vue"
```

and:

```vue
<InstanceListSkeleton v-if="isLoading" />
```

This gives you a much nicer initial experience.

---

# 8. Consider a card layout instead of a table

For a PaaS, I actually prefer **cards** if you only expect a relatively small number of instances.

For example:

```text
┌─────────────────────────────────────────────────────┐
│ my-postgres-db                         ● Running    │
│ inst-7f3a2b                                          │
│                                                     │
│ PostgreSQL 16       10 GiB                          │
│ Created Aug 4, 2026                                 │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│ production-db                         ● Provisioning │
│ inst-81f52a                                          │
│                                                     │
│ PostgreSQL 15       20 GiB                          │
│ Created Aug 5, 2026                                 │
└─────────────────────────────────────────────────────┘
```

It's particularly good for responsive/mobile layouts.

If you're building something more like:

```text
AWS Console
Render
Railway
Heroku
DigitalOcean
```

then I'd use a table.

For your PaaS, I'd probably start with the **table**, because you'll eventually want columns such as:

```text
Name
Status
PostgreSQL
Instances
Storage
Region
Created
```

---

# 9. Add refresh

I'd add a refresh button to the page header.

Change the header:

```vue
<div
  class="
    flex
    flex-col
    gap-4
    sm:flex-row
    sm:items-center
    sm:justify-between
  "
>
  <div>
    <h1
      class="
        text-2xl
        font-semibold
        tracking-tight
      "
    >
      Instances
    </h1>

    <p
      class="
        mt-1
        text-sm
        text-muted-foreground
      "
    >
      Manage your PostgreSQL instances.
    </p>
  </div>

  <div class="flex gap-2">
    <Button
      variant="outline"
      :disabled="isLoading"
      @click="load"
    >
      <RefreshCw
        class="mr-2 h-4 w-4"
        :class="{
          'animate-spin': isLoading,
        }"
      />

      Refresh
    </Button>

    <Button @click="createInstance">
      <Plus class="mr-2 h-4 w-4" />

      Create instance
    </Button>
  </div>
</div>
```

---

# 10. Automatic refresh is useful here too

There is an important interaction between the list page and your asynchronous POST.

Suppose the user does:

```text
Create instance
       │
       ▼
POST /instances
       │
       ▼
202 Accepted
       │
       ▼
redirect to /instances/abc
```

The detail page polls until:

```text
Running
```

When the user eventually returns to `/instances`, they should see:

```text
my-postgres-db       ● Running
```

So I wouldn't continuously poll the **list page** by default.

Instead:

* detail page → poll while provisioning
* list page → fetch on mount
* user can manually refresh
* optionally refresh when browser window becomes visible

That keeps API traffic reasonable.

---

# 11. Add refresh-on-focus

This is a nice little improvement.

In `useInstances.ts`:

```ts
import {
  onMounted,
  onUnmounted,
  ref,
  computed,
} from "vue"
```

Then:

```ts
function handleVisibilityChange() {
  if (
    document.visibilityState === "visible"
  ) {
    fetchInstances()
  }
}

onMounted(() => {
  document.addEventListener(
    "visibilitychange",
    handleVisibilityChange,
  )
})

onUnmounted(() => {
  document.removeEventListener(
    "visibilitychange",
    handleVisibilityChange,
  )
})
```

Now if someone leaves your dashboard open while provisioning happens and comes back to the browser later, the list updates automatically.

---

# 12. Design the backend for pagination now

This is the biggest API recommendation I'd make.

Don't make your API permanently:

```http
GET /api/v1/instances
```

→

```json
[
  {...},
  {...},
  {...}
]
```

Even if you only have five instances today.

Instead, I'd use:

```http
GET /api/v1/instances
```

with:

```http
GET /api/v1/instances?limit=20&cursor=abc
```

and return:

```json
{
  "items": [
    {
      "id": "inst-7f3a2b",
      "name": "my-postgres-db",
      "version": 16,
      "storage": "10Gi",
      "status": "Running",
      "createdAt": "2026-08-04T09:12:00Z"
    }
  ],
  "nextCursor": "eyJpZCI6I..."
}
```

Then your frontend types become:

```ts
export interface InstanceListResponse {
  items: Instance[]
  nextCursor: string | null
}
```

And your API function:

```ts
export interface ListInstancesParams {
  limit?: number
  cursor?: string
}

export async function listInstances(
  params?: ListInstancesParams,
): Promise<InstanceListResponse> {
  const searchParams = new URLSearchParams()

  if (params?.limit) {
    searchParams.set(
      "limit",
      String(params.limit),
    )
  }

  if (params?.cursor) {
    searchParams.set(
      "cursor",
      params.cursor,
    )
  }

  const query =
    searchParams.toString()

  return apiFetch<InstanceListResponse>(
    `/instances${query ? `?${query}` : ""}`,
  )
}
```

This means your UI doesn't need to be redesigned later when a customer has 500 instances.

---

# 13. Then `useInstances()` can eventually support pagination

You'd eventually have:

```ts
const instances = ref<Instance[]>([])
const nextCursor = ref<string | null>(null)
```

and:

```ts
async function fetchInstances() {
  const response =
    await listInstances({
      limit: 20,
    })

  instances.value = response.items
  nextCursor.value =
    response.nextCursor
}
```

Then:

```ts
async function loadMore() {
  if (!nextCursor.value) {
    return
  }

  const response =
    await listInstances({
      limit: 20,
      cursor: nextCursor.value,
    })

  instances.value.push(
    ...response.items,
  )

  nextCursor.value =
    response.nextCursor
}
```

And your UI can eventually have:

```text
┌──────────────────────────────────────────────┐
│ Instances                                    │
│                                              │
│ ...20 instances...                           │
│                                              │
│             [ Load more ]                    │
└──────────────────────────────────────────────┘
```

No architectural rewrite needed.

---

# 14. One more thing: search/filter

I wouldn't implement filtering yet unless your backend supports it.

But I'd design the endpoint so it can eventually handle:

```http
GET /api/v1/instances
    ?search=production
    &status=Running
    &limit=20
    &cursor=...
```

Then your frontend can have:

```text
┌─────────────────────────────────────────────────────┐
│ Instances                         [+ Create]         │
│                                                     │
│ [ 🔍 Search instances... ] [Status ▾]               │
│                                                     │
│ Name              Status       Version      Storage │
│ ─────────────────────────────────────────────────── │
│ production-db     ● Running    PostgreSQL 16  20Gi │
│ staging-db        ● Running    PostgreSQL 16  10Gi │
│ test-db           ● Pending    PostgreSQL 15   5Gi │
└─────────────────────────────────────────────────────┘
```

But don't build all of that prematurely.

---

# 15. Your resulting structure

At this point your frontend has:

```text
src/
├── api/
│   ├── client.ts
│   ├── errors.ts
│   └── instances.ts
│
├── components/
│   ├── ui/
│   │   └── ...
│   │
│   └── instances/
│       ├── InstanceForm.vue
│       ├── InstanceStatus.vue
│       └── InstanceListSkeleton.vue
│
├── composables/
│   ├── useCreateInstance.ts
│   ├── useInstance.ts
│   └── useInstances.ts
│
├── lib/
│   ├── date.ts
│   └── instance-status.ts
│
├── pages/
│   ├── InstancesPage.vue
│   ├── CreateInstancePage.vue
│   └── InstanceDetailPage.vue
│
├── types/
│   └── instance.ts
│
└── router/
    └── index.ts
```

And the resource flow is now:

```text
                         ┌────────────────────┐
                         │   /instances       │
                         │                    │
                         │ InstancesPage.vue  │
                         └─────────┬──────────┘
                                   │
                             useInstances()
                                   │
                                   ▼
                          GET /instances
                                   │
                     ┌─────────────┴─────────────┐
                     │                           │
                     ▼                           ▼
              instance list              /instances/new
                     │                           │
                     │                           ▼
                     │                    POST /instances
                     │                           │
                     │                           ▼
                     └──────────────► /instances/:id
                                              │
                                              ▼
                                      GET /instances/:id
                                              │
                                              ▼
                                      Poll while Pending
```

### The important architectural rule

I'd keep these responsibilities strict:

| Layer          | Responsibility                         |
| -------------- | -------------------------------------- |
| `api/`         | HTTP                                   |
| `types/`       | TypeScript contracts                   |
| `composables/` | async state + business-ish UI behavior |
| `pages/`       | page orchestration                     |
| `components/`  | reusable UI                            |
| `lib/`         | pure utilities                         |
| `schemas/`     | Zod validation                         |
| `ui/`          | shadcn primitives                      |

That gives you a frontend that's easy to unit test and very straightforward to exercise with Playwright.

For the next endpoints, I'd follow exactly the same pattern rather than introducing Pinia yet. **For only five endpoints, composables + a small API layer are cleaner than adding global state management.** If the PaaS later grows into many resources and cross-page shared state, then Pinia becomes worth considering.
