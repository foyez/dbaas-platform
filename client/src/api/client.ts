import { ApiError } from './errors'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,

    headers: {
      Accept: 'application/json',
      ...options.headers,
    },
  })

  const contentType = response.headers.get('content-type')

  const body = contentType?.includes('application/json') ? await response.json() : null

  if (!response.ok) {
    const error = body as {
      error?: {
        code?: string
        message?: string
      }
    } | null

    throw new ApiError(
      response.status,
      error?.error?.code ?? 'UNKNOWN_ERROR',
      error?.error?.message ?? 'Something went wrong.',
    )
  }

  return body as T
}
