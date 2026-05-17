import { computed, ref } from 'vue'

import { updateAPI } from '@/api/update'
import type { UpdateCheckResult } from '@/types/update'
import {
  readUpdateNoticeVersion,
  shouldShowUpdateNotice,
  writeUpdateNoticeVersion,
} from '@/utils/updateNotice'

const checkResult = ref<UpdateCheckResult | null>(null)
const checking = ref(false)
const readLatestVersion = ref(readUpdateNoticeVersion(typeof window !== 'undefined' ? window.localStorage : null))

function browserStorage() {
  return typeof window !== 'undefined' ? window.localStorage : null
}

export function useUpdateNotice() {
  const latestVersion = computed(() => checkResult.value?.latest || '')
  const hasUnreadUpdate = computed(() =>
    shouldShowUpdateNotice(checkResult.value, readLatestVersion.value),
  )

  function setUpdateCheckResult(result: UpdateCheckResult | null) {
    checkResult.value = result
  }

  async function checkForUpdate() {
    if (checking.value) return checkResult.value
    checking.value = true
    try {
      const result = await updateAPI.check(false)
      setUpdateCheckResult(result)
      return result
    } catch {
      return null
    } finally {
      checking.value = false
    }
  }

  function markUpdateNoticeRead() {
    const latest = latestVersion.value
    if (!latest) return
    writeUpdateNoticeVersion(latest, browserStorage())
    readLatestVersion.value = latest
  }

  return {
    checkResult,
    checking,
    latestVersion,
    hasUnreadUpdate,
    checkForUpdate,
    setUpdateCheckResult,
    markUpdateNoticeRead,
  }
}
