import type { MemoryTarget } from "../types.js";

export type MemoryConsoleMode = "actions" | "edit" | "reset" | "view";
export type MemoryConsoleEditField = "memoryCharLimit" | "memoryUserCharLimit" | "memoryNudgeInterval";

export type MemoryConsoleAction =
  | "toggle-memory"
  | "toggle-user-profile"
  | "edit-memory-budget"
  | "edit-user-budget"
  | "edit-review-interval"
  | "view-memory"
  | "view-user"
  | "reset";

export interface MemoryConsoleActionItem {
  action: MemoryConsoleAction;
  title: string;
  desc: string;
}

export const MEMORY_CONSOLE_ACTIONS: MemoryConsoleActionItem[] = [
  { action: "toggle-memory", title: "Toggle long-term memory", desc: "Turn memory writes and context injection on or off" },
  { action: "toggle-user-profile", title: "Toggle user profile", desc: "Store preferences, background, and recurring habits" },
  { action: "edit-memory-budget", title: "Edit personal notes budget", desc: "Change the character budget for personal notes" },
  { action: "edit-user-budget", title: "Edit user profile budget", desc: "Change the character budget for user profile memory" },
  { action: "edit-review-interval", title: "Edit review interval", desc: "Change how often memory review runs, in turns" },
  { action: "view-memory", title: "View personal notes", desc: "Show saved personal note entries" },
  { action: "view-user", title: "View user profile", desc: "Show saved user profile entries" },
  { action: "reset", title: "Reset memory...", desc: "Clear personal notes, user profile, or all memory" },
];

export type MemoryConsoleResetAction = MemoryTarget | "all" | "cancel";

export interface MemoryConsoleResetItem {
  action: MemoryConsoleResetAction;
  title: string;
  desc: string;
}

export const MEMORY_CONSOLE_RESET_ACTIONS: MemoryConsoleResetItem[] = [
  { action: "memory", title: "Personal notes", desc: "Clear only personal note entries" },
  { action: "user", title: "User profile", desc: "Clear only user profile entries" },
  { action: "all", title: "All memory", desc: "Clear personal notes and user profile entries" },
  { action: "cancel", title: "Cancel", desc: "Return to Memory Console without clearing anything" },
];

export const MEMORY_CONSOLE_EDIT_LIMITS: Record<MemoryConsoleEditField, { min: number; max: number; label: string; unit: string }> = {
  memoryCharLimit: { min: 200, max: 20000, label: "Personal notes budget", unit: "chars" },
  memoryUserCharLimit: { min: 200, max: 20000, label: "User profile budget", unit: "chars" },
  memoryNudgeInterval: { min: 1, max: 100, label: "Review interval", unit: "turns" },
};

export function sanitizeMemoryConsoleDraft(value: string): string {
  return value.replace(/\D/g, "");
}

export function parseMemoryConsoleDraft(field: MemoryConsoleEditField, rawValue: string): { value: number | null; error: string } {
  const limits = MEMORY_CONSOLE_EDIT_LIMITS[field];
  const sanitized = sanitizeMemoryConsoleDraft(rawValue);
  if (!sanitized) {
    return { value: null, error: `${limits.label} must be a number.` };
  }
  const value = Number(sanitized);
  if (!Number.isFinite(value)) {
    return { value: null, error: `${limits.label} must be a number.` };
  }
  return { value: Math.max(limits.min, Math.min(limits.max, value)), error: "" };
}

export function memoryConsoleActionCount(mode: MemoryConsoleMode): number {
  if (mode === "reset") return MEMORY_CONSOLE_RESET_ACTIONS.length;
  if (mode === "view" || mode === "edit") return 1;
  return MEMORY_CONSOLE_ACTIONS.length;
}
