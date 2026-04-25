import assert from 'node:assert/strict'
import test from 'node:test'
import type { SessionHistoryThinkingItem, ToolCallItem } from '../src/api/chat'
import { buildInterleavedTimeline } from '../src/utils/replyBatchBuilder'

test('buildInterleavedTimeline renders plan marker content as a plan item', () => {
  const timeline = buildInterleavedTimeline(
    [],
    [
      'Intro text.',
      '<!-- PLAN_START -->',
      '# Plan',
      '',
      '- Step one',
      '<!-- PLAN_END -->',
      'After text.',
    ].join('\n'),
  )

  assert.deepEqual(
    timeline.map((entry) => entry.kind),
    ['text', 'plan', 'text'],
  )
  assert.equal(timeline[1]!.kind, 'plan')
  assert.match('content' in timeline[1]! ? timeline[1].content : '', /# Plan/)
  assert.match('content' in timeline[1]! ? timeline[1].content : '', /Step one/)
})

test('buildInterleavedTimeline keeps an unclosed plan marker as a plan item', () => {
  const timeline = buildInterleavedTimeline(
    [],
    [
      'Before.',
      '<!-- PLAN_START -->',
      'Still a plan.',
    ].join('\n'),
  )

  assert.deepEqual(
    timeline.map((entry) => entry.kind),
    ['text', 'plan'],
  )
  assert.equal(timeline[1]!.kind, 'plan')
  assert.match('content' in timeline[1]! ? timeline[1].content : '', /Still a plan/)
})

test('buildInterleavedTimeline preserves thinking and tool ordering around plans', () => {
  const toolCalls: ToolCallItem[] = [{
    toolCallId: 'tool-1',
    toolName: 'web_search',
    command: 'search',
    params: {},
    requiresApproval: false,
    status: 'completed',
    output: 'ok',
  }]
  const thinkingRecords: SessionHistoryThinkingItem[] = [{
    thinkingId: 'think-1',
    content: 'reasoning',
    status: 'completed',
    durationMs: 1200,
  }]

  const timeline = buildInterleavedTimeline(
    toolCalls,
    [
      '<!-- THINKING:think-1 -->',
      '<!-- PLAN_START -->',
      'Plan body.',
      '<!-- PLAN_END -->',
      '<!-- TOOL_CALL:tool-1 -->',
      'Done.',
    ].join('\n'),
    thinkingRecords,
  )

  assert.deepEqual(
    timeline.map((entry) => entry.kind),
    ['thinking', 'plan', 'tool_start', 'tool_result', 'text'],
  )
})
