/**
 * BlinkDot — blinking dot indicator.
 * When blinkOn=true renders a colored ●; when false, a fixed-width space.
 * Same visual width (one cell) so layout does not jitter.
 */

import React from "react";
import { Text } from "ink";
import { DOT } from "../utils/terminal.js";

interface BlinkDotProps {
  color: string;
  blinkOn: boolean;
}

/** Dot color map */
const DOT_COLORS: Record<string, string> = {
  white: "white",
  yellow: "yellow",
  green: "green",
  red: "red",
};

export function BlinkDot({ color, blinkOn }: BlinkDotProps): React.ReactElement {
  const inkColor = DOT_COLORS[color] || "white";

  return (
    <Text bold color={blinkOn ? inkColor : undefined}>
      {blinkOn ? DOT : " "}
    </Text>
  );
}

/** Dot state while the assistant is waiting */
export function assistantDotState(waiting: boolean): { color: string; blinkOn: boolean } {
  return { color: "white", blinkOn: waiting };
}

/** Dot state for tool calls */
export function toolDotState(status: string): { color: string; blinkOn: boolean } {
  switch (status.trim()) {
    case "pending":
      return { color: "yellow", blinkOn: false };
    case "executing":
      return { color: "yellow", blinkOn: true };
    case "completed":
      return { color: "green", blinkOn: false };
    case "error":
    case "rejected":
      return { color: "red", blinkOn: false };
    default:
      return { color: "white", blinkOn: false };
  }
}
