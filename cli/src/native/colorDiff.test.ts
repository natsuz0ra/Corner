import assert from "node:assert/strict";
import test from "node:test";
import {
  renderColorDiff,
  renderColorDiffRows,
  setNativeColorDiffForTest,
} from "./colorDiff";
import { stripAnsi } from "../utils/terminal";

test("renderColorDiffRows uses native renderer when available", () => {
  setNativeColorDiffForTest({
    renderColorDiffJson: () => JSON.stringify([{ gutter: "N 1", content: "native" }]),
  });

  const rows = renderColorDiffRows({
    filePath: "src/example.ts",
    width: 80,
    lines: [{ kind: "added", newLine: 1, text: "const ok = true" }],
  });

  assert.deepEqual(rows, [{ gutter: "N 1", content: "native" }]);
  setNativeColorDiffForTest(null);
});

test("renderColorDiffRows falls back to TypeScript renderer when native fails", () => {
  setNativeColorDiffForTest({
    renderColorDiffJson: () => {
      throw new Error("native unavailable");
    },
  });

  const rows = renderColorDiffRows({
    filePath: "src/example.go",
    width: 80,
    lines: [{ kind: "added", newLine: 1, text: "func main() {}" }],
  });

  assert.equal(stripAnsi(rows[0]!.gutter).trim(), "+ 1");
  assert.ok(stripAnsi(rows[0]!.content).includes("func main()"));
  assert.match(rows[0]!.content, /\x1b\[/);
  setNativeColorDiffForTest(null);
});

test("renderColorDiff returns ANSI lines with plain content intact", () => {
  setNativeColorDiffForTest(null);
  const lines = renderColorDiff({
    filePath: "src/example.ts",
    width: 80,
    lines: [
      { kind: "removed", oldLine: 1, text: "const oldName = 1" },
      { kind: "added", newLine: 1, text: "const newName = 1" },
    ],
  });
  const plain = lines.map((line) => stripAnsi(line));

  assert.ok(plain[0]!.includes("- 1"));
  assert.ok(plain[0]!.includes("const oldName = 1"));
  assert.ok(plain[1]!.includes("+ 1"));
  assert.ok(plain[1]!.includes("const newName = 1"));
  assert.ok(lines.some((line) => /\x1b\[/.test(line)));
});
