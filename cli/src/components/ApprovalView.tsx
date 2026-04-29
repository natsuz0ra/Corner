/**
 * ApprovalView — tool approval dialog.
 * Shows tool name, command, and args for user approval or denial.
 */

import React from "react";
import { Box, Text } from "ink";

interface ApprovalViewProps {
  toolName: string;
  command: string;
  params: Record<string, string>;
  items?: Array<{
    toolCallId: string;
    toolName: string;
    command: string;
    params: Record<string, string>;
  }>;
  cursor?: number;
}

export function ApprovalView({
  toolName,
  command,
  params,
  items,
  cursor = 0,
}: ApprovalViewProps): React.ReactElement {
  const approvalItems = items && items.length > 0
    ? items
    : [{ toolCallId: "", toolName, command, params }];
  return (
    <Box flexDirection="column">
      <Text bold color="yellow">
        Tool Approval Required{approvalItems.length > 1 ? ` (${approvalItems.length})` : ""}
      </Text>
      {approvalItems.map((item, index) => {
        const itemParamStr = Object.keys(item.params).length > 0
          ? Object.entries(item.params).map(([k, v]) => `${k}=${v}`).join(", ")
          : "";
        const selected = index === cursor;
        return (
          <Box key={item.toolCallId || `${item.toolName}-${index}`} flexDirection="column">
            <Text>
              <Text color={selected ? "cyan" : "gray"}>{selected ? "❯ " : "  "}</Text>
              Tool: <Text bold>{item.toolName}.{item.command}</Text>
            </Text>
            {itemParamStr && (
              <Text color="gray">
                {"  "}Params: {itemParamStr}
              </Text>
            )}
          </Box>
        );
      })}
      <Text color="gray">
        ↑/↓ select | y approve | n/Esc reject | a approve all | r reject all
      </Text>
    </Box>
  );
}
