import assert from 'node:assert/strict'
import test from 'node:test'
import { formatMCPPreview } from '../src/utils/mcpPreview'

const translate = (key: string, params?: Record<string, unknown>) => {
  if (key === 'mcpStdioPreview') return `${params?.transport} | ${params?.command}`
  if (key === 'mcpJsonInvalid') return 'MCP config JSON is invalid.'
  return key
}

test('stdio MCP preview includes command and args', () => {
  assert.equal(
    formatMCPPreview(JSON.stringify({ command: 'python', args: ['-m', 'your_module'] }), translate),
    'stdio | python -m your_module',
  )
})

test('stdio MCP preview keeps command visible when translator treats pipe as plural separator', () => {
  const pipeSplittingTranslate = (key: string, params?: Record<string, unknown>) => {
    if (key === 'mcpStdioPreview') return `${params?.transport} | ${params?.command}`.split('|')[0].trim()
    if (key === 'mcpJsonInvalid') return 'MCP config JSON is invalid.'
    return key
  }

  assert.equal(
    formatMCPPreview(JSON.stringify({ command: 'python', args: ['-m', 'your_module'] }), pipeSplittingTranslate),
    'stdio | python -m your_module',
  )
})

test('explicit stdio MCP preview includes command and args', () => {
  assert.equal(
    formatMCPPreview(JSON.stringify({ transport: 'stdio', command: 'node', args: ['server.js'] }), translate),
    'stdio | node server.js',
  )
})

test('stdio MCP preview uses a readable placeholder without command', () => {
  assert.equal(formatMCPPreview(JSON.stringify({ transport: 'stdio' }), translate), 'stdio | -')
})

test('remote MCP preview still shows transport and url', () => {
  assert.equal(
    formatMCPPreview(JSON.stringify({ transport: 'sse', url: 'https://example.test/mcp' }), translate),
    'sse | https://example.test/mcp',
  )
})

test('invalid MCP preview still shows JSON error message', () => {
  assert.equal(formatMCPPreview('{', translate), 'MCP config JSON is invalid.')
})
