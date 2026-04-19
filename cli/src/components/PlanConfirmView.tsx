/**
 * PlanConfirmView — plan confirmation dialog for CLI.
 * Shows three options (Execute, Modify, Cancel) with arrow key navigation.
 */

import React from "react";
import { Box, Text } from "ink";

const OPTIONS = [
  { label: "Execute Plan", color: "green" as const },
  { label: "Modify Plan", color: "yellow" as const },
  { label: "Cancel", color: "red" as const },
];

interface PlanConfirmViewProps {
  cursor: number;
  modifying: boolean;
  modifyInput: string;
}

export function PlanConfirmView({
  cursor,
  modifying,
  modifyInput,
}: PlanConfirmViewProps): React.ReactElement {
  if (modifying) {
    return (
      <Box flexDirection="column">
        <Text bold color="cyan">
          Modify Plan — Enter your feedback
        </Text>
        <Text color="gray">
          Type your feedback below, Enter to submit, Esc to go back.
        </Text>
        <Text>
          {"> "}
          <Text color="white">{modifyInput}</Text>
          <Text color="gray">_</Text>
        </Text>
      </Box>
    );
  }

  return (
    <Box flexDirection="column">
      <Text bold color="cyan">
        Plan Generated — Choose an action
      </Text>
      {OPTIONS.map((opt, i) => (
        <Text key={opt.label}>
          {i === cursor ? (
            <Text bold color={opt.color}>
              {"  > "}
              {opt.label}
            </Text>
          ) : (
            <Text color="gray">
              {"    "}
              {opt.label}
            </Text>
          )}
        </Text>
      ))}
      <Text color="gray">
        Arrow keys to navigate, Enter to select, Esc to cancel
      </Text>
    </Box>
  );
}
