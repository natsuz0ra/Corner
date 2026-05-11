import assert from 'node:assert/strict'
import test from 'node:test'
import { mdiBookOpenOutline, mdiCogPlayOutline, mdiConsoleLine, mdiFileDocumentOutline, mdiFileEditOutline, mdiFilePlusOutline, mdiFileSearchOutline, mdiFormatListChecks, mdiHelpCircleOutline, mdiWeb, mdiWebBox } from '@mdi/js'
import { getToolCallIcon, getToolCallLabel, getToolCallStatusLabel, getToolCallStatusTone } from '../src/composables/chat/useToolCallDisplay'

const t = (key: string) => `t:${key}`

test('tool call display maps known tools to labels and icons', () => {
  assert.equal(getToolCallLabel('exec', t), 't:toolExec')
  assert.equal(getToolCallLabel('http_request', t), 't:toolHttpRequest')
  assert.equal(getToolCallLabel('file_read', t), 't:toolFileRead')
  assert.equal(getToolCallLabel('file_edit', t), 't:toolFileEdit')
  assert.equal(getToolCallLabel('file_write', t), 't:toolFileWrite')
  assert.equal(getToolCallLabel('search_files', t), 't:toolSearchFiles')
  assert.equal(getToolCallLabel('web_extract', t), 't:toolWebExtract')
  assert.equal(getToolCallLabel('skills', t), 't:toolSkills')
  assert.equal(getToolCallLabel('todo', t), 't:toolTodo')
  assert.equal(getToolCallLabel('process', t), 't:toolProcess')
  assert.equal(getToolCallLabel('custom_tool', t), 'custom_tool')
  assert.equal(getToolCallIcon('web_search'), mdiWeb)
  assert.equal(getToolCallIcon('search_files'), mdiFileSearchOutline)
  assert.equal(getToolCallIcon('web_extract'), mdiWebBox)
  assert.equal(getToolCallIcon('skills'), mdiBookOpenOutline)
  assert.equal(getToolCallIcon('todo'), mdiFormatListChecks)
  assert.equal(getToolCallIcon('process'), mdiCogPlayOutline)
  assert.equal(getToolCallIcon('ask_questions'), mdiHelpCircleOutline)
  assert.equal(getToolCallIcon('file_read'), mdiFileDocumentOutline)
  assert.equal(getToolCallIcon('file_edit'), mdiFileEditOutline)
  assert.equal(getToolCallIcon('file_write'), mdiFilePlusOutline)
  assert.equal(getToolCallIcon('custom_tool'), mdiConsoleLine)
})

test('tool call display maps statuses to labels and tones', () => {
  assert.equal(getToolCallStatusLabel('pending', t), 't:toolCallPending')
  assert.equal(getToolCallStatusTone('completed'), 'success')
  assert.equal(getToolCallStatusTone('rejected'), 'error')
  assert.equal(getToolCallStatusTone('error'), 'error')
})
