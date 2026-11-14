import { afterEach, describe, expect, it, vi } from 'vitest'

import { apiFetch } from './client'
import { ApiError } from './errors'

describe('apiFetch', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('turns an error response into an ApiError with its details', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          error: {
            code: 'QUERY_FAILED',
            message: 'relation "ordrs" does not exist',
            details: { sqlState: '42P01', position: 15 },
          },
        }),
        { status: 422, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    const error = await apiFetch('/instances/abc/query').catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 422,
      code: 'QUERY_FAILED',
      message: 'relation "ordrs" does not exist',
      details: { sqlState: '42P01', position: 15 },
    })
  })
})
