<script setup lang="ts">
import { Database, Plus } from 'lucide-vue-next'

import { Card, CardContent } from '@/components/ui/card'

import { Button } from '@/components/ui/button'

defineEmits<{
  create: []
}>()
</script>

<template>
  <Card>
    <CardContent class="flex flex-col items-center justify-center py-12 text-center">
      <div
        class="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary"
      >
        <Database class="h-5 w-5" />
      </div>

      <h2 class="mt-4 text-lg font-semibold">Create your first instance</h2>

      <p class="mt-1 max-w-md text-sm text-muted-foreground">
        Deploy a managed PostgreSQL instance and let CNPG handle the underlying PostgreSQL cluster.
      </p>

      <Button class="mt-6" @click="$emit('create')">
        <Plus class="mr-2 h-4 w-4" />

        Create instance
      </Button>
    </CardContent>
  </Card>
</template>
