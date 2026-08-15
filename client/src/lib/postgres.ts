export function buildPostgresUri(credentials: {
  host: string
  port: number
  database: string
  username: string
  password: string
}): string {
  const username = encodeURIComponent(credentials.username)
  const password = encodeURIComponent(credentials.password)
  const database = encodeURIComponent(credentials.database)

  return `postgresql://${username}:${password}@${credentials.host}:${credentials.port}/${database}`
}