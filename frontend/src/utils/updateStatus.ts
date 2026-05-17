import type { UpdateCheckResult, UpdateJobStatus, UpdatePhase } from '@/types/update'

export type UpdateTone = 'neutral' | 'info' | 'success' | 'danger'

const phases = new Set<UpdatePhase>([
  'idle',
  'checking',
  'downloading',
  'installing',
  'restarting',
  'succeeded',
  'failed',
])

export function normalizeUpdatePhase(value: unknown): UpdatePhase {
  return typeof value === 'string' && phases.has(value as UpdatePhase) ? value as UpdatePhase : 'idle'
}

export function normalizeUpdateCheck(payload: Partial<UpdateCheckResult> | null | undefined): UpdateCheckResult {
  return {
    current: payload?.current || '',
    latest: payload?.latest || '',
    updateAvailable: Boolean(payload?.updateAvailable),
    canApply: Boolean(payload?.canApply),
    reason: payload?.reason || '',
    releaseName: payload?.releaseName || '',
    releaseNotes: payload?.releaseNotes || '',
    releaseUrl: payload?.releaseUrl || '',
    publishedAt: payload?.publishedAt || '',
    assetName: payload?.assetName || '',
    manualHint: payload?.manualHint || '',
  }
}

export function normalizeUpdateJob(payload: Partial<UpdateJobStatus> | null | undefined): UpdateJobStatus {
  return {
    phase: normalizeUpdatePhase(payload?.phase),
    current: payload?.current || '',
    target: payload?.target || '',
    message: payload?.message || '',
    error: payload?.error || '',
    manualHint: payload?.manualHint || '',
    updatedAt: payload?.updatedAt || '',
  }
}

export function isUpdateJobActive(phase: UpdatePhase): boolean {
  return phase === 'checking' || phase === 'downloading' || phase === 'installing' || phase === 'restarting'
}

export function updatePhaseTone(phase: UpdatePhase): UpdateTone {
  if (phase === 'succeeded') return 'success'
  if (phase === 'failed') return 'danger'
  if (isUpdateJobActive(phase)) return 'info'
  return 'neutral'
}
