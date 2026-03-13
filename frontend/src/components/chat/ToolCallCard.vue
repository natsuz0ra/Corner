<script setup lang="ts">
import { computed } from 'vue'
import { mdiCheck, mdiClose, mdiConsoleLine, mdiWeb } from '@mdi/js'
import { useI18n } from 'vue-i18n'
import MdiIcon from '@/components/MdiIcon.vue'
import type { ToolCallItem } from '@/api/chat'

const props = defineProps<{
  item: ToolCallItem & { preamble?: string }
  showPreamble?: boolean
}>()

const emit = defineEmits<{
  approve: [toolCallId: string]
  reject: [toolCallId: string]
}>()

const { t } = useI18n()

const toolIcon = computed(() => {
  if (props.item.toolName === 'exec') return mdiConsoleLine
  if (props.item.toolName === 'http_request') return mdiWeb
  if (props.item.toolName === 'web_search') return mdiWeb
  return mdiConsoleLine
})

const toolLabel = computed(() => {
  if (props.item.toolName === 'exec') return t('toolExec')
  if (props.item.toolName === 'http_request') return t('toolHttpRequest')
  if (props.item.toolName === 'web_search') return t('toolWebSearch')
  return props.item.toolName
})

const statusLabel = computed(() => {
  switch (props.item.status) {
    case 'pending': return t('toolCallPending')
    case 'executing': return t('toolCallExecuting')
    case 'completed': return t('toolCallCompleted')
    case 'rejected': return t('toolCallRejected')
    case 'error': return t('toolCallError')
    default: return ''
  }
})

const statusColorClass = computed(() => {
  switch (props.item.status) {
    case 'pending': return 'text-amber-600'
    case 'executing': return 'text-blue-600'
    case 'completed': return 'text-green-600'
    case 'rejected': return 'text-red-500'
    case 'error': return 'text-red-500'
    default: return 'text-gray-500'
  }
})

const cardBorderClass = computed(() => {
  switch (props.item.status) {
    case 'pending': return 'border-amber-200 bg-amber-50/60'
    case 'executing': return 'border-blue-200 bg-blue-50/60'
    case 'completed': return 'border-green-200 bg-green-50/60'
    case 'rejected':
    case 'error': return 'border-red-200 bg-red-50/60'
    default: return 'border-gray-200 bg-gray-50'
  }
})

const paramsDisplay = computed(() => {
  const entries = Object.entries(props.item.params)
  if (entries.length === 0) return ''
  return entries.map(([k, v]) => `${k}: ${v}`).join('\n')
})

const showActions = computed(() => props.item.status === 'pending')
const showResult = computed(() => props.item.status === 'completed' || props.item.status === 'error')
const shouldShowPreamble = computed(() => !!props.showPreamble && !!props.item.preamble)
</script>

<template>
  <div
    class="w-full rounded-xl border px-4 py-3 text-sm box-border overflow-hidden"
    :class="cardBorderClass"
  >
    <!-- 头部 -->
    <div class="flex items-center flex-wrap gap-2 font-medium text-gray-700">
      <MdiIcon :path="toolIcon" :size="15" class="flex-shrink-0 text-gray-500" />
      <span class="text-blue-600 font-medium">{{ toolLabel }}</span>
      <code v-if="item.command" class="text-xs text-gray-500 font-mono break-all">{{ item.command }}</code>
      <span class="ml-auto text-xs font-normal" :class="statusColorClass">
        {{ statusLabel }}
      </span>
      <svg
        v-if="item.status === 'executing'"
        class="animate-spin-icon w-3.5 h-3.5 flex-shrink-0"
        :class="statusColorClass"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
      </svg>
    </div>

    <!-- 参数 -->
    <div v-if="paramsDisplay" class="mt-2">
      <pre class="text-xs font-mono bg-black/[0.05] rounded-lg px-3 py-2 text-gray-600 whitespace-pre-wrap break-all">{{ paramsDisplay }}</pre>
    </div>

    <!-- 前言 -->
    <div v-if="shouldShowPreamble" class="mt-2 text-xs text-gray-600 leading-relaxed whitespace-pre-wrap break-words">
      {{ item.preamble }}
    </div>

    <!-- 操作按钮（待审批） -->
    <div v-if="showActions" class="flex items-center gap-2 mt-3">
      <button
        type="button"
        class="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg border border-green-300 bg-white text-green-700 hover:bg-green-50 transition-colors duration-150 cursor-pointer"
        @click="emit('approve', item.toolCallId)"
      >
        <MdiIcon :path="mdiCheck" :size="12" />
        {{ t('toolCallApprove') }}
      </button>
      <button
        type="button"
        class="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg border border-red-200 bg-white text-red-600 hover:bg-red-50 transition-colors duration-150 cursor-pointer"
        @click="emit('reject', item.toolCallId)"
      >
        <MdiIcon :path="mdiClose" :size="12" />
        {{ t('toolCallReject') }}
      </button>
    </div>

    <!-- 执行结果 -->
    <div v-if="showResult" class="mt-2">
      <div v-if="item.error" class="text-xs text-red-500 mb-1">{{ item.error }}</div>
      <details v-if="item.output" class="text-xs">
        <summary class="cursor-pointer text-gray-400 hover:text-gray-600 select-none transition-colors duration-150">
          {{ t('toolCallOutput') }}
        </summary>
        <pre class="mt-1.5 px-3 py-2 bg-gray-900 text-gray-100 rounded-lg text-xs font-mono whitespace-pre-wrap break-all leading-relaxed">{{ item.output }}</pre>
      </details>
    </div>
  </div>
</template>
