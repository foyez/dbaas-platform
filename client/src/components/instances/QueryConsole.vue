<script setup lang="ts">
import { ref } from 'vue'
import { Play, Terminal } from 'lucide-vue-next'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

import { executeQuery } from '@/api/instances'
import type { QueryResult } from '@/types/instance'
import type { ApiError } from '@/api/errors'

const props = defineProps<{
  instanceId: string
}>()

const sql = ref('')
const loading = ref(false)
const result = ref<QueryResult | null>(null)
const error = ref<ApiError | null>(null)

async function run() {
  if (!sql.value.trim() || loading.value) {
    return
  }

  loading.value = true
  error.value = null

  try {
    result.value = await executeQuery(props.instanceId, sql.value)
  } catch (e) {
    error.value = e as ApiError
    result.value = null
  } finally {
    loading.value = false
  }
}

function formatCell(value: unknown): string {
  if (value === null || value === undefined) return 'NULL'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function onKeydown(e: KeyboardEvent) {
  // Cmd/Ctrl+Enter to run, the standard shortcut for query editors.
  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
    e.preventDefault()
    run()
  }
}
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between space-y-0">
      <div>
        <CardTitle>Query console</CardTitle>
        <CardDescription>
          Run SQL directly against this instance's database.
        </CardDescription>
      </div>

      <Button size="sm" :disabled="loading || !sql.trim()" @click="run">
        <Play class="mr-2 h-4 w-4" />
        {{ loading ? 'Running...' : 'Run' }}
      </Button>
    </CardHeader>

    <CardContent class="space-y-4">
      <textarea
        v-model="sql"
        rows="6"
        placeholder="select * from your_table limit 100;"
        spellcheck="false"
        class="w-full rounded-md border bg-background p-3 font-mono text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        @keydown="onKeydown"
      />

      <p class="text-xs text-muted-foreground">
        ⌘/Ctrl + Enter to run. Results are capped at 500 rows and each query
        times out after 10s.
      </p>

      <!-- Error -->
      <div
        v-if="error"
        class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive"
      >
        {{ error.message }}
      </div>

      <!-- No result set (INSERT/UPDATE/DELETE/DDL) -->
      <div
        v-else-if="result && result.columns.length === 0"
        class="rounded-lg border bg-muted/50 p-4 text-sm"
      >
        Query OK. {{ result.rowCount }} row{{ result.rowCount === 1 ? '' : 's' }}
        affected ({{ result.durationMs }}ms).
      </div>

      <!-- Result table -->
      <div v-else-if="result" class="space-y-2">
        <p class="text-xs text-muted-foreground">
          {{ result.rowCount }} row{{ result.rowCount === 1 ? '' : 's' }}
          <span v-if="result.truncated">(truncated at 500)</span>
          · {{ result.durationMs }}ms
        </p>

        <div class="max-h-96 overflow-auto rounded-lg border">
          <table class="w-full text-left text-sm">
            <thead class="sticky top-0 bg-muted/80 backdrop-blur">
              <tr>
                <th
                  v-for="col in result.columns"
                  :key="col"
                  class="whitespace-nowrap px-3 py-2 font-medium"
                >
                  {{ col }}
                </th>
              </tr>
            </thead>

            <tbody>
              <tr
                v-for="(row, i) in result.rows"
                :key="i"
                class="border-t hover:bg-muted/40"
              >
                <td
                  v-for="(cell, j) in row"
                  :key="j"
                  class="whitespace-nowrap px-3 py-2 font-mono"
                >
                  {{ formatCell(cell) }}
                </td>
              </tr>
            </tbody>
          </table>

          <div
            v-if="result.rows.length === 0"
            class="py-6 text-center text-sm text-muted-foreground"
          >
            No rows returned.
          </div>
        </div>
      </div>

      <div v-else class="flex flex-col items-center justify-center gap-2 py-12 text-sm text-muted-foreground">
        <Terminal class="h-6 w-6" />
        Write a query above and run it to see results here.
      </div>
    </CardContent>
  </Card>
</template>