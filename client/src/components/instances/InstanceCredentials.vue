<script setup lang="ts">
import { computed, ref } from 'vue'
import { Check, Copy, Eye, EyeOff } from 'lucide-vue-next'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import { buildPostgresUri } from '@/lib/postgres'

interface InstanceCredentials {
  host: string
  port: number
  database: string
  username: string
  password: string
}

const props = withDefaults(
  defineProps<{
    credentials: InstanceCredentials | null
    loading?: boolean
    error?: Error | null
  }>(),
  {
    loading: false,
    error: null,
  },
)

const emit = defineEmits<{
  reveal: []
}>()

const showPassword = ref(false)
const copied = ref<string | null>(null)

const connectionUri = computed(() => {
  if (!props.credentials) {
    return ''
  }

  return buildPostgresUri(props.credentials)
})

async function copy(value: string, field: string) {
  try {
    await navigator.clipboard.writeText(value)

    copied.value = field

    window.setTimeout(() => {
      copied.value = null
    }, 1500)
  } catch {
    // Clipboard errors are intentionally ignored.
  }
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>Database credentials</CardTitle>

      <CardDescription>
        Use these credentials to connect to your PostgreSQL instance.
      </CardDescription>
    </CardHeader>

    <CardContent>
      <!-- Credentials not revealed -->

      <div
        v-if="!credentials && !error"
        class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
      >
        <div>
          <p class="text-sm font-medium">
            Credentials are hidden
          </p>

          <p class="text-sm text-muted-foreground">
            Reveal the credentials to connect to this database.
          </p>
        </div>

        <Button
          :disabled="loading"
          @click="emit('reveal')"
        >
          {{ loading ? 'Loading...' : 'Reveal credentials' }}
        </Button>
      </div>

      <!-- Error -->

      <div
        v-else-if="error"
        class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
      >
        <p class="text-sm text-destructive">
          {{ error.message }}
        </p>

        <Button
          variant="outline"
          :disabled="loading"
          @click="emit('reveal')"
        >
          {{ loading ? 'Loading...' : 'Try again' }}
        </Button>
      </div>

      <!-- Credentials -->

      <div v-else-if="credentials" class="space-y-6">
        <!-- Connection URI -->

        <div class="space-y-2">
          <div class="flex items-center justify-between gap-2">
            <div>
              <label class="text-sm font-medium">
                PostgreSQL connection URI
              </label>

              <p class="text-xs text-muted-foreground">
                Copy this URI directly into your application.
              </p>
            </div>

            <Button
              variant="outline"
              size="sm"
              @click="copy(connectionUri, 'uri')"
            >
              <Check
                v-if="copied === 'uri'"
                class="mr-2 h-4 w-4"
              />

              <Copy
                v-else
                class="mr-2 h-4 w-4"
              />

              {{ copied === 'uri' ? 'Copied' : 'Copy' }}
            </Button>
          </div>

          <div class="rounded-md border bg-muted/50 p-3">
            <code class="block break-all font-mono text-sm">
              postgresql://{{ credentials.username }}:••••••••@{{ credentials.host }}:{{
                credentials.port
              }}/{{ credentials.database }}
            </code>
          </div>
        </div>

        <!-- Host -->

        <div class="space-y-2">
          <label class="text-sm font-medium">
            Host
          </label>

          <div class="flex gap-2">
            <Input
              :model-value="credentials.host"
              readonly
              class="font-mono"
            />

            <Button
              variant="outline"
              size="icon"
              title="Copy host"
              @click="copy(credentials.host, 'host')"
            >
              <Check
                v-if="copied === 'host'"
                class="h-4 w-4"
              />

              <Copy
                v-else
                class="h-4 w-4"
              />
            </Button>
          </div>
        </div>

        <!-- Port -->

        <div class="space-y-2">
          <label class="text-sm font-medium">
            Port
          </label>

          <Input
            :model-value="String(credentials.port)"
            readonly
            class="font-mono"
          />
        </div>

        <!-- Database -->

        <div class="space-y-2">
          <label class="text-sm font-medium">
            Database
          </label>

          <div class="flex gap-2">
            <Input
              :model-value="credentials.database"
              readonly
              class="font-mono"
            />

            <Button
              variant="outline"
              size="icon"
              title="Copy database name"
              @click="copy(credentials.database, 'database')"
            >
              <Check
                v-if="copied === 'database'"
                class="h-4 w-4"
              />

              <Copy
                v-else
                class="h-4 w-4"
              />
            </Button>
          </div>
        </div>

        <!-- Username -->

        <div class="space-y-2">
          <label class="text-sm font-medium">
            Username
          </label>

          <div class="flex gap-2">
            <Input
              :model-value="credentials.username"
              readonly
              class="font-mono"
            />

            <Button
              variant="outline"
              size="icon"
              title="Copy username"
              @click="copy(credentials.username, 'username')"
            >
              <Check
                v-if="copied === 'username'"
                class="h-4 w-4"
              />

              <Copy
                v-else
                class="h-4 w-4"
              />
            </Button>
          </div>
        </div>

        <!-- Password -->

        <div class="space-y-2">
          <label class="text-sm font-medium">
            Password
          </label>

          <div class="flex gap-2">
            <Input
              :model-value="credentials.password"
              :type="showPassword ? 'text' : 'password'"
              readonly
              autocomplete="off"
              class="font-mono"
            />

            <Button
              variant="outline"
              size="icon"
              :title="showPassword ? 'Hide password' : 'Show password'"
              @click="showPassword = !showPassword"
            >
              <EyeOff
                v-if="showPassword"
                class="h-4 w-4"
              />

              <Eye
                v-else
                class="h-4 w-4"
              />
            </Button>

            <Button
              variant="outline"
              size="icon"
              title="Copy password"
              @click="copy(credentials.password, 'password')"
            >
              <Check
                v-if="copied === 'password'"
                class="h-4 w-4"
              />

              <Copy
                v-else
                class="h-4 w-4"
              />
            </Button>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>