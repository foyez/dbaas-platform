# UI Development

For `POST /api/v1/instances` contract, I would build this as a **vertical slice** first: everything needed to create an instance, from the Vue form all the way through validation, API handling, unit tests, and one Playwright E2E test.

That gives you a pattern you can then copy for your other four endpoints.

I’ll use **Vue 3 + TypeScript + Vite + Tailwind v4 + shadcn-vue + Zod + Vitest + Vue Test Utils + Playwright**. The current shadcn-vue Vite setup uses Tailwind's Vite plugin and `@import "tailwindcss"`, and its CLI can add individual components to your project. ([Shadcn Vue][1])

---

# 1. First, define what we're building

Your API is:

```http
POST /api/v1/instances
Idempotency-Key: <uuid>
Content-Type: application/json
```

with:

```json
{
  "name": "my-db-2",
  "instances": 2,
  "version": 15,
  "storage": "7Gi"
}
```

and successful creation returns:

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

The frontend flow should therefore be:

```text
Create Instance page
        │
        ▼
InstanceForm.vue
        │
        │ validate
        ▼
Zod schema
        │
        │ valid
        ▼
instancesApi.create()
        │
        │ POST + Idempotency-Key
        ▼
Backend
        │
        ├── 202 → success
        ├── 200 → idempotent replay
        ├── 400 → bad request
        ├── 401 → authentication
        ├── 409 → name conflict
        ├── 422 → semantic validation
        └── 500 → server error
```

One important design decision:

**The form should not know anything about HTTP.**

It shouldn't contain:

```ts
fetch(...)
```

or:

```ts
axios.post(...)
```

Instead:

```text
InstanceForm
      ↓
useCreateInstance
      ↓
instancesApi.create
      ↓
apiClient
```

That makes the whole thing much easier to test.

---

# 2. Create the Vue project

If you're starting from scratch:

```bash
pnpm create vite@latest paas-frontend --template vue-ts
cd paas-frontend
pnpm install
```

The shadcn-vue documentation currently recommends the Vite + Vue TypeScript template for this setup. ([Shadcn Vue][1])

Install the core dependencies:

```bash
pnpm add vue-router pinia zod
```

And testing:

```bash
pnpm add -D vitest @vue/test-utils jsdom
```

Vue Test Utils is the official Vue 3 testing utility and recommends Vitest for Vue projects. ([Vue Test Utils][2])

---

# 3. Install Tailwind CSS

Install:

```bash
pnpm add tailwindcss @tailwindcss/vite
```

Then `vite.config.ts`:

```ts
import path from "node:path"
import { defineConfig } from "vite"
import vue from "@vitejs/plugin-vue"
import tailwindcss from "@tailwindcss/vite"

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
  ],

  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
})
```

Tailwind's current Vite integration uses `@tailwindcss/vite`; you don't need the old Tailwind v3 `tailwind.config.js` setup for this basic configuration. ([Tailwind CSS][3])

Create:

```text
src/styles/globals.css
```

with:

```css
@import "tailwindcss";
```

Then in `main.ts`:

```ts
import { createApp } from "vue"
import App from "./App.vue"
import "./styles/globals.css"

createApp(App).mount("#app")
```

---

# 4. Install shadcn-vue

Run:

```bash
pnpm dlx shadcn-vue@latest init
```

The current CLI initializes the component configuration, dependencies, `cn` helper, and CSS variables. ([Shadcn Vue][4])

I'd choose:

```text
Base color: Neutral
```

For our form, install:

```bash
pnpm dlx shadcn-vue@latest add button
pnpm dlx shadcn-vue@latest add input
pnpm dlx shadcn-vue@latest add label
pnpm dlx shadcn-vue@latest add card
pnpm dlx shadcn-vue@latest add alert
pnpm dlx shadcn-vue@latest add badge
```

shadcn-vue provides these as source components in your project rather than making you depend on a huge monolithic UI framework. ([Shadcn Vue][1])

---

# 5. Create the folder structure

Now I'd make:

