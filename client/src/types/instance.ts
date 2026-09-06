export const INSTANCE_STATUSES = [
  'Pending',
  'Provisioning',
  'Running',
  'Failed',
  'Deleting',
  'Deleted',
] as const

export type InstanceStatus = (typeof INSTANCE_STATUSES)[number]

// Statuses that will never change on their own without user action - once
// an instance reaches one of these, there's nothing left to poll for.
const TERMINAL_STATUSES: readonly InstanceStatus[] = ['Running', 'Failed', 'Deleted']

export function isTerminalStatus(status: InstanceStatus): boolean {
  return TERMINAL_STATUSES.includes(status)
}

export interface Instance {
  id: string
  name: string
  database: string
  version: number
  storage: string
  instances: number
  readyInstances: number
  status: InstanceStatus
  createdAt: string
}

export interface InstanceCredentials {
  host: string
  port: number
  database: string
  username: string
  password: string
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
  nextCursor?: string
}

export interface LogLine {
  timestamp: string
  line: string
}

export interface InstanceLogsRespose {
  instanceId: string
  logs: LogLine[]
}

export interface QueryResult {
  columns: string[]
  rows: unknown[][]
  rowCount: number
  truncated: boolean
  durationMs: number
}