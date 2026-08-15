<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Pencil, Trash2 } from 'lucide-vue-next'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

import { Button } from '@/components/ui/button'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'

import InstanceStatus from '@/components/instances/InstanceStatus.vue'
import InstanceForm from '@/components/instances/InstanceForm.vue'
import InstanceCredentials from '@/components/instances/InstanceCredentials.vue'

import { useInstance } from '@/composables/useInstance'
import type { UpdateInstanceForm } from '@/schemas/instance'
import { formatDate } from '@/lib/date'

const route = useRoute()
const router = useRouter()

const {
  instance,
  instanceCredentials,
  isLoading,
  isLoadingCredentials,
  error,
  credentialsError,
  fetchInstance,
  fetchCredentials,
  updateInstance,
  deleteInstance,
} = useInstance()

const instanceId = String(route.params.id)

const isEditing = ref(false)
const isDeleting = ref(false)

async function load() {
  try {
    await fetchInstance(instanceId)
  } catch {
    // Error is exposed by useInstance().
  }
}

async function revealCredentials() {
  try {
    await fetchCredentials(instanceId)
  } catch {
    // Error is exposed by useInstance().
  }
}

async function saveChanges(data: UpdateInstanceForm) {
  try {
    await updateInstance(instanceId, data)

    isEditing.value = false
  } catch {
    // Error is exposed by useInstance().
  }
}

async function removeInstance() {
  isDeleting.value = true

  try {
    await deleteInstance(instanceId)

    await router.push({
      name: 'instances',
    })
  } catch {
    // Error is exposed by useInstance().
  } finally {
    isDeleting.value = false
  }
}

function goBack() {
  router.push({
    name: 'instances',
  })
}

onMounted(load)
</script>

<template>
  <main class="min-h-screen bg-muted/40 px-4 py-8">
    <div class="mx-auto max-w-4xl space-y-6">
      <!-- Back -->

      <Button variant="ghost" class="-ml-2" @click="goBack"> ← Back </Button>

      <!-- Loading -->

      <Card v-if="isLoading && !instance">
        <CardContent class="py-12">
          <div class="flex items-center justify-center text-sm text-muted-foreground">
            Loading instance...
          </div>
        </CardContent>
      </Card>

      <!-- Error -->

      <Card v-else-if="error && !instance">
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

          <div class="flex gap-2">
            <!-- Edit -->

            <Button variant="outline" :disabled="isDeleting" @click="isEditing = !isEditing">
              <Pencil class="mr-2 h-4 w-4" />

              {{ isEditing ? 'Cancel' : 'Edit' }}
            </Button>

            <!-- Delete -->

            <AlertDialog>
              <AlertDialogTrigger as-child>
                <Button variant="destructive" :disabled="isDeleting">
                  <Trash2 class="mr-2 h-4 w-4" />

                  Delete
                </Button>
              </AlertDialogTrigger>

              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle> Delete {{ instance.name }}? </AlertDialogTitle>

                  <AlertDialogDescription>
                    This action cannot be undone. The PostgreSQL instance and its associated
                    resources will be deleted.
                  </AlertDialogDescription>
                </AlertDialogHeader>

                <AlertDialogFooter>
                  <AlertDialogCancel> Cancel </AlertDialogCancel>

                  <AlertDialogAction
                    class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                    :disabled="isDeleting"
                    @click="removeInstance"
                  >
                    {{ isDeleting ? 'Deleting...' : 'Delete instance' }}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>

        <!-- Update error -->

        <Card v-if="error">
          <CardHeader>
            <CardTitle> Unable to update instance </CardTitle>

            <CardDescription>
              {{ error.message }}
            </CardDescription>
          </CardHeader>
        </Card>

        <!-- Edit -->

        <Card v-if="isEditing">
          <CardHeader>
            <CardTitle> Edit instance </CardTitle>

            <CardDescription>
              Update the PostgreSQL version or storage configuration.
            </CardDescription>
          </CardHeader>

          <CardContent>
            <InstanceForm
              mode="edit"
              :initial-value="{
                version: instance.version,
                storage: instance.storage,
              }"
              :loading="isLoading"
              @submit="saveChanges"
            />
          </CardContent>
        </Card>

        <!-- Configuration -->

        <Card>
          <CardHeader>
            <CardTitle>Configuration</CardTitle>

            <CardDescription> PostgreSQL instance configuration. </CardDescription>
          </CardHeader>

          <CardContent>
            <dl class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
              <!-- PostgreSQL version -->

              <div>
                <dt class="text-sm text-muted-foreground">PostgreSQL version</dt>

                <dd class="mt-1 text-lg font-medium">
                  {{ instance.version }}
                </dd>
              </div>

              <!-- Storage -->

              <div>
                <dt class="text-sm text-muted-foreground">Storage</dt>

                <dd class="mt-1 text-lg font-medium">
                  {{ instance.storage }}
                </dd>
              </div>

              <!-- Instances -->

              <div>
                <dt class="text-sm text-muted-foreground">Instances</dt>

                <dd class="mt-1 text-lg font-medium">
                  {{ instance.readyInstances }}/{{ instance.instances }}
                </dd>
              </div>

              <!-- Status -->

              <div>
                <dt class="text-sm text-muted-foreground">Status</dt>

                <dd class="mt-2">
                  <InstanceStatus :status="instance.status" />
                </dd>
              </div>
            </dl>
          </CardContent>
        </Card>

        <InstanceCredentials
          :credentials="instanceCredentials"
          :loading="isLoadingCredentials"
          :error="credentialsError"
          @reveal="revealCredentials"
        />

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