```text
src/
├── api/
│   ├── client.ts
│   ├── errors.ts
│   └── instances.ts
│
├── components/
│   ├── ui/
│   │   └── ... shadcn components
│   │
│   └── instances/
│       ├── InstanceForm.vue
│       └── InstanceStatus.vue
│
├── composables/
│   └── useCreateInstance.ts
│
├── pages/
│   └── CreateInstancePage.vue
│
├── schemas/
│   └── instance.ts
│
├── types/
│   └── instance.ts
│
├── router/
│   └── index.ts
│
├── styles/
│   └── globals.css
│
├── App.vue
└── main.ts
```

Notice something important:

**I'm not creating `services/`, `repositories/`, `domain/`, `usecases/`, etc.**

You have five endpoints.

Don't overengineer it.

---

# 6. Define the domain types

Create:

```text
src/types/instance.ts
```

```ts
export type InstanceStatus =
  | "Pending"
  | "Provisioning"
  | "Running"
  | "Failed"
  | "Deleting"
  | "Deleted"

export interface CreateInstanceRequest {
  name: string
  instances: number
  version: number
  storage: string
}

export interface Instance {
  id: string
  name: string
  version: number
  storage: string
  status: InstanceStatus
  createdAt: string
}
```

This gives us strong TypeScript types throughout the application.

---

# 7. Define the Zod schema

Now create:

```text
src/schemas/instance.ts
```

I'd put **frontend validation** here.

```ts
import * as z from "zod"

export const createInstanceSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Name is required")
    .max(63, "Name must be 63 characters or fewer")
    .regex(
      /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
      "Use lowercase letters, numbers, and hyphens only",
    ),

  instances: z
    .number()
    .int("Instances must be a whole number")
    .min(1, "At least 1 instance is required")
    .max(20, "Maximum 20 instances"),

  version: z
    .number()
    .int("Version must be a whole number")
    .min(12, "Minimum PostgreSQL version is 12")
    .max(18, "Maximum PostgreSQL version is 18"),

  storage: z
    .string()
    .trim()
    .regex(
      /^\d+(Mi|Gi|Ti)$/,
      "Storage must look like 512Mi, 7Gi, or 1Ti",
    ),
})

export type CreateInstanceForm = z.infer<
  typeof createInstanceSchema
>
```

Zod 4 is currently stable and provides schemas that map closely to TypeScript types, including `safeParse()` for validation without throwing. ([Zod][5])

### Why Zod here?

Because you now have one place saying:

```text
name
instances
version
storage
```

are valid.

Your Vue component doesn't need to contain:

```ts
if (name.length < 1)
if (name.length > 63)
if (!/regex/.test(name))
...
```

---

# 8. Important: frontend validation isn't backend validation

This distinction is important.

Your backend should **still validate everything**.

The frontend schema is for:

```text
UX
```

The backend schema is for:

```text
security + correctness
```

Don't assume that because Vue/Zod says:

```text
version = 15
```

the backend can trust it.

The backend should remain authoritative.

---

# 9. Create your API error

Create:

```text
src/api/errors.ts
```

```ts
export interface ApiErrorResponse {
  error: {
    code: string
    message: string
  }
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message)
    this.name = "ApiError"
  }
}
```

This means the rest of your frontend doesn't have to understand raw HTTP responses.

---

# 10. Build the API client

Create:

```text
src/api/client.ts
```

```ts
import { ApiError } from "./errors"

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ?? "/api/v1"

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const response = await fetch(
    `${API_BASE_URL}${path}`,
    {
      ...options,

      headers: {
        Accept: "application/json",
        ...options.headers,
      },
    },
  )

  const contentType =
    response.headers.get("content-type")

  const body = contentType?.includes("application/json")
    ? await response.json()
    : null

  if (!response.ok) {
    const error = body as
      | {
          error?: {
            code?: string
            message?: string
          }
        }
      | null

    throw new ApiError(
      response.status,
      error?.error?.code ?? "UNKNOWN_ERROR",
      error?.error?.message ??
        "Something went wrong.",
    )
  }

  return body as T
}
```

This gives you one central place for HTTP behavior.

Later, authentication can go here.

For example:

```text
Authorization
Accept
Content-Type
request IDs
error handling
401 handling
```

You don't want those duplicated in every endpoint.

---

# 11. Create the instances API

Now:

