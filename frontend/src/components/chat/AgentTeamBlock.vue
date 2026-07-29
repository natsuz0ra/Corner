<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { mdiAccountGroupOutline, mdiChevronRight } from '@mdi/js'
import MdiIcon from '@/components/ui/MdiIcon.vue'
import AgentTeamDetailDialog from '@/components/chat/AgentTeamDetailDialog.vue'
import type { ToolCallItem } from '@/api/chat'
import type { AgentTeamMember, AgentTeamRun } from '@/utils/agentTeam'
import {
  findAgentTeamPendingTool,
  formatAgentTeamDuration,
  getAgentTeamDurationMs,
  selectVisibleAgentTeamMembers,
} from '@/utils/agentTeam'

const props = defineProps<{
  team: AgentTeamRun
  toolCalls: ToolCallItem[]
}>()

const emit = defineEmits<{
  approve: [toolCallId: string]
  reject: [toolCallId: string]
}>()

const { t } = useI18n()
const dialogVisible = ref(false)
const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | undefined

const active = computed(() => props.team.status === 'running')
const terminalCount = computed(() => props.team.members.filter((member) => !['queued', 'running'].includes(member.status)).length)
const duration = computed(() => formatAgentTeamDuration(getAgentTeamDurationMs(props.team, now.value)))
const hasPendingApproval = computed(() => props.team.members.some((member) => findAgentTeamPendingTool(member, props.toolCalls)))
const statusLabel = computed(() => hasPendingApproval.value
  ? t('agentTeamApprovalRequired')
  : t(`agentTeamStatus_${props.team.status}`))
const statusSymbol = computed(() => hasPendingApproval.value ? '\u23F3' : ({
  running: '\u25CF',
  succeeded: '\u2713',
  partial_failed: '!',
  failed: '\u2717',
  canceled: '\u2212',
  interrupted: '\u2016',
}[props.team.status] || '\u25CB'))
const desktopMembers = computed(() => selectVisibleAgentTeamMembers(props.team.members, props.toolCalls, 4))
const mobileMembers = computed(() => selectVisibleAgentTeamMembers(props.team.members, props.toolCalls, 2))

function memberStatusSymbol(member: AgentTeamMember) {
  if (findAgentTeamPendingTool(member, props.toolCalls)) return '\u23F3'
  return ({
    queued: '\u25CB',
    running: '\u25CF',
    succeeded: '\u2713',
    failed: '\u2717',
    canceled: '\u2212',
    interrupted: '\u2016',
  }[member.status] || '\u25CB')
}

function memberStatusLabel(member: AgentTeamMember) {
  return findAgentTeamPendingTool(member, props.toolCalls)
    ? t('agentTeamApprovalRequired')
    : t(`agentTeamMemberStatus_${member.status}`)
}

function restartTimer(isActive: boolean) {
  if (timer) clearInterval(timer)
  timer = undefined
  if (!isActive) return
  timer = setInterval(() => { now.value = Date.now() }, 1000)
}

watch(active, restartTimer, { immediate: true })
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<template>
  <section class="agent-team" :class="[`agent-team--${team.status}`, { 'agent-team--active': active }]">
    <button
      type="button"
      class="agent-team-summary"
      aria-haspopup="dialog"
      :aria-label="t('agentTeamViewDetails')"
      @click="dialogVisible = true"
    >
      <span class="agent-team-primary-row">
        <MdiIcon :path="mdiAccountGroupOutline" :size="17" class="agent-team-icon" />
        <span class="agent-team-title">{{ t('agentTeamTitle') }}</span>
        <span class="agent-team-status" :class="`agent-team-status--${team.status}`">
          <span class="agent-team-status-symbol" aria-hidden="true">{{ statusSymbol }}</span>
          {{ statusLabel }}
        </span>
        <span class="agent-team-count">{{ terminalCount }}/{{ team.members.length }}</span>
        <span class="agent-team-duration">{{ duration }}</span>
        <span class="agent-team-detail-hint">{{ t('agentTeamViewDetails') }}</span>
        <MdiIcon :path="mdiChevronRight" :size="15" class="agent-team-chevron" />
      </span>

      <span v-if="team.members.length > 0" class="agent-team-member-row agent-team-member-row--desktop">
        <span
          v-for="member in desktopMembers"
          :key="member.id"
          class="agent-team-member-summary"
          :class="`agent-team-member-summary--${member.status}`"
        >
          <span aria-hidden="true">{{ memberStatusSymbol(member) }}</span>
          <span class="agent-team-member-name">{{ member.title || t('agentTeamMemberFallback') }}</span>
          <span class="agent-team-member-status">{{ memberStatusLabel(member) }}</span>
        </span>
        <span v-if="team.members.length > desktopMembers.length" class="agent-team-more">
          +{{ team.members.length - desktopMembers.length }}
        </span>
      </span>

      <span v-if="team.members.length > 0" class="agent-team-member-row agent-team-member-row--mobile">
        <span
          v-for="member in mobileMembers"
          :key="member.id"
          class="agent-team-member-summary"
          :class="`agent-team-member-summary--${member.status}`"
        >
          <span aria-hidden="true">{{ memberStatusSymbol(member) }}</span>
          <span class="agent-team-member-name">{{ member.title || t('agentTeamMemberFallback') }}</span>
          <span class="agent-team-member-status">{{ memberStatusLabel(member) }}</span>
        </span>
        <span v-if="team.members.length > mobileMembers.length" class="agent-team-more">
          +{{ team.members.length - mobileMembers.length }}
        </span>
      </span>

      <span v-else class="agent-team-member-row agent-team-empty">{{ t('agentTeamWaitingMembers') }}</span>
    </button>

    <AgentTeamDetailDialog
      v-if="dialogVisible"
      v-model:visible="dialogVisible"
      :team="team"
      :tool-calls="toolCalls"
      @approve="emit('approve', $event)"
      @reject="emit('reject', $event)"
    />
  </section>
