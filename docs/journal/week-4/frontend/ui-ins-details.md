Yes. Since `POST /instances` already established the structure, I'd make the detail page the next vertical slice and keep it similarly thin:

```text
Route
  /instances/:id
       │
       ▼
InstanceDetailPage.vue
       │
       ▼
useInstance(id)
       │
       ▼
instancesApi.get(id)
       │
       ▼
GET /api/v1/instances/:id
```

I'll assume the GET response is the same `Instance` object you showed for POST:

```json
{
  "id": "inst-7f3a2b",
  "name": "my-postgres-db",
  "version": 16,
  "storage": "10Gi",
  "status": "Pending",
  "createdAt": "2026-08-04T09:12:00Z"
}
```

No tests this time.

---

# 1. Add the GET endpoint to the API layer

You already have:

```text
src/api/instances.ts
```

Keep the existing `createInstance()` and add `getInstance()`.

```ts
import { apiFetch } from "./client"
import type {
  CreateInstanceRequest,
  Instance,
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
```

### Why `encodeURIComponent()`?

Don't do:

```ts
`/instances/${id}`
```

blindly.

Your ID currently looks like:

```text
inst-7f3a2b
```

which is safe, but encoding the path parameter makes the API function more robust if the ID format ever changes.

---

# 2. Create `useInstance`

Create:

```text
src/composables/useInstance.ts
```

```ts
import { ref } from "vue"

import { getInstance } from "@/api/instances"
import { ApiError } from "@/api/errors"
import type { Instance } from "@/types/instance"

export function useInstance() {
  const instance = ref<Instance | null>(null)
  const isLoading = ref(false)
  const error = ref<ApiError | null>(null)

  async function fetchInstance(id: string) {
    isLoading.value = true
    error.value = null

    try {
      instance.value = await getInstance(id)

      return instance.value
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
    instance,
    isLoading,
    error,
    fetchInstance,
  }
}
```

The important thing is that this composable doesn't know anything about Vue Router.

It just knows:

> Give me an instance ID and I'll load that instance.

---

# 3. Add the route

Update:

```text
src/router/index.ts
```

```ts
import {
  createRouter,
  createWebHistory,
} from "vue-router"

import CreateInstancePage from "@/pages/CreateInstancePage.vue"
import InstanceDetailPage from "@/pages/InstanceDetailPage.vue"

const router = createRouter({
  history: createWebHistory(),

  routes: [
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

Now:

```text
/instances/new
```

is your creation page and:

```text
/instances/inst-7f3a2b
```

is your detail page.

---

# 4. Create the detail page

Create:

```text
src/pages/InstanceDetailPage.vue
```

I'd make the page responsible for orchestration:

* get route parameter
* load instance
* handle loading
* handle error
* display instance

The actual visual pieces can remain components.

```vue
<script setup lang="ts">
import { onMounted } from "vue"
import { useRoute, useRouter } from "vue-router"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

import { Button } from "@/components/ui/button"

import InstanceStatus from "@/components/instances/InstanceStatus.vue"

import { useInstance } from "@/composables/useInstance"

const route = useRoute()
const router = useRouter()

const {
  instance,
  isLoading,
  error,
  fetchInstance,
} = useInstance()

const instanceId = String(route.params.id)

async function load() {
  try {
    await fetchInstance(instanceId)
  } catch {
    // Error is exposed by useInstance().
  }
}