```text
src/api/instances.ts
```

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
```

That's intentionally boring.

And that's good.

The API module knows:

```text
POST /instances
```

and nothing about Vue.

Therefore it is extremely easy to unit test.

---

# 12. Generate the idempotency key in the frontend

Your endpoint specifically requires:

```http
Idempotency-Key: 8ad1fdf6-4b6e-48b0-bec2-3c6c4f386635
```

This is important.

You **should not generate a new key every time the user retries the request**.

The same logical create operation should keep the same key.

For example:

```text
User clicks Create
       │
       ▼
Generate UUID
       │
       ▼
POST /instances
Idempotency-Key: abc
       │
       ▼
network timeout
       │
       ▼
retry
       │
       ▼
POST /instances
Idempotency-Key: abc
```

Not:

```text
POST → key abc
POST → key xyz
```

because then you've turned a retry into a potentially new creation.

---

# 13. Build the composable

Create:

```text
src/composables/useCreateInstance.ts
```

```ts
import { ref } from "vue"
import { createInstance } from "@/api/instances"
import type {
  CreateInstanceRequest,
  Instance,
} from "@/types/instance"
import { ApiError } from "@/api/errors"

export function useCreateInstance() {
  const isLoading = ref(false)
  const instance = ref<Instance | null>(null)
  const error = ref<ApiError | null>(null)

  let idempotencyKey: string | null = null

  async function execute(
    request: CreateInstanceRequest,
  ) {
    isLoading.value = true
    error.value = null

    if (!idempotencyKey) {
      idempotencyKey = crypto.randomUUID()
    }

    try {
      instance.value = await createInstance(
        request,
        idempotencyKey,
      )

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

  function reset() {
    instance.value = null
    error.value = null
    idempotencyKey = null
  }

  return {
    execute,
    reset,
    isLoading,
    instance,
    error,
  }
}
```

Now the Vue component doesn't need to understand:

```text
HTTP
idempotency
API errors
loading state
```

It just says:

```ts
await execute(form)
```

This is exactly the sort of separation that makes Vue components easier to test; Vue Test Utils' own guidance recommends extracting HTTP requests and complex methods from components. ([Vue Test Utils][6])

---

# 14. Build the form

Now create:

```text
src/components/instances/InstanceForm.vue
```

I'd start with this:

```vue
<script setup lang="ts">
import { ref } from "vue"
import { createInstanceSchema } from "@/schemas/instance"
import type { CreateInstanceForm } from "@/schemas/instance"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

const emit = defineEmits<{
  submit: [value: CreateInstanceForm]
}>()

const form = ref<CreateInstanceForm>({
  name: "",
  instances: 1,
  version: 16,
  storage: "10Gi",
})

const errors = ref<
  Partial<Record<keyof CreateInstanceForm, string>>
>({})

function submit() {
  const result =
    createInstanceSchema.safeParse(form.value)

  if (!result.success) {
    const fieldErrors: typeof errors.value = {}

    for (const issue of result.error.issues) {
      const field = issue.path[0]

      if (
        typeof field === "string" &&
        field in form.value
      ) {
        fieldErrors[
          field as keyof CreateInstanceForm
        ] = issue.message
      }
    }

    errors.value = fieldErrors
    return
  }

  errors.value = {}

  emit("submit", result.data)
}
</script>

<template>
  <form
    class="space-y-6"
    @submit.prevent="submit"
  >
    <!-- Name -->

    <div class="space-y-2">
      <Label for="name">
        Instance name
      </Label>

      <Input
        id="name"
        v-model="form.name"
        name="name"
        placeholder="my-postgres-db"
        autocomplete="off"
        :aria-invalid="!!errors.name"
      />

      <p
        v-if="errors.name"
        class="text-sm text-destructive"
      >
        {{ errors.name }}
      </p>
    </div>

    <!-- Instances -->

    <div class="space-y-2">
      <Label for="instances">
        PostgreSQL instances
      </Label>

      <Input
        id="instances"
        v-model.number="form.instances"
        name="instances"
        type="number"
        min="1"
        max="20"
        :aria-invalid="!!errors.instances"
      />

      <p
        v-if="errors.instances"
        class="text-sm text-destructive"
      >
        {{ errors.instances }}
      </p>
    </div>

    <!-- Version -->

    <div class="space-y-2">
      <Label for="version">
        PostgreSQL version
      </Label>

      <Input
        id="version"
        v-model.number="form.version"
        name="version"
        type="number"
        min="12"
        max="18"
        :aria-invalid="!!errors.version"
      />

      <p
        v-if="errors.version"
        class="text-sm text-destructive"
      >
        {{ errors.version }}
      </p>
    </div>

    <!-- Storage -->

    <div class="space-y-2">
      <Label for="storage">
        Storage
      </Label>

      <Input
        id="storage"
        v-model="form.storage"
        name="storage"
        placeholder="10Gi"
        :aria-invalid="!!errors.storage"
      />

      <p
        v-if="errors.storage"
        class="text-sm text-destructive"
      >
        {{ errors.storage }}
      </p>
    </div>

    <Button
      type="submit"
      class="w-full"
    >
      Create instance
    </Button>
  </form>
</template>
```

Notice that this component **does not call the API**.

That is intentional.

---

# 15. Create the page

Now:

```text
src/pages/CreateInstancePage.vue
```

```vue
<script setup lang="ts">
import { useRouter } from "vue-router"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

import InstanceForm from "@/components/instances/InstanceForm.vue"
import { useCreateInstance } from "@/composables/useCreateInstance"

const router = useRouter()

const {
  execute,
  isLoading,
  error,
} = useCreateInstance()

async function createInstance(form: any) {
  try {
    const instance = await execute(form)

    await router.push(
      `/instances/${instance.id}`,
    )
  } catch {
    // Error is already exposed by the composable.
  }
}
</script>

<template>
  <main
    class="
      min-h-screen
      bg-muted/40
      px-4
      py-12
    "
  >
    <div class="mx-auto max-w-2xl">
      <Card>
        <CardHeader>
          <CardTitle>
            Create PostgreSQL instance
          </CardTitle>

          <CardDescription>
            Create a new CloudNativePG instance.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <div
            v-if="error"
            class="
              mb-6
              rounded-md
              border
              border-destructive/30
              bg-destructive/10
              p-4
              text-sm
              text-destructive
            "
          >
            <strong>
              {{ error.code }}
            </strong>

            <p>
              {{ error.message }}
            </p>
          </div>

          <InstanceForm
            @submit="createInstance"
          />

          <p
            v-if="isLoading"
            class="mt-4 text-sm text-muted-foreground"
          >
            Creating your PostgreSQL instance...
          </p>
        </CardContent>
      </Card>
    </div>
  </main>
</template>
```

There's one improvement I'd make immediately: don't use `any`.

Import the type:

```ts
import type { CreateInstanceForm } from "@/schemas/instance"
```

and:

```ts
async function createInstance(
  form: CreateInstanceForm,
) {
```

So the complete function becomes:

```ts
async function createInstance(
  form: CreateInstanceForm,
) {
  try {
    const instance = await execute(form)

    await router.push(
      `/instances/${instance.id}`,
    )
  } catch {
    // Error is already exposed by the composable.
  }
}
```

---

# 16. Add the route

Create:

```text
src/router/index.ts
```

```ts
import {
  createRouter,
  createWebHistory,
} from "vue-router"

import CreateInstancePage from "@/pages/CreateInstancePage.vue"

const router = createRouter({
  history: createWebHistory(),

  routes: [
    {
      path: "/instances/new",
      component: CreateInstancePage,
    },
  ],
})

export default router
```

Then:

```ts
// main.ts

import { createApp } from "vue"
import App from "./App.vue"
import router from "./router"
import "./styles/globals.css"

createApp(App)
  .use(router)
  .mount("#app")
```

And:

```vue
<!-- App.vue -->

<template>
  <RouterView />
</template>
```

---

# 17. Add an API proxy for local development

Your backend is:

```text
http://localhost:8080
```

I wouldn't hardcode that into your components.

Use Vite's proxy.

`vite.config.ts`:

```ts
import path from "node:path"
import { defineConfig } from "vite"
import vue from "@vitejs/plugin-vue"
import tailwindcss from "@tailwindcss/vite"

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
  ],

  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },

  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
})
```

Now:

```text
Browser
   │
   │ /api/v1/instances
   ▼
