<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

import { Button } from '@/components/ui/button'

import InstanceStatus from '@/components/instances/InstanceStatus.vue'

import { useInstance } from '@/composables/useInstance'
import { formatDate } from '@/lib/date'

const route = useRoute()
const router = useRouter()

const { instance, isLoading, error, fetchInstance } = useInstance()

const instanceId = String(route.params.id)

async function load() {
  try {
    await fetchInstance(instanceId)
  } catch {
    // Error is exposed by useInstance().
  }
}

function goBack() {
  router.push('/instances/new')
}

onMounted(load)
</script>

<template>
  <main class="min-h-screen bg-muted/40 px-4 py-8">
    <div class="mx-auto max-w-4xl space-y-6">
      <!-- Back -->

      <Button variant="ghost" class="-ml-2" @click="goBack"> ← Back </Button>

      <!-- Loading -->

      <Card v-if="isLoading">
        <CardContent class="py-12">
          <div class="flex items-center justify-center text-sm text-muted-foreground">
            Loading instance...
          </div>
        </CardContent>
      </Card>

      <!-- Error -->

      <Card v-else-if="error">
        <CardHeader>
          <CardTitle> Unable to load instance </CardTitle>

          <CardDescription>
            {{ error.message }}
          </CardDescription>
        </CardHeader>

        <CardContent>
          <Button @click="load"> Try again </Button>
        </CardContent>
      </Card>

      <!-- Instance -->

      <template v-else-if="instance">
        <!-- Header -->

        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <div class="flex items-center gap-3">
              <h1 class="text-2xl font-semibold tracking-tight">
                {{ instance.name }}
              </h1>

              <InstanceStatus :status="instance.status" />
            </div>

            <p class="mt-1 font-mono text-sm text-muted-foreground">
              {{ instance.id }}
            </p>
          </div>
        </div>

        <!-- Configuration -->

        <Card>
          <CardHeader>
            <CardTitle> Configuration </CardTitle>

            <CardDescription> PostgreSQL instance configuration. </CardDescription>
          </CardHeader>

          <CardContent>
            <dl class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              <div>
                <dt class="text-sm text-muted-foreground">PostgreSQL version</dt>

                <dd class="mt-1 text-lg font-medium">
                  {{ instance.version }}
                </dd>
              </div>

              <div>
                <dt class="text-sm text-muted-foreground">Storage</dt>

                <dd class="mt-1 text-lg font-medium">
                  {{ instance.storage }}
                </dd>
              </div>

              <div>
                <dt class="text-sm text-muted-foreground">Status</dt>

                <dd class="mt-2">
                  <InstanceStatus :status="instance.status" />
                </dd>
              </div>
            </dl>
          </CardContent>
        </Card>

        <!-- Metadata -->

        <Card>
          <CardHeader>
            <CardTitle> Details </CardTitle>
          </CardHeader>

          <CardContent>
            <dl class="space-y-4">
              <div class="flex flex-col gap-1 sm:flex-row sm:justify-between">
                <dt class="text-sm text-muted-foreground">Instance ID</dt>

                <dd class="font-mono text-sm">
                  {{ instance.id }}
                </dd>
              </div>

              <div class="flex flex-col gap-1 sm:flex-row sm:justify-between">
                <dt class="text-sm text-muted-foreground">Created</dt>

                <dd class="text-sm">
                  {{ formatDate(instance.createdAt) }}
                </dd>
              </div>
            </dl>
          </CardContent>
        </Card>
      </template>
    </div>
  </main>
</template>
