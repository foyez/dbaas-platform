export type LogLevel = 'debug' | 'info' | 'warning' | 'error' | 'fatal' | 'unknown'

export interface ParsedLog {
  time: string | null
  level: LogLevel
  message: string
  source: string // "postgres" or "wal-archive" etc - useful for filtering noise
  raw: string
}

function normalizeLevel(value: unknown): LogLevel {
  const v = String(value ?? '').toLowerCase()

  if (v.startsWith('debug')) return 'debug'          // DEBUG1..DEBUG5
  if (['log', 'info', 'notice'].includes(v)) return 'info'
  if (['warning', 'warn'].includes(v)) return 'warning'
  if (v === 'error') return 'error'
  if (['fatal', 'panic'].includes(v)) return 'fatal'

  return 'unknown'
}

export function parseInstanceLog(line: string): ParsedLog {
  try {
    const obj = JSON.parse(line)

    if (obj.record) {
      return {
        time: obj.record.log_time ?? obj.ts ?? null,
        level: normalizeLevel(obj.record.error_severity),
        message: obj.record.message ?? line,
        source: 'postgres',
        raw: line,
      }
    }

    return {
      time: obj.ts ?? null,
      level: normalizeLevel(obj.level),
      message: obj.msg ?? line,
      source: obj.logger ?? 'instance-manager',
      raw: line,
    }
  } catch {
    return { time: null, level: 'unknown', message: line, source: 'unknown', raw: line }
  }
}