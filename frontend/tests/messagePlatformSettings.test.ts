import assert from 'node:assert/strict'
import test from 'node:test'
import { createMessagePlatformThinkingOptions, shouldSaveMessagePlatformDefaultModel } from '../src/utils/messagePlatformSettings'

test('shouldSaveMessagePlatformDefaultModel allows clearing the platform model', () => {
  assert.equal(shouldSaveMessagePlatformDefaultModel(''), true)
  assert.equal(shouldSaveMessagePlatformDefaultModel('model-1'), true)
})

test('createMessagePlatformThinkingOptions returns the supported thinking levels in UI order', () => {
  const got = createMessagePlatformThinkingOptions((key) => key)

  assert.deepEqual(got, [
    { value: 'off', label: 'thinkingOff' },
    { value: 'low', label: 'thinkingLow' },
    { value: 'medium', label: 'thinkingMedium' },
    { value: 'high', label: 'thinkingHigh' },
    { value: 'max', label: 'thinkingMax' },
  ])
})
