<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { mdiCheckCircleOutline, mdiDownload, mdiGithub, mdiOpenInNew, mdiRefresh, mdiAlertCircleOutline } from '@mdi/js'

import MdiIcon from '@/components/ui/MdiIcon.vue'
import { updateAPI } from '@/api/update'
import type { UpdateCheckResult, UpdateJobStatus } from '@/types/update'
import { isUpdateJobActive, normalizeUpdateJob, updatePhaseTone } from '@/utils/updateStatus'

const { t } = useI18n()

const githubUrl = 'https://github.com/natsuz0ra/SlimeBot'
const version = `v${__APP_VERSION__}`
const checkResult = ref<UpdateCheckResult | null>(null)
const job = ref<UpdateJobStatus>(normalizeUpdateJob(null))
const checking = ref(false)
const applying = ref(false)
const errorMessage = ref('')
let pollTimer: ReturnType<typeof setInterval> | null = null

const hasUpdate = computed(() => Boolean(checkResult.value?.updateAvailable))
const canApplyUpdate = computed(() => Boolean(checkResult.value?.canApply) && !checking.value && !applying.value && !isUpdateJobActive(job.value.phase))
const releaseNotes = computed(() => {
  const notes = checkResult.value?.releaseNotes?.trim() || ''
  if (!notes) return ''
  return notes.split(/\r?\n/).filter(Boolean).slice(0, 4).join('\n')
})
const statusTone = computed(() => updatePhaseTone(job.value.phase))
const statusIcon = computed(() => {
  if (statusTone.value === 'success') return mdiCheckCircleOutline
  if (statusTone.value === 'danger') return mdiAlertCircleOutline
  return mdiRefresh
})
const statusText = computed(() => {
  if (errorMessage.value) return errorMessage.value
  if (job.value.message) return job.value.message
  if (!checkResult.value) return t('updateNotChecked')
  if (!hasUpdate.value) return t('updateAlreadyLatest')
  if (!checkResult.value.canApply) return checkResult.value.reason || t('updateManualOnly')
  return t('updateAvailable')
})
const manualHint = computed(() => job.value.manualHint || checkResult.value?.manualHint || '')
const versionLine = computed(() => {
  if (!checkResult.value?.latest) return version
  return `${checkResult.value.current || version} → ${checkResult.value.latest}`
})

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function loadJob() {
  job.value = await updateAPI.job()
  applying.value = isUpdateJobActive(job.value.phase)
  if (!applying.value) {
    stopPolling()
  }
}

function startPolling() {
  stopPolling()
  pollTimer = setInterval(() => {
    void loadJob().catch(() => {
      applying.value = false
      stopPolling()
    })
  }, 1500)
}

async function checkUpdate(force = true) {
  checking.value = true
  errorMessage.value = ''
  try {
    checkResult.value = await updateAPI.check(force)
    await loadJob()
  } catch (err: unknown) {
    const response = err as { response?: { data?: { error?: string } } }
    errorMessage.value = response.response?.data?.error || t('updateCheckFailed')
  } finally {
    checking.value = false
  }
}

async function applyUpdate() {
  if (!checkResult.value?.latest || !canApplyUpdate.value) return
  applying.value = true
  errorMessage.value = ''
  try {
    job.value = await updateAPI.apply(checkResult.value.latest)
    startPolling()
  } catch (err: unknown) {
    const response = err as { response?: { data?: { error?: string } } }
    errorMessage.value = response.response?.data?.error || t('updateApplyFailed')
    applying.value = false
  }
}

onMounted(() => {
  void checkUpdate(false)
})

onUnmounted(stopPolling)
</script>

