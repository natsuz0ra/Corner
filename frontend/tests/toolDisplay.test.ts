import assert from 'node:assert/strict'
import test from 'node:test'
import type { ToolCallItem } from '../src/api/chat'
import {
  buildToolCallSummary,
  filterToolParamsForDetail,
} from '../src/utils/toolDisplay'

function tool(overrides: Partial<ToolCallItem>): ToolCallItem {
  return {
    toolCallId: 'call-1',
    toolName: 'exec',
    command: 'run',
    params: {},
    requiresApproval: true,
    status: 'pending',
    ...overrides,
  }
}

test('buildToolCallSummary uses exec description', () => {
  const item = tool({ params: { command: 'go test ./...', description: 'Run Go tests' } })

  assert.equal(buildToolCallSummary(item), 'Run Go tests')
})

test('buildToolCallSummary uses query and http request fields', () => {
  assert.equal(
    buildToolCallSummary(tool({ toolName: 'web_search', command: 'search', params: { query: 'SlimeBot latest' } })),
    'query: SlimeBot latest',
  )
  assert.equal(
    buildToolCallSummary(tool({ toolName: 'http_request', command: 'request', params: { method: 'post', url: 'https://example.test/api' } })),
    'POST https://example.test/api',
  )
})

test('buildToolCallSummary hides missing legacy exec description', () => {
  assert.equal(buildToolCallSummary(tool({ params: { command: 'go test ./...' } })), '')
})

test('filterToolParamsForDetail removes params already shown in summary', () => {
  assert.deepEqual(
    filterToolParamsForDetail(tool({ params: { command: 'go test ./...', description: 'Run Go tests' } })),
    { command: 'go test ./...' },
  )
  assert.deepEqual(
    filterToolParamsForDetail(tool({ toolName: 'web_search', command: 'search', params: { query: 'SlimeBot latest' } })),
    {},
  )
})
