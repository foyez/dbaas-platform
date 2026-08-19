import type { LogLine } from '@/types/instance'
import { apiFetch } from './client'

export interface AuditLogsResponse {
  logs: LogLine[]
}

export async function getAuditLogs(resourceId?: string): Promise<AuditLogsResponse> {
  const query = resourceId ? `?resourceId=${encodeURIComponent(resourceId)}` : ''

  return apiFetch<AuditLogsResponse>(`/audit-logs${query}`)
}
