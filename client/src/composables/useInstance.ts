import { ref } from 'vue'

import {
  deleteInstance as deleteInstanceApi,
  getInstance,
  updateInstance as updateInstanceApi,
} from '@/api/instances'
import { ApiError } from '@/api/errors'
import type { Instance, UpdateInstanceRequest } from '@/types/instance'

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

  async function updateInstance(id: string, data: UpdateInstanceRequest) {
    isLoading.value = true
    error.value = null

    try {
      instance.value = await updateInstanceApi(id, data)

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

  async function deleteInstance(id: string) {
    isLoading.value = true
    error.value = null

    try {
      await deleteInstanceApi(id)
      instance.value = null
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
    updateInstance,
    deleteInstance,
  }
}
