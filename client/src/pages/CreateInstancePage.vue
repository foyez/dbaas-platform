<script setup lang="ts">
import { useRouter } from 'vue-router'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

import CreateInstanceForm from '@/components/instances/CreateInstanceForm.vue'
import { useCreateInstance } from '@/composables/useCreateInstance'
import type { CreateInstanceInput } from '@/schemas/instance'

const router = useRouter()

const { execute, isLoading, error } = useCreateInstance()

async function createInstance(form: CreateInstanceInput) {
  try {
    const instance = await execute(form)

    await router.push({
      name: 'instance-detail',
      params: {
        id: instance.id,
      },
    })
  } catch {
    // Error is already exposed by the composable.
  }
}
</script>

<template>
  <main class="min-h-screen bg-muted/40 px-4 py-12">
    <div class="mx-auto max-w-2xl">
      <Card>
        <CardHeader>
          <CardTitle> Create PostgreSQL instance </CardTitle>

          <CardDescription> Create a new CloudNativePG instance. </CardDescription>
        </CardHeader>

        <CardContent>
          <div
            v-if="error"
            class="mb-6 rounded-md border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive"
          >
            <strong>
              {{ error.code }}
            </strong>

            <p>
              {{ error.message }}
            </p>
          </div>

          <CreateInstanceForm :loading="isLoading" @submit="createInstance" />

          <p v-if="isLoading" class="mt-4 text-sm text-muted-foreground">
            Creating your PostgreSQL instance...
          </p>
        </CardContent>
      </Card>
    </div>
  </main>
</template>
