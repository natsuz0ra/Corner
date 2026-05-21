import assert from "node:assert/strict";
import test from "node:test";
import {
  MEMORY_CONSOLE_ACTIONS,
  MEMORY_CONSOLE_RESET_ACTIONS,
  memoryConsoleActionCount,
  parseMemoryConsoleDraft,
  sanitizeMemoryConsoleDraft,
} from "./memoryConsole.js";

test("memory console actions expose settings and reset entries", () => {
  assert.deepEqual(
    MEMORY_CONSOLE_ACTIONS.map((item) => item.action),
    [
      "toggle-memory",
      "toggle-user-profile",
      "edit-memory-budget",
      "edit-user-budget",
      "edit-review-interval",
      "view-memory",
      "view-user",
      "reset",
    ],
  );
  assert.deepEqual(MEMORY_CONSOLE_RESET_ACTIONS.map((item) => item.action), ["memory", "user", "all", "cancel"]);
});

test("memory console draft parsing accepts only valid numeric ranges", () => {
  assert.equal(sanitizeMemoryConsoleDraft("12a3"), "123");
  assert.deepEqual(parseMemoryConsoleDraft("memoryCharLimit", "2200"), { value: 2200, error: "" });
  assert.equal(parseMemoryConsoleDraft("memoryCharLimit", "").value, null);
  assert.equal(parseMemoryConsoleDraft("memoryCharLimit", "199").value, 200);
  assert.equal(parseMemoryConsoleDraft("memoryCharLimit", "20001").value, 20000);
  assert.deepEqual(parseMemoryConsoleDraft("memoryNudgeInterval", "10"), { value: 10, error: "" });
  assert.equal(parseMemoryConsoleDraft("memoryNudgeInterval", "0").value, 1);
});

test("memory console action count follows mode", () => {
  assert.equal(memoryConsoleActionCount("actions"), 8);
  assert.equal(memoryConsoleActionCount("reset"), 4);
  assert.equal(memoryConsoleActionCount("edit"), 1);
  assert.equal(memoryConsoleActionCount("view"), 1);
});
