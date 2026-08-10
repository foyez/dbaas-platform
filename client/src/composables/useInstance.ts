import { ref } from 'vue'

import { getInstance } from '@/api/instances'
import { ApiError } from '@/api/errors'
import type { Instance } from '@/types/instance'

export function useInstance() {
  const instance = ref<Instance | null>(null)
  const isLoading = ref(false)
  const error = ref<ApiError | null>(null)

  async function fetchInstance(id: string) {
    isLoading.value = true
    error.value = null

    try {
      instance.value = await getInstance(id)

      return instance.value
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
    instance,
    isLoading,
    error,
    fetchInstance,
  }
}
