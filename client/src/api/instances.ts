import { apiFetch } from './client'

import type { CreateInstanceRequest, Instance, InstanceListResponse } from '@/types/instance'

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
