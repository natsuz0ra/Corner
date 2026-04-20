/**
 * Splits plan text into narration (before first heading) and plan body.
 * Mirrors the Go backend splitNarrationAndPlan logic.
 */
export function splitNarrationAndPlan(fullText: string): { narration: string; planBody: string } {
  const lines = fullText.split("\n");
  for (let i = 0; i < lines.length; i++) {
    if (lines[i]!.trimStart().startsWith("# ")) {
      return {
        narration: lines.slice(0, i).join("\n").trimEnd(),
        planBody: lines.slice(i).join("\n"),
      };
    }
  }
  return { narration: "", planBody: fullText };
}
