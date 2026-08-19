<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { RotateCw } from 'lucide-vue-next'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

import { useAuditLogs } from '@/composables/useAuditLogs'
import { formatDate } from '@/lib/date'

interface ParsedAuditEvent {
  time?: string
  actorId?: string
  action?: string
  resourceType?: string
  resourceId?: string
  status?: number
}

const route = useRoute()
const router = useRouter()

const { logs, isLoading, error, fetchLogs } = useAuditLogs()

const resourceId = ref(String(route.query.resourceId ?? ''))

function parse(line: string): ParsedAuditEvent {
  try {
    const obj = JSON.parse(line)
    return {
      time: obj.time,
      actorId: obj.actor_id,
      action: obj.action,
      resourceType: obj.resource_type,
      resourceId: obj.resource_id,
      status: obj.status,
    }
  } catch {
    return {}
  }
}

function load() {
  fetchLogs(resourceId.value || undefined).catch(() => {})
}

function applyFilter() {
  router.replace({ query: resourceId.value ? { resourceId: resourceId.value } : {} })
  load()
}

onMounted(load)
</script>

<template>
  <main class="min-h-screen bg-muted/40 px-4 py-8">
    <div class="mx-auto max-w-6xl space-y-6">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">Audit log</h1>
          <p class="mt-1 text-sm text-muted-foreground">A record of actions taken across the platform.</p>
        </div>

        <Button variant="outline" size="sm" :disabled="isLoading" @click="load">
          <RotateCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': isLoading }" />
          Refresh
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Filter</CardTitle>
          <CardDescription>Narrow results to a specific instance.</CardDescription>
        </CardHeader>

        <CardContent class="flex gap-2">
          <Input v-model="resourceId" placeholder="Instance ID" class="max-w-sm" @keyup.enter="applyFilter" />
          <Button variant="secondary" @click="applyFilter">Apply</Button>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="p-0">
          <div v-if="error" class="p-6 text-sm text-destructive">{{ error.message }}</div>

          <div v-else-if="isLoading && !logs" class="p-12 text-center text-sm text-muted-foreground">
            Loading audit log...
          </div>

          <div v-else-if="logs && logs.length === 0" class="p-12 text-center text-sm text-muted-foreground">
            No audit events found.
          </div>

          <Table v-else-if="logs">
            <TableHeader>
              <TableRow>
                <TableHead>Time</TableHead>
                <TableHead>Actor</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              <TableRow v-for="(entry, i) in logs.map((l) => parse(l.line))" :key="i">
                <TableCell class="whitespace-nowrap text-sm text-muted-foreground">
                  {{ entry.time ? formatDate(entry.time) : '—' }}
                </TableCell>
                <TableCell class="font-mono text-sm">{{ entry.actorId ?? '—' }}</TableCell>
                <TableCell class="text-sm">{{ entry.action ?? '—' }}</TableCell>
                <TableCell class="font-mono text-sm">{{ entry.resourceType }}/{{ entry.resourceId }}</TableCell>
                <TableCell class="text-sm">{{ entry.status ?? '—' }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  </main>
</template>