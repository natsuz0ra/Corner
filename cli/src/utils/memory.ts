import type { MemorySnapshot, MemoryTarget, MemoryTargetState } from "../types.js";

export function normalizeMemoryResetTarget(value: string): MemoryTarget | "all" | null {
  const trimmed = value.trim().toLowerCase();
  if (trimmed === "memory" || trimmed === "user" || trimmed === "all") {
    return trimmed;
  }
  return null;
}

export function formatMemorySnapshot(snapshot: MemorySnapshot): string {
  const lines = [
    "Memory",
    `Status: ${snapshot.memoryEnabled ? "enabled" : "disabled"} | user profile: ${snapshot.memoryUserProfileEnabled ? "enabled" : "disabled"} | review interval: ${snapshot.memoryNudgeInterval}`,
  ];
  if (snapshot.memoryDirectory) {
    lines.push(`Path: ${snapshot.memoryDirectory}`);
  }
  lines.push("", ...formatTarget("Personal notes", snapshot.memory), "", ...formatTarget("User profile", snapshot.user));
  return lines.join("\n");
}

function formatTarget(title: string, state: MemoryTargetState): string[] {
  const percent = state.charLimit > 0 ? Math.round((state.usageChars / state.charLimit) * 100) : 0;
  const lines = [`${title}: ${state.usageChars}/${state.charLimit} chars (${percent}%)`];
  if (!state.enabled) {
    lines[0] += " [disabled]";
  }
  if (state.entries.length === 0) {
    lines.push("  (empty)");
    return lines;
  }
  state.entries.forEach((entry, index) => {
    lines.push(`  ${index + 1}. ${entry.replace(/\n/g, "\n     ")}`);
  });
  return lines;
}
