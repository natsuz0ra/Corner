import assert from 'node:assert/strict'
import test from 'node:test'

import { createClientId } from '../src/utils/uuid'

const originalCrypto = globalThis.crypto

test.afterEach(() => {
  Object.defineProperty(globalThis, 'crypto', {
    configurable: true,
    value: originalCrypto,
  })
})

test('createClientId falls back when crypto.randomUUID is unavailable', () => {
  Object.defineProperty(globalThis, 'crypto', {
    configurable: true,
    value: {},
  })

  const first = createClientId()
  const second = createClientId()

  assert.match(first, /^client-[a-z0-9]+-[a-z0-9]+$/)
  assert.notEqual(first, second)
})

test('createClientId uses crypto.randomUUID when available', () => {
  Object.defineProperty(globalThis, 'crypto', {
    configurable: true,
    value: {
      randomUUID: () => 'native-uuid',
    },
  })

  assert.equal(createClientId(), 'native-uuid')
})