</template>

<style scoped>
.agent-team {
  width: min(100%, 680px);
  min-width: 0;
  border: 1px solid var(--tool-card-border);
  border-radius: 8px;
  background: var(--card-bg);
  color: var(--text-primary);
}

.agent-team-summary {
  display: grid;
  width: 100%;
  min-width: 0;
  grid-template-rows: minmax(22px, auto) minmax(20px, auto);
  gap: 3px;
  padding: 7px 10px;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  text-align: left;
}
.agent-team-summary:hover { background: var(--tool-summary-bg); }
.agent-team-summary:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: -2px;
}

.agent-team-primary-row {
  display: grid;
  min-width: 0;
  grid-template-columns: 19px minmax(82px, auto) minmax(74px, auto) auto auto minmax(0, 1fr) 16px;
  align-items: center;
  gap: 7px;
}
.agent-team-icon { color: var(--tool-meta-icon); }
.agent-team-title { font-size: 13px; font-weight: 650; }
.agent-team-status {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  font-size: 12px;
  font-weight: 650;
  white-space: nowrap;
}
.agent-team-status--running { color: var(--tool-pending-text); }
.agent-team-status--succeeded { color: var(--tool-success-text); }
.agent-team-status--partial_failed,
.agent-team-status--failed { color: var(--tool-error-text); }
.agent-team-status--canceled,
.agent-team-status--interrupted { color: var(--text-secondary); }
.agent-team-count,
.agent-team-duration,
.agent-team-detail-hint {
  color: var(--text-secondary);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.agent-team-detail-hint { justify-self: end; }
.agent-team-chevron { color: var(--tool-summary-text); }

.agent-team-member-row {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 6px;
  padding-left: 26px;
  overflow: hidden;
}
.agent-team-member-row--mobile { display: none; }
.agent-team-member-summary {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 4px;
  color: var(--text-secondary);
  font-size: 11px;
  white-space: nowrap;
}
.agent-team-member-summary--running { color: var(--tool-pending-text); }
.agent-team-member-summary--succeeded { color: var(--tool-success-text); }
.agent-team-member-summary--failed { color: var(--tool-error-text); }
.agent-team-member-name {
  max-width: 112px;
  overflow: hidden;
  color: var(--text-primary);
  font-weight: 600;
  text-overflow: ellipsis;
}
.agent-team-member-status { color: inherit; }
.agent-team-more,
.agent-team-empty {
  color: var(--text-secondary);
  font-size: 11px;
  white-space: nowrap;
}
.agent-team-empty {
  min-width: 0;
  overflow-wrap: anywhere;
}

.agent-team--active .agent-team-status-symbol { animation: agent-team-pulse 1.6s ease-in-out infinite; }
@keyframes agent-team-pulse { 50% { opacity: 0.4; } }

@media (max-width: 480px) {
  .agent-team-summary { padding: 7px 8px; }
  .agent-team-primary-row {
    grid-template-columns: 19px minmax(0, 1fr) auto auto 16px;
    gap: 5px;
  }
  .agent-team-title { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .agent-team-status { grid-column: 3; }
  .agent-team-count { grid-column: 4; }
  .agent-team-duration,
  .agent-team-detail-hint { display: none; }
  .agent-team-chevron { grid-column: 5; }
  .agent-team-member-row { padding-left: 24px; }
  .agent-team-member-row--desktop { display: none; }
  .agent-team-member-row--mobile { display: flex; }
  .agent-team-member-name { max-width: 86px; }
  .agent-team-member-status { display: none; }
}

@media (prefers-reduced-motion: reduce) {
  .agent-team-status-symbol { animation: none !important; }
}
</style>
