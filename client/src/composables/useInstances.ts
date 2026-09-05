import { computed, ref } from 'vue'

import { listInstances } from '@/api/instances'
import { ApiError } from '@/api/errors'

import type { Instance } from '@/types/instance'

const PAGE_SIZE = 20

export function useInstances() {
  const instances = ref<Instance[]>([])
  const nextCursor = ref<string | null>(null)

  const isLoading = ref(false)
  const isLoadingMore = ref(false)
  const error = ref<ApiError | null>(null)

  const isEmpty = computed(() => !isLoading.value && !error.value && instances.value.length === 0)
  const hasMore = computed(() => nextCursor.value !== null)

  function toApiError(err: unknown): ApiError {
    if (err instanceof ApiError) return err
    return new ApiError(0, 'NETWORK_ERROR', 'Unable to connect to the server.')
  }

  async function fetchInstances() {
    isLoading.value = true
    error.value = null

    try {
      const response = await listInstances({ limit: PAGE_SIZE })

      instances.value = response.items
      nextCursor.value = response.nextCursor ?? null

      return response
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
      const response = await listInstances({
        limit: PAGE_SIZE,
        cursor: nextCursor.value,
      })

      instances.value = [...instances.value, ...response.items]
      nextCursor.value = response.nextCursor ?? null
    } catch (err) {
      error.value = toApiError(err)
      throw err
    } finally {
      isLoadingMore.value = false
    }
  }

  // Re-fetches exactly however many instances are currently visible, in a
  // single request from the start (cursor-based pagination has no "refresh
  // page 2 in place" - only "continue from the start" or "continue from
  // cursor X"). Used by polling so a background refresh can pick up status
  // changes without discarding pages already loaded via loadMore().
  // Deliberately swallows errors - a transient failure on a background
  // poll tick shouldn't surface as a page-level error.
  async function refresh() {
    const limit = Math.max(instances.value.length, PAGE_SIZE)

    try {
      const response = await listInstances({ limit })

      instances.value = response.items
      nextCursor.value = response.nextCursor ?? null
    } catch {
      // Swallowed - see doc comment above.
    }
  }

  return {
    instances,
    isLoading,
    isLoadingMore,
    error,
    isEmpty,
    hasMore,
    fetchInstances,
    loadMore,
    refresh,
  }
}