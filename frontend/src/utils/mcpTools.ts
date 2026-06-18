import type { MCPToolItem, MCPToolListResponse } from '@/types/settings'

export type MCPToolsBadgeTone = 'ok' | 'warn' | 'muted'

export interface MCPToolsBadge {
  label: string
  tone: MCPToolsBadgeTone
}

export function formatMCPToolParameters(tool: Pick<MCPToolItem, 'parameterCount' | 'requiredParameters'>) {
  const count = `${tool.parameterCount} 参数`
  const requiredParameters = Array.isArray(tool.requiredParameters) ? tool.requiredParameters : []
  if (!requiredParameters.length) {
    return count
  }
  return `${count} · 必填 ${requiredParameters.join(', ')}`
}

export function getMCPToolsBadge(enabled: boolean, loading: boolean, response?: MCPToolListResponse): MCPToolsBadge {
  if (!enabled) {
    return { label: '不加载 tools', tone: 'muted' }
  }
  if (loading) {
    return { label: '加载中', tone: 'muted' }
  }
  if (!response) {
    return { label: '未加载 tools', tone: 'muted' }
  }
  if (response.status === 'error') {
    return { label: 'tools 加载失败', tone: 'warn' }
  }
  if (response.status === 'disabled') {
    return { label: '不加载 tools', tone: 'muted' }
  }
  return { label: `${response.toolCount} tools`, tone: 'ok' }
}

export function filterMCPTools(tools: MCPToolItem[], query: string) {
  const needle = query.trim().toLowerCase()
  if (!needle) return tools
  return tools.filter((tool) => {
    return tool.name.toLowerCase().includes(needle) || (tool.description || '').toLowerCase().includes(needle)
  })
}
