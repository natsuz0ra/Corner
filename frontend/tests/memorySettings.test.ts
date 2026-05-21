import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import test from 'node:test'

test('settings panel exposes a dedicated memory tab and API', () => {
  const panelSource = readFileSync(resolve(import.meta.dirname, '../src/components/settings/SettingsPanel.vue'), 'utf8')
  const tabSource = readFileSync(resolve(import.meta.dirname, '../src/components/settings/SettingsMemoryTab.vue'), 'utf8')
  const apiSource = readFileSync(resolve(import.meta.dirname, '../src/api/memory.ts'), 'utf8')
  const i18nSource = readFileSync(resolve(import.meta.dirname, '../src/i18n.ts'), 'utf8')

  assert.match(panelSource, /SettingsMemoryTab/)
  assert.match(panelSource, /key: 'memory'/)
  assert.match(panelSource, /memoryAPI\.get/)
  assert.match(tabSource, /memoryEnabled/)
  assert.match(tabSource, /memoryUserProfileEnabled/)
  assert.match(tabSource, /memoryNudgeInterval/)
  assert.match(tabSource, /memoryContextNotice/)
  assert.match(i18nSource, /会进入模型上下文|model context/)
  assert.match(apiSource, /\/api\/memory/)
})