Vite
   │
   │ proxy
   ▼
localhost:8080
```

Your frontend doesn't need to care where the backend lives.

---

# 18. Test the Zod schema first

This is our first unit test.

Create:

```text
src/schemas/instance.test.ts
```

```ts
import { describe, expect, it } from "vitest"

import {
  createInstanceSchema,
} from "./instance"

describe("createInstanceSchema", () => {
  it("accepts a valid instance", () => {
    const result =
      createInstanceSchema.safeParse({
        name: "my-db-2",
        instances: 2,
        version: 15,
        storage: "7Gi",
      })

    expect(result.success).toBe(true)
  })

  it("rejects an empty name", () => {
    const result =
      createInstanceSchema.safeParse({
        name: "",
        instances: 2,
        version: 15,
        storage: "7Gi",
      })

    expect(result.success).toBe(false)
  })

  it("rejects invalid storage", () => {
    const result =
      createInstanceSchema.safeParse({
        name: "my-db",
        instances: 2,
        version: 15,
        storage: "seven gigs",
      })

    expect(result.success).toBe(false)
  })

  it("rejects zero instances", () => {
    const result =
      createInstanceSchema.safeParse({
        name: "my-db",
        instances: 0,
        version: 15,
        storage: "7Gi",
      })

    expect(result.success).toBe(false)
  })
})
```

Run:

```bash
pnpm vitest
```

Vitest is specifically designed around Vite and is the recommended test runner for Vue Test Utils. ([Vitest][7])

---

# 19. Test the API independently

This is where the architecture starts paying off.

Create:

```text
src/api/instances.test.ts
```

```ts
import {
  afterEach,
  describe,
  expect,
  it,
  vi,
} from "vitest"

