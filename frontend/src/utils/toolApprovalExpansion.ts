import type { ToolCallItem } from '@/api/chat'

export function hasPendingNestedApproval(item: ToolCallItem, nestedTools: ToolCallItem[]): boolean {
  if (item.toolName !== 'run_subagent') return false
  return nestedTools.some((tool) => tool.status === 'pending')
}

export function shouldAutoExpandToolCall(item: ToolCallItem, nestedTools: ToolCallItem[]): boolean {
  if (item.status === 'pending') return true
  return hasPendingNestedApproval(item, nestedTools)
}
