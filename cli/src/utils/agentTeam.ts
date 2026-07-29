import type { AgentTeamMemberRun, AgentTeamRun, TimelineEntry } from "../types.js";
import { wrapText } from "./format.js";

export type AgentTeamDisplayLine = {
  text: string;
  color: "gray" | "green" | "yellow" | "red" | "cyan";
  active?: boolean;
};

function compareTimeAndId(leftAt: string | undefined, leftId: string, rightAt: string | undefined, rightId: string): number {
  const left = Date.parse(leftAt || "");
  const right = Date.parse(rightAt || "");
  return (Number.isFinite(left) ? left : 0) - (Number.isFinite(right) ? right : 0) || leftId.localeCompare(rightId);
}

function sortMembers(members: AgentTeamMemberRun[]): AgentTeamMemberRun[] {
  return [...members].sort((left, right) => compareTimeAndId(left.createdAt || left.startedAt, left.id, right.createdAt || right.startedAt, right.id));
}

function mergeText(previous: string | undefined, next: string | undefined): string | undefined {
  return next === undefined || next === "" ? previous : next;
}

function placeholderRun(teamRunId: string, member?: AgentTeamMemberRun): AgentTeamRun {
  const startedAt = member?.createdAt || member?.startedAt || "";
  return {
    id: teamRunId,
    sessionId: "",
    requestId: "",
    status: "running",
    maxMembers: 8,
    maxParallel: 4,
    startedAt,
    createdAt: startedAt,
    updatedAt: member?.updatedAt || startedAt,
    members: [],
  };
}

function teamEntryIndex(entries: TimelineEntry[], teamRunId: string): number {
  return entries.findIndex((entry) => entry.kind === "team" && entry.teamRun?.id === teamRunId);
}

export function upsertAgentTeamRun(
  entries: TimelineEntry[],
  incoming: Omit<AgentTeamRun, "members"> & { members?: AgentTeamMemberRun[] },
): TimelineEntry[] {
  const next = [...entries];
  const index = teamEntryIndex(next, incoming.id);
  if (index < 0) {
    next.push({ kind: "team", content: "", teamRun: { ...incoming, members: sortMembers(incoming.members || []) } });
    return next;
  }
  const previous = next[index]?.teamRun || placeholderRun(incoming.id);
  let members = previous.members;
  if (incoming.members) {
    for (const member of incoming.members) {
      const memberIndex = members.findIndex((item) => item.id === member.id);
      members = memberIndex < 0
        ? [...members, member]
        : members.map((item, itemIndex) => itemIndex === memberIndex ? { ...item, ...member } : item);
    }
  }
  next[index] = {
    kind: "team",
    content: "",
    teamRun: {
      ...previous,
      ...incoming,
      sessionId: mergeText(previous.sessionId, incoming.sessionId) || "",
      requestId: mergeText(previous.requestId, incoming.requestId) || "",
      createdAt: incoming.createdAt || previous.createdAt || incoming.startedAt,
      updatedAt: incoming.updatedAt || incoming.finishedAt || previous.updatedAt,
      members: sortMembers(members),
    },
  };
  return next;
}

export function upsertAgentTeamMember(entries: TimelineEntry[], incoming: AgentTeamMemberRun): TimelineEntry[] {
  let next = [...entries];
  let index = teamEntryIndex(next, incoming.teamRunId);
  if (index < 0) {
    next = upsertAgentTeamRun(next, placeholderRun(incoming.teamRunId, incoming));
    index = teamEntryIndex(next, incoming.teamRunId);
  }
  const run = next[index]?.teamRun || placeholderRun(incoming.teamRunId, incoming);
  const memberIndex = run.members.findIndex((member) => member.id === incoming.id);
  const members = [...run.members];
  if (memberIndex < 0) {
    members.push(incoming);
  } else {
    const previous = members[memberIndex]!;
    members[memberIndex] = {
      ...previous,
      ...incoming,
      title: mergeText(previous.title, incoming.title) || "",
      task: mergeText(previous.task, incoming.task) || "",
      subagentRunId: mergeText(previous.subagentRunId, incoming.subagentRunId),
      createdAt: incoming.createdAt || previous.createdAt || incoming.startedAt,
      updatedAt: incoming.updatedAt || incoming.finishedAt || previous.updatedAt,
    };
  }
  next[index] = { kind: "team", content: "", teamRun: { ...run, members: sortMembers(members) } };
  return next;
}

