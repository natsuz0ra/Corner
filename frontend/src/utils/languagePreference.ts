import { ref, type Ref } from 'vue'

export type LanguageCode = 'zh-CN' | 'en-US'

export const LANGUAGE_STORAGE_KEY = 'slimebot.language'
export const FALLBACK_LANGUAGE: LanguageCode = 'zh-CN'

const LANGUAGE_VALUES: LanguageCode[] = ['zh-CN', 'en-US']

export interface LanguagePreferenceState {
  language: Ref<LanguageCode>
  savingLanguage: Ref<boolean>
}

interface LoadLanguageOptions {
  allowRemote?: boolean
}

interface ChangeLanguageOptions {
  allowRemote?: boolean
  showSuccessToast?: boolean
}

interface SyncLanguageOptions {
  showSuccessToast?: boolean
  silentOnError?: boolean
}

interface LanguagePreferenceControllerDependencies {
  state?: LanguagePreferenceState
  getLocale: () => string | null | undefined
  setLocale: (language: LanguageCode) => void
  readLocalLanguage: () => string | null | undefined
  writeLocalLanguage: (language: LanguageCode) => void
  canUseRemote: (allowRemote: boolean) => boolean
  fetchRemoteLanguage: () => Promise<string | null | undefined>
  updateRemoteLanguage: (language: LanguageCode) => Promise<void>
  onSaveSuccess?: () => void
  onSaveError?: () => void
}

export function normalizeLanguage(value: string | null | undefined): LanguageCode | null {
  return LANGUAGE_VALUES.includes(value as LanguageCode) ? (value as LanguageCode) : null
}

function createLanguagePreferenceState(initialLanguage: LanguageCode): LanguagePreferenceState {
  return {
    language: ref(initialLanguage),
    savingLanguage: ref(false),
  }
}

function buildLanguagePreferenceController(deps: LanguagePreferenceControllerDependencies) {
  const state = deps.state ?? createLanguagePreferenceState(
    normalizeLanguage(deps.getLocale()) || FALLBACK_LANGUAGE,
  )

  function applyLanguage(nextLanguage: LanguageCode) {
    state.language.value = nextLanguage
    deps.setLocale(nextLanguage)
    deps.writeLocalLanguage(nextLanguage)
  }

  function canUseRemote(allowRemote: boolean) {
    return allowRemote && deps.canUseRemote(allowRemote)
  }

  async function loadLanguage(options?: LoadLanguageOptions) {
    const allowRemote = options?.allowRemote ?? true
    const localLanguage = normalizeLanguage(deps.readLocalLanguage())
    if (localLanguage) {
      applyLanguage(localLanguage)
    } else {
      applyLanguage(normalizeLanguage(deps.getLocale()) || FALLBACK_LANGUAGE)
    }

    if (!canUseRemote(allowRemote)) return state.language.value

    try {
      const remoteLanguage = normalizeLanguage(await deps.fetchRemoteLanguage()) || FALLBACK_LANGUAGE
      applyLanguage(remoteLanguage)
      return remoteLanguage
    } catch {
      return state.language.value
    }
  }

  async function changeLanguage(nextLanguage: LanguageCode, options?: ChangeLanguageOptions) {
    if (state.savingLanguage.value) return false
    if (nextLanguage === state.language.value) return true

    const previousLanguage = state.language.value
    const allowRemote = options?.allowRemote ?? true
    const showSuccessToast = options?.showSuccessToast ?? false

    applyLanguage(nextLanguage)

    if (!canUseRemote(allowRemote)) return true

    state.savingLanguage.value = true
    try {
      await deps.updateRemoteLanguage(nextLanguage)
      if (showSuccessToast) deps.onSaveSuccess?.()
      return true
    } catch {
      applyLanguage(previousLanguage)
      deps.onSaveError?.()
      return false
    } finally {
      state.savingLanguage.value = false
    }
  }

  async function syncLanguageToServer(options?: SyncLanguageOptions) {
    if (state.savingLanguage.value) return false
    const showSuccessToast = options?.showSuccessToast ?? false
    const silentOnError = options?.silentOnError ?? true
    if (!canUseRemote(true)) return false

    state.savingLanguage.value = true
    try {
      await deps.updateRemoteLanguage(state.language.value)
      if (showSuccessToast) deps.onSaveSuccess?.()
      return true
    } catch {
      if (!silentOnError) deps.onSaveError?.()
      return false
    } finally {
      state.savingLanguage.value = false
    }
  }

  return {
    language: state.language,
    savingLanguage: state.savingLanguage,
    loadLanguage,
    changeLanguage,
    syncLanguageToServer,
  }
}

export const createLanguagePreferenceController = Object.assign(buildLanguagePreferenceController, {
  createState: createLanguagePreferenceState,
})
