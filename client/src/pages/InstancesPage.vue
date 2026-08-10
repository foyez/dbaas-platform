<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, RefreshCw } from 'lucide-vue-next'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

import { Button } from '@/components/ui/button'

import InstanceStatus from '@/components/instances/InstanceStatus.vue'

import { useInstances } from '@/composables/useInstances'

import { formatDate } from '@/lib/date'

const router = useRouter()

const { instances, isLoading, error, isEmpty, fetchInstances } = useInstances()

function createInstance() {
  router.push('/instances/new')
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
  <main class="min-h-screen bg-muted/40 px-4 py-8">
    <div class="mx-auto max-w-6xl space-y-6">
      <!-- Header -->

      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">Instances</h1>

          <p class="mt-1 text-sm text-muted-foreground">Manage your PostgreSQL instances.</p>
        </div>

        <Button @click="createInstance">
          <Plus class="mr-2 h-4 w-4" />

          Create instance
        </Button>
      </div>

      <!-- Loading -->

      <Card v-if="isLoading">
        <CardContent class="py-12">
          <div class="flex items-center justify-center text-sm text-muted-foreground">
            Loading instances...
          </div>
        </CardContent>
      </Card>

      <!-- Error -->

      <Card v-else-if="error">
        <CardHeader>
          <CardTitle> Unable to load instances </CardTitle>

          <CardDescription>
            {{ error.message }}
          </CardDescription>
        </CardHeader>

        <CardContent>
          <Button variant="outline" @click="load">
            <RefreshCw class="mr-2 h-4 w-4" />

            Try again
          </Button>
        </CardContent>
      </Card>

      <!-- Empty -->

      <Card v-else-if="isEmpty">
        <CardContent class="flex flex-col items-center justify-center py-16 text-center">
          <div class="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
            <Plus class="h-5 w-5" />
          </div>

          <h2 class="mt-4 font-semibold">No instances yet</h2>

          <p class="mt-1 max-w-sm text-sm text-muted-foreground">
            Create your first PostgreSQL instance to get started.
          </p>

          <Button class="mt-6" @click="createInstance">
            <Plus class="mr-2 h-4 w-4" />

            Create instance
          </Button>
        </CardContent>
      </Card>

      <!-- Instances -->

      <Card v-else>
        <CardHeader>
          <CardTitle> Your instances </CardTitle>

          <CardDescription>
            {{ instances.length }}
            {{ instances.length === 1 ? 'instance' : 'instances' }}
          </CardDescription>
        </CardHeader>

        <CardContent class="p-0">
          <div class="overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b text-left text-sm text-muted-foreground">
                  <th class="px-6 py-3 font-medium">Name</th>

                  <th class="px-6 py-3 font-medium">Status</th>

                  <th class="hidden px-6 py-3 font-medium sm:table-cell">PostgreSQL</th>

                  <th class="hidden px-6 py-3 font-medium md:table-cell">Storage</th>

                  <th class="hidden px-6 py-3 font-medium lg:table-cell">Created</th>
                </tr>
              </thead>

              <tbody>
                <tr
                  v-for="instance in instances"
                  :key="instance.id"
                  class="cursor-pointer border-b last:border-0 hover:bg-muted/50"
                  @click="openInstance(instance.id)"
                >
                  <td class="px-6 py-4">
                    <RouterLink
                      :to="{
                        name: 'instance-detail',
                        params: {
                          id: instance.id,
                        },
                      }"
                      class="font-medium text-foreground hover:underline"
                    >
                      {{ instance.name }}
                    </RouterLink>

                    <div class="mt-1 font-mono text-xs text-muted-foreground">
                      {{ instance.id }}
                    </div>
                  </td>

                  <td class="px-6 py-4">
                    <InstanceStatus :status="instance.status" />
                  </td>

                  <td class="hidden px-6 py-4 text-sm sm:table-cell">
                    PostgreSQL
                    {{ instance.version }}
                  </td>

                  <td class="hidden px-6 py-4 text-sm md:table-cell">
                    {{ instance.storage }}
                  </td>

                  <td class="hidden px-6 py-4 text-sm text-muted-foreground lg:table-cell">
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
