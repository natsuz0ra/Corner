import assert from "node:assert/strict";
import test from "node:test";
import { stringWidth } from "./stringWidth";
import {
  clampAgentTeamCursor,
  formatAgentTeamMemberDetail,
  formatAgentTeamRows,
  getAgentTeamRuns,
  getAgentTeamStatusLabel,
  upsertAgentTeamMember,
  upsertAgentTeamRun,
} from "./agentTeam";
import type { AgentTeamMemberRun, AgentTeamRun, TimelineEntry } from "../types";

function team(overrides: Partial<AgentTeamRun> = {}): AgentTeamRun {
  return {
    id: "team-1", sessionId: "session-1", requestId: "request-1", status: "running",
    maxMembers: 8, maxParallel: 4, startedAt: "2026-07-29T00:00:00Z",
    createdAt: "2026-07-29T00:00:00Z", updatedAt: "2026-07-29T00:00:00Z", members: [],
    ...overrides,
  };
}

function member(id: string, createdAt: string, overrides: Partial<AgentTeamMemberRun> = {}): AgentTeamMemberRun {
  return {
    id, teamRunId: "team-1", toolCallId: `tool-${id}`, title: id,
    task: `Investigate the implementation details for ${id}`, status: "running",
    createdAt, updatedAt: createdAt, ...overrides,
  };
}

test("Agent Team reducer helpers merge out-of-order members in stable server order", () => {
  let entries: TimelineEntry[] = [];
  entries = upsertAgentTeamMember(entries, member("member-b", "2026-07-29T00:00:02Z"));
  entries = upsertAgentTeamMember(entries, member("member-a", "2026-07-29T00:00:01Z"));
  entries = upsertAgentTeamRun(entries, team({ requestId: "request-from-server" }));

  const entry = entries.find((item) => item.kind === "team");
  assert.equal(entry?.teamRun?.requestId, "request-from-server");
  assert.deepEqual(entry?.teamRun?.members.map((item) => item.id), ["member-a", "member-b"]);
});

test("formatAgentTeamRows keeps Team output to two lines without task bodies", () => {
  const run = team({
    status: "partial_failed",
    finishedAt: "2026-07-29T00:00:05Z",
    members: [
      member("member-a", "2026-07-29T00:00:01Z", { status: "succeeded", finishedAt: "2026-07-29T00:00:03Z" }),
      member("member-b", "2026-07-29T00:00:02Z", { status: "failed", error: "boom", finishedAt: "2026-07-29T00:00:04Z" }),
    ],
  });
  const lines = formatAgentTeamRows(run, 80);

  assert.equal(lines.length, 2);
  assert.doesNotMatch(lines.map((line) => line.text).join("\n"), /Investigate the implementation/);
  assert.match(lines[0]?.text || "", /Partial failure/);
});

test("formatAgentTeamRows preserves failures and +N in narrow terminals", () => {
  const run = team({
    status: "partial_failed",
    members: [
      member("done-a", "2026-07-29T00:00:01Z", { status: "succeeded" }),
      member("running-a", "2026-07-29T00:00:02Z"),
      member("failed-member", "2026-07-29T00:00:03Z", { status: "failed" }),
      member("done-b", "2026-07-29T00:00:04Z", { status: "succeeded" }),
      member("done-c", "2026-07-29T00:00:05Z", { status: "succeeded" }),
    ],
  });
  const lines = formatAgentTeamRows(run, 40);

  assert.equal(lines.length, 2);
  assert.match(lines[1]!.text, /failed-member/);
  assert.match(lines[1]!.text, /\+\d/);
  assert.ok(lines.every((line) => stringWidth(line.text) <= 40));
});

test("formatAgentTeamMemberDetail shows task result error duration and tool count only", () => {
  const target = member("member-a", "2026-07-29T00:00:01Z", {
    status: "succeeded",
    startedAt: "2026-07-29T00:00:01Z",
    finishedAt: "2026-07-29T00:00:04Z",
    answer: "Finished the requested implementation.\n\nHidden second paragraph.",
  });
  const childTools: TimelineEntry[] = [
    { kind: "tool", content: "tool output", toolCallId: "child-a", parentToolCallId: target.toolCallId, params: { secret: true } },
    { kind: "tool", content: "tool output", toolCallId: "child-b", parentToolCallId: target.toolCallId },
    { kind: "thinking", content: "Thinking should stay hidden", parentToolCallId: target.toolCallId },
  ];
  const lines = formatAgentTeamMemberDetail(target, childTools, 40);
  const text = lines.map((line) => line.text).join("\n");

  assert.match(text, /Task/);
  assert.match(text, /Result/);
  assert.match(text, /2 tools/);
  assert.doesNotMatch(text, /Thinking|Params|tool output|Hidden second paragraph/);
  assert.ok(lines.every((line) => stringWidth(line.text) <= 40));
});

test("Agent Team detail helpers select timeline runs and clamp cursors", () => {
  const first = team({ id: "team-a" });
  const second = team({ id: "team-b" });
  const entries: TimelineEntry[] = [
    { kind: "team", content: "", teamRun: first },
    { kind: "assistant", content: "done" },
    { kind: "team", content: "", teamRun: second },
  ];

  assert.deepEqual(getAgentTeamRuns(entries).map((run) => run.id), ["team-a", "team-b"]);
  assert.equal(clampAgentTeamCursor(-2, 2), 0);
  assert.equal(clampAgentTeamCursor(5, 2), 1);
  assert.equal(clampAgentTeamCursor(5, 0), 0);
});

test("all Agent Team terminal statuses have visible labels", () => {
  assert.deepEqual(
    ["running", "succeeded", "partial_failed", "failed", "canceled", "interrupted"].map(getAgentTeamStatusLabel),
    ["Running", "Completed", "Partial failure", "Failed", "Canceled", "Interrupted"],
  );
});
