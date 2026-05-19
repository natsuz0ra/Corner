import assert from 'node:assert/strict'
import test from 'node:test'
import { buildSettingsPayload, normalizeSettingsPayload } from '../src/utils/settingsPayload'

test('normalizeSettingsPayload exposes platform runtime settings with defaults', () => {
  const got = normalizeSettingsPayload({
    language: 'zh-CN',
    messagePlatformDefaultModel: 'model-platform',
    messagePlatformThinkingLevel: 'high',
    messagePlatformApprovalMode: 'auto_review',
  })

  assert.equal(got.messagePlatformDefaultModel, 'model-platform')
  assert.equal(got.messagePlatformThinkingLevel, 'high')
  assert.equal(got.messagePlatformApprovalMode, 'auto_review')
  assert.equal(got.approvalMode, 'standard')
})

test('buildSettingsPayload sends platform runtime settings and empty default model', () => {
  const got = buildSettingsPayload({
    messagePlatformDefaultModel: '',
    messagePlatformThinkingLevel: 'medium',
    messagePlatformApprovalMode: 'auto',
  })

  assert.deepEqual(got, {
    messagePlatformDefaultModel: '',
    messagePlatformThinkingLevel: 'medium',
    messagePlatformApprovalMode: 'auto',
  })
})

test('settings payload normalizes and sends memory settings', () => {
  const normalized = normalizeSettingsPayload({ language: 'zh-CN' })
  assert.equal(normalized.memoryEnabled, true)
  assert.equal(normalized.memoryUserProfileEnabled, true)
  assert.equal(normalized.memoryCharLimit, 2200)
  assert.equal(normalized.memoryUserCharLimit, 1375)
  assert.equal(normalized.memoryNudgeInterval, 10)

  const payload = buildSettingsPayload({
    memoryEnabled: false,
    memoryUserProfileEnabled: true,
    memoryCharLimit: 3000,
    memoryUserCharLimit: 1500,
    memoryNudgeInterval: 5,
  })
  assert.deepEqual(payload, {
    memoryEnabled: false,
    memoryUserProfileEnabled: true,
    memoryCharLimit: 3000,
    memoryUserCharLimit: 1500,
    memoryNudgeInterval: 5,
  })
})
