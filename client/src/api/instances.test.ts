import { afterEach, describe, expect, it, vi } from 'vitest'

import { createInstance } from './instances'

describe('createInstance', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('sends the correct request', async () => {
    const response = {
      id: 'inst-7f3a2b',
      name: 'my-db-2',
      version: 15,
      storage: '7Gi',
      status: 'Pending',
      createdAt: '2026-08-04T09:12:00Z',
    }

    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify(response), {
        status: 202,
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    )

    await createInstance(
      {
        name: 'my-db-2',
        instances: 2,
        version: 15,
        storage: '7Gi',
      },
      'test-idempotency-key',
    )

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/instances',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
          'Idempotency-Key': 'test-idempotency-key',
        }),
      }),
    )
  })

  it('returns the created instance', async () => {
    const response = {
      id: 'inst-7f3a2b',
      name: 'my-db-2',
      version: 15,
      storage: '7Gi',
      status: 'Pending',
      createdAt: '2026-08-04T09:12:00Z',
    }

    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify(response), {
        status: 202,
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    )

    const result = await createInstance(
      {
        name: 'my-db-2',
        instances: 2,
        version: 15,
        storage: '7Gi',
      },
      'test-key',
    )

    expect(result).toEqual(response)
  })

  it('throws an ApiError for a conflict', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          error: {
            code: 'INSTANCE_ALREADY_EXISTS',
            message: 'An instance with this name already exists.',
          },
        }),
        {
          status: 409,
          headers: {
            'Content-Type': 'application/json',
          },
        },
      ),
    )

    await expect(
      createInstance(
        {
          name: 'my-db',
          instances: 1,
          version: 16,
          storage: '10Gi',
        },
        'test-key',
      ),
    ).rejects.toMatchObject({
      status: 409,
      code: 'INSTANCE_ALREADY_EXISTS',
    })
  })
})
