import type { LogLine } from '@/types/instance'
import { apiFetch } from './client'

export interface AuditLogsResponse {
  logs: LogLine[]
  nextCursor?: string
}

export async function getAuditLogs(params: {
  resourceId?: string
  limit?: number
  cursor?: string
} = {}): Promise<AuditLogsResponse> {
  const query = new URLSearchParams()

  if (params.resourceId) query.set('resourceId', params.resourceId)
  if (params.limit) query.set('limit', String(params.limit))
  if (params.cursor) query.set('cursor', params.cursor)

  const qs = query.toString()

  return apiFetch<AuditLogsResponse>(`/audit-logs${qs ? `?${qs}` : ''}`)
}