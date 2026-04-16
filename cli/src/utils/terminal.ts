/**
 * Terminal helpers.
 */

/** Clear screen and move cursor to top-left */
export function clearScreen(): void {
  process.stdout.write("\x1b[2J\x1b[3J\x1b[H");
}

/** Terminal width (default 80) */
export function terminalWidth(): number {
  return process.stdout.columns || 80;
}

/** Terminal height (default 24) */
export function terminalHeight(): number {
  return process.stdout.rows || 24;
}

/**
 * Indent multiline text with a prefix; continuation lines are padded to align with the prefix.
 * Same idea as Go indentMultilineANSI.
 */
export function indentMultiline(prefix: string, body: string): string {
  if (!body) return prefix.trimEnd();
  const lines = body.split("\n");
  const pad = " ".repeat(stripAnsi(prefix).length);
  return lines.map((line, i) => (i === 0 ? prefix : pad) + line).join("\n");
}

/** Strip ANSI escapes for approximate visible width */
export function stripAnsi(str: string): string {
  // eslint-disable-next-line no-control-regex
  return str.replace(/\x1b\[[0-9;]*[a-zA-Z]/g, "");
}

/** Visible width of text (length after stripping ANSI) */
export function visualWidth(str: string): number {
  return stripAnsi(str).length;
}

/** Whether the terminal supports truecolor */
export function supportsTrueColor(): boolean {
  const term = process.env.TERM || "";
  const colorterm = process.env.COLORTERM || "";
  return colorterm === "truecolor" || colorterm === "24bit" || term === "xterm-256color";
}

/** Set terminal title (OSC sequence) */
export function setTerminalTitle(title: string): void {
  if (!process.stdout.isTTY) return;
  process.stdout.write(`\x1b]2;${title}\x07`);
}

export const DOT = "\u25CF"; // ●
