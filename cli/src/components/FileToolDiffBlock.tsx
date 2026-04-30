import React from "react";
import { Box, Text } from "ink";
import type { TimelineEntry } from "../types.js";
import { buildFileToolDisplay } from "../utils/fileToolDisplay.js";
import { renderColorDiffRows } from "../native/colorDiff.js";

const MAX_PREVIEW_LINES = 8;
const BORDER_COLOR = "#64748b";
const SUMMARY_COLOR = "#cbd5e1";
const HINT_COLOR = "#94a3b8";

interface FileToolDiffBlockProps {
  entry: TimelineEntry;
  maxWidth: number;
  expanded: boolean;
}

export function FileToolDiffBlock({
  entry,
  maxWidth,
  expanded,
}: FileToolDiffBlockProps): React.ReactElement | null {
  const display = buildFileToolDisplay(entry);
  if (!display) return null;

  if (entry.status === "error" || entry.status === "rejected") {
    return (
      <Box marginLeft={3} flexDirection="column">
        <Text color="red">{entry.error || entry.content || "File tool failed"}</Text>
      </Box>
    );
  }

  if (entry.status !== "completed") return null;

  if (display.toolName === "file_read" || display.diffLines.length === 0) {
    return (
      <Box marginLeft={3}>
        <Text color={SUMMARY_COLOR}>└─ {display.summary}</Text>
      </Box>
    );
  }

  const width = Math.max(24, Math.min(maxWidth - 3, maxWidth));
  const previewLines = expanded ? display.diffLines : display.diffLines.slice(0, MAX_PREVIEW_LINES);
  const remaining = display.diffLines.length - previewLines.length;
  const renderedRows = renderColorDiffRows({
    filePath: display.filePath,
    lines: previewLines,
    width: Math.max(12, width - 2),
  });
  return (
    <Box marginLeft={3} flexDirection="column">
      <Text color={SUMMARY_COLOR}>└─ {display.summary}</Text>
      <Box flexDirection="column" borderStyle="single" borderColor={BORDER_COLOR} borderLeft={false} borderRight={false}>
        <Box flexDirection="row">
          <Box flexDirection="column" flexShrink={0}>
            {renderedRows.map((row, index) => (
              <Text key={`gutter-${index}`}>{row.gutter}</Text>
            ))}
          </Box>
          <Box flexDirection="column" marginLeft={1}>
            {renderedRows.map((row, index) => (
              <Text key={`content-${index}`}>{row.content}</Text>
            ))}
          </Box>
        </Box>
        {remaining > 0 ? (
          <Text color={HINT_COLOR}>... +{remaining} more changed lines (ctrl+o to expand)</Text>
        ) : expanded && display.diffLines.length > MAX_PREVIEW_LINES ? (
          <Text color={HINT_COLOR}>... (ctrl+o to collapse)</Text>
        ) : null}
      </Box>
    </Box>
  );
}
