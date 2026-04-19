/**
 * Banner — top bar showing product name, version, model, and working directory.
 */

import React from "react";
import { Box, Text } from "ink";

interface BannerProps {
  version: string;
  modelName: string;
  cwd: string;
  approvalMode?: string;
  thinkingLevel?: string;
  planMode?: boolean;
}

export function Banner({ version, modelName, cwd, approvalMode, thinkingLevel, planMode }: BannerProps): React.ReactElement {
  const logoLines = [
    "██████████",
    "███ ██ ███",
    "██████████",
  ];

  return (
    <Box flexDirection="row">
      <Box flexDirection="column" marginRight={2}>
        {logoLines.map((line, i) => (
          <Text key={i} color="#a78bfa">
            {line}
          </Text>
        ))}
      </Box>

      <Box flexDirection="column">
        <Text>
          <Text bold color="white">
            SlimeBot CLI{" "}
          </Text>
          <Text color="#94a3b8">v{version}</Text>
          {approvalMode === "auto" && (
            <Text color="#eab308"> [auto]</Text>
          )}
        </Text>
        <Text color="#9ca3af">{modelName || "(none)"}{planMode ? <Text color="#22d3ee"> [plan]</Text> : ""}{thinkingLevel && thinkingLevel !== "off" ? <Text color="#a78bfa"> [think:{thinkingLevel}]</Text> : ""}</Text>
        <Text color="#9ca3af">{cwd}</Text>
      </Box>
    </Box>
  );
}
