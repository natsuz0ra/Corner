import type { AppSettings, ApprovalMode, SandboxMode, ThinkingLevel } from '../types/settings'

export type SettingsPayload = {
  language: 'zh-CN' | 'en-US'
  defaultModel?: string
  messagePlatformDefaultModel?: string
  messagePlatformThinkingLevel?: ThinkingLevel
  messagePlatformApprovalMode?: ApprovalMode
  webSearchApiKey?: string
  approvalMode?: ApprovalMode
  thinkingLevel?: ThinkingLevel
  sandboxMode?: SandboxMode
  sandboxWritableRoots?: string[]
  sandboxNetworkEnabled?: boolean
  sandboxNetworkAllowedDomains?: string[]
  cliSandboxMode?: SandboxMode
  cliSandboxWritableRoots?: string[]
  cliSandboxNetworkEnabled?: boolean
  cliSandboxNetworkAllowedDomains?: string[]
}

export function normalizeSettingsPayload(data: Partial<SettingsPayload>): AppSettings {
  return {
    ...(data as SettingsPayload),
    language: data.language || 'zh-CN',
    webSearchKey: data.webSearchApiKey,
    approvalMode: data.approvalMode || 'standard',
    thinkingLevel: data.thinkingLevel || 'off',
    messagePlatformThinkingLevel: data.messagePlatformThinkingLevel || 'off',
    messagePlatformApprovalMode: data.messagePlatformApprovalMode || 'standard',
    sandboxMode: data.sandboxMode || 'workspace-write',
    sandboxWritableRoots: Array.isArray(data.sandboxWritableRoots) ? data.sandboxWritableRoots : [],
    sandboxNetworkEnabled: data.sandboxNetworkEnabled !== undefined ? data.sandboxNetworkEnabled : true,
    sandboxNetworkAllowedDomains: Array.isArray(data.sandboxNetworkAllowedDomains) ? data.sandboxNetworkAllowedDomains : [],
    cliSandboxMode: data.cliSandboxMode || 'workspace-write',
    cliSandboxWritableRoots: Array.isArray(data.cliSandboxWritableRoots) ? data.cliSandboxWritableRoots : [],
    cliSandboxNetworkEnabled: data.cliSandboxNetworkEnabled !== undefined ? data.cliSandboxNetworkEnabled : true,
    cliSandboxNetworkAllowedDomains: Array.isArray(data.cliSandboxNetworkAllowedDomains) ? data.cliSandboxNetworkAllowedDomains : [],
  }
}

export function buildSettingsPayload(payload: Partial<AppSettings>): Partial<SettingsPayload> {
  const wirePayload: Partial<SettingsPayload> = {}
  if (payload.language !== undefined) wirePayload.language = payload.language
  if (payload.defaultModel !== undefined) wirePayload.defaultModel = payload.defaultModel
  if (payload.messagePlatformDefaultModel !== undefined) wirePayload.messagePlatformDefaultModel = payload.messagePlatformDefaultModel
  if (payload.messagePlatformThinkingLevel !== undefined) wirePayload.messagePlatformThinkingLevel = payload.messagePlatformThinkingLevel
  if (payload.messagePlatformApprovalMode !== undefined) wirePayload.messagePlatformApprovalMode = payload.messagePlatformApprovalMode
  if (payload.webSearchKey !== undefined) wirePayload.webSearchApiKey = payload.webSearchKey
  if (payload.approvalMode !== undefined) wirePayload.approvalMode = payload.approvalMode
  if (payload.thinkingLevel !== undefined) wirePayload.thinkingLevel = payload.thinkingLevel
  if (payload.sandboxMode !== undefined) wirePayload.sandboxMode = payload.sandboxMode
  if (payload.sandboxWritableRoots !== undefined) wirePayload.sandboxWritableRoots = payload.sandboxWritableRoots
  if (payload.sandboxNetworkEnabled !== undefined) wirePayload.sandboxNetworkEnabled = payload.sandboxNetworkEnabled
  if (payload.sandboxNetworkAllowedDomains !== undefined) wirePayload.sandboxNetworkAllowedDomains = payload.sandboxNetworkAllowedDomains
  if (payload.cliSandboxMode !== undefined) wirePayload.cliSandboxMode = payload.cliSandboxMode
  if (payload.cliSandboxWritableRoots !== undefined) wirePayload.cliSandboxWritableRoots = payload.cliSandboxWritableRoots
  if (payload.cliSandboxNetworkEnabled !== undefined) wirePayload.cliSandboxNetworkEnabled = payload.cliSandboxNetworkEnabled
  if (payload.cliSandboxNetworkAllowedDomains !== undefined) wirePayload.cliSandboxNetworkAllowedDomains = payload.cliSandboxNetworkAllowedDomains
  return wirePayload
}
