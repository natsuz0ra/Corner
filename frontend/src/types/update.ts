export type UpdatePhase =
  | 'idle'
  | 'checking'
  | 'downloading'
  | 'installing'
  | 'restarting'
  | 'succeeded'
  | 'failed'

export interface UpdateCheckResult {
  current: string
  latest: string
  updateAvailable: boolean
  canApply: boolean
  reason: string
  releaseName: string
  releaseNotes: string
  releaseUrl: string
  publishedAt: string
  assetName: string
  manualHint: string
}

export interface UpdateJobStatus {
  phase: UpdatePhase
  current: string
  target: string
  message: string
  error: string
  manualHint: string
  updatedAt: string
}