import { createInstance } from "./instances"

describe("createInstance", () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("sends the correct request", async () => {
    const response = {
      id: "inst-7f3a2b",
      name: "my-db-2",
      version: 15,
      storage: "7Gi",
      status: "Pending",
      createdAt: "2026-08-04T09:12:00Z",
    }

    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(
        new Response(
          JSON.stringify(response),
          {
            status: 202,
            headers: {
              "Content-Type":
                "application/json",
            },
          },
        ),
      )

    await createInstance(
      {
        name: "my-db-2",
        instances: 2,
        version: 15,
        storage: "7Gi",
      },
      "test-idempotency-key",
    )

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/instances",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          "Content-Type":
            "application/json",
          "Idempotency-Key":
            "test-idempotency-key",
        }),
      }),
    )
  })

  it("returns the created instance", async () => {
    const response = {
      id: "inst-7f3a2b",
      name: "my-db-2",
      version: 15,
      storage: "7Gi",
      status: "Pending",
      createdAt: "2026-08-04T09:12:00Z",
    }

    vi.spyOn(globalThis, "fetch")
      .mockResolvedValue(
        new Response(
          JSON.stringify(response),
          {
            status: 202,
            headers: {
              "Content-Type":
                "application/json",
            },
          },
        ),
      )

    const result = await createInstance(
      {
        name: "my-db-2",
        instances: 2,
        version: 15,
        storage: "7Gi",
      },
      "test-key",
    )

    expect(result).toEqual(response)
  })
})
```

This test doesn't need Vue.

That's exactly what we want.

---

# 20. Test 409

Your API specifically has:

```text
409 Conflict
```

for an existing instance name.

That's worth testing.

```ts
it("throws an ApiError for a conflict", async () => {
  vi.spyOn(globalThis, "fetch")
    .mockResolvedValue(
      new Response(
        JSON.stringify({
          error: {
            code: "INSTANCE_ALREADY_EXISTS",
            message:
              "An instance with this name already exists.",
          },
        }),
        {
          status: 409,
          headers: {
            "Content-Type":
              "application/json",
          },
        },
      ),
    )

  await expect(
    createInstance(
      {
        name: "my-db",
        instances: 1,
        version: 16,
        storage: "10Gi",
      },
      "test-key",
    ),
  ).rejects.toMatchObject({
    status: 409,
    code: "INSTANCE_ALREADY_EXISTS",
  })
})
```

Now your API contract has a test.

---

# 21. Test the Vue form

Now test the actual user interaction.

Create:

```text
src/components/instances/InstanceForm.test.ts
```

```ts
import {
  describe,
  expect,
  it,
} from "vitest"

import {
  mount,
} from "@vue/test-utils"

import InstanceForm from "./InstanceForm.vue"

