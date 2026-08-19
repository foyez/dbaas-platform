<script setup lang="ts">
import { computed, ref } from 'vue'
import { RotateCw, Terminal } from 'lucide-vue-next'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'

import { parseInstanceLog, type LogLevel } from '@/lib/parseInstanceLog'
import type { LogLine } from '@/types/instance'
import type { ApiError } from '@/api/errors'

const props = defineProps<{
  logs: LogLine[] | null
  loading: boolean
  error: ApiError | null
}>()

const emit = defineEmits<{ reveal: [] }>()

const search = ref('')
const levelFilter = ref<LogLevel | 'all'>('all')
const sourceFilter = ref<'all' | string>('all')

const levelStyles: Record<LogLevel, string> = {
  debug: 'bg-muted text-muted-foreground',
  info: 'bg-blue-100 text-blue-700',
  warning: 'bg-amber-100 text-amber-700',
  error: 'bg-red-100 text-red-700',
  fatal: 'bg-red-200 text-red-900',
  unknown: 'bg-muted text-muted-foreground',
}

const parsedLogs = computed(() => (props.logs ?? []).map((l) => parseInstanceLog(l.line)))

const filteredLogs = computed(() =>
  parsedLogs.value.filter((entry) => {
    const matchesLevel = levelFilter.value === 'all' || entry.level === levelFilter.value
    const matchesSource = sourceFilter.value === 'all' || entry.source === sourceFilter.value
    const matchesSearch =
      !search.value || entry.message.toLowerCase().includes(search.value.toLowerCase())
    return matchesLevel && matchesSource && matchesSearch
  }),
)

function formatTime(time: string | null): string {
  if (!time) return '—'
  const d = new Date(time)
  return isNaN(d.getTime()) ? time : d.toLocaleTimeString()
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

    <CardContent class="space-y-4">
      <div
        v-if="error"
        class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive"
      >
        {{ error.message }}
      </div>

      <div
        v-else-if="loading && !logs"
        class="flex items-center justify-center py-12 text-sm text-muted-foreground"
      >
        Loading logs...
      </div>

      <div
        v-else-if="logs && logs.length === 0"
        class="flex flex-col items-center justify-center gap-2 py-12 text-sm text-muted-foreground"
      >
        <Terminal class="h-6 w-6" />
        No log entries yet.
      </div>

      <template v-else-if="logs">
        <!-- Filters -->
        <div class="flex gap-2">
          <Input v-model="search" placeholder="Search messages..." class="max-w-xs" />

          <select v-model="levelFilter" class="h-9 rounded-md border bg-background px-3 text-sm">
            <option value="all">All levels</option>
            <option value="error">Error</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
            <option value="debug">Debug</option>
          </select>

          <select v-model="sourceFilter" class="h-9 rounded-md border bg-background px-3 text-sm">
            <option value="all">All sources</option>
            <option value="postgres">Postgres</option>
            <option value="wal-archive">WAL archive</option>
            <option value="instance-manager">Instance manager</option>
          </select>
        </div>

        <!-- Log lines -->
        <div class="max-h-96 space-y-1 overflow-y-auto rounded-lg bg-muted/60 p-3">
          <div
            v-for="(entry, i) in filteredLogs"
            :key="i"
            class="flex items-start gap-3 rounded px-2 py-1.5 font-mono text-xs hover:bg-muted"
          >
            <span class="shrink-0 text-muted-foreground">{{ formatTime(entry.time) }}</span>
            <Badge :class="levelStyles[entry.level]" class="shrink-0">{{ entry.level }}</Badge>
            <span class="whitespace-pre-wrap break-all">{{ entry.message }}</span>
          </div>

          <div
            v-if="filteredLogs.length === 0"
            class="py-6 text-center text-sm text-muted-foreground"
          >
            No entries match your filter.
          </div>
        </div>
      </template>

      <div v-else class="py-12 text-center text-sm text-muted-foreground">
        Click "Load logs" to view recent activity for this instance.
      </div>
    </CardContent>
  </Card>
</template>
