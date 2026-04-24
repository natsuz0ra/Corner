<script setup lang="ts">
import { renderMarkdown } from '@/utils/markdown'

defineProps<{
  content: string
}>()
</script>

<template>
  <section class="plan-block" aria-label="Plan">
    <header class="plan-block-header">
      <span class="plan-block-dot" aria-hidden="true" />
      <span class="plan-block-title">Plan</span>
    </header>
    <div class="plan-block-body bubble-markdown" v-html="renderMarkdown(content)" />
  </section>
</template>

<style scoped>
.plan-block {
  overflow: hidden;
  border: 1px solid rgba(14, 165, 233, 0.24);
  border-left: 3px solid #0ea5e9;
  border-radius: 8px;
  background:
    linear-gradient(90deg, rgba(14, 165, 233, 0.08), rgba(14, 165, 233, 0.02)),
    var(--card-bg);
}

.plan-block-header {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 34px;
  padding: 8px 12px 6px;
  color: #0284c7;
}

.plan-block-dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: currentColor;
  box-shadow: 0 0 0 3px rgba(14, 165, 233, 0.12);
}

.plan-block-title {
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0;
}

.plan-block-body {
  padding: 0 12px 12px 23px;
  color: var(--text-primary);
}

.plan-block-body :deep(h1),
.plan-block-body :deep(h2),
.plan-block-body :deep(h3) {
  margin-top: 4px;
}

.dark .plan-block {
  border-color: rgba(56, 189, 248, 0.28);
  border-left-color: #38bdf8;
  background:
    linear-gradient(90deg, rgba(56, 189, 248, 0.13), rgba(56, 189, 248, 0.04)),
    var(--card-bg);
}

.dark .plan-block-header {
  color: #7dd3fc;
}
</style>
