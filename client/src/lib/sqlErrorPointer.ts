/**
 * Formats a psql-style pointer to a SQL error: the line containing the error
 * and a caret under it. `position` is PostgreSQL's 1-based character offset
 * into the statement. Returns null when there's no position to point at.
 */
export function sqlErrorPointer(sql: string, position: number | undefined): string | null {
  if (!position || position > sql.length + 1) return null

  const before = sql.slice(0, position - 1)
  const lineNumber = before.split('\n').length
  const lineStart = before.lastIndexOf('\n') + 1
  const lineEnd = sql.indexOf('\n', lineStart)
  const line = sql.slice(lineStart, lineEnd === -1 ? undefined : lineEnd)
  const prefix = `LINE ${lineNumber}: `
  // Keep tabs so the caret lines up with tab-indented SQL.
  const padding = (prefix + before.slice(lineStart)).replace(/[^\t]/g, ' ')

  return `${prefix}${line}\n${padding}^`
}