describe("InstanceForm", () => {
  it("submits valid data", async () => {
    const wrapper = mount(InstanceForm)

    await wrapper
      .get('input[name="name"]')
      .setValue("my-db-2")

    await wrapper
      .get('input[name="instances"]')
      .setValue("2")

    await wrapper
      .get('input[name="version"]')
      .setValue("15")

    await wrapper
      .get('input[name="storage"]')
      .setValue("7Gi")

    await wrapper
      .get("form")
      .trigger("submit")

    expect(
      wrapper.emitted("submit"),
    ).toEqual([
      [
        {
          name: "my-db-2",
          instances: 2,
          version: 15,
          storage: "7Gi",
        },
      ],
    ])
  })

  it("shows validation errors", async () => {
    const wrapper = mount(InstanceForm)

    await wrapper
      .get("form")
      .trigger("submit")

    expect(
      wrapper.text(),
    ).toContain("Name is required")

    expect(
      wrapper.text(),
    ).toContain(
      "At least 1 instance is required",
    )

    expect(
      wrapper.emitted("submit"),
    ).toBeUndefined()
  })
})
```

This is the kind of test you want: **interaction → observable result**.

Not:

```ts
expect(wrapper.vm.form.name).toBe(...)
```

Vue Test Utils explicitly recommends testing inputs, interactions, rendered output and side effects rather than implementation details. ([Vue Test Utils][6])

---

# 22. Add Playwright

Install it:

```bash
pnpm create playwright
```

Or add it to the existing project using the Playwright installer. The official setup creates `playwright.config.ts`, a test directory and browser configuration. ([Playwright][8])

I'd use:

```text
e2e/
├── instances/
│   └── create-instance.spec.ts
└── fixtures/
```

---

# 23. Your first E2E test

Create:

```text
e2e/instances/create-instance.spec.ts
```

```ts
import {
  test,
  expect,
} from "@playwright/test"

test("user can create a PostgreSQL instance", async ({
  page,
}) => {
  await page.goto("/instances/new")

  await page
    .getByLabel("Instance name")
    .fill("my-db-2")

  await page
    .getByLabel("PostgreSQL instances")
    .fill("2")

  await page
    .getByLabel("PostgreSQL version")
    .fill("15")

  await page
    .getByLabel("Storage")
    .fill("7Gi")

  await page
    .getByRole("button", {
      name: "Create instance",
    })
    .click()

  await expect(
    page,
  ).toHaveURL(
    /\/instances\/inst-/,
  )
})
```

Playwright's locator strategy is designed around user-facing selectors such as roles and labels, which is exactly what we want here. ([Playwright][8])

---

# 24. But don't let E2E hit your production-like database

For this test:

```text
POST /instances
```

you need a deterministic backend.

I'd eventually run:

```text
Playwright
    │
    ▼
Frontend
    │
    ▼
Test API
    │
    ▼
Test PostgreSQL
```

For your PaaS, you could eventually have a dedicated test Kubernetes namespace:

```text
paas-e2e
├── api
├── postgres
└── cnpg operator
```

But don't start there.

For local frontend development, you can mock the backend request with Playwright:

```ts
await page.route(
  "**/api/v1/instances",
  async (route) => {
    await route.fulfill({
      status: 202,

      contentType:
        "application/json",

      body: JSON.stringify({
        id: "inst-test-123",
        name: "my-db-2",
        version: 15,
        storage: "7Gi",
        status: "Pending",
        createdAt:
          "2026-08-04T09:12:00Z",
      }),
    })
  },
)
```

Then the test is testing:

```text
Vue
routing
form
validation
HTTP request
success state
```

without needing Kubernetes.

Later, add a smaller number of true integration E2E tests against the real backend.

---

# 25. Test your idempotency behavior

This endpoint has a particularly interesting feature:

```text
Idempotency-Key
```

I'd absolutely test it.

At the API layer:

```text
same request
same idempotency key
        ↓
same instance
```

and:

```text
different request
same idempotency key
        ↓
should NOT create another instance
```

That is primarily a **backend integration test**, not a Vue test.

But the frontend should ensure the same key survives a retry.

That's why I put the key inside:

```ts
useCreateInstance()
```

rather than:

```ts
createInstance()
```

being responsible for generating it every time.

---

# 26. Handle your `202` response properly

Your API says:

```http
202 Accepted
```

and:

```json
{
  "status": "Pending"
}
```

That tells me your creation is asynchronous.

So don't make the frontend pretend that:

```text
POST /instances
```

means:

```text
database is ready
```

It means:

```text
request accepted
provisioning started
```

Your UI should say something like:

```text
Instance created

