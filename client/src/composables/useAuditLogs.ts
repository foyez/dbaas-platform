import { ref } from 'vue'

import { getAuditLogs } from '@/api/audit'
import { ApiError } from '@/api/errors'
import type { LogLine } from '@/types/instance'

export function useAuditLogs() {
  const logs = ref<LogLine[] | null>(null)
  const isLoading = ref(false)
  const error = ref<ApiError | null>(null)

  async function fetchLogs(resourceId?: string) {
    isLoading.value = true
    error.value = null

    try {
      const res = await getAuditLogs(resourceId)
      logs.value = res.logs
      return logs.value
    } catch (err) {
      error.value =
        err instanceof ApiError
          ? err
          : new ApiError(0, 'NETWORK_ERROR', 'Unable to connect to the server.')
      throw err
    } finally {
      isLoading.value = false
    }
  }

  return { logs, isLoading, error, fetchLogs }
}
