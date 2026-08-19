<script setup lang="ts">
import { RotateCw, Terminal } from 'lucide-vue-next'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

import type { LogLine } from '@/types/instance'
import type { ApiError } from '@/api/errors'

defineProps<{
  logs: LogLine[] | null
  loading: boolean
  error: ApiError | null
}>()

const emit = defineEmits<{ reveal: [] }>()

function formatLogTimestamp(ns: string): string {
  // Loki timestamps are unix nanoseconds as a string
  return new Date(Number(ns) / 1_000_000).toLocaleTimeString()
}
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between space-y-0">
      <div>
        <CardTitle>Logs</CardTitle>
        <CardDescription>Recent log output from this instance.</CardDescription>
      </div>

      <Button variant="outline" size="sm" :disabled="loading" @click="emit('reveal')">
        <RotateCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': loading }" />
        {{ logs ? 'Refresh' : 'Load logs' }}
      </Button>
    </CardHeader>

    <CardContent>
      <div
        v-if="error"
        class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive"
      >
        {{ error.message }}
      </div>

      <div v-else-if="loading && !logs" class="flex items-center justify-center py-12 text-sm text-muted-foreground">
        Loading logs...
      </div>

      <div
        v-else-if="logs && logs.length === 0"
        class="flex flex-col items-center justify-center gap-2 py-12 text-sm text-muted-foreground"
      >
        <Terminal class="h-6 w-6" />
        No log entries yet.
      </div>

      <div
        v-else-if="logs"
        class="max-h-96 overflow-y-auto rounded-lg bg-muted/60 p-4 font-mono text-xs leading-relaxed"
      >
        <div v-for="(entry, i) in logs" :key="i" class="whitespace-pre-wrap break-all">
          <span class="text-muted-foreground">{{ formatLogTimestamp(entry.timestamp) }}</span>
          <span class="ml-2">{{ entry.line }}</span>
        </div>
      </div>

      <div v-else class="py-12 text-center text-sm text-muted-foreground">
        Click "Load logs" to view recent activity for this instance.
      </div>
    </CardContent>
  </Card>
</template>