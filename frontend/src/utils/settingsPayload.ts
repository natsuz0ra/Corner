import type { AppSettings, ApprovalMode, ThinkingLevel } from '../types/settings'

export type SettingsPayload = {
  language: 'zh-CN' | 'en-US'
  defaultModel?: string
  messagePlatformDefaultModel?: string
  messagePlatformThinkingLevel?: ThinkingLevel
  messagePlatformApprovalMode?: ApprovalMode
  webSearchApiKey?: string
  approvalMode?: ApprovalMode
  thinkingLevel?: ThinkingLevel
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
  return wirePayload
}
