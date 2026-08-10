<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'

import DashboardStats from '@/components/dashboard/DashboardStats.vue'
import RecentInstances from '@/components/dashboard/RecentInstances.vue'
import GettingStarted from '@/components/dashboard/GettingStarted.vue'

import { useInstances } from '@/composables/useInstances'

const router = useRouter()

const { instances, isLoading, error, fetchInstances } = useInstances()

function createInstance() {
  router.push({
    name: 'instance-create',
  })
}

function viewInstances() {
  router.push({
    name: 'instances',
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

      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">Dashboard</h1>

          <p class="mt-1 text-sm text-muted-foreground">Manage your PostgreSQL instances.</p>
        </div>

        <Button @click="createInstance">
          <Plus class="mr-2 h-4 w-4" />

          Create instance
        </Button>
      </div>

      <!-- Error -->

      <div
        v-if="error"
        class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive"
      >
        {{ error.message }}
      </div>

      <!-- Statistics -->

      <DashboardStats :instances="instances" :is-loading="isLoading" />

      <!-- Recent instances -->

      <RecentInstances :instances="instances" :is-loading="isLoading" @view-all="viewInstances" />

      <!-- Getting started -->

      <GettingStarted v-if="!isLoading && instances.length === 0" @create="createInstance" />
    </div>
  </div>
</template>
