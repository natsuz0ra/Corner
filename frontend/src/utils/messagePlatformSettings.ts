import type { ThinkingLevel } from '../types/settings'

type Translate = (key: string) => string

export function shouldSaveMessagePlatformDefaultModel(_modelId: string) {
  return true
}

export function createMessagePlatformThinkingOptions(t: Translate): { value: ThinkingLevel; label: string }[] {
  return [
    { value: 'off', label: t('thinkingOff') },
    { value: 'low', label: t('thinkingLow') },
    { value: 'medium', label: t('thinkingMedium') },
    { value: 'high', label: t('thinkingHigh') },
    { value: 'max', label: t('thinkingMax') },
  ]
}
