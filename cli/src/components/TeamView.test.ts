import assert from "node:assert/strict";
import test from "node:test";
import type React from "react";
import type { AgentTeamRun, TimelineEntry } from "../types";

function textContent(node: React.ReactNode): string {
  if (typeof node === "string" || typeof node === "number") return String(node);
  if (!node || typeof node !== "object") return "";
  if (Array.isArray(node)) return node.map(textContent).join("");
  if ("props" in node) return textContent((node as React.ReactElement<{ children?: React.ReactNode }>).props.children);
  return "";
}

async function loadTeamView() {
  const module = await import("./TeamView.js").catch(() => null);
  assert.ok(module, "TeamView component should exist");
  return module.default;
}

function teamRun(): AgentTeamRun {
  return {
    id: "team-1",
    sessionId: "session-1",
    requestId: "request-1",
    status: "partial_failed",
    maxMembers: 8,
    maxParallel: 4,
    startedAt: "2026-07-29T00:00:00Z",
    finishedAt: "2026-07-29T00:00:05Z",
    members: [
      {
        id: "member-1",
        teamRunId: "team-1",
        toolCallId: "tool-1",
        title: "Research",
        task: "Inspect runtime",
        status: "succeeded",
        answer: "Runtime is healthy.",
        startedAt: "2026-07-29T00:00:00Z",
        finishedAt: "2026-07-29T00:00:02Z",
      },
      {
        id: "member-2",
        teamRunId: "team-1",
        toolCallId: "tool-2",
        title: "Frontend",
        task: "Inspect frontend rendering",
        status: "failed",
        error: "Layout check failed.",
        startedAt: "2026-07-29T00:00:01Z",
        finishedAt: "2026-07-29T00:00:04Z",
      },
    ],
  };
}

test("TeamView renders an empty state", async () => {
  const TeamView = await loadTeamView();
  const text = textContent(TeamView({
    runs: [],
    entries: [],
    teamCursor: 0,
    memberCursor: 0,
    columns: 80,
  }));

  assert.match(text, /No Agent Team runs in this session\./);
});

test("TeamView renders selected member details without nested execution data", async () => {
  const TeamView = await loadTeamView();
  const entries: TimelineEntry[] = [
    { kind: "thinking", content: "private chain of thought" },
    {
      kind: "tool",
      content: "",
      toolCallId: "child-tool",
      parentToolCallId: "tool-2",
      toolName: "exec",
      params: { command: "secret params" },
      output: "hidden tool output",
      status: "completed",
    },
  ];
  const text = textContent(TeamView({
    runs: [teamRun()],
    entries,
    teamCursor: 0,
    memberCursor: 1,
    columns: 80,
    now: Date.parse("2026-07-29T00:00:05Z"),
  }));

  assert.match(text, /Agent Team 1\/1/);
  assert.match(text, /Frontend/);
  assert.match(text, /Task: Inspect frontend rendering/);
  assert.match(text, /Error: Layout check failed\./);
  assert.match(text, /Tools: 1 tool/);
  assert.doesNotMatch(text, /private chain of thought|secret params|hidden tool output|Thinking|Params/);
});
