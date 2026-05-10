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
