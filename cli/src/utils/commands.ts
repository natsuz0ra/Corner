/**
 * Command definitions, matching, and tab completion.
 */

import { SUPPORTED_COMMANDS, type CommandMeta } from "../types.js";

const MAX_HINTS = 5;

/** Return command hints matching the prefix */
export function matchCommandHints(input: string): CommandMeta[] {
  const trimmed = input.trim();
  if (!trimmed.startsWith("/")) return [];
  const matched: CommandMeta[] = [];
  for (const cmd of SUPPORTED_COMMANDS) {
    if (cmd.command.startsWith(trimmed)) {
      matched.push(cmd);
      if (matched.length >= MAX_HINTS) break;
    }
  }
  return matched;
}

/** Tab completion: first matching full command */
export function completeCommand(input: string): string | null {
  const matched = matchCommandHints(input);
  if (matched.length === 0) return null;
  return matched[0].command;
}

/** Whether input is a command (starts with /) */
export function isCommand(input: string): boolean {
  return input.trim().startsWith("/");
}
