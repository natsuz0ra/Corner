<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import MdiIcon from '@/components/ui/MdiIcon.vue'
import type { LightweightToolDisplay } from '@/utils/toolDisplay'
import { buildLightweightToolGroupSummary } from '@/utils/toolDisplay'
import { getToolCallStatusLabel, getToolCallStatusTone } from '@/composables/chat/useToolCallDisplay'
import { mdiChevronDown, mdiFileSearchOutline } from '@mdi/js'

const props = defineProps<{
  items: LightweightToolDisplay[]
}>()

const { locale, t } = useI18n()
const expanded = ref(false)

const summary = computed(() =>
  buildLightweightToolGroupSummary(props.items, locale.value.startsWith('zh') ? 'zh' : 'en'),
)

const hasActive = computed(() => props.items.some((item) => item.status === 'pending' || item.status === 'reviewing' || item.status === 'executing'))
const hasFailure = computed(() => props.items.some((item) => item.status === 'error' || item.status === 'rejected'))
const groupTone = computed(() => hasActive.value ? 'running' : hasFailure.value ? 'failed' : 'completed')

function toggleExpanded() {
  expanded.value = !expanded.value
}
</script>

<template>
  <section :class="['light-tool-group', `light-tool-group--${groupTone}`]">
    <button
      type="button"
      class="light-tool-summary"
      :aria-expanded="expanded ? 'true' : 'false'"
      @click="toggleExpanded"
    >
      <MdiIcon :path="mdiFileSearchOutline" :size="14" class="light-tool-icon" />
      <span class="light-tool-title">{{ summary || t('lightweightToolGroupFallback') }}</span>
      <span class="light-tool-count">{{ items.length }}</span>
      <MdiIcon
        :path="mdiChevronDown"
        :size="14"
        class="light-tool-chevron"
        :class="{ 'light-tool-chevron--open': expanded }"
      />
    </button>

    <Transition name="light-tool-expand">
      <div v-if="expanded" class="light-tool-list">
        <div v-for="item in items" :key="item.toolCallId" class="light-tool-item">
          <span class="light-tool-item-label">{{ item.label }}</span>
          <span class="light-tool-item-target" :title="item.target">{{ item.target }}</span>
          <span :class="['light-tool-status', `light-tool-status--${getToolCallStatusTone(item.status)}`]">
            {{ getToolCallStatusLabel(item.status, (key) => t(key)) }}
          </span>
          <span v-if="item.error" class="light-tool-error">{{ item.error }}</span>
        </div>
      </div>
    </Transition>
  </section>
</template>

<style scoped>
.light-tool-group {
  width: min(100%, 760px);
  border-radius: 8px;
  border: 1px solid var(--tool-card-border, rgba(100, 116, 139, 0.15));
  background: var(--card-bg);
  overflow: hidden;
}

.light-tool-summary {
  width: 100%;
  min-height: 34px;
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr) auto 18px;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border: 0;
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  text-align: left;
}

.light-tool-icon {
  color: var(--accent-primary);
}

.light-tool-title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 650;
}

.light-tool-count {
  min-width: 22px;
  height: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 0 7px;
  background: var(--tool-summary-bg);
  color: var(--tool-summary-text);
  font-size: 12px;
  font-weight: 650;
}

.light-tool-chevron {
  color: var(--text-secondary);
  transition: transform 0.16s ease;
}

.light-tool-chevron--open {
  transform: rotate(180deg);
}

.light-tool-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 0 10px 10px 36px;
}

.light-tool-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 12px;
}

.light-tool-item-label {
  color: var(--text-primary);
  font-weight: 650;
}

.light-tool-item-target {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.light-tool-status {
  font-size: 12px;
  font-weight: 650;
}

.light-tool-status--success {
  color: var(--tool-success-text);
}

.light-tool-status--error {
  color: var(--tool-error-text);
}

.light-tool-status--pending,
.light-tool-status--executing {
  color: var(--tool-pending-text);
}

.light-tool-error {
  grid-column: 2 / 4;
  color: var(--tool-error-text);
  overflow-wrap: anywhere;
}

.light-tool-expand-enter-active,
.light-tool-expand-leave-active {
  transition: opacity 0.14s ease;
}

.light-tool-expand-enter-from,
.light-tool-expand-leave-to {
  opacity: 0;
}
</style>
