export const INSTANCE_STATUSES = [
  'Pending',
  'Provisioning',
  'Running',
  'Failed',
  'Deleting',
  'Deleted',
] as const

export type InstanceStatus = (typeof INSTANCE_STATUSES)[number]

export interface Instance {
  id: string
  name: string
  version: number
  storage: string
  status: InstanceStatus
  createdAt: string
}

export interface CreateInstanceRequest {
  name: string
  instances: number
  version: number
  storage: string
}

export interface UpdateInstanceRequest {
  version?: number
  storage?: string
}

export interface InstanceListResponse {
  items: Instance[]
}
