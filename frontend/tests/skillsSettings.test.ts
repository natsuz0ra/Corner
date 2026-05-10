import assert from 'node:assert/strict'
import test from 'node:test'
import { canDeleteSkill, sortSkillRows } from '../src/utils/skills'
import type { SkillItem } from '../src/types/settings'

function skill(overrides: Partial<SkillItem>): SkillItem {
  return {
    id: 'local',
    name: 'local',
    relativePath: 'skills/local',
    description: 'local skill',
    uploadedAt: '2026-01-01T00:00:00Z',
    source: 'local',
    sourceLabel: 'SlimeBot',
    provider: 'slimebot',
    readOnly: false,
    enabled: true,
    ...overrides,
  }
}

test('sortSkillRows keeps source/provider metadata for rendering external skills', () => {
  const rows = sortSkillRows([
    skill({ id: 'codex:global:alpha', provider: 'codex', sourceLabel: 'Codex', readOnly: true }),
    skill({ id: 'claude:global:alpha', provider: 'claude', sourceLabel: 'Claude Code', readOnly: true, enabled: false }),
  ])

  assert.equal(rows[0].id, 'claude:global:alpha')
  assert.equal(rows[0].sourceLabel, 'Claude Code')
  assert.equal(rows[0].readOnly, true)
  assert.equal(rows[0].enabled, false)
})

test('canDeleteSkill only allows local writable skills', () => {
  assert.equal(canDeleteSkill(skill({ readOnly: false })), true)
  assert.equal(canDeleteSkill(skill({ readOnly: true })), false)
})
