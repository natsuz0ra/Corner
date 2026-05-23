type Translate = (key: string, params?: Record<string, unknown>) => string

function stringifyValue(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

function formatStdioCommand(cfg: Record<string, unknown>): string {
  const command = stringifyValue(cfg.command)
  if (!command) return '-'

  const args = Array.isArray(cfg.args)
    ? cfg.args.map((arg) => String(arg).trim()).filter(Boolean)
    : []

  return [command, ...args].join(' ')
}

export function formatMCPPreview(config: string, t: Translate): string {
  try {
    const cfg = JSON.parse(config || '{}') as Record<string, unknown>
    const transport = stringifyValue(cfg.transport) || 'stdio'
    if (transport === 'stdio') {
      return `${transport} | ${formatStdioCommand(cfg)}`
    }
    return `${transport} | ${stringifyValue(cfg.url) || '-'}`
  } catch {
    return t('mcpJsonInvalid')
  }
}