<template>
  <div>
    <p class="section-label">{{ t('aboutSettings') }}</p>

    <div class="settings-card about-card rounded-xl px-5 py-6">
      <div class="about-hero">
        <div class="about-logo-wrap">
          <img src="/slime-icon.svg" alt="SlimeBot" class="about-logo" />
        </div>
        <h2 class="about-title">SlimeBot</h2>
      </div>

      <div class="about-info">
        <div class="about-row">
          <span class="text-sm settings-field-label">{{ t('appVersion') }}</span>
          <span class="text-sm settings-item-sub">{{ version }}</span>
        </div>

        <div class="about-row">
          <span class="text-sm settings-field-label">{{ t('githubRepository') }}</span>
          <a
            class="about-link"
            :href="githubUrl"
            target="_blank"
            rel="noreferrer"
          >
            <MdiIcon :path="mdiGithub" :size="16" />
            <span>natsuz0ra/SlimeBot</span>
          </a>
        </div>
      </div>
    </div>

    <div class="settings-card update-card rounded-xl px-5 py-5 mt-4">
      <div class="update-header">
        <div>
          <div class="section-label mb-1">{{ t('updateCenterTitle') }}</div>
          <div class="update-version-line">{{ versionLine }}</div>
        </div>
        <span class="update-status-pill" :data-tone="statusTone">
          <MdiIcon :path="statusIcon" :size="15" />
          <span>{{ statusText }}</span>
        </span>
      </div>

      <div v-if="checkResult" class="update-meta">
        <div class="about-row">
          <span class="text-sm settings-field-label">{{ t('updateLatestVersion') }}</span>
          <span class="text-sm settings-item-sub">{{ checkResult.latest || '-' }}</span>
        </div>
        <div v-if="checkResult.releaseUrl" class="about-row">
          <span class="text-sm settings-field-label">{{ t('updateReleasePage') }}</span>
          <a class="about-link" :href="checkResult.releaseUrl" target="_blank" rel="noreferrer">
            <MdiIcon :path="mdiOpenInNew" :size="15" />
            <span>{{ t('open') }}</span>
          </a>
        </div>
      </div>

      <pre v-if="releaseNotes" class="update-notes">{{ releaseNotes }}</pre>

      <div v-if="manualHint" class="update-manual">
        <span>{{ t('updateManualHint') }}</span>
        <code>{{ manualHint }}</code>
      </div>

      <div class="update-actions">
        <button type="button" class="update-btn update-btn-secondary" :disabled="checking || applying" @click="checkUpdate(true)">
          <MdiIcon :path="mdiRefresh" :size="15" />
          <span>{{ checking ? t('updateChecking') : t('checkUpdate') }}</span>
        </button>
        <button v-if="hasUpdate" type="button" class="update-btn update-btn-primary" :disabled="!canApplyUpdate" @click="applyUpdate">
          <MdiIcon :path="mdiDownload" :size="15" />
          <span>{{ applying ? t('updateApplying') : t('applyUpdate') }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.about-card {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.about-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 8px 0 2px;
}

.about-logo-wrap {
  width: 88px;
  height: 88px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 22px;
  background: var(--primary-alpha-08);
  border: 1px solid var(--primary-alpha-15);
}

.about-logo {
  width: 68px;
  height: 68px;
  display: block;
}

.about-title {
  color: var(--text-primary);
  font-size: 22px;
  line-height: 1.25;
  font-weight: 700;
  margin: 0;
}

.about-info {
  display: flex;
  flex-direction: column;
  border-top: 1px solid var(--card-border);
}

.about-row {
  min-height: 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid var(--card-border);
}

.about-row:last-child {
  border-bottom: none;
}

.about-link {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--sb-brand);
  font-size: 14px;
  line-height: 1.3;
  text-decoration: none;
  transition: color 0.15s;
}

.about-link:hover {
  color: var(--sb-brand-hover);
}

.update-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.update-header,
.update-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.update-version-line {
  color: var(--text-primary);
  font-size: 14px;
  line-height: 1.35;
  font-weight: 600;
}

.update-status-pill {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 55%;
  padding: 6px 9px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 1.2;
  color: var(--text-secondary);
  background: var(--hover-bg);
  border: 1px solid var(--card-border);
}

.update-status-pill span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.update-status-pill[data-tone="info"] {
  color: var(--sb-brand);
  background: var(--primary-alpha-08);
  border-color: var(--primary-alpha-15);
}

.update-status-pill[data-tone="success"] {
  color: #16a34a;
  background: rgba(34, 197, 94, 0.1);
  border-color: rgba(34, 197, 94, 0.18);
}

.update-status-pill[data-tone="danger"] {
  color: #dc2626;
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.18);
}

.update-meta {
  display: flex;
  flex-direction: column;
  border-top: 1px solid var(--card-border);
}

.update-notes {
  max-height: 132px;
  overflow: auto;
  margin: 0;
  padding: 12px;
  border-radius: 10px;
  color: var(--text-secondary);
  background: var(--hover-bg);
  border: 1px solid var(--card-border);
  font-size: 12px;
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-word;
}

.update-manual {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: var(--text-secondary);
  font-size: 12px;
}

.update-manual code {
  display: block;
  overflow-x: auto;
  padding: 9px 10px;
  border-radius: 9px;
  color: var(--text-primary);
  background: var(--hover-bg);
  border: 1px solid var(--card-border);
  font-family: var(--font-mono);
}

.update-btn {
  min-height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 0 12px;
  border-radius: 10px;
  font-size: 13px;
  line-height: 1.2;
  border: 1px solid transparent;
  transition: all 0.15s;
  cursor: pointer;
}

.update-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.update-btn-primary {
  color: white;
  background: var(--sb-brand);
}

.update-btn-secondary {
  color: var(--text-primary);
  background: var(--hover-bg);
  border-color: var(--card-border);
}

@media (max-width: 640px) {
  .about-card {
    padding: 20px 16px;
  }

  .about-row {
    align-items: flex-start;
    flex-direction: column;
    justify-content: center;
    gap: 6px;
    padding: 10px 0;
  }

  .update-header,
  .update-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .update-status-pill {
    max-width: 100%;
    justify-content: center;
  }

  .update-btn {
    width: 100%;
  }
}
</style>
