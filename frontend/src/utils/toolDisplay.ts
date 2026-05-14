import type { ToolCallItem } from '@/api/chat'
import type { ToolCallStatus } from '@/types/chat'
import { fileToolSummaryFromParams, isFileToolName } from './fileToolDisplay'

export interface ExecOutputPayload {
  stdout: string
  stderr: string
  exit_code: number
  timed_out: boolean
  truncated: boolean
  shell: string
  working_directory: string
  duration_ms: number
  sandbox_permissions?: string
  sandbox_mode?: string
}

export interface WebSearchResult {
  title: string
  url: string
  content: string
}

export interface WebSearchPayload {
  query: string
  results: WebSearchResult[]
}

export interface ToolResultDisplay {
  mode: 'text' | 'exec' | 'web_search' | 'ask_questions'
  outputText: string
  exec?: ExecOutputPayload
  webSearch?: WebSearchPayload
}

export interface AskQuestionsAnswer {
  questionId: string
  selectedOption: number
  customAnswer: string
}

export interface AskQuestionsReadableAnswer {
  id: string
  question: string
  answer: string
}

export interface AskQuestionsQuestion {
  id: string
  question: string
  options: string[]
  option_descriptions?: string[]
}

export interface ToolCallSummaryInput {
  toolName: string
  command: string
  params?: Record<string, unknown>
  subagentTitle?: string
  subagentTask?: string
}

export type LightweightToolKind = 'search' | 'web' | 'file_read'

export interface LightweightToolDisplay {
  toolCallId: string
  toolName: string
  command: string
  kind: LightweightToolKind
  label: string
  target: string
  status: ToolCallStatus
  error: string
  count: number
}

export type LightweightToolTimelineRow<T> =
  | { kind: 'timeline'; entry: T }
  | { kind: 'lightweight_tool_group'; id: string; items: LightweightToolDisplay[]; trailing: boolean }

export type LightweightToolItemRow<T> =
  | { kind: 'item'; item: T }
  | { kind: 'lightweight_tool_group'; id: string; items: LightweightToolDisplay[]; trailing: boolean }

export function parseAskQuestionsAnswers(raw: string): AskQuestionsAnswer[] | null {
  const parsed = tryParseJSON(raw)
  if (!Array.isArray(parsed)) return null
  const answers: AskQuestionsAnswer[] = []
  for (const item of parsed) {
    if (!isRecord(item)) return null
    const questionId = item.questionId
    const selectedOption = item.selectedOption
    const customAnswer = item.customAnswer
    if (typeof questionId !== 'string' || typeof selectedOption !== 'number' || typeof customAnswer !== 'string') return null
    answers.push({ questionId, selectedOption, customAnswer })
  }
  return answers.length > 0 ? answers : null
}

export function parseAskQuestionsReadableAnswers(raw: string): AskQuestionsReadableAnswer[] | null {
  const parsed = tryParseJSON(raw)
  if (!Array.isArray(parsed)) return null
  const answers: AskQuestionsReadableAnswer[] = []
  for (const item of parsed) {
    if (!isRecord(item)) return null
    const id = item.id
    const question = item.question
    const answer = item.answer
    if (typeof id !== 'string' || typeof question !== 'string' || typeof answer !== 'string') return null
    answers.push({ id, question, answer })
  }
  return answers.length > 0 ? answers : null
}

export function parseAskQuestionsQuestions(raw: string): AskQuestionsQuestion[] | null {
  const parsed = tryParseJSON(raw)
  if (!Array.isArray(parsed)) return null
  const questions: AskQuestionsQuestion[] = []
  for (const item of parsed) {
    if (!isRecord(item)) return null
    const id = item.id
    const question = item.question
    const options = item.options
    if (typeof id !== 'string' || typeof question !== 'string' || !Array.isArray(options)) return null
    questions.push({ id, question, options: options.filter((o): o is string => typeof o === 'string'), option_descriptions: Array.isArray(item.option_descriptions) ? item.option_descriptions.filter((d: unknown): d is string => typeof d === 'string') : undefined })
  }
  return questions.length > 0 ? questions : null
}