function goBack() {
  router.push("/instances/new")
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
    <div class="mx-auto max-w-4xl space-y-6">

      <!-- Back -->

      <Button
        variant="ghost"
        class="-ml-2"
        @click="goBack"
      >
        ← Back
      </Button>

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
            Loading instance...
          </div>
        </CardContent>
      </Card>

      <!-- Error -->

      <Card v-else-if="error">
        <CardHeader>
          <CardTitle>
            Unable to load instance
          </CardTitle>

          <CardDescription>
            {{ error.message }}
          </CardDescription>
        </CardHeader>

        <CardContent>
          <Button @click="load">
            Try again
          </Button>
        </CardContent>
      </Card>

      <!-- Instance -->

      <template v-else-if="instance">

        <!-- Header -->

        <div
          class="
            flex
            flex-col
            gap-4
            sm:flex-row
            sm:items-start
            sm:justify-between
          "
        >
          <div>
            <div
              class="
                flex
                items-center
                gap-3
              "
            >
              <h1
                class="
                  text-2xl
                  font-semibold
                  tracking-tight
                "
              >
                {{ instance.name }}
              </h1>

              <InstanceStatus
                :status="instance.status"
              />
            </div>

            <p
              class="
                mt-1
                font-mono
                text-sm
                text-muted-foreground
              "
            >
              {{ instance.id }}
            </p>
          </div>
        </div>

        <!-- Configuration -->

        <Card>
          <CardHeader>
            <CardTitle>
              Configuration
            </CardTitle>

            <CardDescription>
              PostgreSQL instance configuration.
            </CardDescription>
          </CardHeader>

          <CardContent>
            <dl
              class="
                grid
                gap-6
                sm:grid-cols-2
                lg:grid-cols-3
              "
            >
              <div>
                <dt
                  class="
                    text-sm
                    text-muted-foreground
                  "
                >
                  PostgreSQL version
                </dt>

                <dd
                  class="
                    mt-1
                    text-lg
                    font-medium
                  "
                >
                  {{ instance.version }}
                </dd>
              </div>

              <div>
                <dt
                  class="
                    text-sm
                    text-muted-foreground
                  "
                >
                  Storage
                </dt>

                <dd
                  class="
                    mt-1
                    text-lg
                    font-medium
                  "
                >
                  {{ instance.storage }}
                </dd>
              </div>

              <div>
                <dt
                  class="
                    text-sm
                    text-muted-foreground
                  "
                >
                  Status
                </dt>

                <dd class="mt-2">
                  <InstanceStatus
                    :status="instance.status"
                  />
                </dd>
              </div>
            </dl>
          </CardContent>
        </Card>

        <!-- Metadata -->

        <Card>
          <CardHeader>
            <CardTitle>
              Details
            </CardTitle>
          </CardHeader>

          <CardContent>
            <dl class="space-y-4">
              <div
                class="
                  flex
                  flex-col
                  gap-1
                  sm:flex-row
                  sm:justify-between
                "
              >
                <dt
                  class="
                    text-sm
                    text-muted-foreground
                  "
                >
                  Instance ID
                </dt>

                <dd
                  class="
                    font-mono
                    text-sm
                  "
                >
                  {{ instance.id }}
                </dd>
              </div>

              <div
                class="
                  flex
                  flex-col
                  gap-1
                  sm:flex-row
                  sm:justify-between
                "
              >
                <dt
                  class="
                    text-sm
                    text-muted-foreground
                  "
                >
                  Created
                </dt>

                <dd class="text-sm">
                  {{ instance.createdAt }}
                </dd>
              </div>
            </dl>
          </CardContent>
        </Card>

      </template>
    </div>
  </main>
</template>
```

---

# 5. Improve the date display

I wouldn't leave:

```vue
{{ instance.createdAt }}
```

in the final UI.

The API gives us:

```text
2026-08-04T09:12:00Z
```

but users want something like:

```text
August 4, 2026 at 11:12 AM
```

Create:

```text
src/lib/date.ts
```

```ts
export function formatDate(
  value: string,
): string {
  return new Intl.DateTimeFormat(
    undefined,
    {
      dateStyle: "medium",
      timeStyle: "short",
    },
  ).format(new Date(value))
}
```

Then:

```ts
import { formatDate } from "@/lib/date"
```

and:

```vue
<dd class="text-sm">
  {{ formatDate(instance.createdAt) }}
</dd>
```

This is preferable to doing formatting directly inside the template.

---

# 6. Improve the status component

You already have:

```text
src/components/instances/InstanceStatus.vue
```

I'd make the statuses more intentional.

```vue
<script setup lang="ts">
import { computed } from "vue"

import { Badge } from "@/components/ui/badge"

import type {
  InstanceStatus,
} from "@/types/instance"

const props = defineProps<{
  status: InstanceStatus
}>()

const variant = computed(() => {
  switch (props.status) {
    case "Running":
      return "default"

    case "Failed":
      return "destructive"

    case "Pending":
    case "Provisioning":
    case "Deleting":
      return "secondary"

    default:
      return "outline"
  }
})
</script>

<template>
  <Badge :variant="variant">
    {{ status }}
  </Badge>
</template>
```

Later, you could add an animated indicator for transitional states:

```text
● Pending
● Provisioning
● Running
● Failed
```

---

# 7. Add a status description

For a PaaS, status alone isn't particularly friendly.

For example:

```text
Pending
```

doesn't tell the user what's happening.

I'd add:

```text
Pending
Your instance has been accepted and provisioning will begin shortly.
```

Create:

```text
src/lib/instance-status.ts
```

```ts
import type {
  InstanceStatus,
} from "@/types/instance"

export function getInstanceStatusDescription(
  status: InstanceStatus,
): string {
  switch (status) {
    case "Pending":
      return "Your instance has been accepted and provisioning will begin shortly."

    case "Provisioning":
      return "Your PostgreSQL instance is currently being provisioned."

    case "Running":
      return "Your PostgreSQL instance is running normally."

    case "Failed":
      return "Provisioning failed. Please check the instance details or try again."

    case "Deleting":
      return "Your PostgreSQL instance is being deleted."

    case "Deleted":
      return "This instance has been deleted."

    default:
      return ""
  }
}
```

Then on the page:

```ts
import {
  getInstanceStatusDescription,
} from "@/lib/instance-status"
```

and:

```vue
<p
  class="
    mt-2
    text-sm
    text-muted-foreground
  "
>
  {{
    getInstanceStatusDescription(
      instance.status,
    )
  }}
</p>
```

So the header becomes:

```text
my-postgres-db                 [Pending]

inst-7f3a2b

Your instance has been accepted and
provisioning will begin shortly.
```

Much better PaaS UX.

---

# 8. Add a refresh button

Because your POST is asynchronous, this page will eventually need to refresh.

I'd add a refresh button now.

Import:

```ts
import { RefreshCw } from "lucide-vue-next"
```

If you haven't already installed Lucide:

```bash
pnpm add lucide-vue-next
```

Then:

```vue
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
```

I'd put this in the page header:

```vue
<div
  class="
    flex
    flex-col
    gap-4
    sm:flex-row
    sm:items-start
    sm:justify-between
"
>
  <div>
    ...
  </div>

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
</div>
```

---

# 9. Then add automatic polling

This is the important part for your architecture.

Your POST returns:

```http
202 Accepted
```

with:

```json
{
  "status": "Pending"
}
```

Therefore, the detail page is a natural place to poll.

I would **not** put polling directly inside `InstanceDetailPage.vue`.

Put it inside `useInstance()`.

Change it to:

```ts
import {
  onUnmounted,
  ref,
} from "vue"

import { getInstance } from "@/api/instances"
import { ApiError } from "@/api/errors"
import type { Instance } from "@/types/instance"

export function useInstance() {
  const instance = ref<Instance | null>(null)
  const isLoading = ref(false)
  const error = ref<ApiError | null>(null)

  let pollTimer: ReturnType<
    typeof setTimeout
  > | null = null

  async function fetchInstance(id: string) {
    isLoading.value = true
    error.value = null

    try {
      instance.value = await getInstance(id)

      return instance.value
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

  function startPolling(id: string) {
    stopPolling()

    async function poll() {
      try {
        const result =
          await fetchInstance(id)

        if (
          result.status === "Pending" ||
          result.status === "Provisioning"
        ) {
          pollTimer = setTimeout(
            poll,
            3000,
          )
        }
      } catch {
        stopPolling()
      }
    }

    pollTimer = setTimeout(
      poll,
      3000,
    )
  }

  function stopPolling() {
    if (pollTimer) {
      clearTimeout(pollTimer)
      pollTimer = null
    }
  }

  onUnmounted(stopPolling)

  return {
    instance,
    isLoading,
    error,
    fetchInstance,
    startPolling,
    stopPolling,
  }
}
```

Then on the page:

```ts
const {
  instance,
  isLoading,
  error,
  fetchInstance,
  startPolling,
} = useInstance()

async function load() {
  try {
    const result =
      await fetchInstance(instanceId)

    if (
      result.status === "Pending" ||
      result.status === "Provisioning"
    ) {
      startPolling(instanceId)
    }
  } catch {
    // Error exposed by composable.
  }
}
```

Now:

```text
POST
 │
 │ 202
 ▼
/instances/:id
 │
 │ Pending
 ▼
wait 3 seconds
 │
 ▼
GET /instances/:id
 │
 │ Provisioning
 ▼
wait 3 seconds
 │
 ▼
GET /instances/:id
 │
 │ Running
 ▼
STOP
```

That's a very natural architecture for your PaaS.

---

# 10. Don't poll forever

I'd actually put a limit on this before production.

For example:

```text
maximum polling duration = 10 minutes
```

or:

```text
maximum attempts = 200
```

Otherwise a broken backend could cause a browser to poll indefinitely.

A more production-ready `useInstance()` could eventually have:

```ts
const maxAttempts = 200
const pollingInterval = 3000
```

and stop once either:

```text
Running
Failed
Deleted
```

is reached.

---

# 11. Add a proper not-found state

Your GET endpoint presumably has a `404`.

The page should distinguish:

```text
404
```

from:

```text
500
```

and:

```text
network failure
```

For example:

```vue
<Card v-else-if="error">
  <CardHeader>
    <CardTitle>
      {{
        error.status === 404
          ? "Instance not found"
          : "Unable to load instance"
      }}
    </CardTitle>

    <CardDescription>
      {{
        error.status === 404
          ? "This instance may have been deleted or the URL is incorrect."
          : error.message
      }}
    </CardDescription>
  </CardHeader>

  <CardContent class="flex gap-2">
    <Button
      v-if="error.status !== 404"
      @click="load"
    >
      Try again
    </Button>

    <Button
      variant="outline"
      @click="goBack"
    >
      Back
    </Button>
  </CardContent>
</Card>
```

This is why we created `ApiError` with a `status`.

---

# 12. One thing I'd change from the previous page

Previously I suggested:

```ts
function goBack() {
  router.push("/instances/new")
}
```

For the detail page, I actually wouldn't do that.

You eventually want:

```text
/instances
```

as the list page.

So I'd establish that URL now:

```text
/instances
```

Then:

```ts
function goBack() {
  router.push("/instances")
}
```

Your eventual navigation becomes:

```text
/instances
     │
     ├── Create
     │      │
     │      ▼
     │  /instances/new
     │      │
     │      ▼
     │  /instances/:id
     │
     └── existing instance
            │
            ▼
       /instances/:id
```

That's a much cleaner application structure.

---

# 13. I would also add a breadcrumb

Once you have more than a couple pages, the detail page benefits from:

```text
Instances / my-postgres-db
```

For example:

```vue
<nav
  class="
    text-sm
    text-muted-foreground
  "
  aria-label="Breadcrumb"
>
  <button
    class="hover:text-foreground"
    @click="goBack"
  >
    Instances
  </button>

  <span class="mx-2">/</span>

  <span class="text-foreground">
    {{ instance.name }}
  </span>
</nav>
```

This will become useful once you have:

```text
Instances
Instance
Backups
Connections
Settings
```

---

# 14. Final component structure

For this page I'd end up with:

```text
src/
├── api/
│   ├── client.ts
│   ├── errors.ts
│   └── instances.ts
│
├── components/
│   ├── ui/
│   │
│   └── instances/
│       ├── InstanceForm.vue
│       └── InstanceStatus.vue
│
├── composables/
│   ├── useCreateInstance.ts
│   └── useInstance.ts
│
├── lib/
│   ├── date.ts
│   └── instance-status.ts
│
├── pages/
│   ├── CreateInstancePage.vue
│   └── InstanceDetailPage.vue
│
├── schemas/
│   └── instance.ts
│
├── types/
│   └── instance.ts
│
└── router/
    └── index.ts
```

The dependency flow stays very clean:

```text
                    InstanceDetailPage
                           │
                           ▼
                     useInstance()
                           │
                           ▼
                  instances.getInstance()
                           │
                           ▼
                       apiFetch()
                           │
                           ▼
                 GET /api/v1/instances/:id
```

And importantly, **nothing below the page knows about the page**.

---

# 15. What the resulting UI should look like

Conceptually:

```text
Instances / my-postgres-db

┌──────────────────────────────────────────────────────┐
│                                                      │
│  my-postgres-db                         [Pending]    │
│  inst-7f3a2b                                         │
│                                                      │
│  Your instance has been accepted and provisioning    │
│  will begin shortly.                     [Refresh]   │
│                                                      │
├──────────────────────────────────────────────────────┤
│                                                      │
│  Configuration                                       │
│  PostgreSQL instance configuration.                  │
│                                                      │
│  PostgreSQL version    Storage       Status          │
│  16                    10Gi          [Pending]       │
│                                                      │
├──────────────────────────────────────────────────────┤
│                                                      │
│  Details                                             │
│                                                      │
│  Instance ID                inst-7f3a2b              │
│  Created                    Aug 4, 2026, 11:12 AM    │
│                                                      │
└──────────────────────────────────────────────────────┘
```

Then, when CNPG finishes:

```text
my-postgres-db                         [Running]

Your PostgreSQL instance is running normally.
```

This is a good foundation because your **GET detail endpoint becomes the central resource page**. Later, when you add endpoints for backups, connections, scaling, deletion, logs, etc., they'll naturally attach to this page rather than creating unrelated pages and state management.
