import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import test from 'node:test'

const projectRoot = resolve(import.meta.dirname, '..')

test('Team 摘要使用 dialog 入口且不再内嵌完整工具树', () => {
  const source = readFileSync(resolve(projectRoot, 'src/components/chat/AgentTeamBlock.vue'), 'utf8')

  assert.match(source, /aria-haspopup="dialog"/)
  assert.match(source, /AgentTeamDetailDialog/)
  assert.doesNotMatch(source, /<ToolCallInline/)
})

test('Team 详情弹窗只在待审批时复用工具审批组件', () => {
  const path = resolve(projectRoot, 'src/components/chat/AgentTeamDetailDialog.vue')
  assert.ok(existsSync(path), 'AgentTeamDetailDialog.vue should exist')
  const source = readFileSync(path, 'utf8')

  assert.match(source, /pendingTool/)
  assert.match(source, /<ToolCallInline/)
  assert.match(source, /buildAgentTeamResultPreview/)
  assert.doesNotMatch(source, /nested-tools=/)
})

test('AppDialog 管理焦点锁定和关闭后的焦点归还', () => {
  const source = readFileSync(resolve(projectRoot, 'src/components/ui/AppDialog.vue'), 'utf8')

  assert.match(source, /previouslyFocused/)
  assert.match(source, /focusableElements/)
  assert.match(source, /e\.key === 'Tab'/)
  assert.match(source, /panelRef/)
})
