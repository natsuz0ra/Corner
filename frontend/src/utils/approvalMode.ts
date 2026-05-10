import type { ApprovalMode } from '@/types/settings'

export type ApprovalModeTone = 'safe' | 'review' | 'auto'

export type ApprovalModeOption = {
  value: ApprovalMode
  labelKey: string
  descriptionKey: string
  tone: ApprovalModeTone
}

const STANDARD_APPROVAL_MODE_OPTION: ApprovalModeOption = {
  value: 'standard',
  labelKey: 'approvalModeStandard',
  descriptionKey: 'approvalModeStandardDesc',
  tone: 'safe',
}

export const APPROVAL_MODE_OPTIONS: ApprovalModeOption[] = [
  STANDARD_APPROVAL_MODE_OPTION,
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
]

export function getApprovalModeOptions() {
  return APPROVAL_MODE_OPTIONS
}

function getApprovalModeOption(mode: string) {
  return APPROVAL_MODE_OPTIONS.find((item) => item.value === mode) || STANDARD_APPROVAL_MODE_OPTION
}

export function getApprovalModeLabelKey(mode: string) {
  return getApprovalModeOption(mode).labelKey
}

export function getApprovalModeDescriptionKey(mode: string) {
  return getApprovalModeOption(mode).descriptionKey
}

export function getApprovalModeTone(mode: string) {
  return getApprovalModeOption(mode).tone
}