export function getAgentTeamStatusLabel(status: string): string {
  return ({
    running: "Running",
    succeeded: "Completed",
    partial_failed: "Partial failure",
    failed: "Failed",
    canceled: "Canceled",
    interrupted: "Interrupted",
  } as Record<string, string>)[status] || status;
}

function memberStatusLabel(status: string): string {
  return ({ queued: "Queued", running: "Running", succeeded: "Completed", failed: "Failed", canceled: "Canceled", interrupted: "Interrupted" } as Record<string, string>)[status] || status;
}

function statusSymbol(status: string): string {
  return status === "running" ? "●" : status === "queued" ? "○" : status === "succeeded" ? "✓" : status === "failed" || status === "partial_failed" ? "✗" : "−";
}

function statusColor(status: string): AgentTeamDisplayLine["color"] {
  if (status === "succeeded") return "green";
  if (status === "running" || status === "queued") return "yellow";
  if (status === "failed" || status === "partial_failed") return "red";
  return "gray";
}

function durationMs(startedAt: string | undefined, finishedAt: string | undefined, now: number): number {
  const start = Date.parse(startedAt || "");
  if (!Number.isFinite(start)) return 0;
  const finish = Date.parse(finishedAt || "");
  return Math.max(0, (Number.isFinite(finish) ? finish : now) - start);
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${Math.round(ms)}ms`;
  if (ms < 60_000) return `${(ms / 1000).toFixed(1).replace(/\.0$/, "")}s`;
  const minutes = Math.floor(ms / 60_000);
  const seconds = Math.floor((ms % 60_000) / 1000);
  return seconds ? `${minutes}m ${seconds}s` : `${minutes}m`;
}

function wrapTreeLine(prefix: string, body: string, continuation: string, maxWidth: number): string[] {
  const available = Math.max(1, maxWidth - prefix.length);
  const wrapped = wrapText(body, available).split("\n");
  return wrapped.map((line, index) => `${index === 0 ? prefix : continuation}${line}`);
}

export function formatAgentTeamRows(run: AgentTeamRun, maxWidth: number, expanded = false, now = Date.now()): AgentTeamDisplayLine[] {
  const width = Math.max(10, maxWidth);
  const completed = run.members.filter((member) => member.status !== "queued" && member.status !== "running").length;
  const status = getAgentTeamStatusLabel(run.status);
  const duration = formatDuration(durationMs(run.startedAt, run.finishedAt, now));
  const summary = width < 36
    ? `${statusSymbol(run.status)} ${status} · ${completed}/${run.members.length}`
    : `${statusSymbol(run.status)} Agent Team · ${status} · ${completed}/${run.members.length} · ${duration}`;
  const lines: AgentTeamDisplayLine[] = wrapText(summary, width).split("\n").map((text) => ({
    text,
    color: statusColor(run.status),
    active: run.status === "running",
  }));

  run.members.forEach((member, index) => {
    const last = index === run.members.length - 1;
    const branch = last ? "└─ " : "├─ ";
    const continuation = last ? "   " : "│  ";
    const body = `${statusSymbol(member.status)} ${member.title || "Team member"} · ${memberStatusLabel(member.status)} · ${formatDuration(durationMs(member.startedAt || member.createdAt, member.finishedAt, now))}${member.task ? ` · ${member.task}` : ""}`;
    for (const text of wrapTreeLine(branch, body, continuation, width)) {
      lines.push({ text, color: statusColor(member.status), active: member.status === "running" });
    }
    if (expanded && (member.error || member.answer)) {
      const detailPrefix = last ? "   " : "│  ";
      for (const text of wrapTreeLine(`${detailPrefix}  `, member.error || member.answer || "", `${detailPrefix}  `, width)) {
        lines.push({ text, color: member.error ? "red" : "gray" });
      }
    }
  });
  return lines;
}
