import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import test from 'node:test'
import {
  createWebSandboxModeOptions,
  toWebSandboxMode,
} from '../src/utils/sandboxSettings'

test('web sandbox mode options hide workspace-write', () => {
  const options = createWebSandboxModeOptions((key) => key)

  assert.deepEqual(options, [
    { value: 'read-only', label: 'sandboxReadOnly' },
    { value: 'danger-full-access', label: 'sandboxDangerFullAccess' },
  ])
})

test('web sandbox mode normalizes workspace-write to read-only for display only', () => {
  assert.equal(toWebSandboxMode('workspace-write'), 'read-only')
  assert.equal(toWebSandboxMode('danger-full-access'), 'danger-full-access')
})

test('settings basic tab uses shared select and toggle controls for sandbox settings', () => {
  const source = readFileSync(resolve(import.meta.dirname, '../src/components/settings/SettingsBasicTab.vue'), 'utf8')

  assert.match(source, /import AppSelect/)
  assert.match(source, /import ToggleSwitch/)
  assert.match(source, /<AppSelect[\s\S]*sandboxMode/)
  assert.match(source, /<ToggleSwitch[\s\S]*sandboxNetworkEnabled/)
  assert.doesNotMatch(source, /<select class="settings-select"/)
  assert.doesNotMatch(source, /type="checkbox"/)
})

test('settings panel rolls sandbox setting changes back after save failures', () => {
  const source = readFileSync(resolve(import.meta.dirname, '../src/components/settings/SettingsPanel.vue'), 'utf8')

  assert.match(source, /const previousMode = sandboxMode\.value/)
  assert.match(source, /sandboxMode\.value = previousMode/)
  assert.match(source, /const previousEnabled = sandboxNetworkEnabled\.value/)
  assert.match(source, /sandboxNetworkEnabled\.value = previousEnabled/)
  assert.match(source, /sandboxSaveFailed/)
})
