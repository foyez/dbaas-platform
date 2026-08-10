<script setup lang="ts">
import { Database, CircleCheck, LoaderCircle } from 'lucide-vue-next'

import { Card, CardContent } from '@/components/ui/card'

import { Skeleton } from '@/components/ui/skeleton'

import type { Instance } from '@/types/instance'

const props = defineProps<{
  instances: Instance[]
  isLoading: boolean
}>()

const runningCount = () =>
  props.instances.filter((instance) => instance.status === 'Running').length

const provisioningCount = () =>
  props.instances.filter(
    (instance) => instance.status === 'Pending' || instance.status === 'Provisioning',
  ).length
</script>

<template>
  <div class="grid gap-4 sm:grid-cols-3">
    <!-- Total -->

    <Card>
      <CardContent class="p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted-foreground">Instances</p>

            <Skeleton v-if="isLoading" class="mt-2 h-8 w-12" />

            <p v-else class="mt-1 text-3xl font-semibold">
              {{ instances.length }}
            </p>
          </div>

          <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
            <Database class="h-5 w-5" />
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Running -->

    <Card>
      <CardContent class="p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted-foreground">Running</p>

            <Skeleton v-if="isLoading" class="mt-2 h-8 w-12" />

            <p v-else class="mt-1 text-3xl font-semibold">
              {{ runningCount() }}
            </p>
          </div>

          <div
            class="flex h-10 w-10 items-center justify-center rounded-lg bg-green-100 text-green-700 dark:bg-green-950 dark:text-green-400"
          >
            <CircleCheck class="h-5 w-5" />
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Provisioning -->

    <Card>
      <CardContent class="p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted-foreground">Provisioning</p>

            <Skeleton v-if="isLoading" class="mt-2 h-8 w-12" />

            <p v-else class="mt-1 text-3xl font-semibold">
              {{ provisioningCount() }}
            </p>
          </div>

          <div
            class="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-400"
          >
            <LoaderCircle class="h-5 w-5" />
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
