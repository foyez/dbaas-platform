<script setup lang="ts">
import { Skeleton } from '@/components/ui/skeleton'
</script>

<template>
  <div class="space-y-3">
    <div
      v-for="index in 5"
      :key="index"
      class="flex items-center justify-between rounded-lg border bg-background p-4"
    >
      <div class="space-y-2">
        <Skeleton class="h-4 w-40" />
        <Skeleton class="h-3 w-24" />
      </div>

      <Skeleton class="h-6 w-20" />
    </div>
  </div>
</template>