my-db-2

Status
Pending

Your PostgreSQL instance is being provisioned.
```

Then your **next endpoint** should probably be something like:

```http
GET /api/v1/instances/{id}
```

and the frontend can poll or refresh the instance status.

That means your UX becomes:

```text
Create
  │
  ▼
202 Accepted
  │
  ▼
Instance Details
  │
  ├── Pending
  │
  ├── Provisioning
  │
  └── Running
```

This is much better than keeping the user on the form.

---

# 27. One API-contract issue I'd fix

Your documentation says:

```text
200
Idempotent replay
```

and:

```text
202
Request accepted
```

That's perfectly reasonable.

But your `Location` example is:

```text
/v1/instances/inst-7f3a2b
```

while your actual API appears to be:

```text
/api/v1/instances
```

I'd make the API documentation consistent:

```http
Location: /api/v1/instances/inst-7f3a2b
```

or, if `/v1` is intentionally exposed differently, document that explicitly.

Also, I'd make your error examples distinct.

Right now everything says:

```json
{
  "error": {
    "code": "INSTANCE_NOT_FOUND",
    "message": "Product instance does not exist."
  }
}
```

even for:

```text
400
401
409
422
500
```

That's going to make frontend error handling unnecessarily difficult.

I'd instead define something like:

### 400

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Request body is malformed."
  }
}
```

### 401

```json
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "Authentication is required."
  }
}
```

### 409

```json
{
  "error": {
    "code": "INSTANCE_NAME_CONFLICT",
    "message": "An instance with this name already exists."
  }
}
```

### 422

```json
{
  "error": {
    "code": "INVALID_INSTANCE_CONFIGURATION",
    "message": "The requested PostgreSQL version is not supported."
  }
}
```

### 500

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred."
  }
}
```

That lets the frontend make meaningful decisions.

---

# 28. Add a better error mapper

Once your backend has stable error codes, you can have:

```text
src/api/error-message.ts
```

```ts
import type { ApiError } from "./errors"

export function getInstanceErrorMessage(
  error: ApiError,
): string {
  switch (error.code) {
    case "INSTANCE_NAME_CONFLICT":
      return "An instance with this name already exists."

    case "INVALID_INSTANCE_CONFIGURATION":
      return error.message

    case "UNAUTHENTICATED":
      return "Please sign in again."

    default:
      return "Unable to create the instance. Please try again."
  }
}
```

Your page then doesn't need to know API codes.

---

# 29. Add loading state to the button

One thing I'd change in our initial form: let the parent tell it that submission is happening.

In `InstanceForm.vue`:

```ts
const props = defineProps<{
  loading?: boolean
}>()
```

Then:

```vue
<Button
  type="submit"
  class="w-full"
  :disabled="props.loading"
>
  <span v-if="props.loading">
    Creating...
  </span>

  <span v-else>
    Create instance
  </span>
</Button>
```

And:

```vue
<InstanceForm
  :loading="isLoading"
  @submit="createInstance"
/>
```

Now you can't accidentally submit the form ten times.

Your idempotency key is still the safety net, but disabling the button is good UX.

---

# 30. Add a status component

Create:

```text
src/components/instances/InstanceStatus.vue
```

```vue
<script setup lang="ts">
import { Badge } from "@/components/ui/badge"
import type { InstanceStatus } from "@/types/instance"

defineProps<{
  status: InstanceStatus
}>()
</script>

<template>
  <Badge
    :variant="
      status === 'Running'
        ? 'default'
        : status === 'Failed'
          ? 'destructive'
          : 'secondary'
    "
  >
    {{ status }}
  </Badge>
