Yes. I would update the app in a small vertical slice rather than introducing a large dashboard framework.

The goal is:

```text
/                       Dashboard
/instances              Instance list
/instances/new          Create instance
/instances/:id          Instance detail
```

And the dashboard will initially use your existing `GET /instances` endpoint instead of creating a new `/dashboard` backend endpoint.

---

# 1. Target structure

After this change, I'd have:

```text
src/
├── api/
│   ├── client.ts
│   ├── errors.ts
│   └── instances.ts
│
├── components/
│   ├── dashboard/
│   │   ├── DashboardStats.vue
│   │   ├── RecentInstances.vue
│   │   └── GettingStarted.vue
│   │
│   ├── instances/
│   │   ├── InstanceListSkeleton.vue
│   │   └── InstanceStatus.vue
│   │
│   ├── layout/
│   │   ├── AppLayout.vue
│   │   ├── AppSidebar.vue
│   │   └── AppHeader.vue
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
│   ├── DashboardPage.vue
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

The important part is that the dashboard is composed from small components rather than becoming a giant `DashboardPage.vue`.

---

# 2. Make sure your router has the four pages

Update:

```text
src/router/index.ts
```

```ts
import {
  createRouter,
  createWebHistory,
} from "vue-router"

import DashboardPage from "@/pages/DashboardPage.vue"
import InstancesPage from "@/pages/InstancesPage.vue"
import CreateInstancePage from "@/pages/CreateInstancePage.vue"
import InstanceDetailPage from "@/pages/InstanceDetailPage.vue"

const router = createRouter({
  history: createWebHistory(),

  routes: [
    {
      path: "/",
      name: "dashboard",
      component: DashboardPage,
    },

    {
      path: "/instances",
      name: "instances",
      component: InstancesPage,
    },

    {
      path: "/instances/new",
      name: "instance-create",
      component: CreateInstancePage,
    },

    {
      path: "/instances/:id",
      name: "instance-detail",
      component: InstanceDetailPage,
    },
  ],
})

export default router
```

This gives you a clean resource hierarchy.

---

# 3. Create the application layout

Now I would stop putting the full page shell inside individual pages.

Create:

```text
src/components/layout/AppLayout.vue
```

```vue
<script setup lang="ts">
import AppSidebar from "./AppSidebar.vue"
import AppHeader from "./AppHeader.vue"
</script>

<template>
  <div class="min-h-screen bg-muted/40">
    <div class="flex min-h-screen">

      <AppSidebar />

      <div class="flex min-w-0 flex-1 flex-col">

        <AppHeader />

        <main class="flex-1">
          <RouterView />
        </main>

      </div>
    </div>
  </div>
</template>
```

Now your application has:

```text
AppLayout
├── Sidebar
├── Header
└── Page
```

---

# 4. Create the sidebar

Create:

```text
src/components/layout/AppSidebar.vue
```

```vue
<script setup lang="ts">
import {
  Database,
  Home,
  BookOpen,
} from "lucide-vue-next"

import { RouterLink } from "vue-router"
</script>

<template>
  <aside
    class="
      hidden
      w-64
      shrink-0
      border-r
      bg-background
      md:block
    "
  >
    <div
      class="
        flex
        h-16
        items-center
        border-b
        px-6
      "
    >
      <RouterLink
        to="/"
        class="
          flex
          items-center
          gap-2
          font-semibold
        "
      >
        <Database class="h-5 w-5" />

        PostgreSQL PaaS
      </RouterLink>
    </div>

    <nav class="space-y-1 p-4">
      <RouterLink
        to="/"
        class="
          flex
          items-center
          gap-3
          rounded-md
          px-3
          py-2
          text-sm
          font-medium
          text-muted-foreground
          transition-colors
          hover:bg-muted
          hover:text-foreground
        "
        active-class="
          bg-muted
          text-foreground
        "
      >
        <Home class="h-4 w-4" />

        Dashboard
      </RouterLink>

      <RouterLink
        to="/instances"
        class="
          flex
          items-center
          gap-3
          rounded-md
          px-3
          py-2
          text-sm
          font-medium
          text-muted-foreground
          transition-colors
          hover:bg-muted
          hover:text-foreground
        "
        active-class="
          bg-muted
          text-foreground
        "
      >
        <Database class="h-4 w-4" />

        Instances
      </RouterLink>
    </nav>

    <div class="mt-4 border-t p-4">
      <a
        href="#"
        class="
          flex
          items-center
          gap-3
          rounded-md
          px-3
          py-2
          text-sm
          text-muted-foreground
          hover:bg-muted
          hover:text-foreground
        "
      >
        <BookOpen class="h-4 w-4" />

        Documentation
      </a>
    </div>
  </aside>