function tryParseJSON(raw: string): unknown | null {
  const trimmed = raw.trim()
  if (!trimmed) return null
  try {
    return JSON.parse(trimmed)
  } catch {
    return null
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export function decodeCommonEscapes(raw: string): string {
  if (!raw.includes('\\')) return raw
  return raw
    .replace(/\\r\\n/g, '\n')
    .replace(/\\n/g, '\n')
    .replace(/\\r/g, '\n')
    .replace(/\\t/g, '\t')
    .replace(/\\\\"/g, '"')
    .replace(/\\\\/g, '\\')
}

export function formatDisplayText(raw: string): string {
  const parsed = tryParseJSON(raw)
  if (parsed !== null) {
    if (typeof parsed === 'string') {
      return decodeCommonEscapes(parsed)
    }
    try {
      return JSON.stringify(parsed, null, 2)
    } catch {
      return raw
    }
  }
  const decoded = decodeCommonEscapes(raw)
  // Filter consecutive empty lines in display only
  return decoded.replace(/\n{2,}/g, '\n').trim()
}

export function parseExecOutputPayload(raw: string): ExecOutputPayload | null {
  const parsed = tryParseJSON(raw)
  if (!isRecord(parsed)) return null

  const stdout = parsed.stdout
  const stderr = parsed.stderr
  const exitCode = parsed.exit_code
  const timedOut = parsed.timed_out
  const truncated = parsed.truncated
  const shell = parsed.shell
  const workingDirectory = parsed.working_directory
  const durationMs = parsed.duration_ms
  const sandboxPermissions = parsed.sandbox_permissions
  const sandboxMode = parsed.sandbox_mode

  if (
    typeof stdout !== 'string' ||
    typeof stderr !== 'string' ||
    typeof exitCode !== 'number' ||
    typeof timedOut !== 'boolean' ||
    typeof truncated !== 'boolean' ||
    typeof shell !== 'string' ||
    typeof workingDirectory !== 'string' ||
    typeof durationMs !== 'number' ||
    (sandboxPermissions !== undefined && typeof sandboxPermissions !== 'string') ||
    (sandboxMode !== undefined && typeof sandboxMode !== 'string')
  ) {
    return null
  }

  return {
    stdout,
    stderr,
    exit_code: exitCode,
    timed_out: timedOut,
    truncated,
    shell,
    working_directory: workingDirectory,
    duration_ms: durationMs,
    sandbox_permissions: sandboxPermissions,
    sandbox_mode: sandboxMode,
  }
}

export function parseWebSearchPayload(raw: string): WebSearchPayload | null {
  const parsed = tryParseJSON(raw)
  if (!isRecord(parsed)) return null

  const query = parsed.query
  const results = parsed.results
  if (typeof query !== 'string' || !Array.isArray(results)) return null

  const normalizedResults: WebSearchResult[] = results
    .map((item) => {
      if (!isRecord(item)) return null
      const title = item.title
      const url = item.url
      const content = typeof item.content === 'string'
        ? item.content
        : typeof item.snippet === 'string'
          ? item.snippet
          : ''
      if (typeof title !== 'string' || typeof url !== 'string') return null
      return { title, url, content }
    })
    .filter((item): item is WebSearchResult => item !== null)

  return {
    query,
    results: normalizedResults,
  }
}

export function formatToolParams(params: Record<string, unknown>): Array<{ key: string; value: string }> {
  return Object.keys(params)
    .sort()
    .map((key) => ({ key, value: formatDisplayText(String(params[key] ?? '')) }))
}

function normalizedParam(params: Record<string, unknown> | undefined, key: string): string {
  return String(params?.[key] ?? '').trim()
}

function firstNonEmptyParam(params: Record<string, unknown> | undefined): { key: string; value: string } | null {
  if (!params) return null
  for (const key of Object.keys(params).sort()) {
    const value = normalizedParam(params, key)
    if (value !== '') return { key, value }
  }
  return null
}

function compactURL(raw: string): string {
  try {
    const parsed = new URL(raw)
    return `${parsed.host}${parsed.pathname}`.replace(/\/$/, '') || parsed.host
  } catch {
    return raw
  }
}

function arrayParamLength(value: unknown): number {
  if (Array.isArray(value)) return value.length
  if (typeof value !== 'string') return 0
  const parsed = tryParseJSON(value)
  return Array.isArray(parsed) ? parsed.length : 0
}

function parseArrayParam(value: unknown): unknown[] {
  if (Array.isArray(value)) return value
  if (typeof value !== 'string') return []
  const parsed = tryParseJSON(value)
  return Array.isArray(parsed) ? parsed : []
}

function baseName(path: string): string {
  const normalized = path.replace(/\\/g, '/').replace(/\/$/, '')
  return normalized.split('/').pop() || path
}

export function isLightweightToolName(toolName?: string) {
  const name = (toolName || '').trim().toLowerCase()
  return name === 'http_request' ||
    name === 'web_extract' ||
    name === 'web_search' ||
    name === 'file_read' ||
    name === 'search_files' ||
    name === 'search_file'
}

export function isLightweightToolCall(item: Pick<ToolCallItem, 'toolName'>) {
  return isLightweightToolName(item.toolName)
}

function lightweightToolKind(toolName: string): LightweightToolKind {
  const name = toolName.trim().toLowerCase()
  if (name === 'file_read') return 'file_read'
  if (name === 'web_extract' || name === 'http_request') return 'web'
  return 'search'
}

function fileReadTargetAndCount(params: Record<string, unknown> | undefined): { target: string; count: number } {
  const requests = parseArrayParam(params?.requests)
  if (requests.length > 0) {
    const paths = requests
      .map((item) => isRecord(item) ? String(item.file_path ?? '').trim() : '')
      .filter(Boolean)
    return {
      target: paths.length === 1 ? paths[0]! : `${requests.length} files`,
      count: requests.length,
    }
  }
  const filePath = normalizedParam(params, 'file_path')
  return { target: filePath || 'file', count: 1 }
}

export function buildLightweightToolDisplay(item: Pick<ToolCallItem, 'toolCallId' | 'toolName' | 'command' | 'params' | 'status' | 'error'>): LightweightToolDisplay | null {
  const toolName = item.toolName.trim().toLowerCase()
  if (!isLightweightToolName(toolName)) return null
  const params = item.params || {}
  const kind = lightweightToolKind(toolName)
  let target = ''
  let label = ''
  let count = 1

  if (toolName === 'web_search') {
    label = 'Search'
    target = normalizedParam(params, 'query') || 'web'
  } else if (toolName === 'search_files' || toolName === 'search_file') {
    label = 'Search files'
    target = normalizedParam(params, 'query') || normalizedParam(params, 'pattern') || 'files'
    const path = normalizedParam(params, 'path')
    const pattern = normalizedParam(params, 'pattern')
    if (path) target += ` in ${path}`
    if (pattern && !target.includes(pattern)) target += ` (${pattern})`
  } else if (toolName === 'web_extract') {
    label = 'Browse'
    target = compactURL(normalizedParam(params, 'url')) || 'web page'
  } else if (toolName === 'http_request') {
    label = 'Request'
    const method = normalizedParam(params, 'method').toUpperCase()
    const url = compactURL(normalizedParam(params, 'url'))
    target = [method, url].filter(Boolean).join(' ') || 'URL'
  } else {
    label = 'Read'
    const read = fileReadTargetAndCount(params)
    target = read.count === 1 ? baseName(read.target) : read.target
    count = read.count
  }

  return {
    toolCallId: item.toolCallId,
    toolName: item.toolName,
    command: item.command,
    kind,
    label,
    target,
    status: item.status,
    error: item.error || '',
    count,
  }
}

export function buildLightweightToolGroupSummary(items: LightweightToolDisplay[], locale: 'zh' | 'en' = 'zh') {
  let searches = 0
  let webPages = 0
  let files = 0
  for (const item of items) {
    if (item.kind === 'search') searches += 1
    if (item.kind === 'web') webPages += 1
    if (item.kind === 'file_read') files += Math.max(1, item.count)
  }
  const parts: string[] = []
  if (locale === 'en') {
    if (searches > 0) parts.push(`${searches} ${searches === 1 ? 'search' : 'searches'}`)
    if (webPages > 0) parts.push(`${webPages} web ${webPages === 1 ? 'page' : 'pages'}`)
    if (files > 0) parts.push(`${files} ${files === 1 ? 'file' : 'files'} read`)
    return parts.join(', ')
  }
  if (searches > 0) parts.push(`搜索 ${searches} 次`)
  if (webPages > 0) parts.push(`浏览 ${webPages} 个网页`)
  if (files > 0) parts.push(`读取 ${files} 个文件`)
  return parts.join('，')
}

export function buildLightweightToolTimelineRows<T extends { kind: string; id: string; toolCallId?: string }>(
  timeline: T[],
  getTool: (toolCallId: string) => ToolCallItem | undefined,
): LightweightToolTimelineRow<T>[] {
  const rows: LightweightToolTimelineRow<T>[] = []
  let pending: LightweightToolDisplay[] = []

  const flush = (trailing: boolean) => {
    if (pending.length === 0) return
    rows.push({
      kind: 'lightweight_tool_group',
      id: `lightweight-${pending.map((item) => item.toolCallId).join('-')}`,
      items: pending,
      trailing,
    })
    pending = []
  }

  for (const entry of timeline) {
    if ((entry.kind === 'tool_start' || entry.kind === 'tool_result') && entry.toolCallId) {
      const tool = getTool(entry.toolCallId)
      const display = tool ? buildLightweightToolDisplay(tool) : null
      if (display) {
        if (entry.kind === 'tool_start') pending.push(display)
        continue
      }
    }
    flush(false)
    rows.push({ kind: 'timeline', entry })
  }
  flush(true)
  return rows
}

export function buildLightweightToolRows<T extends { id: string }>(
  items: T[],
  getTool: (item: T) => ToolCallItem | undefined,
): LightweightToolItemRow<T>[] {
  const rows: LightweightToolItemRow<T>[] = []
  let pending: LightweightToolDisplay[] = []

  const flush = (trailing: boolean) => {
    if (pending.length === 0) return
    rows.push({
      kind: 'lightweight_tool_group',
      id: `lightweight-${pending.map((item) => item.toolCallId).join('-')}`,
      items: pending,
      trailing,
    })
    pending = []
  }

  for (const item of items) {
    const tool = getTool(item)
    const display = tool ? buildLightweightToolDisplay(tool) : null
    if (display) {
      pending.push(display)
      continue
    }
    flush(false)
    rows.push({ kind: 'item', item })
  }
  flush(true)
  return rows
}

export function getToolSummaryParamKeys(toolCall: ToolCallSummaryInput): string[] {
  const toolName = toolCall.toolName.trim().toLowerCase()
  const command = toolCall.command.trim().toLowerCase()
  const params = toolCall.params || {}

  if (isFileToolName(toolName)) {
    const keys: string[] = []
    if (normalizedParam(params, 'file_path') !== '') keys.push('file_path')
    if (normalizedParam(params, 'requests') !== '') keys.push('requests')
    if (normalizedParam(params, 'edits') !== '') keys.push('edits')
    if (normalizedParam(params, 'writes') !== '') keys.push('writes')
    return keys
  }

  if (toolName === 'exec' && command === 'run' && normalizedParam(params, 'description') !== '') {
    return ['description']
  }
  if (toolName === 'web_search' && normalizedParam(params, 'query') !== '') {
    return ['query']
  }
  if (toolName === 'search_files') {
    const keys: string[] = []
    if (normalizedParam(params, 'query') !== '') keys.push('query')
    if (normalizedParam(params, 'path') !== '') keys.push('path')
    if (normalizedParam(params, 'pattern') !== '') keys.push('pattern')
    if (keys.length > 0) return keys
  }
  if (toolName === 'web_extract' && normalizedParam(params, 'url') !== '') {
    return ['url']
  }
  if (toolName === 'skills' && command === 'view' && normalizedParam(params, 'name') !== '') {
    return ['name']
  }
  if (toolName === 'todo' && command === 'update' && normalizedParam(params, 'items') !== '') {
    return ['items']
  }
  if (toolName === 'process' && (command === 'status' || command === 'stop') && normalizedParam(params, 'process_id') !== '') {
    return ['process_id']
  }
  if (toolName === 'run_subagent') {
    const keys: string[] = []
    if (normalizedParam(params, 'title') !== '') keys.push('title')
    if (normalizedParam(params, 'task') !== '') keys.push('task')
    if (keys.length > 0) return keys
  }
  if (toolName === 'http_request' && command === 'request') {
    const keys: string[] = []
    if (normalizedParam(params, 'method') !== '') keys.push('method')
    if (normalizedParam(params, 'url') !== '') keys.push('url')
    if (keys.length > 0) return keys
  }

  const fallback = firstNonEmptyParam(params)
  return fallback ? [fallback.key] : []
}

export function buildToolCallSummary(toolCall: ToolCallSummaryInput): string {
  const toolName = toolCall.toolName.trim().toLowerCase()
  const command = toolCall.command.trim().toLowerCase()
  const params = toolCall.params || {}

  if (isFileToolName(toolName)) {
    return fileToolSummaryFromParams(toolName, params)
  }

  if (toolName === 'exec' && command === 'run') {
    return normalizedParam(params, 'description')
  }
  if (toolName === 'web_search') {
    const query = normalizedParam(params, 'query')
    return query
  }
  if (toolName === 'search_files') {
    const query = normalizedParam(params, 'query')
    if (!query) return ''
    const path = normalizedParam(params, 'path')
    const pattern = normalizedParam(params, 'pattern')
    let summary = query
    if (path) summary += ` in ${path}`
    if (pattern) summary += ` (${pattern})`
    return summary
  }
  if (toolName === 'web_extract') {
    return compactURL(normalizedParam(params, 'url'))
  }
  if (toolName === 'skills') {
    if (command === 'list') return 'List skills'
    if (command === 'view') return normalizedParam(params, 'name')
  }
  if (toolName === 'todo') {
    if (command === 'list') return 'List todos'
    if (command === 'update') {
      const count = arrayParamLength(params.items)
      return count > 0 ? `Update ${count} ${count === 1 ? 'todo' : 'todos'}` : 'Update todos'
    }
  }
  if (toolName === 'process') {
    if (command === 'list') return 'List processes'
    if (command === 'status' || command === 'stop') return normalizedParam(params, 'process_id')
  }
  if (toolName === 'run_subagent') {
    const title = String(toolCall.subagentTitle ?? '').trim() || normalizedParam(params, 'title')
    if (title) return title
    const task = normalizedParam(params, 'task') || String(toolCall.subagentTask ?? '').trim()
    return task ? `task: ${task}` : ''
  }
  if (toolName === 'http_request' && command === 'request') {
    const method = normalizedParam(params, 'method').toUpperCase()
    const url = normalizedParam(params, 'url')
    return [method, url].filter(Boolean).join(' ')
  }

  const fallback = firstNonEmptyParam(params)
  return fallback ? `${fallback.key}: ${fallback.value}` : ''
}

export function filterToolParamsForDetail(toolCall: ToolCallSummaryInput): Record<string, unknown> {
  const hidden = new Set(getToolSummaryParamKeys(toolCall))
  const result: Record<string, unknown> = {}
  for (const [key, value] of Object.entries(toolCall.params || {})) {
    if (!hidden.has(key)) result[key] = value
  }
  return result
}

export function buildToolResultDisplay(toolName: string, command: string, output?: string): ToolResultDisplay {
  const raw = output || ''
  if (toolName === 'exec' && command === 'run') {
    const exec = parseExecOutputPayload(raw)
    if (exec) {
      return {
        mode: 'exec',
        outputText: '',
        exec,
      }
    }
  }

  if (toolName === 'web_search') {
    const webSearch = parseWebSearchPayload(raw)
    if (webSearch) {
      return {
        mode: 'web_search',
        outputText: formatDisplayText(raw),
        webSearch,
      }
    }
  }

  if (toolName === 'ask_questions') {
    const answers = parseAskQuestionsAnswers(raw)
    if (answers) {
      return {
        mode: 'ask_questions',
        outputText: raw,
      }
    }
  }

  return {
    mode: 'text',
    outputText: formatDisplayText(raw),
  }
}
