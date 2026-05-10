import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { settingAPI } from '@/api/settings'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import {
  createLanguagePreferenceController,
  FALLBACK_LANGUAGE,
  LANGUAGE_STORAGE_KEY,
  normalizeLanguage,
  type LanguageCode,
} from '@/utils/languagePreference'

const sharedLanguageState = createLanguagePreferenceController.createState(FALLBACK_LANGUAGE)

function readLocalLanguage(): LanguageCode | null {
  return normalizeLanguage(window.localStorage.getItem(LANGUAGE_STORAGE_KEY))
}

function writeLocalLanguage(language: LanguageCode) {
  window.localStorage.setItem(LANGUAGE_STORAGE_KEY, language)
}

export function useLanguagePreference() {
  const { t, locale } = useI18n()
  const toast = useToast()
  const authStore = useAuthStore()

  const languageOptions: Array<{ value: LanguageCode; labelKey: 'chinese' | 'english' }> = [
    { value: 'zh-CN', labelKey: 'chinese' },
    { value: 'en-US', labelKey: 'english' },
  ]

  const languageSelectOptions = computed(() =>
    languageOptions.map((option) => ({
      value: option.value,
      label: t(option.labelKey),
    })),
  )

  function ensureAuthHydrated() {
    if (!authStore.initialized) authStore.hydrate()
  }

  function canUseRemote() {
    ensureAuthHydrated()
    return !!authStore.isAuthenticated
  }

  const controller = createLanguagePreferenceController({
    state: sharedLanguageState,
    getLocale: () => locale.value as string,
    setLocale: (nextLanguage) => {
      locale.value = nextLanguage
    },
    readLocalLanguage,
    writeLocalLanguage,
    canUseRemote,
    fetchRemoteLanguage: async () => {
      const settings = await settingAPI.get()
      return settings.language
    },
    updateRemoteLanguage: async (nextLanguage) => {
      await settingAPI.update({ language: nextLanguage })
    },
    onSaveSuccess: () => toast.success(t('saveSuccess')),
    onSaveError: () => toast.error(t('languageSaveFailed')),
  })

  const currentLanguageLabel = computed(() => t(controller.language.value === 'zh-CN' ? 'chinese' : 'english'))

  return {
    language: controller.language,
    languageOptions,
    languageSelectOptions,
    currentLanguageLabel,
    savingLanguage: controller.savingLanguage,
    loadLanguage: controller.loadLanguage,
    changeLanguage: controller.changeLanguage,
    syncLanguageToServer: controller.syncLanguageToServer,
  }
}

export type { LanguageCode }