</template>
```

For now, don't build a complicated responsive sidebar. We can add a mobile navigation later.

---

# 5. Create the header

Create:

```text
src/components/layout/AppHeader.vue
```

```vue
<script setup lang="ts">
import { Button } from "@/components/ui/button"
</script>

<template>
  <header
    class="
      flex
      h-16
      items-center
      justify-between
      border-b
      bg-background
      px-4
      sm:px-6
    "
  >
    <div class="md:hidden">
      <span class="font-semibold">
        PostgreSQL PaaS
      </span>
    </div>

    <div class="hidden md:block" />

    <div class="flex items-center gap-2">
      <Button
        variant="ghost"
        size="sm"
      >
        Account
      </Button>
    </div>
  </header>
</template>
```

Don't worry about the account button yet. Later it can become your user menu.

---

# 6. Make the router use the layout

Now change your router.

Instead of having the pages directly at the top level:

```ts
{
  path: "/",
  component: DashboardPage,
}
```

use the layout:

```ts
import AppLayout from "@/components/layout/AppLayout.vue"

const router = createRouter({
  history: createWebHistory(),

  routes: [
    {
      path: "/",
      component: AppLayout,

      children: [
        {
          path: "",
          name: "dashboard",
          component: DashboardPage,
        },

        {
          path: "instances",
          name: "instances",
          component: InstancesPage,
        },

        {
          path: "instances/new",
          name: "instance-create",
          component: CreateInstancePage,
        },

        {
          path: "instances/:id",
          name: "instance-detail",
          component: InstanceDetailPage,
        },
      ],
    },
  ],
})
```

This is a much better structure.

The layout is now shared automatically.

---

# 7. Create `DashboardPage.vue`

Now create:

```text
src/pages/DashboardPage.vue
```

Keep this page very simple:

```vue
<script setup lang="ts">
import { onMounted } from "vue"
import { useRouter } from "vue-router"
import { Plus } from "lucide-vue-next"

import { Button } from "@/components/ui/button"

import DashboardStats from "@/components/dashboard/DashboardStats.vue"
import RecentInstances from "@/components/dashboard/RecentInstances.vue"
import GettingStarted from "@/components/dashboard/GettingStarted.vue"

import { useInstances } from "@/composables/useInstances"

const router = useRouter()

const {
  instances,
  isLoading,
  error,
  fetchInstances,
} = useInstances()

function createInstance() {
  router.push({
    name: "instance-create",
  })
}

function viewInstances() {
  router.push({
    name: "instances",
  })
}

onMounted(() => {
  fetchInstances()
})
</script>

<template>
  <div class="px-4 py-8 sm:px-6">
    <div class="mx-auto max-w-6xl space-y-8">

      <!-- Page header -->

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
            Dashboard
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

      <!-- Error -->

      <div
        v-if="error"
        class="
          rounded-lg
          border
          border-destructive/50
          bg-destructive/10
          p-4
          text-sm
          text-destructive
        "
      >
        {{ error.message }}
      </div>

      <!-- Statistics -->

      <DashboardStats
        :instances="instances"
        :is-loading="isLoading"
      />

      <!-- Recent instances -->

      <RecentInstances
        :instances="instances"
        :is-loading="isLoading"
        @view-all="viewInstances"
      />

      <!-- Getting started -->

      <GettingStarted
        v-if="!isLoading && instances.length === 0"
        @create="createInstance"
      />

    </div>
  </div>
</template>
```

Notice how little logic is in the page.

That's intentional.

---

# 8. Create the dashboard statistics component

Create:

```text
src/components/dashboard/DashboardStats.vue
```

```vue
<script setup lang="ts">
import {
  Database,
  CircleCheck,
  LoaderCircle,
} from "lucide-vue-next"

