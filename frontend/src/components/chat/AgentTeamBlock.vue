<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { mdiAccountGroupOutline, mdiChevronDown } from '@mdi/js'
import MdiIcon from '@/components/ui/MdiIcon.vue'
import ToolCallInline from '@/components/chat/ToolCallInline.vue'
import type { ToolCallItem } from '@/api/chat'
import type { AgentTeamMember, AgentTeamRun } from '@/utils/agentTeam'
import { formatAgentTeamDuration, getAgentTeamDurationMs } from '@/utils/agentTeam'

const props = defineProps<{
  team: AgentTeamRun
  toolCalls: ToolCallItem[]
}>()

const emit = defineEmits<{
  approve: [toolCallId: string]
  reject: [toolCallId: string]
}>()

const { t } = useI18n()
const expanded = ref(props.team.status === 'running')
const expandedMemberIds = ref<string[]>([])
const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | undefined

const active = computed(() => props.team.status === 'running')
const terminalCount = computed(() => props.team.members.filter((item) => !['queued', 'running'].includes(item.status)).length)
const duration = computed(() => formatAgentTeamDuration(getAgentTeamDurationMs(props.team, now.value)))
const statusLabel = computed(() => t(`agentTeamStatus_${props.team.status}`))
const statusSymbol = computed(() => ({
  running: '\u25CF',
  succeeded: '\u2713',
  partial_failed: '!',
  failed: '\u2717',
  canceled: '\u2212',
  interrupted: '\u2016',
}[props.team.status] || '\u25CB'))

function memberStatusLabel(member: AgentTeamMember) {
  return t(`agentTeamMemberStatus_${member.status}`)
}

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

function toolForMember(member: AgentTeamMember) {
  return props.toolCalls.find((item) => item.toolCallId === member.toolCallId)
}

function nestedToolsForMember(member: AgentTeamMember) {
  return props.toolCalls.filter((item) => item.parentToolCallId === member.toolCallId)
}

function isMemberExpanded(memberId: string) {
  return expandedMemberIds.value.includes(memberId)
}

function toggleMember(memberId: string) {
  expandedMemberIds.value = isMemberExpanded(memberId)
    ? expandedMemberIds.value.filter((id) => id !== memberId)
    : [...expandedMemberIds.value, memberId]
}

function restartTimer(isActive: boolean) {
  if (timer) clearInterval(timer)
  timer = undefined
  if (!isActive) return
  timer = setInterval(() => { now.value = Date.now() }, 250)
}

watch(active, restartTimer, { immediate: true })
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<template>
  <section class="agent-team" :class="[`agent-team--${team.status}`, { 'agent-team--active': active }]">
    <button
      type="button"
      class="agent-team-summary"
      :aria-expanded="expanded"
      :aria-controls="`agent-team-members-${team.id}`"
      @click="expanded = !expanded"
    >
      <MdiIcon :path="mdiAccountGroupOutline" :size="18" class="agent-team-icon" />
      <span class="agent-team-title">{{ t('agentTeamTitle') }}</span>
      <span class="agent-team-status" :class="`agent-team-status--${team.status}`">
        <span class="agent-team-status-symbol" aria-hidden="true">{{ statusSymbol }}</span>
        {{ statusLabel }}
      </span>
      <span class="agent-team-count">{{ terminalCount }}/{{ team.members.length }}</span>
      <span class="agent-team-duration">{{ duration }}</span>
      <MdiIcon
        :path="mdiChevronDown"
        :size="16"
        class="agent-team-chevron"
        :class="{ 'agent-team-chevron--open': expanded }"
      />
    </button>

    <div v-if="expanded" :id="`agent-team-members-${team.id}`" class="agent-team-members">
      <article v-for="member in team.members" :key="member.id" class="agent-team-member">
        <button
          type="button"
          class="agent-team-member-summary"
          :aria-expanded="isMemberExpanded(member.id)"
          :aria-controls="`agent-team-member-detail-${member.id}`"
          @click="toggleMember(member.id)"
        >
          <span class="agent-team-member-status" :class="`agent-team-member-status--${member.status}`">
            <span aria-hidden="true">{{ memberStatusSymbol(member) }}</span>
            {{ memberStatusLabel(member) }}
          </span>
          <span class="agent-team-member-copy">
            <strong>{{ member.title || t('agentTeamMemberFallback') }}</strong>
            <span>{{ member.task }}</span>
          </span>
          <span class="agent-team-member-duration">{{ memberDuration(member) }}</span>
          <MdiIcon
            :path="mdiChevronDown"
            :size="15"
            class="agent-team-chevron"
            :class="{ 'agent-team-chevron--open': isMemberExpanded(member.id) }"
          />
        </button>
        <div
          v-if="isMemberExpanded(member.id)"
          :id="`agent-team-member-detail-${member.id}`"
          class="agent-team-member-detail"
        >
          <ToolCallInline
            v-if="toolForMember(member)"
            :item="toolForMember(member)!"
            :nested-tools="nestedToolsForMember(member)"
            @approve="emit('approve', $event)"
            @reject="emit('reject', $event)"
          />
          <p v-else-if="member.error" class="agent-team-member-error">{{ member.error }}</p>
          <p v-else-if="member.answer" class="agent-team-member-answer">{{ member.answer }}</p>
        </div>
      </article>
      <p v-if="team.members.length === 0" class="agent-team-empty">{{ t('agentTeamWaitingMembers') }}</p>
    </div>
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

