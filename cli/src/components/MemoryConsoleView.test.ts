import assert from "node:assert/strict";
import test from "node:test";
import type React from "react";
import type { MemorySnapshot } from "../types";
import MemoryConsoleView, { formatMemoryUsageBar, memoryConsoleHint } from "./MemoryConsoleView";

function textContent(node: React.ReactNode): string {
  if (typeof node === "string" || typeof node === "number") return String(node);
  if (!node || typeof node !== "object") return "";
  if (Array.isArray(node)) return node.map(textContent).join("");
  if ("props" in node) return textContent((node as React.ReactElement<{ children?: React.ReactNode }>).props.children);
  return "";
}

function snapshot(): MemorySnapshot {
  return {
    memoryEnabled: true,
    memoryUserProfileEnabled: false,
    memoryNudgeInterval: 10,
    memoryDirectory: "/tmp/memory",
    memory: { target: "memory", entries: ["项目约定"], usageChars: 1100, charLimit: 2200, entryCount: 1, enabled: true },
    user: { target: "user", entries: ["偏好中文"], usageChars: 275, charLimit: 1375, entryCount: 1, enabled: false },
  };
}

test("formatMemoryUsageBar renders a stable text progress bar", () => {
  assert.equal(formatMemoryUsageBar(snapshot().memory), "█████░░░░░ 50%");
});

test("memory console hints change by mode", () => {
  assert.equal(memoryConsoleHint("actions"), "↑/↓ select | Enter apply/edit | R reset | Esc close");
  assert.equal(memoryConsoleHint("edit"), "Type digits | Enter save | Esc cancel");
  assert.equal(memoryConsoleHint("reset"), "↑/↓ select | Enter confirm | Esc cancel");
  assert.equal(memoryConsoleHint("view"), "Esc return");
});

test("memory console view renders settings summary and actions", () => {
  const view = MemoryConsoleView({
    snapshot: snapshot(),
    loading: false,
    cursor: 0,
    mode: "actions",
    editingField: null,
    draft: "",
    viewTarget: null,
    message: "",
    columns: 90,
  });
  const text = textContent(view);
  assert.match(text, /Memory Console/);
  assert.match(text, /Long-term memory: ON/);
  assert.match(text, /User profile:\s+OFF/);
  assert.match(text, /Toggle long-term memory/);
  assert.match(text, /Reset memory/);
});

test("memory console view renders edit and entry modes", () => {
  const editText = textContent(MemoryConsoleView({
    snapshot: snapshot(),
    loading: false,
    cursor: 0,
    mode: "edit",
    editingField: "memoryCharLimit",
    draft: "2200",
    viewTarget: null,
    message: "",
    columns: 90,
  }));
  assert.match(editText, /Edit Personal notes budget/);
  assert.match(editText, /Allowed range: 200 - 20000 chars/);

  const viewText = textContent(MemoryConsoleView({
    snapshot: snapshot(),
    loading: false,
    cursor: 0,
    mode: "view",
    editingField: null,
    draft: "",
    viewTarget: "user",
    message: "",
    columns: 90,
  }));
  assert.match(viewText, /User profile entries/);
  assert.match(viewText, /偏好中文/);
});