</template>
```

Then every page can use the same status presentation.

Eventually you can make this smarter:

```text
Pending       → gray
Provisioning  → blue
Running       → green
Failed        → red
Deleting      → orange
```

This is where a component system starts paying off.

---

# 31. Your resulting architecture

After this one endpoint, you've arrived at:

```text
src/
│
├── api/
│   ├── client.ts              ← HTTP
│   ├── errors.ts              ← HTTP errors
│   └── instances.ts           ← endpoint
│
├── schemas/
│   └── instance.ts            ← Zod validation
│
├── types/
│   └── instance.ts            ← TypeScript types
│
├── composables/
│   └── useCreateInstance.ts   ← application state
│
├── components/
│   ├── ui/                    ← shadcn-vue
│   │
│   └── instances/
│       ├── InstanceForm.vue
│       └── InstanceStatus.vue
│
├── pages/
│   └── CreateInstancePage.vue
│
├── router/
│   └── index.ts
│
└── styles/
    └── globals.css
```

And the responsibility boundaries are:

```text
┌─────────────────────────────┐
│ CreateInstancePage          │
│                             │
│ Orchestration               │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│ useCreateInstance           │
│                             │
│ loading / error / result    │
│ idempotency                 │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│ instances.ts                │
│                             │
│ HTTP endpoint               │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│ api/client.ts               │
│                             │
│ HTTP infrastructure         │
└─────────────────────────────┘
```

while:

```text
InstanceForm
     │
     ▼
Zod schema
     │
     ▼
emit("submit", data)
```

The form doesn't know the backend exists.

---

# 32. Testing pyramid for this endpoint

I'd aim for this:

```text
                   ┌───────────────┐
                   │  Playwright   │
                   │               │
                   │  1–3 tests   │
                   └───────┬───────┘
                           │
                 ┌─────────┴─────────┐
                 │                   │
          component tests       API integration
          2–5 tests             tests
                 │
        ┌────────┴────────┐
        │                 │
     Zod tests         API unit
     4–6 tests         tests
```

You don't need 50 E2E tests for five endpoints.

I'd have **lots of cheap unit tests**, some component tests, and a small number of high-value E2E tests.

---

# 33. Package scripts

I'd eventually have:

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc -b && vite build",
    "preview": "vite preview",

    "lint": "eslint .",

    "test": "vitest run",
    "test:watch": "vitest",

    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",

    "check": "pnpm lint && pnpm test && pnpm build"
  }
}
```

Then locally:

```bash
pnpm check
```

means:

```text
lint
  ↓
unit tests
  ↓
TypeScript
  ↓
production build
```

And CI:

```text
PR
 │
 ├── lint
 ├── unit tests
 ├── build
 └── Playwright
```

---

# 34. What I'd do next

I would **not** implement all five endpoints at once.

Implement this vertical slice completely:

```text
POST /instances
```

including:

```text
✓ Vue page
✓ shadcn form
✓ Tailwind styling
✓ Zod validation
✓ API client
✓ API error handling
✓ Idempotency
✓ loading state
✓ 202 handling
✓ unit tests
✓ component tests
✓ Playwright test
```

Then implement:

```text
GET /instances/:id
```

and connect it to the `202` result.

That will give you the foundation for the entire PaaS:

```text
POST /instances
       │
       ▼
GET /instances/:id
       │
       ▼
status polling
       │
       ▼
Running
```

Then the other endpoints can follow exactly the same architecture.

One final recommendation: **if you already have an OpenAPI document for these endpoints, use it as the source of truth and generate your TypeScript API types/client rather than manually duplicating the request/response types in Vue.** Your current API documentation is already structured enough that this would be a very good fit.

[1]: https://www.shadcn-vue.com/docs/installation/vite?utm_source=chatgpt.com "Vite - shadcn/vue"
[2]: https://test-utils.vuejs.org/installation/?utm_source=chatgpt.com "Installation | Vue Test Utils"
[3]: https://tailwindcss.com/docs/installation/using-vite?trk=public_post-text&utm_source=chatgpt.com "Installing Tailwind CSS with Vite - Tailwind CSS"
[4]: https://www.shadcn-vue.com/docs/cli?utm_source=chatgpt.com "CLI - shadcn/vue"
[5]: https://zod.dev/packages/zod?utm_source=chatgpt.com "Zod | Zod"
[6]: https://test-utils.vuejs.org/guide/essentials/easy-to-test.html?utm_source=chatgpt.com "Write components that are easy to test | Vue Test Utils"
[7]: https://vitest.dev/guide/?utm_source=chatgpt.com "Getting Started | Guide | Vitest"
[8]: https://playwright.dev/docs/intro?utm_source=chatgpt.com "Installation | Playwright"