.agent-team-summary,
.agent-team-member-summary {
  width: 100%;
  min-width: 0;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  text-align: left;
}

.agent-team-summary {
  display: grid;
  grid-template-columns: 20px minmax(88px, auto) minmax(84px, auto) auto auto 18px;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 8px 10px;
}

.agent-team-summary:hover,
.agent-team-member-summary:hover {
  background: var(--tool-summary-bg);
}

.agent-team-summary:focus-visible,
.agent-team-member-summary:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: -2px;
}

.agent-team-icon { color: var(--tool-meta-icon); }
.agent-team-title { font-size: 13px; font-weight: 650; }

.agent-team-status,
.agent-team-member-status {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
  font-size: 12px;
  font-weight: 650;
  white-space: nowrap;
}

.agent-team-status--running,
.agent-team-member-status--running { color: var(--tool-pending-text); }
.agent-team-status--succeeded,
.agent-team-member-status--succeeded { color: var(--tool-success-text); }
.agent-team-status--partial_failed,
.agent-team-status--failed,
.agent-team-member-status--failed { color: var(--tool-error-text); }
.agent-team-status--canceled,
.agent-team-status--interrupted,
.agent-team-member-status--queued,
.agent-team-member-status--canceled,
.agent-team-member-status--interrupted { color: var(--text-secondary); }

.agent-team-count,
.agent-team-duration,
.agent-team-member-duration {
  color: var(--text-secondary);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.agent-team-chevron {
  color: var(--tool-summary-text);
  transition: transform 150ms ease;
}
.agent-team-chevron--open { transform: rotate(180deg); }

.agent-team-members {
  min-width: 0;
  border-top: 1px solid var(--tool-card-border);
}

.agent-team-member + .agent-team-member { border-top: 1px solid var(--tool-section-border); }

.agent-team-member-summary {
  display: grid;
  grid-template-columns: minmax(74px, auto) minmax(0, 1fr) auto 18px;
  align-items: center;
  gap: 10px;
  padding: 8px 10px 8px 16px;
}

.agent-team-member-copy {
  display: grid;
  min-width: 0;
  gap: 2px;
}
.agent-team-member-copy strong { font-size: 13px; font-weight: 620; }
.agent-team-member-copy span {
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.agent-team-member-detail { min-width: 0; padding: 0 10px 10px 16px; }
.agent-team-member-error,
.agent-team-member-answer,
.agent-team-empty {
  margin: 0;
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.5;
  overflow-wrap: anywhere;
}
.agent-team-member-error { color: var(--tool-error-text); }
.agent-team-empty { padding: 10px 16px; }

.agent-team--active .agent-team-status-symbol { animation: agent-team-pulse 1.6s ease-in-out infinite; }
@keyframes agent-team-pulse { 50% { opacity: 0.4; } }

@media (max-width: 480px) {
  .agent-team-summary {
    grid-template-columns: 20px minmax(0, 1fr) auto 18px;
    gap: 6px;
  }
  .agent-team-status { grid-column: 2; }
  .agent-team-count { grid-column: 3; grid-row: 1; }
  .agent-team-duration { grid-column: 3; grid-row: 2; }
  .agent-team-chevron { grid-column: 4; grid-row: 1 / span 2; }
  .agent-team-member-summary {
    grid-template-columns: minmax(0, 1fr) auto 18px;
    align-items: start;
    padding-left: 10px;
  }
  .agent-team-member-status { grid-column: 1; }
  .agent-team-member-copy { grid-column: 1 / -1; grid-row: 2; }
  .agent-team-member-duration { grid-column: 2; grid-row: 1; }
  .agent-team-member-summary > .agent-team-chevron { grid-column: 3; grid-row: 1; }
  .agent-team-member-detail { padding-left: 10px; }
}

@media (prefers-reduced-motion: reduce) {
  .agent-team-status-symbol { animation: none !important; }
  .agent-team-chevron { transition: none; }
}
</style>
