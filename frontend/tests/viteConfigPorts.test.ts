import assert from 'node:assert/strict'
import { mkdtempSync, mkdirSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { pathToFileURL } from 'node:url'
import test from 'node:test'
import type { UserConfig } from 'vite'

const projectRoot = resolve(import.meta.dirname, '..')
const viteConfigPath = resolve(projectRoot, 'vite.config.ts')

async function loadViteConfig(caseName: string): Promise<UserConfig> {
  const url = pathToFileURL(viteConfigPath)
  url.searchParams.set('case', `${caseName}-${Date.now()}-${Math.random()}`)
  const mod = await import(url.href)
  return mod.default as UserConfig
}

function resetPortEnv() {
  delete process.env.SERVER_PORT
  delete process.env.FRONTEND_PORT
  delete process.env.SLIMEBOT_HOME
}

function useEmptyConfigHome() {
  process.env.SLIMEBOT_HOME = mkdtempSync(join(tmpdir(), 'slimebot-vite-empty-'))
}

test('vite config defaults to less common development ports', async () => {
  resetPortEnv()
  useEmptyConfigHome()

  const config = await loadViteConfig('default-ports')

  assert.equal(config.server?.port, 7391)
  assert.equal(config.server?.proxy?.['/api']?.target, 'http://localhost:6247')
  assert.equal(config.server?.proxy?.['/ws']?.target, 'ws://localhost:6247')
})

test('vite config reads development ports from environment variables first', async () => {
  resetPortEnv()
  process.env.SERVER_PORT = '18081'
  process.env.FRONTEND_PORT = '15174'

  const config = await loadViteConfig('env-ports')

  assert.equal(config.server?.port, 15174)
  assert.equal(config.server?.proxy?.['/api']?.target, 'http://localhost:18081')
  assert.equal(config.server?.proxy?.['/ws']?.target, 'ws://localhost:18081')
})

test('vite config reads development ports from SLIMEBOT_HOME env file', async () => {
  resetPortEnv()
  const home = mkdtempSync(join(tmpdir(), 'slimebot-vite-config-'))
  mkdirSync(home, { recursive: true })
  writeFileSync(join(home, '.env'), 'SERVER_PORT=18082\nFRONTEND_PORT=15175\n')
  process.env.SLIMEBOT_HOME = home

  const config = await loadViteConfig('file-ports')

  assert.equal(config.server?.port, 15175)
  assert.equal(config.server?.proxy?.['/api']?.target, 'http://localhost:18082')
  assert.equal(config.server?.proxy?.['/ws']?.target, 'ws://localhost:18082')
})
