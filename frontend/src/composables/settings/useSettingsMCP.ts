import { computed, ref, type Ref } from 'vue'
import { json } from '@codemirror/lang-json'
import { oneDark } from '@codemirror/theme-one-dark'
import { lineNumbers } from '@codemirror/view'
import { mcpAPI } from '@/api/mcp'
import type { MCPConfig, MCPToolListResponse } from '@/types/settings'
import { filterMCPTools } from '@/utils/mcpTools'
import { formatMCPPreview } from '@/utils/mcpPreview'

type ToastLike = {
  error(message: string): void
}

type Translate = (key: string, params?: Record<string, unknown>) => string

type MCPTransport = 'stdio' | 'sse' | 'streamable_http'

export function useSettingsMCP(options: {
  mcpList: Ref<MCPConfig[]>
  mcpDialogVisible: Ref<boolean>
  mcpSubmitting: Ref<boolean>
  toast: ToastLike
  t: Translate
}) {
  const { mcpList, mcpDialogVisible, mcpSubmitting, toast, t } = options
  const mcpEditingID = ref('')
  const mcpForm = ref({ name: '', config: '', isEnabled: true })
  const mcpTemplateType = ref<MCPTransport>('stdio')
  const expandedMCPToolsID = ref('')
  const mcpToolLoading = ref<Record<string, boolean>>({})
  const mcpToolResponses = ref<Record<string, MCPToolListResponse | undefined>>({})
  const mcpToolQueries = ref<Record<string, string>>({})
  const mcpEditorExtensions = [lineNumbers(), json(), oneDark]
  const mcpRows = computed(() => mcpList.value || [])
  const mcpDialogTitle = computed(() => (mcpEditingID.value ? t('editMcp') : t('addMcp')))

  async function refreshMCP() {
    mcpList.value = await mcpAPI.list()
  }

  function buildTemplate(transport: MCPTransport) {
    if (transport === 'stdio') {
      return JSON.stringify({ command: 'python', args: ['-m', 'your_module'] }, null, 2)
    }
    return JSON.stringify(
      { transport, url: 'https://your-mcp-server-url', headers: {}, timeout: 5, sse_read_timeout: 300 },
      null,
      2,
    )
  }

  function applyTemplate(transport: MCPTransport) {
    mcpTemplateType.value = transport
    mcpForm.value.config = buildTemplate(transport)
  }

  function openMCPDialog() {
    mcpEditingID.value = ''
    mcpForm.value = { name: '', config: buildTemplate('stdio'), isEnabled: true }
    mcpTemplateType.value = 'stdio'
    mcpDialogVisible.value = true
  }

  function openMCPEditDialog(item: MCPConfig) {
    mcpEditingID.value = item.id
    mcpForm.value = { name: item.name, config: item.config, isEnabled: item.isEnabled }
    mcpDialogVisible.value = true
  }

  async function saveMCP() {
    if (!mcpForm.value.name || !mcpForm.value.config) {
      toast.error(t('mcpFormIncomplete'))
      return
    }
    let parsed: unknown
    try {
      parsed = JSON.parse(mcpForm.value.config)
    } catch {
      toast.error(t('mcpJsonInvalid'))
      return
    }
    if (typeof parsed === 'object' && parsed !== null && 'mcpServers' in parsed) {
      toast.error(t('mcpWrapperNotSupported'))
      return
    }

    mcpSubmitting.value = true
    try {
      const payload = {
        name: mcpForm.value.name.trim(),
        config: JSON.stringify(parsed, null, 2),
        isEnabled: mcpForm.value.isEnabled,
      }
      if (mcpEditingID.value) {
        await mcpAPI.update(mcpEditingID.value, payload)
      } else {
        await mcpAPI.create(payload)
      }
      mcpForm.value = { name: '', config: buildTemplate('stdio'), isEnabled: true }
      await refreshMCP()
      mcpDialogVisible.value = false
    } finally {
      mcpSubmitting.value = false
    }
  }

  async function updateMCP(item: MCPConfig) {
    await mcpAPI.update(item.id, { name: item.name, config: item.config, isEnabled: item.isEnabled })
    mcpToolResponses.value[item.id] = undefined
  }

  async function deleteMCP(id: string) {
    await mcpAPI.remove(id)
    delete mcpToolLoading.value[id]
    delete mcpToolResponses.value[id]
    delete mcpToolQueries.value[id]
    if (expandedMCPToolsID.value === id) {
      expandedMCPToolsID.value = ''
    }
    await refreshMCP()
  }

  function mcpPreview(item: MCPConfig) {
    return formatMCPPreview(item.config, t)
  }

  async function loadMCPTools(item: MCPConfig, force = false) {
    if (!item.isEnabled) return
    if (!force && mcpToolResponses.value[item.id]) return
    mcpToolLoading.value[item.id] = true
    try {
      mcpToolResponses.value[item.id] = await mcpAPI.tools(item.id)
    } catch (err: unknown) {
      const response = err as { response?: { data?: { error?: string } } }
      toast.error(response.response?.data?.error || 'MCP tools 加载失败')
    } finally {
      mcpToolLoading.value[item.id] = false
    }
  }

  async function toggleMCPTools(item: MCPConfig) {
    if (expandedMCPToolsID.value === item.id) {
      expandedMCPToolsID.value = ''
      return
    }
    expandedMCPToolsID.value = item.id
    await loadMCPTools(item)
  }

  function setMCPToolsQuery(id: string, query: string) {
    mcpToolQueries.value[id] = query
  }

  function visibleMCPTools(id: string) {
    return filterMCPTools(mcpToolResponses.value[id]?.tools || [], mcpToolQueries.value[id] || '')
  }

  return {
    mcpForm,
    mcpRows,
    mcpDialogTitle,
    mcpTemplateType,
    expandedMCPToolsID,
    mcpToolLoading,
    mcpToolResponses,
    mcpToolQueries,
    mcpEditorExtensions,
    applyTemplate,
    openMCPDialog,
    openMCPEditDialog,
    saveMCP,
    updateMCP,
    deleteMCP,
    mcpPreview,
    loadMCPTools,
    toggleMCPTools,
    setMCPToolsQuery,
    visibleMCPTools,
  }
}
