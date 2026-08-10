import { ref } from 'vue'
import { createInstance } from '@/api/instances'
import type { CreateInstanceRequest, Instance } from '@/types/instance'
import { ApiError } from '@/api/errors'

export function useCreateInstance() {
  const isLoading = ref(false)
  const instance = ref<Instance | null>(null)
  const error = ref<ApiError | null>(null)

  let idempotencyKey: string | null = null

  async function execute(request: CreateInstanceRequest) {
    isLoading.value = true
    error.value = null

    if (!idempotencyKey) {
      idempotencyKey = crypto.randomUUID()
    }

    try {
      instance.value = await createInstance(request, idempotencyKey)

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

  function reset() {
    instance.value = null
    error.value = null
    idempotencyKey = null
  }

  return {
    execute,
    reset,
    isLoading,
    instance,
    error,
  }
}
