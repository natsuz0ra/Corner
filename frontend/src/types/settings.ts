export type ApprovalMode = 'standard' | 'auto_review' | 'auto'
export type ThinkingLevel = 'off' | 'low' | 'medium' | 'high' | 'max'
export type SandboxMode = 'read-only' | 'workspace-write' | 'danger-full-access'
export type SettingsTabKey = 'basic' | 'llm' | 'mcp' | 'skills' | 'agents' | 'platform' | 'about'

export interface AppSettings {
  language: 'zh-CN' | 'en-US'
  defaultModel?: string
  messagePlatformDefaultModel?: string
  messagePlatformThinkingLevel?: ThinkingLevel
  messagePlatformApprovalMode?: ApprovalMode
  webSearchKey?: string
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

export interface LLMConfig {
  id: string
  name: string
  provider: 'openai' | 'anthropic' | 'deepseek'
  baseUrl: string
  apiKey: string
  model: string
  contextSize?: number
}

export interface MCPConfig {
  id: string
  name: string
  config: string
  isEnabled: boolean
  createdAt?: string
  updatedAt?: string
}

export interface SkillItem {
  id: string
  name: string
  relativePath: string
  description: string
  source: string
  sourceLabel: string
  provider: string
  readOnly: boolean
  enabled: boolean
  absolutePath?: string
  uploadedAt: string
  createdAt?: string
  updatedAt?: string
}

export interface MessagePlatformConfig {
  id: string
  platform: string
  displayName: string
  authConfigJson: string
  isEnabled: boolean
  createdAt?: string
  updatedAt?: string
}
