import { ref, computed } from 'vue'

import { listInstances } from '@/api/instances'
import { ApiError } from '@/api/errors'
import type { Instance } from '@/types/instance'

export function useInstances() {
  const instances = ref<Instance[]>([])
  const isLoading = ref(false)
  const error = ref<ApiError | null>(null)

  const isEmpty = computed(() => !isLoading.value && !error.value && instances.value.length === 0)

  async function fetchInstances() {
    isLoading.value = true
    error.value = null

    try {
      const response = await listInstances()

      instances.value = response.items

      return response
    } catch (err) {
      if (err instanceof ApiError) {
        error.value = err
      } else {
        error.value = new ApiError(0, 'NETWORK_ERROR', 'Unable to connect to the server.')
      }

      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    instances,
    isLoading,
    error,
    isEmpty,
    fetchInstances,
  }
}
