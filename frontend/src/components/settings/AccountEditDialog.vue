<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import BaseDialog from '@/components/ui/BaseDialog.vue'
import { useToast } from '@/composables/useToast'
import { useI18n } from 'vue-i18n'
import { authAPI } from '@/api/auth'

const props = withDefaults(defineProps<{
  visible: boolean
  forceMode?: boolean
}>(), {
  forceMode: false,
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  success: []
}>()

const { t } = useI18n()
const toast = useToast()

const submitting = ref(false)
const username = ref('')
const oldPassword = ref('')
const newPassword = ref('')

const title = computed(() => (props.forceMode ? t('forcePasswordChangeTitle') : t('accountEditTitle')))

function resetForm() {
  username.value = ''
  oldPassword.value = ''
  newPassword.value = ''
}

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      resetForm()
    }
  },
)

async function onConfirm() {
  const nextUsername = username.value.trim()
  const nextNewPassword = newPassword.value.trim()
  const nextOldPassword = oldPassword.value

  if (props.forceMode && nextNewPassword === '') {
    toast.error(t('newPasswordRequired'))
    return
  }
  if (nextUsername === '' && nextNewPassword === '') {
    toast.error(t('accountEditNeedOneField'))
    return
  }
  if (nextNewPassword !== '' && nextOldPassword.trim() === '') {
    toast.error(t('oldPasswordRequired'))
    return
  }

  submitting.value = true
  try {
    await authAPI.updateAccount({
      username: nextUsername || undefined,
      oldPassword: nextOldPassword || undefined,
      newPassword: nextNewPassword || undefined,
    })
    toast.success(t('saveSuccess'))
    emit('success')
    emit('update:visible', false)
  } catch (error: any) {
    toast.error(error?.response?.data?.error || t('accountEditFailed'))
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  emit('update:visible', false)
}
</script>

<template>
  <BaseDialog
    :visible="visible"
    :title="title"
    :confirm-text="t('confirm')"
    :cancel-text="t('cancel')"
    :confirm-loading="submitting"
    :show-close="!forceMode"
    :show-cancel="!forceMode"
    :close-on-mask="!forceMode"
    :close-on-esc="!forceMode"
    width="420px"
    @update:visible="emit('update:visible', $event)"
    @confirm="onConfirm"
    @cancel="onCancel"
  >
    <div class="flex flex-col gap-4">
      <div v-if="forceMode" class="account-force-tip text-xs">
        {{ t('forcePasswordChangeTip') }}
      </div>
      <div class="flex flex-col gap-1.5">
        <label class="text-xs font-medium account-field-label">{{ t('usernameOptional') }}</label>
        <input
          v-model="username"
          type="text"
          class="account-input px-3 py-2.5 text-sm rounded-xl outline-none transition-all duration-150"
          :placeholder="t('usernameOptionalPlaceholder')"
        />
      </div>

      <div class="flex flex-col gap-1.5">
        <label class="text-xs font-medium account-field-label">{{ t('oldPassword') }}</label>
        <input
          v-model="oldPassword"
          type="password"
          class="account-input px-3 py-2.5 text-sm rounded-xl outline-none transition-all duration-150"
          :placeholder="t('oldPasswordPlaceholder')"
        />
      </div>

      <div class="flex flex-col gap-1.5">
        <label class="text-xs font-medium account-field-label">{{ t('newPassword') }}</label>
        <input
          v-model="newPassword"
          type="password"
          class="account-input px-3 py-2.5 text-sm rounded-xl outline-none transition-all duration-150"
          :placeholder="t('newPasswordPlaceholder')"
        />
      </div>
    </div>
  </BaseDialog>
</template>

<style scoped>
.account-force-tip {
  color: #6366f1;
  background: rgba(99, 102, 241, 0.08);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 12px;
  padding: 10px 12px;
}

.account-field-label {
  color: var(--text-muted);
}

.account-input {
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  color: var(--text-primary);
  width: 100%;
}

.account-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}

.account-input::placeholder {
  color: var(--text-muted);
}
</style>
