import assert from 'node:assert/strict'
import test from 'node:test'
import { createLanguagePreferenceController, type LanguageCode } from '../src/utils/languagePreference'

function createMemoryStorage(initial?: LanguageCode) {
  let value: string | null = initial ?? null
  return {
    read: () => value,
    write: (next: LanguageCode) => {
      value = next
    },
    get value() {
      return value
    },
  }
}

test('loadLanguage applies authenticated remote language to shared state and locale', async () => {
  let locale = 'zh-CN'
  const storage = createMemoryStorage('zh-CN')
  const controller = createLanguagePreferenceController({
    getLocale: () => locale,
    setLocale: (next) => {
      locale = next
    },
    readLocalLanguage: storage.read,
    writeLocalLanguage: storage.write,
    canUseRemote: () => true,
    fetchRemoteLanguage: async () => 'en-US',
    updateRemoteLanguage: async () => {},
  })

  const loaded = await controller.loadLanguage({ allowRemote: true })

  assert.equal(loaded, 'en-US')
  assert.equal(controller.language.value, 'en-US')
  assert.equal(locale, 'en-US')
  assert.equal(storage.value, 'en-US')
})

test('loadLanguage skips remote settings when remote use is not allowed', async () => {
  let locale = 'zh-CN'
  let fetchCount = 0
  const storage = createMemoryStorage('en-US')
  const controller = createLanguagePreferenceController({
    getLocale: () => locale,
    setLocale: (next) => {
      locale = next
    },
    readLocalLanguage: storage.read,
    writeLocalLanguage: storage.write,
    canUseRemote: () => true,
    fetchRemoteLanguage: async () => {
      fetchCount += 1
      return 'zh-CN'
    },
    updateRemoteLanguage: async () => {},
  })

  const loaded = await controller.loadLanguage({ allowRemote: false })

  assert.equal(loaded, 'en-US')
  assert.equal(fetchCount, 0)
  assert.equal(controller.language.value, 'en-US')
  assert.equal(locale, 'en-US')
})

test('loadLanguage uses the fallback language without local or remote language', async () => {
  let locale = 'unexpected'
  let fetchCount = 0
  const storage = createMemoryStorage()
  const controller = createLanguagePreferenceController({
    getLocale: () => locale,
    setLocale: (next) => {
      locale = next
    },
    readLocalLanguage: storage.read,
    writeLocalLanguage: storage.write,
    canUseRemote: () => true,
    fetchRemoteLanguage: async () => {
      fetchCount += 1
      return 'en-US'
    },
    updateRemoteLanguage: async () => {},
  })

  const loaded = await controller.loadLanguage({ allowRemote: false })

  assert.equal(loaded, 'zh-CN')
  assert.equal(fetchCount, 0)
  assert.equal(controller.language.value, 'zh-CN')
  assert.equal(locale, 'zh-CN')
  assert.equal(storage.value, 'zh-CN')
})

test('language state is shared between controller users', async () => {
  let locale = 'zh-CN'
  const storage = createMemoryStorage('zh-CN')
  const sharedState = createLanguagePreferenceController.createState('zh-CN')
  const first = createLanguagePreferenceController({
    state: sharedState,
    getLocale: () => locale,
    setLocale: (next) => {
      locale = next
    },
    readLocalLanguage: storage.read,
    writeLocalLanguage: storage.write,
    canUseRemote: () => false,
    fetchRemoteLanguage: async () => 'zh-CN',
    updateRemoteLanguage: async () => {},
  })
  const second = createLanguagePreferenceController({
    state: sharedState,
    getLocale: () => locale,
    setLocale: (next) => {
      locale = next
    },
    readLocalLanguage: storage.read,
    writeLocalLanguage: storage.write,
    canUseRemote: () => false,
    fetchRemoteLanguage: async () => 'zh-CN',
    updateRemoteLanguage: async () => {},
  })

  await first.changeLanguage('en-US', { allowRemote: false })

  assert.equal(second.language.value, 'en-US')
  assert.equal(locale, 'en-US')
  assert.equal(storage.value, 'en-US')
})

test('loadLanguage preserves local language when remote loading fails', async () => {
  let locale = 'zh-CN'
  const storage = createMemoryStorage('en-US')
  const controller = createLanguagePreferenceController({
    getLocale: () => locale,
    setLocale: (next) => {
      locale = next
    },
    readLocalLanguage: storage.read,
    writeLocalLanguage: storage.write,
    canUseRemote: () => true,
    fetchRemoteLanguage: async () => {
      throw new Error('network down')
    },
    updateRemoteLanguage: async () => {},
  })

  const loaded = await controller.loadLanguage({ allowRemote: true })

  assert.equal(loaded, 'en-US')
  assert.equal(controller.language.value, 'en-US')
  assert.equal(locale, 'en-US')
  assert.equal(storage.value, 'en-US')
})
