import { describe, expect, it } from 'vitest'

import { sqlErrorPointer } from './sqlErrorPointer'

describe('sqlErrorPointer', () => {
  it('points at the error on a single line', () => {
    expect(sqlErrorPointer('SELECT * FROM ordrs', 15)).toBe(
      'LINE 1: SELECT * FROM ordrs\n' + '                      ^',
    )
  })

  it('points at the right line of a multi-line statement', () => {
    expect(sqlErrorPointer('SELECT id\nFROM ordrs', 16)).toBe(
      'LINE 2: FROM ordrs\n' + '             ^',
    )
  })

  it('keeps tabs so the caret lines up', () => {
    expect(sqlErrorPointer('SELECT\n\titme', 9)).toBe('LINE 2: \titme\n' + '        \t^')
  })

  it('returns null without a usable position', () => {
    expect(sqlErrorPointer('SELECT 1', undefined)).toBeNull()
    expect(sqlErrorPointer('SELECT 1', 0)).toBeNull()
    expect(sqlErrorPointer('SELECT 1', 99)).toBeNull()
  })
})
