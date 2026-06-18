import assert from 'node:assert/strict'
import test from 'node:test'
import {
  filterMCPTools,
  formatMCPToolParameters,
  getMCPToolsBadge,
} from '../src/utils/mcpTools'
import type { MCPToolListResponse } from '../src/types/settings'

test('formatMCPToolParameters prefers required parameter names', () => {
  assert.equal(
    formatMCPToolParameters({
      name: 'search',
      functionName: 'mcp_1__search',
      description: '',
      parameterCount: 3,
      requiredParameters: ['query', 'owner'],
      parameters: [],
      inputSchema: {},
    }),
    '3 参数 · 必填 query, owner',
  )
})

test('formatMCPToolParameters tolerates null required parameters from API responses', () => {
  assert.equal(
    formatMCPToolParameters({
      parameterCount: 1,
      requiredParameters: null as unknown as string[],
    }),
    '1 参数',
  )
})

test('getMCPToolsBadge describes loaded, error, disabled and idle states', () => {
  const loaded: MCPToolListResponse = {
    configId: 'mcp-1',
    name: 'github',
    isEnabled: true,
    status: 'loaded',
    toolCount: 2,
    loadedAt: '',
    tools: [],
    error: '',
  }
  assert.deepEqual(getMCPToolsBadge(true, false, loaded), { label: '2 tools', tone: 'ok' })
  assert.deepEqual(getMCPToolsBadge(true, true, undefined), { label: '加载中', tone: 'muted' })
  assert.deepEqual(getMCPToolsBadge(true, false, { ...loaded, status: 'error', error: 'failed' }), { label: 'tools 加载失败', tone: 'warn' })
  assert.deepEqual(getMCPToolsBadge(false, false, undefined), { label: '不加载 tools', tone: 'muted' })
})

test('filterMCPTools matches name and description case insensitively', () => {
  const tools = [
    { name: 'search_repositories', functionName: '', description: 'Find repos', parameterCount: 0, requiredParameters: [], parameters: [], inputSchema: {} },
    { name: 'create_issue', functionName: '', description: 'Open a ticket', parameterCount: 0, requiredParameters: [], parameters: [], inputSchema: {} },
  ]

  assert.equal(filterMCPTools(tools, 'REPOS').length, 1)
  assert.equal(filterMCPTools(tools, 'ticket')[0].name, 'create_issue')
  assert.equal(filterMCPTools(tools, '').length, 2)
})