import {
  Card,
  CardContent,
} from "@/components/ui/card"

import { Skeleton } from "@/components/ui/skeleton"

import type { Instance } from "@/types/instance"

const props = defineProps<{
  instances: Instance[]
  isLoading: boolean
}>()

const runningCount = () =>
  props.instances.filter(
    instance =>
      instance.status === "Running",
  ).length

const provisioningCount = () =>
  props.instances.filter(
    instance =>
      instance.status === "Pending" ||
      instance.status === "Provisioning",
  ).length
</script>

<template>
  <div
    class="
      grid
      gap-4
      sm:grid-cols-3
    "
  >
    <!-- Total -->

    <Card>
      <CardContent class="p-6">
        <div
          class="
            flex
            items-center
            justify-between
          "
        >
          <div>
            <p
              class="
                text-sm
                text-muted-foreground
              "
            >
              Instances
            </p>

            <Skeleton
              v-if="isLoading"
              class="mt-2 h-8 w-12"
            />

            <p
              v-else
              class="
                mt-1
                text-3xl
                font-semibold
              "
            >
              {{ instances.length }}
            </p>
          </div>

          <div
            class="
              flex
              h-10
              w-10
              items-center
              justify-center
              rounded-lg
              bg-muted
            "
          >
            <Database class="h-5 w-5" />
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Running -->

    <Card>
      <CardContent class="p-6">
        <div
          class="
            flex
            items-center
            justify-between
          "
        >
          <div>
            <p
              class="
                text-sm
                text-muted-foreground
              "
            >
              Running
            </p>

            <Skeleton
              v-if="isLoading"
              class="mt-2 h-8 w-12"
            />

            <p
              v-else
              class="
                mt-1
                text-3xl
                font-semibold
              "
            >
              {{ runningCount() }}
            </p>
          </div>

          <div
            class="
              flex
              h-10
              w-10
              items-center
              justify-center
              rounded-lg
              bg-green-100
              text-green-700
              dark:bg-green-950
              dark:text-green-400
            "
          >
            <CircleCheck class="h-5 w-5" />
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Provisioning -->

    <Card>
      <CardContent class="p-6">
        <div
          class="
            flex
            items-center
            justify-between
          "
        >
          <div>
            <p
              class="
                text-sm
                text-muted-foreground
              "
            >
              Provisioning
            </p>

            <Skeleton
              v-if="isLoading"
              class="mt-2 h-8 w-12"
            />

            <p
              v-else
              class="
                mt-1
                text-3xl
                font-semibold
              "
            >
              {{ provisioningCount() }}
            </p>
          </div>

          <div
            class="
              flex
              h-10
              w-10
              items-center
              justify-center
              rounded-lg
              bg-blue-100
              text-blue-700
              dark:bg-blue-950
              dark:text-blue-400
            "
          >
            <LoaderCircle class="h-5 w-5" />
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
```

This is intentionally simple.

You don't need charts yet.

---

# 9. Create the recent instances component

Create:

```text
src/components/dashboard/RecentInstances.vue
```

```vue
<script setup lang="ts">
import {
  ArrowRight,
} from "lucide-vue-next"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"

import InstanceStatus from "@/components/instances/InstanceStatus.vue"

import { formatDate } from "@/lib/date"

import type { Instance } from "@/types/instance"

const props = defineProps<{
  instances: Instance[]
  isLoading: boolean
}>()

defineEmits<{
  viewAll: []
}>()

function recentInstances(): Instance[] {
  return [...props.instances]
    .sort(
      (a, b) =>
        new Date(b.createdAt).getTime() -
        new Date(a.createdAt).getTime(),
    )
    .slice(0, 5)
}
</script>

