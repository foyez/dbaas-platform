import { describe, expect, it } from 'vitest'

import { createInstanceSchema } from './instance'

describe('createInstanceSchema', () => {
  it('accepts a valid instance', () => {
    const result = createInstanceSchema.safeParse({
      name: 'my-db-2',
      instances: 2,
      version: 15,
      storage: '7Gi',
    })

    expect(result.success).toBe(true)
  })

  it('rejects an empty name', () => {
    const result = createInstanceSchema.safeParse({
      name: '',
      instances: 2,
      version: 15,
      storage: '7Gi',
    })

    expect(result.success).toBe(false)
  })

  it('rejects invalid storage', () => {
    const result = createInstanceSchema.safeParse({
      name: 'my-db',
      instances: 2,
      version: 15,
      storage: 'seven gigs',
    })

    expect(result.success).toBe(false)
  })

  it('rejects zero instances', () => {
    const result = createInstanceSchema.safeParse({
      name: 'my-db',
      instances: 0,
      version: 15,
      storage: '7Gi',
    })

    expect(result.success).toBe(false)
  })
})
