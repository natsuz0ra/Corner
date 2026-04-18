<script setup lang="ts">
import ToolCallInline from '@/components/chat/ToolCallInline.vue'
import ThinkingBlock from '@/components/chat/ThinkingBlock.vue'
import TypingDots from '@/components/chat/TypingDots.vue'
import { renderMarkdown } from '@/utils/markdown'
import type { MessageItem } from '@/api/chat'
import { useChatContext } from '@/composables/chat/useChatContext'

defineProps<{
  item: MessageItem
}>()

const ctx = useChatContext()
</script>

<template>
  <div class="text-sm leading-relaxed w-full">
    <template v-for="entry in ctx.getReplyTimeline(item.id)" :key="entry.id">
      <ThinkingBlock
        v-if="entry.kind === 'thinking'"
        :content="entry.content"
        :done="entry.done"
        :duration-ms="entry.durationMs"
      />

      <div v-else-if="entry.kind === 'text'" class="bubble-markdown sb-text-primary" v-html="renderMarkdown(entry.content)" />

      <div v-else-if="entry.kind === 'tool_start'" class="my-1.5">
        <ToolCallInline
          v-if="ctx.getReplyToolItem(item.id, entry.toolCallId)"
          :item="ctx.getReplyToolItem(item.id, entry.toolCallId)!"
          :nested-tools="ctx.getSubagentChildTools(item.id, entry.toolCallId)"
          @approve="ctx.approveToolCall($event, true)"
          @reject="ctx.approveToolCall($event, false)"
        />
      </div>
    </template>

    <TypingDots v-if="ctx.isEmptyPlaceholder(item.id) && ctx.waiting" />
  </div>
</template>
