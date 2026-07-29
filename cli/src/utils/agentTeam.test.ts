import assert from "node:assert/strict";
import test from "node:test";
import { stringWidth } from "./stringWidth";
import {
  formatAgentTeamRows,
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

test("formatAgentTeamRows fits tree member labels into narrow terminals", () => {
  const run = team({
    status: "partial_failed",
    finishedAt: "2026-07-29T00:00:05Z",
    members: [
      member("member-a", "2026-07-29T00:00:01Z", { status: "succeeded", finishedAt: "2026-07-29T00:00:03Z" }),
      member("member-b", "2026-07-29T00:00:02Z", { status: "failed", error: "boom", finishedAt: "2026-07-29T00:00:04Z" }),
    ],
  });
  const lines = formatAgentTeamRows(run, 28, false);

  assert.ok(lines.every((line) => stringWidth(line.text) <= 28));
  assert.match(lines.map((line) => line.text).join("\n"), /├─|└─/);
  assert.match(lines[0]?.text || "", /Partial failure/);
});

test("all Agent Team terminal statuses have visible labels", () => {
  assert.deepEqual(
    ["running", "succeeded", "partial_failed", "failed", "canceled", "interrupted"].map(getAgentTeamStatusLabel),
    ["Running", "Completed", "Partial failure", "Failed", "Canceled", "Interrupted"],
  );
});
