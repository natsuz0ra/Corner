<script setup lang="ts">
import CodeMirror from 'vue-codemirror6'
import { useI18n } from 'vue-i18n'

defineProps<{
  content: string
  path: string
  saving: boolean
}>()

const emit = defineEmits<{
  'update:content': [value: string]
  save: []
}>()

const { t } = useI18n()

function onContentChange(value: unknown) {
  emit('update:content', typeof value === 'string' ? value : String(value ?? ''))
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between gap-3 mb-3">
      <div class="min-w-0">
        <p class="section-label mb-1">{{ t('agentsSettings') }}</p>
        <p class="text-xs settings-item-meta truncate">{{ path }}</p>
      </div>
      <button
        type="button"
        class="px-3 py-1.5 text-xs rounded-lg cursor-pointer account-edit-btn"
        :disabled="saving"
        @click="emit('save')"
      >
        {{ saving ? t('saving') : t('save') }}
      </button>
    </div>
    <div class="settings-card rounded-xl overflow-hidden">
      <CodeMirror
        :model-value="content"
        class="json-codemirror agents-codemirror"
        :indent-with-tab="true"
        :tab-size="2"
        :style="{ height: '520px' }"
        @update:model-value="onContentChange"
      />
    </div>
  </div>
</template>
