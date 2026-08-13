import { userManager } from '@/auth/oidc'
import { ApiError } from './errors'

const API_BASE_URL = import.meta.env.VITE_API_URL

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const user = await userManager.getUser()

  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')

  if (user?.access_token) {
    headers.set('Authorization', `Bearer ${user.access_token}`)
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
  })

  const contentType = response.headers.get('content-type')
  const body = contentType?.includes('application/json') ? await response.json() : null

  if (!response.ok) {
    if (response.status === 401) {
      await userManager.signinRedirect()
      throw new ApiError(401, 'UNAUTHORIZED', 'Authentication required.')
    }

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
