import type { SandboxMode } from '../types/settings'

type SelectOption = {
  value: string
  label: string
}

export type WebSandboxMode = Exclude<SandboxMode, 'workspace-write'>

export function createWebSandboxModeOptions(t: (key: string) => string): SelectOption[] {
  return [
    { value: 'read-only', label: t('sandboxReadOnly') },
    { value: 'danger-full-access', label: t('sandboxDangerFullAccess') },
  ]
}

export function toWebSandboxMode(mode: SandboxMode): WebSandboxMode {
  return mode === 'danger-full-access' ? 'danger-full-access' : 'read-only'
}