<template>
  <Card>
    <CardHeader
      class="
        flex
        flex-row
        items-center
        justify-between
      "
    >
      <div>
        <CardTitle>
          Recent instances
        </CardTitle>

        <CardDescription>
          Your most recently created instances.
        </CardDescription>
      </div>

      <Button
        variant="ghost"
        size="sm"
        @click="$emit('viewAll')"
      >
        View all

        <ArrowRight class="ml-2 h-4 w-4" />
      </Button>
    </CardHeader>

    <CardContent>

      <!-- Loading -->

      <div
        v-if="isLoading"
        class="space-y-4"
      >
        <div
          v-for="index in 3"
          :key="index"
          class="
            flex
            items-center
            justify-between
          "
        >
          <div class="space-y-2">
            <Skeleton class="h-4 w-40" />
            <Skeleton class="h-3 w-24" />
          </div>

          <Skeleton class="h-6 w-20" />
        </div>
      </div>

      <!-- Empty -->

      <div
        v-else-if="instances.length === 0"
        class="
          py-8
          text-center
          text-sm
          text-muted-foreground
        "
      >
        No instances yet.
      </div>

      <!-- Instances -->

      <div
        v-else
        class="divide-y"
      >
        <RouterLink
          v-for="instance in recentInstances()"
          :key="instance.id"
          :to="{
            name: 'instance-detail',
            params: {
              id: instance.id,
            },
          }"
          class="
            flex
            items-center
            justify-between
            gap-4
            py-4
            first:pt-0
            last:pb-0
            hover:bg-muted/50
          "
        >
          <div class="min-w-0">
            <p
              class="
                truncate
                font-medium
              "
            >
              {{ instance.name }}
            </p>

            <div
              class="
                mt-1
                flex
                flex-wrap
                gap-x-3
                gap-y-1
                text-xs
                text-muted-foreground
              "
            >
              <span class="font-mono">
                {{ instance.id }}
              </span>

              <span>
                PostgreSQL {{ instance.version }}
              </span>

              <span>
                {{ formatDate(instance.createdAt) }}
              </span>
            </div>
          </div>

          <InstanceStatus
            :status="instance.status"
          />
        </RouterLink>
      </div>

    </CardContent>
  </Card>
</template>
```

---

# 10. Create the getting-started component

Create:

```text
src/components/dashboard/GettingStarted.vue
```

```vue
<script setup lang="ts">
import { Database, Plus } from "lucide-vue-next"

import {
  Card,
  CardContent,
} from "@/components/ui/card"

import { Button } from "@/components/ui/button"

defineEmits<{
  create: []
}>()
</script>

<template>
  <Card>
    <CardContent
      class="
        flex
        flex-col
        items-center
        justify-center
        py-12
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
          bg-primary/10
          text-primary
        "
      >
        <Database class="h-5 w-5" />
      </div>

      <h2
        class="
          mt-4
          text-lg
          font-semibold
        "
      >
        Create your first instance
      </h2>

      <p
        class="
          mt-1
          max-w-md
          text-sm
          text-muted-foreground
        "
      >
        Deploy a managed PostgreSQL instance
        and let CNPG handle the underlying
        PostgreSQL cluster.
      </p>

      <Button
        class="mt-6"
        @click="$emit('create')"
      >
        <Plus class="mr-2 h-4 w-4" />

        Create instance
      </Button>
    </CardContent>
  </Card>
</template>
```

This means a brand-new user doesn't get an empty dashboard full of meaningless statistics.

They get:

```text
Create your first instance

Deploy a managed PostgreSQL instance
and let CNPG handle the underlying
PostgreSQL cluster.

