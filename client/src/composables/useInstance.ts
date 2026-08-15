import { ref } from 'vue'

import {
  deleteInstance as deleteInstanceApi,
  getInstance,
  getInstanceCredentials,
  updateInstance as updateInstanceApi,
} from '@/api/instances'
import { ApiError } from '@/api/errors'
import type { Instance, InstanceCredentials, UpdateInstanceRequest } from '@/types/instance'

export function useInstance() {
  const instance = ref<Instance | null>(null)
  const instanceCredentials = ref<InstanceCredentials | null>(null)

  const isLoading = ref(false)
  const isLoadingCredentials = ref(false)
  const error = ref<ApiError | null>(null)
  const credentialsError = ref<Error | null>(null)

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

  async function fetchCredentials(id: string) {
    isLoadingCredentials.value = true
    credentialsError.value = null

    try {
      instanceCredentials.value = await getInstanceCredentials(id)

      return instanceCredentials.value
    } catch (err) {
      if (err instanceof ApiError) {
        credentialsError.value = err
      } else {
        credentialsError.value = new ApiError(
          0,
          'NETWORK_ERROR',
          'Unable to connect to the server.',
        )
      }

      throw err
    } finally {
      isLoadingCredentials.value = false
    }
  }

  function clearCredentials() {
    instanceCredentials.value = null
    credentialsError.value = null
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
    instanceCredentials,

    isLoading,
    isLoadingCredentials,

    error,
    credentialsError,

    fetchInstance,
    fetchCredentials,
    clearCredentials,

    updateInstance,
    deleteInstance,
  }
}
