import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const projectRoot = resolve(import.meta.dirname, '..')

test('settings model list shows an OpenAI compatible badge', () => {
  const source = readFileSync(resolve(projectRoot, 'src/components/settings/SettingsLLMTab.vue'), 'utf8')

  assert.match(source, /item\.provider === 'openai'/)
  assert.match(source, /t\('providerOpenAI'\)/)
})
