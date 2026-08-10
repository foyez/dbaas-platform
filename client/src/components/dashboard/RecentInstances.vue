<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

import InstanceStatus from '@/components/instances/InstanceStatus.vue'

import { formatDate } from '@/lib/date'

import type { Instance } from '@/types/instance'

const props = defineProps<{
  instances: Instance[]
  isLoading: boolean
}>()

defineEmits<{
  viewAll: []
}>()

function recentInstances(): Instance[] {
  return [...props.instances]
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, 5)
}
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between">
      <div>
        <CardTitle> Recent instances </CardTitle>

        <CardDescription> Your most recently created instances. </CardDescription>
      </div>

      <Button variant="ghost" size="sm" @click="$emit('viewAll')">
        View all

        <ArrowRight class="ml-2 h-4 w-4" />
      </Button>
    </CardHeader>

    <CardContent>
      <!-- Loading -->

      <div v-if="isLoading" class="space-y-4">
        <div v-for="index in 3" :key="index" class="flex items-center justify-between">
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
        class="py-8 text-center text-sm text-muted-foreground"
      >
        No instances yet.
      </div>

      <!-- Instances -->

      <div v-else class="divide-y">
        <RouterLink
          v-for="instance in recentInstances()"
          :key="instance.id"
          :to="{
            name: 'instance-detail',
            params: {
              id: instance.id,
            },
          }"
          class="flex items-center justify-between gap-4 py-4 first:pt-0 last:pb-0 hover:bg-muted/50"
        >
          <div class="min-w-0">
            <p class="truncate font-medium">
              {{ instance.name }}
            </p>

            <div class="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
              <span class="font-mono">
                {{ instance.id }}
              </span>

              <span> PostgreSQL {{ instance.version }} </span>

              <span>
                {{ formatDate(instance.createdAt) }}
              </span>
            </div>
          </div>

          <InstanceStatus :status="instance.status" />
        </RouterLink>
      </div>
    </CardContent>
  </Card>
</template>
