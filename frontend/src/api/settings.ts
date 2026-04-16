import { apiClient } from './client'
import type { AppSettings } from '@/types/settings'

export type { AppSettings, LLMConfig, MCPConfig, MessagePlatformConfig, SkillItem } from '@/types/settings'

type SettingsPayload = {
  language: 'zh-CN' | 'en-US'
  defaultModel?: string
  messagePlatformDefaultModel?: string
  webSearchApiKey?: string
  approvalMode?: 'standard' | 'auto'
}

export const settingAPI = {
  get: async (): Promise<AppSettings> => {
    const data = (await apiClient.get<SettingsPayload>('/api/settings')).data
    return {
      ...data,
      webSearchKey: data.webSearchApiKey,
      approvalMode: data.approvalMode || 'standard',
    }
  },
  update: async (payload: Partial<AppSettings>) => {
    const wirePayload: Partial<SettingsPayload> = {
      ...payload,
      webSearchApiKey: payload.webSearchKey,
      approvalMode: payload.approvalMode,
    }
    return apiClient.put('/api/settings', wirePayload)
  },
}
