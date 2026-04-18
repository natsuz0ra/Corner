<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  visible: boolean
  toolName: string
  command: string
  params: Record<string, string>
}>()

const emit = defineEmits<{
  approve: []
  reject: []
}>()

const { t } = useI18n()
const approveBtnRef = ref<HTMLButtonElement | null>(null)

const paramsEntries = ref<{ key: string; value: string }[]>([])

watch(
  () => props.params,
  (newParams) => {
    paramsEntries.value = Object.entries(newParams).map(([key, value]) => ({ key, value }))
  },
  { immediate: true },
)

watch(
  () => props.visible,
  async (isOpen) => {
    if (isOpen) {
      await nextTick()
      approveBtnRef.value?.focus()
    }
  },
)

function onKeydown(e: KeyboardEvent) {
  if (!props.visible) return
  if (e.key === 'Enter') {
    e.preventDefault()
    emit('approve')
  } else if (e.key === 'Escape') {
    e.preventDefault()
    emit('reject')
  }
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer-slide">
      <div v-if="visible" class="drawer-overlay" @click.self="emit('reject')">
        <div class="drawer-panel" role="dialog" aria-modal="true" :aria-label="t('approvalDrawerTitle')">
          <header class="drawer-header">
            <h3 class="drawer-title">{{ t('approvalDrawerTitle') }}</h3>
          </header>

          <section class="drawer-body">
            <div class="drawer-tool-info">
              <span class="drawer-tool-gear" aria-hidden="true">&#9881;</span>
              <code class="drawer-tool-command">{{ toolName }}.{{ command }}()</code>
            </div>

            <div v-if="paramsEntries.length > 0" class="drawer-params">
              <p class="drawer-params-title">{{ t('toolCallParams') }}</p>
              <div class="drawer-params-list">
                <div v-for="row in paramsEntries" :key="row.key" class="drawer-params-row">
                  <span class="drawer-params-key">{{ row.key }}</span>
                  <pre class="drawer-params-value">{{ row.value }}</pre>
                </div>
              </div>
            </div>
          </section>

          <footer class="drawer-footer">
            <button
              type="button"
              class="drawer-btn drawer-btn--reject"
              @click="emit('reject')"
            >
              {{ t('toolCallReject') }}
            </button>
            <button
              ref="approveBtnRef"
              type="button"
              class="drawer-btn drawer-btn--approve"
              @click="emit('approve')"
            >
              {{ t('approvalDrawerApprove') }}
            </button>
          </footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.drawer-overlay {
  position: fixed;
  inset: 0;
  z-index: 300;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: rgba(0, 0, 0, 0.18);
  backdrop-filter: blur(2px);
}

.drawer-panel {
  width: 100%;
  max-width: 520px;
  max-height: 70vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 20px 20px 0 0;
  background: var(--bg-main);
  border: 1px solid var(--tool-card-border);
  border-bottom: none;
  box-shadow:
    0 -8px 40px rgba(0, 0, 0, 0.2),
    0 0 0 1px var(--primary-alpha-08),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(20px) saturate(1.4);
}

:root:not(.dark) .drawer-panel {
  background: rgba(255, 255, 255, 0.82);
}

.dark .drawer-panel {
  background: rgba(24, 24, 48, 0.88);
  box-shadow:
    0 -8px 40px rgba(0, 0, 0, 0.5),
    0 0 0 1px rgba(255, 255, 255, 0.06),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.drawer-header {
  padding: 16px 20px 12px;
  border-bottom: 1px solid var(--tool-section-border);
}

.drawer-title {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: 0.01em;
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.drawer-tool-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.drawer-tool-gear {
  font-size: 18px;
  color: var(--tool-summary-text);
  flex-shrink: 0;
}

.drawer-tool-command {
  display: inline-block;
  color: var(--tool-command-text);
  background: var(--tool-command-bg);
  border: 1px solid var(--tool-command-border);
  border-radius: 8px;
  padding: 4px 10px;
  font-size: 14px;
  font-family: var(--font-mono);
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.drawer-params {
  border: 1px solid var(--tool-section-border);
  border-radius: 10px;
  padding: 10px;
  background: var(--tool-section-bg);
}

.drawer-params-title {
  margin: 0 0 8px 0;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--tool-summary-text);
  text-transform: uppercase;
}

.drawer-params-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.drawer-params-row {
  border: 1px solid var(--tool-section-border);
  border-radius: 8px;
  padding: 6px 10px;
  background: rgba(0, 0, 0, 0.6);
}

.dark .drawer-params-row {
  background: rgba(0, 0, 0, 0.5);
}

.drawer-params-key {
  display: block;
  color: var(--tool-summary-text);
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.02em;
  margin-bottom: 3px;
}

.drawer-params-value {
  margin: 0;
  color: var(--tool-detail-body-text);
  font-size: 12px;
  line-height: 1.45;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.drawer-footer {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 20px 18px;
  border-top: 1px solid var(--tool-section-border);
}

.drawer-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-height: 36px;
  padding: 8px 18px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 180ms ease, color 180ms ease, box-shadow 180ms ease, border-color 180ms ease;
}

.drawer-btn:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 2px;
}

.drawer-btn--reject {
  background: var(--tool-error-bg);
  border: 1px solid var(--tool-error-border);
  color: var(--tool-error-text);
}

.drawer-btn--reject:hover {
  background: var(--tool-error-bg-hover);
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.18);
}

.drawer-btn--approve {
  flex: 1;
  background: var(--tool-success-bg);
  border: 1px solid var(--tool-success-border);
  color: var(--tool-success-text);
}

.drawer-btn--approve:hover {
  background: var(--tool-success-bg-hover);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.22);
}

/* Slide-up / slide-down transition */
.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition: opacity 300ms ease-out;
}

.drawer-slide-enter-active .drawer-panel,
.drawer-slide-leave-active .drawer-panel {
  transition: transform 300ms cubic-bezier(0.16, 1, 0.3, 1);
}

.drawer-slide-enter-from,
.drawer-slide-leave-to {
  opacity: 0;
}

.drawer-slide-enter-from .drawer-panel,
.drawer-slide-leave-to .drawer-panel {
  transform: translateY(100%);
}

@media (max-width: 640px) {
  .drawer-panel {
    max-width: 100%;
    border-radius: 16px 16px 0 0;
  }

  .drawer-footer {
    padding: 12px 16px 16px;
  }

  .drawer-body {
    padding: 12px 16px;
  }

  .drawer-header {
    padding: 14px 16px 10px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .drawer-slide-enter-active,
  .drawer-slide-leave-active,
  .drawer-slide-enter-active .drawer-panel,
  .drawer-slide-leave-active .drawer-panel {
    transition: none;
  }
}
</style>
