import assert from "node:assert/strict";
import test from "node:test";
import { formatMemorySnapshot, normalizeMemoryResetTarget } from "./memory.js";

test("normalizeMemoryResetTarget accepts memory user and all", () => {
  assert.equal(normalizeMemoryResetTarget("memory"), "memory");
  assert.equal(normalizeMemoryResetTarget("user"), "user");
  assert.equal(normalizeMemoryResetTarget("all"), "all");
  assert.equal(normalizeMemoryResetTarget("project"), null);
});

test("formatMemorySnapshot renders status and entries", () => {
  const output = formatMemorySnapshot({
    memoryEnabled: true,
    memoryUserProfileEnabled: true,
    memoryNudgeInterval: 10,
    memoryDirectory: "/tmp/memories",
    memory: { target: "memory", entries: ["项目约定"], usageChars: 4, charLimit: 2200, entryCount: 1, enabled: true },
    user: { target: "user", entries: [], usageChars: 0, charLimit: 1375, entryCount: 0, enabled: true },
  });

  assert.match(output, /Memory/);
  assert.match(output, /Path: \/tmp\/memories/);
  assert.match(output, /项目约定/);
  assert.match(output, /User profile/);
});
