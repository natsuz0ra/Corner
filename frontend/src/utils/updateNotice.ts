import type { UpdateCheckResult } from '@/types/update'

export const UPDATE_NOTICE_READ_STORAGE_KEY = 'slimebot.update.readLatest'

type StorageLike = Pick<Storage, 'getItem' | 'setItem'>

export function shouldShowUpdateNotice(
  check: Pick<UpdateCheckResult, 'updateAvailable' | 'latest'> | null | undefined,
  readLatestVersion: string | null | undefined,
): boolean {
  const latest = check?.latest?.trim() || ''
  return Boolean(check?.updateAvailable && latest && latest !== (readLatestVersion || ''))
}

export function readUpdateNoticeVersion(storage?: StorageLike | null): string {
  try {
    return storage?.getItem(UPDATE_NOTICE_READ_STORAGE_KEY) || ''
  } catch {
    return ''
  }
}

export function writeUpdateNoticeVersion(version: string, storage?: StorageLike | null): void {
  const latest = version.trim()
  if (!latest) return
  try {
    storage?.setItem(UPDATE_NOTICE_READ_STORAGE_KEY, latest)
  } catch {
    // Storage may be unavailable in private mode; the notice can safely reappear.
  }
}
