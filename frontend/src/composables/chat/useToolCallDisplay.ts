import { computed, type MaybeRefOrGetter, toValue } from 'vue'
import { mdiBookOpenOutline, mdiCogPlayOutline, mdiFileDocumentOutline, mdiFileEditOutline, mdiFilePlusOutline, mdiFileSearchOutline, mdiFormatListChecks, mdiHelpCircleOutline, mdiServerNetwork, mdiSourceBranch, mdiWeb, mdiWebBox } from '@mdi/js'
import type { ToolCallItem } from '../../api/chat'
import { buildToolCallSummary } from '../../utils/toolDisplay'

type Translate = (key: string) => string

export function getToolCallIcon(toolName: string) {
  if (toolName === 'run_subagent') return mdiSourceBranch
  if (toolName === 'http_request' || toolName === 'web_search') return mdiWeb
  if (toolName === 'search_files') return mdiFileSearchOutline
  if (toolName === 'web_extract') return mdiWebBox
  if (toolName === 'skills') return mdiBookOpenOutline
  if (toolName === 'todo') return mdiFormatListChecks
  if (toolName === 'process') return mdiCogPlayOutline
  if (toolName === 'ask_questions') return mdiHelpCircleOutline
  if (toolName === 'file_read') return mdiFileDocumentOutline
  if (toolName === 'file_edit') return mdiFileEditOutline
  if (toolName === 'file_write') return mdiFilePlusOutline
  return mdiServerNetwork
}

export function getToolCallLabel(toolName: string, t: Translate) {
  if (toolName === 'exec') return t('toolExec')
  if (toolName === 'http_request') return t('toolHttpRequest')
  if (toolName === 'web_search') return t('toolWebSearch')
  if (toolName === 'search_files') return t('toolSearchFiles')
  if (toolName === 'web_extract') return t('toolWebExtract')
  if (toolName === 'skills') return t('toolSkills')
  if (toolName === 'todo') return t('toolTodo')
  if (toolName === 'process') return t('toolProcess')
  if (toolName === 'run_subagent') return t('toolRunSubagent')
  if (toolName === 'ask_questions') return t('toolAskQuestions')
  if (toolName === 'file_read') return t('toolFileRead')
  if (toolName === 'file_edit') return t('toolFileEdit')
  if (toolName === 'file_write') return t('toolFileWrite')
  return toolName
}

export function getToolCallCommandLabel(toolName: string, command: string, t: Translate) {
  const label = getToolCallLabel(toolName, t)
  if (label !== toolName) return ''
  return command.trim()
}

export function getToolCallStatusLabel(status: ToolCallItem['status'], t: Translate) {
  switch (status) {
    case 'pending': return t('toolCallPending')
    case 'reviewing': return t('toolCallReviewing')
    case 'executing': return t('toolCallExecuting')
    case 'completed': return t('toolCallCompleted')
    case 'rejected': return t('toolCallRejected')
    case 'error': return t('toolCallError')
    default: return ''
  }
}

export function getToolCallStatusTone(status: ToolCallItem['status']) {
  switch (status) {
    case 'pending': return 'pending'
    case 'reviewing': return 'executing'
    case 'executing': return 'executing'
    case 'completed': return 'success'
    case 'rejected':
    case 'error': return 'error'
    default: return 'default'
  }
}

export function useToolCallDisplay(item: MaybeRefOrGetter<ToolCallItem>, t: Translate) {
  const toolIcon = computed(() => getToolCallIcon(toValue(item).toolName))
  const toolLabel = computed(() => getToolCallLabel(toValue(item).toolName, t))
  const toolCommandLabel = computed(() => getToolCallCommandLabel(toValue(item).toolName, toValue(item).command, t))
  const statusLabel = computed(() => getToolCallStatusLabel(toValue(item).status, t))
  const statusTone = computed(() => getToolCallStatusTone(toValue(item).status))
  const toolSummary = computed(() => buildToolCallSummary(toValue(item)))

  const statusDotClass = computed(() => `status-dot-${statusTone.value}`)
  const statusTextClass = computed(() => `tool-status-text tool-status-text-${statusTone.value}`)

  return {
    toolIcon,
    toolLabel,
    toolCommandLabel,
    toolSummary,
    statusLabel,
    statusDotClass,
    statusTextClass,
  }
}
