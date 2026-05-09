import assert from 'node:assert/strict'
import test from 'node:test'
import {
  getApprovalModeDescriptionKey,
  getApprovalModeLabelKey,
  getApprovalModeOptions,
  getApprovalModeTone,
} from '../src/utils/approvalMode'

test('getApprovalModeOptions returns the three supported modes in menu order', () => {
  assert.deepEqual(getApprovalModeOptions(), [
    {
      value: 'standard',
      labelKey: 'approvalModeStandard',
      descriptionKey: 'approvalModeStandardDesc',
      tone: 'safe',
    },
    {
      value: 'auto_review',
      labelKey: 'approvalModeAutoReview',
      descriptionKey: 'approvalModeAutoReviewDesc',
      tone: 'review',
    },
    {
      value: 'auto',
      labelKey: 'approvalModeAuto',
      descriptionKey: 'approvalModeAutoDesc',
      tone: 'auto',
    },
  ])
})

test('approval mode helpers fall back to standard for unknown values', () => {
  assert.equal(getApprovalModeLabelKey('auto_review'), 'approvalModeAutoReview')
  assert.equal(getApprovalModeLabelKey('auto'), 'approvalModeAuto')
  assert.equal(getApprovalModeDescriptionKey('auto_review'), 'approvalModeAutoReviewDesc')
  assert.equal(getApprovalModeDescriptionKey('auto'), 'approvalModeAutoDesc')
  assert.equal(getApprovalModeTone('auto_review'), 'review')
  assert.equal(getApprovalModeTone('auto'), 'auto')
  assert.equal(getApprovalModeLabelKey('unexpected'), 'approvalModeStandard')
  assert.equal(getApprovalModeDescriptionKey('unexpected'), 'approvalModeStandardDesc')
  assert.equal(getApprovalModeTone('unexpected'), 'safe')
})
