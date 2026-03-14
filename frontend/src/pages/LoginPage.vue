<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import SlimeBotLogo from '@/components/ui/SlimeBotLogo.vue'
import AccountEditDialog from '@/components/settings/AccountEditDialog.vue'
import { useI18n } from 'vue-i18n'
import { useToast } from '@/composables/useToast'
import { authAPI } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const toast = useToast()
const router = useRouter()
const authStore = useAuthStore()

const username = ref('admin')
const password = ref('admin')
const submitting = ref(false)
const accountDialogVisible = ref(false)

const canSubmit = computed(() => !!username.value.trim() && !!password.value)

async function login() {
  if (!canSubmit.value || submitting.value) return
  submitting.value = true
  try {
    const response = await authAPI.login({
      username: username.value.trim(),
      password: password.value,
    })
    authStore.setAuth(response.token, !!response.mustChangePassword)
    if (response.mustChangePassword) {
      accountDialogVisible.value = true
      return
    }
    await router.replace('/')
  } catch (error: any) {
    toast.error(error?.response?.data?.error || t('loginFailed'))
  } finally {
    submitting.value = false
  }
}

async function onAccountUpdated() {
  authStore.markPasswordChanged()
  accountDialogVisible.value = false
  await router.replace('/')
}

onMounted(async () => {
  if (!authStore.initialized) {
    authStore.hydrate()
  }
  if (authStore.isAuthenticated && authStore.mustChangePassword) {
    accountDialogVisible.value = true
    return
  }
  if (authStore.isAuthenticated && !authStore.mustChangePassword) {
    await router.replace('/')
  }
})
</script>

<template>
  <div class="login-page min-h-screen px-4 py-10 sm:py-14">
    <div class="login-card w-full max-w-[420px] mx-auto rounded-2xl p-5 sm:p-7">
      <div class="logo-row flex items-end gap-3.5">
        <SlimeBotLogo :size="44" />
        <span class="brand-tech-font logo-text">SlimeBot</span>
      </div>

      <div class="mt-7 flex flex-col gap-3.5">
        <input
          v-model="username"
          type="text"
          class="login-input w-full rounded-xl px-3.5 py-2.5 text-sm outline-none"
          :placeholder="t('username')"
          autocomplete="username"
        />
        <input
          v-model="password"
          type="password"
          class="login-input w-full rounded-xl px-3.5 py-2.5 text-sm outline-none"
          :placeholder="t('password')"
          autocomplete="current-password"
          @keydown.enter="login"
        />
        <button
          type="button"
          class="login-submit w-full rounded-xl py-2.5 text-sm font-semibold cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="!canSubmit || submitting"
          @click="login"
        >
          <span v-if="!submitting">{{ t('login') }}</span>
          <span v-else>{{ t('loading') }}...</span>
        </button>
        <p class="login-tip text-xs">
          {{ t('loginDefaultTip') }}
        </p>
      </div>
    </div>

    <AccountEditDialog
      v-model:visible="accountDialogVisible"
      :force-mode="true"
      @success="onAccountUpdated"
    />
  </div>
</template>

<style scoped>
.login-page {
  background:
    radial-gradient(1200px 500px at 10% -10%, rgba(99, 102, 241, 0.14), transparent 55%),
    radial-gradient(900px 460px at 100% 0%, rgba(167, 139, 250, 0.13), transparent 60%),
    var(--bg-base);
  display: flex;
  align-items: center;
  justify-content: center;
}

.login-card {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 20px 56px rgba(15, 23, 42, 0.14), 0 2px 8px rgba(99, 102, 241, 0.08);
  backdrop-filter: blur(6px);
}

.logo-text {
  color: var(--text-primary);
  font-size: 28px;
  line-height: 1;
  font-weight: 600;
  letter-spacing: 0.01em;
}

.login-input {
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  color: var(--text-primary);
}

.login-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}

.login-input::placeholder {
  color: var(--text-muted);
}

.login-submit {
  color: #fff;
  background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.35);
}

.login-submit:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.44);
  transform: translateY(-1px);
}

.login-tip {
  color: var(--text-muted);
}
</style>
