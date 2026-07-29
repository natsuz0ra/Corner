<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppDialog from '@/components/ui/AppDialog.vue'
import ToolCallInline from '@/components/chat/ToolCallInline.vue'
import type { ToolCallItem } from '@/api/chat'
import type { AgentTeamMember, AgentTeamRun } from '@/utils/agentTeam'
import {
  buildAgentTeamResultPreview,
  countAgentTeamMemberTools,
  findAgentTeamPendingTool,
  formatAgentTeamDuration,
  getAgentTeamDurationMs,
} from '@/utils/agentTeam'

const props = defineProps<{
  visible: boolean
  team: AgentTeamRun
  toolCalls: ToolCallItem[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  approve: [toolCallId: string]
  reject: [toolCallId: string]
}>()

const { t } = useI18n()
const selectedMemberId = ref('')
const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | undefined

const terminalCount = computed(() => props.team.members.filter((member) => !['queued', 'running'].includes(member.status)).length)
const teamDuration = computed(() => formatAgentTeamDuration(getAgentTeamDurationMs(props.team, now.value)))
const selectedMember = computed(() => props.team.members.find((member) => member.id === selectedMemberId.value))
const pendingTool = computed(() => selectedMember.value
  ? findAgentTeamPendingTool(selectedMember.value, props.toolCalls)
  : undefined)
const resultPreview = computed(() => buildAgentTeamResultPreview(
  selectedMember.value?.error || selectedMember.value?.answer,
))
const selectedToolCount = computed(() => selectedMember.value
  ? countAgentTeamMemberTools(selectedMember.value.toolCallId, props.toolCalls)
  : 0)

function memberStatusSymbol(member: AgentTeamMember) {
  return ({
    queued: '\u25CB',
    running: '\u25CF',
    succeeded: '\u2713',
    failed: '\u2717',
    canceled: '\u2212',
    interrupted: '\u2016',
  }[member.status] || '\u25CB')
}

function memberDuration(member: AgentTeamMember) {
  return formatAgentTeamDuration(getAgentTeamDurationMs(member, now.value))
}

function chooseInitialMember(preferPending: boolean) {
  const pendingMember = props.team.members.find((member) => findAgentTeamPendingTool(member, props.toolCalls))
  const currentExists = props.team.members.some((member) => member.id === selectedMemberId.value)
  selectedMemberId.value = (preferPending ? pendingMember?.id : undefined)
    || (currentExists ? selectedMemberId.value : '')
    || pendingMember?.id
    || props.team.members[0]?.id
    || ''
}

function restartTimer(active: boolean) {
  if (timer) clearInterval(timer)
  timer = undefined
  if (!active) return
  timer = setInterval(() => { now.value = Date.now() }, 1000)
}

watch(() => props.visible, (visible) => {
  if (visible) chooseInitialMember(true)
})
watch(() => props.team.members.map((member) => member.id), () => chooseInitialMember(false))
watch(() => props.visible && props.team.status === 'running', restartTimer, { immediate: true })
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<template>
  <AppDialog
    :visible="visible"
    :title="t('agentTeamDetailTitle')"
    width="760px"
    hide-footer
    @update:visible="emit('update:visible', $event)"
  >
    <div class="agent-team-detail-dialog">
      <div class="agent-team-detail-overview">
        <span class="agent-team-detail-status" :class="`agent-team-detail-status--${team.status}`">
          {{ t(`agentTeamStatus_${team.status}`) }}
        </span>
        <span>{{ terminalCount }}/{{ team.members.length }}</span>
        <span>{{ teamDuration }}</span>
      </div>

      <div v-if="team.members.length > 0" class="agent-team-detail-layout">
        <nav class="agent-team-detail-members" :aria-label="t('agentTeamMembersLabel')">
          <button
            v-for="member in team.members"
            :key="member.id"
            type="button"
            class="agent-team-detail-member"
            :class="{ 'agent-team-detail-member--selected': member.id === selectedMemberId }"
            @click="selectedMemberId = member.id"
          >
            <span class="agent-team-detail-member-name">{{ member.title || t('agentTeamMemberFallback') }}</span>
            <span class="agent-team-detail-member-status" :class="`agent-team-detail-member-status--${member.status}`">
              <span aria-hidden="true">{{ memberStatusSymbol(member) }}</span>
              {{ t(`agentTeamMemberStatus_${member.status}`) }}
            </span>
          </button>
        </nav>

        <article v-if="selectedMember" class="agent-team-detail-content">
          <header class="agent-team-detail-member-header">
            <div>
              <h3>{{ selectedMember.title || t('agentTeamMemberFallback') }}</h3>
              <span class="agent-team-detail-member-status" :class="`agent-team-detail-member-status--${selectedMember.status}`">
                <span aria-hidden="true">{{ memberStatusSymbol(selectedMember) }}</span>
                {{ t(`agentTeamMemberStatus_${selectedMember.status}`) }}
              </span>
            </div>
            <span class="agent-team-detail-duration">{{ memberDuration(selectedMember) }}</span>
          </header>

          <section class="agent-team-detail-section">
            <h4>{{ t('agentTeamTaskLabel') }}</h4>
            <p>{{ selectedMember.task || t('agentTeamNoTask') }}</p>
          </section>

          <section class="agent-team-detail-section">
            <h4>{{ selectedMember.error ? t('agentTeamErrorLabel') : t('agentTeamResultLabel') }}</h4>
            <p :class="{ 'agent-team-detail-error': selectedMember.error }">
              {{ resultPreview || t('agentTeamNoResult') }}
            </p>
          </section>

          <div class="agent-team-detail-tool-count">
            {{ t('agentTeamToolsLabel') }}
            <strong>{{ selectedToolCount }}</strong>
          </div>

          <section v-if="pendingTool" class="agent-team-detail-approval">
            <h4>{{ t('agentTeamApprovalRequired') }}</h4>
            <ToolCallInline
              :item="pendingTool"
              @approve="emit('approve', $event)"
              @reject="emit('reject', $event)"
            />
          </section>
        </article>
      </div>

      <p v-else class="agent-team-detail-empty">{{ t('agentTeamWaitingMembers') }}</p>
    </div>
  </AppDialog>
</template>

<style scoped>
.agent-team-detail-dialog {
  min-width: 0;
  max-height: 80vh;
  overflow: hidden;
}

.agent-team-detail-overview {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
  color: var(--text-secondary);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.agent-team-detail-status,
.agent-team-detail-member-status {
  font-weight: 650;
}

.agent-team-detail-status--running,
.agent-team-detail-member-status--running { color: var(--tool-pending-text); }
.agent-team-detail-status--succeeded,
.agent-team-detail-member-status--succeeded { color: var(--tool-success-text); }
.agent-team-detail-status--partial_failed,
.agent-team-detail-status--failed,
.agent-team-detail-member-status--failed { color: var(--tool-error-text); }

.agent-team-detail-layout {
  display: grid;
  grid-template-columns: minmax(0, 220px) minmax(0, 1fr);
  min-width: 0;
  max-height: calc(80vh - 86px);
  border: 1px solid var(--tool-card-border);
  border-radius: 8px;
  overflow: hidden;
}

.agent-team-detail-members {
  min-width: 0;
  padding: 6px;
  border-right: 1px solid var(--tool-card-border);
  background: var(--tool-summary-bg);
  overflow-y: auto;
}

.agent-team-detail-member {
  display: grid;
  width: 100%;
  min-width: 0;
  gap: 3px;
  padding: 9px 10px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  text-align: left;
}
.agent-team-detail-member:hover { background: var(--primary-alpha-08); }
.agent-team-detail-member--selected { background: var(--primary-alpha-12); }
.agent-team-detail-member:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: -2px;
}

.agent-team-detail-member-name {
  min-width: 0;
  overflow: hidden;
  font-size: 13px;
  font-weight: 620;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.agent-team-detail-member-status { font-size: 11px; }

.agent-team-detail-content {
  min-width: 0;
  padding: 16px;
  overflow-y: auto;
}

.agent-team-detail-member-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--tool-section-border);
}
.agent-team-detail-member-header h3 {
  margin: 0 0 4px;
  font-size: 15px;
  font-weight: 650;
  overflow-wrap: anywhere;
}
.agent-team-detail-duration {
  flex: none;
  color: var(--text-secondary);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.agent-team-detail-section { margin-top: 14px; }
.agent-team-detail-section h4,
.agent-team-detail-approval h4 {
  margin: 0 0 5px;
  color: var(--text-secondary);
  font-size: 11px;
  font-weight: 650;
}
.agent-team-detail-section p {
  margin: 0;
  color: var(--text-primary);
  font-size: 13px;
  line-height: 1.55;
  overflow-wrap: anywhere;
}
.agent-team-detail-section .agent-team-detail-error { color: var(--tool-error-text); }

.agent-team-detail-tool-count {
  display: flex;
  justify-content: space-between;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--tool-section-border);
  color: var(--text-secondary);
  font-size: 12px;
}
.agent-team-detail-tool-count strong { color: var(--text-primary); }

.agent-team-detail-approval {
  min-width: 0;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--tool-section-border);
}
.agent-team-detail-empty {
  margin: 0;
  color: var(--text-secondary);
  font-size: 13px;
  overflow-wrap: anywhere;
}

@media (max-width: 640px) {
  .agent-team-detail-dialog { max-height: calc(100vh - 72px); }
  .agent-team-detail-layout {
    grid-template-columns: minmax(0, 1fr);
    max-height: calc(100vh - 150px);
  }
  .agent-team-detail-members {
    display: flex;
    gap: 4px;
    max-height: 116px;
    border-right: 0;
    border-bottom: 1px solid var(--tool-card-border);
    overflow-y: auto;
  }
  .agent-team-detail-member { min-width: 140px; }
  .agent-team-detail-content { padding: 13px; }
}
</style>
