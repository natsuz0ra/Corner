import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import test from 'node:test'
import {
  isUpdateJobActive,
  normalizeUpdateCheck,
  updatePhaseTone,
} from '../src/utils/updateStatus'

test('normalizeUpdateCheck fills stable defaults', () => {
  const got = normalizeUpdateCheck({
    current: 'v1.26.1',
    latest: 'v1.26.2',
    updateAvailable: true,
    canApply: true,
  })

  assert.equal(got.current, 'v1.26.1')
  assert.equal(got.latest, 'v1.26.2')
  assert.equal(got.releaseNotes, '')
  assert.equal(got.manualHint, '')
})

test('update phase helpers classify active and terminal states', () => {
  assert.equal(isUpdateJobActive('downloading'), true)
  assert.equal(isUpdateJobActive('succeeded'), false)
  assert.equal(updatePhaseTone('failed'), 'danger')
  assert.equal(updatePhaseTone('succeeded'), 'success')
})

test('about settings tab renders update center hooks', () => {
  const source = readFileSync(resolve(import.meta.dirname, '../src/components/settings/SettingsAboutTab.vue'), 'utf8')

  assert.match(source, /updateAPI/)
  assert.match(source, /updateCenterTitle/)
  assert.match(source, /checkUpdate/)
  assert.match(source, /applyUpdate/)
  assert.match(source, /manualHint/)
})

test('i18n contains update center labels in both languages', () => {
  const source = readFileSync(resolve(import.meta.dirname, '../src/i18n.ts'), 'utf8')

  assert.match(source, /updateCenterTitle: '更新中心'/)
  assert.match(source, /updateCenterTitle: 'Update Center'/)
  assert.match(source, /updateRestarting/)
  assert.match(source, /updateManualHint/)
})
