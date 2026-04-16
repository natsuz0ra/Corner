export interface ExecOutputPayload {
  stdout: string
  stderr: string
  exit_code: number
  timed_out: boolean
  truncated: boolean
  shell: string
  working_directory: string
  duration_ms: number
}

export interface ToolResultDisplay {
  mode: 'text' | 'exec'
  outputText: string
  exec?: ExecOutputPayload
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
  return decodeCommonEscapes(raw)
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

  if (
    typeof stdout !== 'string' ||
    typeof stderr !== 'string' ||
    typeof exitCode !== 'number' ||
    typeof timedOut !== 'boolean' ||
    typeof truncated !== 'boolean' ||
    typeof shell !== 'string' ||
    typeof workingDirectory !== 'string' ||
    typeof durationMs !== 'number'
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
  }
}

export function formatToolParams(params: Record<string, string>): Array<{ key: string; value: string }> {
  return Object.keys(params)
    .sort()
    .map((key) => ({ key, value: formatDisplayText(params[key] ?? '') }))
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

  return {
    mode: 'text',
    outputText: formatDisplayText(raw),
  }
}
