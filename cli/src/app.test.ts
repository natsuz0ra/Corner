import assert from "node:assert/strict";
import test from "node:test";
import { mapHistoryMessages } from "./app";
import type { Message, ToolCallHistoryItem } from "./types";

test("mapHistoryMessages inserts tool calls after assistant messages in timeline order", () => {
  const messages: Message[] = [
    {
      id: "u1",
      sessionId: "s1",
      role: "user",
      content: "hello",
      seq: 1,
      createdAt: "2026-01-01T00:00:00Z",
    },
    {
      id: "a1",
      sessionId: "s1",
      role: "assistant",
      content: "running tool",
      seq: 2,
      createdAt: "2026-01-01T00:00:01Z",
    },
    {
      id: "a-stop",
      sessionId: "s1",
      role: "assistant",
      content: "placeholder",
      seq: 3,
      isStopPlaceholder: true,
      createdAt: "2026-01-01T00:00:02Z",
    },
  ];

  const toolCallsByMsgId: Record<string, ToolCallHistoryItem[]> = {
    a1: [
      {
        toolCallId: "tc1",
        toolName: "web_search",
        command: "search",
        params: { q: "hello" },
        status: "completed",
        requiresApproval: false,
        output: "result",
        startedAt: "2026-01-01T00:00:01Z",
        finishedAt: "2026-01-01T00:00:02Z",
      },
    ],
  };

  const entries = mapHistoryMessages(messages, toolCallsByMsgId);

  assert.deepEqual(entries, [
    { kind: "user", content: "hello" },
    {
      kind: "tool",
      content: "result",
      toolCallId: "tc1",
      toolName: "web_search",
      command: "search",
      params: { q: "hello" },
      status: "completed",
      output: "result",
      error: undefined,
    },
    { kind: "assistant", content: "running tool" },
  ]);
});

test("mapHistoryMessages preserves parentToolCallId for nested tool calls", () => {
  const messages: Message[] = [
    {
      id: "a1",
      sessionId: "s1",
      role: "assistant",
      content: "done",
      seq: 1,
      createdAt: "2026-01-01T00:00:01Z",
    },
  ];
  const toolCallsByMsgId: Record<string, ToolCallHistoryItem[]> = {
    a1: [
      {
        toolCallId: "parent",
        toolName: "run_subagent",
        command: "delegate",
        params: { task: "x" },
        status: "completed",
        requiresApproval: false,
        output: "ok",
        startedAt: "2026-01-01T00:00:00Z",
      },
      {
        toolCallId: "child",
        toolName: "web_search",
        command: "search",
        params: { q: "y" },
        status: "completed",
        requiresApproval: false,
        parentToolCallId: "parent",
        subagentRunId: "run-1",
        output: "hits",
        startedAt: "2026-01-01T00:00:01Z",
      },
    ],
  };
  const entries = mapHistoryMessages(messages, toolCallsByMsgId);
  const child = entries.find((e) => e.kind === "tool" && e.toolCallId === "child");
  assert.ok(child && child.kind === "tool");
  assert.equal(child.parentToolCallId, "parent");
  assert.equal(child.subagentRunId, "run-1");
});

