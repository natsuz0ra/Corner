import assert from "node:assert/strict";
import test from "node:test";
import type { Key } from "ink";
import { createInitialState } from "../reducer.js";
import { shouldLetQuestionAnswerViewHandleInput } from "./useCliKeyboard.js";

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
    return: false,
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

test("custom input cursor lets question view handle printable input", () => {
  const state = {
    ...createInitialState("http://127.0.0.1:8080", "token", "/tmp", "1.0.0"),
    view: "question-answer" as const,
    qaStep: "questions" as const,
    qaCurrentIndex: 0,
    qaCursor: 1,
    qaQuestions: [{ id: "q1", question: "Q", options: ["A"] }],
    qaAnswers: [{ questionId: "q1", selectedOption: -1, customAnswer: "" }],
  };

  assert.equal(shouldLetQuestionAnswerViewHandleInput(state, "h", key()), true);
  assert.equal(shouldLetQuestionAnswerViewHandleInput(state, "", key({ backspace: true })), true);
});

test("navigation keys are still owned by global keyboard handler", () => {
  const state = {
    ...createInitialState("http://127.0.0.1:8080", "token", "/tmp", "1.0.0"),
    view: "question-answer" as const,
    qaStep: "questions" as const,
    qaCurrentIndex: 0,
    qaCursor: 1,
    qaQuestions: [{ id: "q1", question: "Q", options: ["A"] }],
    qaAnswers: [{ questionId: "q1", selectedOption: -1, customAnswer: "" }],
  };

  assert.equal(shouldLetQuestionAnswerViewHandleInput(state, "", key({ upArrow: true })), false);
  assert.equal(shouldLetQuestionAnswerViewHandleInput(state, "", key({ tab: true })), false);
  assert.equal(shouldLetQuestionAnswerViewHandleInput(state, "", key({ escape: true })), false);
});
