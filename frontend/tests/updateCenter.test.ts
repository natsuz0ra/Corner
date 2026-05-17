import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import test from 'node:test'
import {
  isUpdateJobActive,
  normalizeUpdateCheck,
  updatePhaseTone,
} from '../src/utils/updateStatus'
import { renderMarkdown } from '../src/utils/markdown'

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
  assert.match(source, /renderMarkdown\(notes\)/)
  assert.match(source, /v-html="releaseNotes"/)
  assert.doesNotMatch(source, /<pre v-if="releaseNotes"/)
})

test('home and settings surfaces render unread update notice dots', () => {
  const homeSource = readFileSync(resolve(import.meta.dirname, '../src/pages/HomePage.vue'), 'utf8')
  const sidebarSource = readFileSync(resolve(import.meta.dirname, '../src/components/home/HomeSidebar.vue'), 'utf8')
  const dialogsSource = readFileSync(resolve(import.meta.dirname, '../src/components/home/HomeDialogs.vue'), 'utf8')
  const settingsSource = readFileSync(resolve(import.meta.dirname, '../src/components/settings/SettingsPanel.vue'), 'utf8')

  assert.match(homeSource, /useUpdateNotice\(\)/)
  assert.match(homeSource, /checkForUpdate\(\)/)
  assert.match(homeSource, /:has-update-notice="hasUnreadUpdate"/)
  assert.match(homeSource, /:mark-update-notice-read="markUpdateNoticeRead"/)
  assert.match(sidebarSource, /hasUpdateNotice:\s*boolean/)
  assert.match(sidebarSource, /v-if="hasUpdateNotice"[\s\S]*update-notice-dot/)
  assert.match(dialogsSource, /hasUpdateNotice:\s*boolean/)
  assert.match(dialogsSource, /markUpdateNoticeRead:\s*\(\) => void/)
  assert.match(settingsSource, /hasUpdateNotice:\s*boolean/)
  assert.match(settingsSource, /markUpdateNoticeRead:\s*\(\) => void/)
  assert.match(settingsSource, /<span>\{\{ t\(item\.labelKey\) \}\}<\/span>[\s\S]*item\.key === 'about' && hasUpdateNotice[\s\S]*update-notice-dot/)
  assert.match(settingsSource, /watch\(tab[\s\S]*markUpdateNoticeRead\(\)/)
})

test('release notes markdown renders as html instead of raw markdown', () => {
  const html = renderMarkdown('## Highlights\n- **A** item\n- [Release](https://example.test/release)')

  assert.match(html, /<h2>Highlights<\/h2>/)
  assert.match(html, /<strong>A<\/strong> item/)
  assert.match(html, /target="_blank"/)
  assert.doesNotMatch(html, /## Highlights/)
  assert.doesNotMatch(html, /\*\*A\*\*/)
})

test('i18n contains update center labels in both languages', () => {
  const source = readFileSync(resolve(import.meta.dirname, '../src/i18n.ts'), 'utf8')

  assert.match(source, /updateCenterTitle: '更新中心'/)
  assert.match(source, /updateCenterTitle: 'Update Center'/)
  assert.match(source, /updateRestarting/)
  assert.match(source, /updateManualHint/)
})
