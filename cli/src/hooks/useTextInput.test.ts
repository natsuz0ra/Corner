import assert from "node:assert/strict";
import test from "node:test";
import type { Key } from "ink";
import { shouldInsertNewlineOnEnter } from "./useTextInput.js";

function key(overrides: Partial<Key> = {}): Key {
  return {
    upArrow: false,
    downArrow: false,
    leftArrow: false,
    rightArrow: false,
    pageDown: false,
    pageUp: false,
    home: false,
    end: false,
    return: true,
    escape: false,
    ctrl: false,
    shift: false,
    tab: false,
    backspace: false,
    delete: false,
    meta: false,
    ...overrides,
  } as Key;
}

test("multiline input without submit handler inserts newline on Enter", () => {
  const shouldInsert = shouldInsertNewlineOnEnter(true, false, key(), "\r");
  assert.equal(shouldInsert, true);
});

test("multiline input with submit handler inserts newline on Ctrl+Enter", () => {
  const shouldInsert = shouldInsertNewlineOnEnter(true, true, key({ ctrl: true }), "\r");
  assert.equal(shouldInsert, true);
});

test("multiline input with submit handler inserts newline on Ctrl+Enter raw LF", () => {
  const shouldInsert = shouldInsertNewlineOnEnter(true, true, key({ ctrl: true }), "\n");
  assert.equal(shouldInsert, true);
});

test("multiline input treats raw newline input as Ctrl+Enter fallback", () => {
  const shouldInsert = shouldInsertNewlineOnEnter(true, true, key(), "\n");
  assert.equal(shouldInsert, true);
});

test("multiline input with submit handler submits on plain Enter", () => {
  const shouldInsert = shouldInsertNewlineOnEnter(true, true, key(), "\r");
  assert.equal(shouldInsert, false);
});

test("single line input never inserts newline", () => {
  const shouldInsert = shouldInsertNewlineOnEnter(false, false, key({ ctrl: true }), "\n");
  assert.equal(shouldInsert, false);
});
