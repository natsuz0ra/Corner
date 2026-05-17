import assert from 'node:assert/strict'
import test from 'node:test'
import {
  UPDATE_NOTICE_READ_STORAGE_KEY,
  readUpdateNoticeVersion,
  shouldShowUpdateNotice,
  writeUpdateNoticeVersion,
} from '../src/utils/updateNotice'

class MemoryStorage {
  private values = new Map<string, string>()

  getItem(key: string) {
    return this.values.get(key) ?? null
  }

  setItem(key: string, value: string) {
    this.values.set(key, value)
  }
}

test('update notice is unread when latest differs from the read version', () => {
  assert.equal(shouldShowUpdateNotice({
    updateAvailable: true,
    latest: 'v1.26.2',
  }, 'v1.26.1'), true)
})

test('update notice is cleared after the same latest version is marked read', () => {
  assert.equal(shouldShowUpdateNotice({
    updateAvailable: true,
    latest: 'v1.26.2',
  }, 'v1.26.2'), false)
})

test('update notice reappears when a newer latest version is detected', () => {
  assert.equal(shouldShowUpdateNotice({
    updateAvailable: true,
    latest: 'v1.26.3',
  }, 'v1.26.2'), true)
})

test('update notice read version is stored under the stable localStorage key', () => {
  const storage = new MemoryStorage()

  writeUpdateNoticeVersion('v1.26.2', storage)

  assert.equal(storage.getItem(UPDATE_NOTICE_READ_STORAGE_KEY), 'v1.26.2')
  assert.equal(readUpdateNoticeVersion(storage), 'v1.26.2')
})
