import { computed, ref } from 'vue'

import { getAuditLogs } from '@/api/audit'
import { ApiError } from '@/api/errors'
import type { LogLine } from '@/types/instance'

const PAGE_SIZE = 20

export function useAuditLogs() {
  const logs = ref<LogLine[] | null>(null)
  const nextCursor = ref<string | null>(null)
  const currentResourceId = ref<string | undefined>(undefined)

  const isLoading = ref(false)
  const isLoadingMore = ref(false)
  const error = ref<ApiError | null>(null)

  const hasMore = computed(() => nextCursor.value !== null)

  function toApiError(err: unknown): ApiError {
    if (err instanceof ApiError) return err
    return new ApiError(0, 'NETWORK_ERROR', 'Unable to connect to the server.')
  }

  async function fetchLogs(resourceId?: string) {
    isLoading.value = true
    error.value = null
    currentResourceId.value = resourceId

    try {
      const res = await getAuditLogs({ resourceId, limit: PAGE_SIZE })
      logs.value = res.logs
      nextCursor.value = res.nextCursor ?? null
      return logs.value
    } catch (err) {
      error.value = toApiError(err)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function loadMore() {
    if (!nextCursor.value || isLoadingMore.value) return

    isLoadingMore.value = true
    error.value = null

    try {
      const res = await getAuditLogs({
        resourceId: currentResourceId.value,
        limit: PAGE_SIZE,
        cursor: nextCursor.value,
      })

      logs.value = [...(logs.value ?? []), ...res.logs]
      nextCursor.value = res.nextCursor ?? null
    } catch (err) {
      error.value = toApiError(err)
      throw err
    } finally {
      isLoadingMore.value = false
    }
  }

  return { logs, isLoading, isLoadingMore, error, hasMore, fetchLogs, loadMore }
}