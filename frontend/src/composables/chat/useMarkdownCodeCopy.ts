import { useI18n } from 'vue-i18n'
import { useToast } from '@/composables/useToast'

export function useMarkdownCodeCopy() {
  const { t } = useI18n()
  const toast = useToast()

  async function handleMarkdownClick(event: MouseEvent) {
    const target = event.target instanceof Element ? event.target : null
    const button = target?.closest<HTMLButtonElement>('.markdown-code-copy-btn')
    if (!button) return
    const codeBlock = button.closest('.markdown-code-block')
    const code = (codeBlock?.querySelector('pre code')?.textContent ?? '').replace(/\n+$/, '')
    if (!code) return
    try {
      await navigator.clipboard.writeText(code)
      toast.success(t('messageCopySuccess'))
    } catch {
      toast.error(t('messageCopyFailed'))
    }
  }

  return { handleMarkdownClick }
}
