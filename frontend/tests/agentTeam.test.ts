import assert from 'node:assert/strict'
import test from 'node:test'
import { dispatchChatSocketMessage } from '../src/api/chatSocket'
import {
  createAgentTeamState,
  formatAgentTeamDuration,
  getAgentTeamStatusLabel,
  mergeAgentTeamMember,
  mergeAgentTeamRun,
  selectAgentTeamMembers,
  type AgentTeamMember,
  type AgentTeamRun,
} from '../src/utils/agentTeam'

function team(overrides: Partial<AgentTeamRun> = {}): AgentTeamRun {
  return {
    id: 'team-1', sessionId: 'session-1', requestId: 'request-1', status: 'running',
    maxMembers: 8, maxParallel: 4, startedAt: '2026-07-29T00:00:00Z',
    createdAt: '2026-07-29T00:00:00Z', updatedAt: '2026-07-29T00:00:00Z', members: [],
    ...overrides,
  }
}

function member(id: string, createdAt: string, overrides: Partial<AgentTeamMember> = {}): AgentTeamMember {
  return {
    id, teamRunId: 'team-1', toolCallId: `tool-${id}`, title: id, task: `task-${id}`,
    status: 'running', createdAt, updatedAt: createdAt, ...overrides,
  }
}

test('Agent Team 成员在乱序事件中仍按 createdAt 和 id 稳定排列', () => {
  const state = createAgentTeamState()
  mergeAgentTeamRun(state, team())
  mergeAgentTeamMember(state, member('member-b', '2026-07-29T00:00:02Z'))
  mergeAgentTeamMember(state, member('member-c', '2026-07-29T00:00:01Z'))
  mergeAgentTeamMember(state, member('member-a', '2026-07-29T00:00:01Z'))

  assert.deepEqual(selectAgentTeamMembers(state, 'team-1').map((item) => item.id), ['member-a', 'member-c', 'member-b'])
})

test('成员事件先于 team_start 到达时会被后续服务端数据补全', () => {
  const state = createAgentTeamState()
  mergeAgentTeamMember(state, member('member-a', '2026-07-29T00:00:01Z'))
  mergeAgentTeamRun(state, team({ requestId: 'request-from-server', maxParallel: 3 }))

  assert.equal(state.runs[0]?.requestId, 'request-from-server')
  assert.equal(state.runs[0]?.maxParallel, 3)
  assert.equal(state.runs[0]?.members[0]?.id, 'member-a')
})

test('重复成员事件保留已有字段并更新终态', () => {
  const state = createAgentTeamState()
  mergeAgentTeamMember(state, member('member-a', '2026-07-29T00:00:01Z', { task: 'inspect API' }))
  mergeAgentTeamMember(state, member('member-a', '2026-07-29T00:00:01Z', { task: '', status: 'failed', error: 'boom' }))

  const got = selectAgentTeamMembers(state, 'team-1')[0]
  assert.equal(got?.task, 'inspect API')
  assert.equal(got?.status, 'failed')
  assert.equal(got?.error, 'boom')
})

test('所有 Team 终态都有可见文字文案', () => {
  assert.deepEqual(
    ['running', 'succeeded', 'partial_failed', 'failed', 'canceled', 'interrupted'].map(getAgentTeamStatusLabel),
    ['进行中', '已完成', '部分失败', '失败', '已取消', '已中断'],
  )
})

test('Team 耗时格式化支持毫秒、秒和分钟', () => {
  assert.equal(formatAgentTeamDuration(640), '640ms')
  assert.equal(formatAgentTeamDuration(12_300), '12.3s')
  assert.equal(formatAgentTeamDuration(125_000), '2m 5s')
})

test('socket 分派 team_start、team_done 并保留成员标识', () => {
  const events: string[] = []
  const handlers = {
    onSession() {}, onStart() {}, onChunk() {}, onSessionTitle() {}, onDone() {}, onError() {},
    onTeamStart(data: AgentTeamRun) { events.push(`start:${data.id}:${data.status}`) },
    onTeamDone(data: AgentTeamRun) { events.push(`done:${data.id}:${data.status}`) },
    onSubagentStart(data: { teamRunId?: string; memberRunId?: string }) {
      events.push(`member:${data.teamRunId}:${data.memberRunId}`)
    },
  }

  dispatchChatSocketMessage(JSON.stringify({
    type: 'team_start', sessionId: 'session-1', id: 'team-1', requestId: 'request-1', status: 'running',
    maxMembers: 8, maxParallel: 4, startedAt: '2026-07-29T00:00:00Z', createdAt: '2026-07-29T00:00:00Z', updatedAt: '2026-07-29T00:00:00Z',
  }), handlers)
  dispatchChatSocketMessage(JSON.stringify({
    type: 'subagent_start', sessionId: 'session-1', parentToolCallId: 'tool-1', subagentRunId: 'sub-1',
    teamRunId: 'team-1', memberRunId: 'member-1', title: '研究', task: '检查后端',
  }), handlers)
  dispatchChatSocketMessage(JSON.stringify({
    type: 'team_done', sessionId: 'session-1', id: 'team-1', requestId: 'request-1', status: 'succeeded',
    maxMembers: 8, maxParallel: 4, startedAt: '2026-07-29T00:00:00Z', finishedAt: '2026-07-29T00:00:05Z',
    createdAt: '2026-07-29T00:00:00Z', updatedAt: '2026-07-29T00:00:05Z',
  }), handlers)

  assert.deepEqual(events, ['start:team-1:running', 'member:team-1:member-1', 'done:team-1:succeeded'])
})