[ Create instance ]
```

That's a much better first-run experience.

---

# 11. Make the dashboard composable reusable

Your existing `useInstances()` is already suitable for this.

I'd make one small improvement.

Currently, if you have:

```ts
async function fetchInstances() {
```

you probably don't need to expose internal implementation details.

Keep the return value:

```ts
return {
  instances,
  isLoading,
  error,
  isEmpty,
  fetchInstances,
}
```

That's enough.

Don't put:

```text
runningCount
provisioningCount
recentInstances
```

inside `useInstances()`.

Those are **dashboard presentation concerns**.

This is a useful distinction:

```text
useInstances()
    │
    └── "Give me instances"

DashboardStats
    │
    └── "How should I summarize instances?"

RecentInstances
    │
    └── "How should I display recent instances?"
```

That keeps your composable reusable on `/instances`.

---

# 12. Improve `useInstances()` error handling

I'd make sure it handles a failed request cleanly:

```ts
import {
  computed,
  ref,
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

---

# 13. Use named routes everywhere

Now that your application has several pages, stop doing this:

```ts
router.push("/instances/new")
```

Prefer:

```ts
router.push({
  name: "instance-create",
})
```

Likewise:

```ts
router.push({
  name: "instances",
})
```

and:

```vue
<RouterLink
  :to="{
    name: 'instance-detail',
    params: {
      id: instance.id,
    },
  }"
>
```

This becomes valuable when URLs change later.

For example, you could change:

```text
/instances/:id
```

to:

```text
/dashboard/instances/:id
```

without having string URLs scattered throughout the application.

---

# 14. Your dashboard should NOT have its own backend endpoint yet

This is an important decision.

Don't create:

```http
GET /api/v1/dashboard
```

just to return:

```json
{
  "totalInstances": 12,
  "runningInstances": 9,
  "provisioningInstances": 2,
  "recentInstances": [...]
}
```

At this stage, the dashboard can derive all of that from:

```http
GET /api/v1/instances
```

This is enough:

```text
GET /instances
       │
       ├── total
       ├── running
       ├── provisioning
       └── recent
```

Later, if your platform has millions of resources or expensive queries, you can introduce a dashboard-specific aggregation endpoint.

But don't optimize for that now.

---

# 15. One change I would make to your GET `/instances`

If you haven't finalized the API contract yet, I'd make the response:

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
  "nextCursor": null
}
```

rather than:

```json
[
  {
    "id": "inst-7f3a2b"
  }
]
```

Because later you can support:

```http
GET /instances?limit=20
GET /instances?limit=20&cursor=abc
GET /instances?search=production
GET /instances?status=Running
```

without changing the basic frontend architecture.

Your TypeScript type:

```ts
export interface InstanceListResponse {
  items: Instance[]
  nextCursor: string | null
}
```

Then the dashboard doesn't care whether there are 5 instances or 50,000.

---

# 16. What the finished dashboard looks like

The result should be intentionally boring:

```text
┌──────────────────────────────────────────────────────────────┐
│ PostgreSQL PaaS             Dashboard  Instances      Account │
├────────────────┬─────────────────────────────────────────────┤
│                │                                             │
│ 🏠 Dashboard   │  Dashboard                   [+ Create]    │
│                │  Manage your PostgreSQL instances.          │
│ ◉ Instances    │                                             │
│                │  ┌───────────┐ ┌──────────┐ ┌───────────┐ │
│                │  │ Instances │ │ Running  │ │Provision. │ │
│                │  │     5     │ │    3     │ │     1     │ │
│                │  └───────────┘ └──────────┘ └───────────┘ │
│                │                                             │
│                │  Recent instances              View all →  │
│                │  ─────────────────────────────────────────  │
│                │  production-db       ● Running             │
│                │  staging-db          ● Running             │
│                │  development-db      ● Provisioning        │
│                │                                             │
│                │  Getting Started                            │
│                │  ┌───────────────────────────────────────┐ │
│                │  │ Create your first PostgreSQL instance │ │
│                │  │                            [Create]   │ │
│                │  └───────────────────────────────────────┘ │
│                │                                             │
└────────────────┴─────────────────────────────────────────────┘
```

## 17. Final architecture

At this point I'd consider your frontend architecture to be:

```text
                         AppLayout
                    ┌────────┴────────┐
                    │                 │
                Sidebar             Header
                    │
                    ▼
               RouterView
                    │
       ┌────────────┼───────────────┐
       │            │               │
       ▼            ▼               ▼
   Dashboard     Instances       Instance
       │            │               │
       │            │               │
       └────────────┼───────────────┘
                    │
                    ▼
              useInstances()
                    │
                    ▼
              API / instances
                    │
                    ▼
          GET /api/v1/instances
```

And I'd **stop here for the dashboard**.

Don't add charts, CPU graphs, cluster health graphs, activity feeds, billing metrics, resource utilization, etc. until you actually have backend data for them.

Your next useful frontend work is probably the **application shell + responsive mobile navigation**, followed by making the instance detail page more PaaS-like with **connection information, status/provisioning events, and actions such as delete/scale**.
