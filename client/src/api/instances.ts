import { apiFetch } from './client'

import type {
  CreateInstanceRequest,
  Instance,
  InstanceListResponse,
  UpdateInstanceRequest,
} from '@/types/instance'

export async function createInstance(
  request: CreateInstanceRequest,
  idempotencyKey: string,
): Promise<Instance> {
  return apiFetch<Instance>('/instances', {
    method: 'POST',

    headers: {
      'Content-Type': 'application/json',
      'Idempotency-Key': idempotencyKey,
    },

    body: JSON.stringify(request),
  })
}

export async function getInstance(id: string): Promise<Instance> {
  return apiFetch<Instance>(`/instances/${encodeURIComponent(id)}`)
}

export async function listInstances(): Promise<InstanceListResponse> {
  return apiFetch<InstanceListResponse>('/instances')
}

export async function updateInstance(id: string, data: UpdateInstanceRequest): Promise<Instance> {
  return apiFetch<Instance>(`/instances/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export async function deleteInstance(id: string): Promise<void> {
  await apiFetch<void>(`/instances/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}
