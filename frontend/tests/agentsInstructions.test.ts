import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import test from 'node:test'

test('settings payload ignores AGENTS instructions because they use a dedicated API', () => {
  const source = readFileSync(resolve(import.meta.dirname, '../src/utils/settingsPayload.ts'), 'utf8')

  assert.doesNotMatch(source, /agentsInstructions/)
  assert.doesNotMatch(source, /agentsContent/)
})

test('settings panel exposes a dedicated AGENTS editor tab', () => {
  const source = readFileSync(resolve(import.meta.dirname, '../src/components/settings/SettingsPanel.vue'), 'utf8')

  assert.match(source, /agentsInstructionsAPI/)
  assert.match(source, /SettingsAgentsTab/)
  assert.match(source, /key: 'agents'/)
  assert.match(source, /saveAgentsInstructions/)
})
