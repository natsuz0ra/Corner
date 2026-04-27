import assert from "node:assert/strict";
import test from "node:test";
import type { TimelineEntry } from "../types";
import {
	findStableTimelineFlushCount,
	formatTimelineEntriesForTranscript,
	splitLiveAssistantForTranscript,
} from "./transcript";

test("stable flush boundary keeps an executing child tool with its parent", () => {
	const entries: TimelineEntry[] = [
		{ kind: "user", content: "inspect repo" },
		{
			kind: "tool",
			content: "parent done",
			toolCallId: "parent",
			toolName: "run_subagent",
			command: "delegate",
			status: "completed",
			output: "parent done",
		},
		{
			kind: "tool",
			content: "",
			toolCallId: "child",
			parentToolCallId: "parent",
			toolName: "exec",
			command: "npm test",
			status: "executing",
		},
		{ kind: "assistant", content: "later stable text must not jump ahead" },
	];

	assert.equal(findStableTimelineFlushCount(entries), 1);
});

test("completed tool transcript formatting follows ctrl-o expansion state", () => {
	const output = Array.from(
		{ length: 10 },
		(_, index) => `line${index + 1}`,
	).join("\n");
	const entries: TimelineEntry[] = [
		{
			kind: "tool",
			content: "",
			toolCallId: "tool-1",
			toolName: "exec",
			command: "run",
			status: "completed",
			output,
		},
	];

	const collapsed = formatTimelineEntriesForTranscript(entries, {
		maxWidth: 100,
		compact: true,
		toolOutputExpanded: false,
	});
	const expanded = formatTimelineEntriesForTranscript(entries, {
		maxWidth: 100,
		compact: true,
		toolOutputExpanded: true,
	});

	assert.match(collapsed, /ctrl\+o to expand/);
	assert.doesNotMatch(collapsed, /line10/);
	assert.match(expanded, /line10/);
	assert.match(expanded, /ctrl\+o to collapse/);
});

test("streaming assistant split keeps only a short newline-bounded active tail", () => {
	const liveAssistant = `${Array.from({ length: 20 }, (_, index) => `line ${index + 1}`).join("\n")}\n`;

	const split = splitLiveAssistantForTranscript(liveAssistant, 0, 5);

	assert.ok(split.flushText.startsWith("line 1\n"));
	assert.equal(split.nextFlushedLength, split.flushText.length);
	assert.equal(split.activeContinuation, true);
	assert.equal(
		split.activeText,
		"line 16\nline 17\nline 18\nline 19\nline 20\n",
	);
});

test("transcript thinking numbering continues after flushed entries", () => {
	const entries: TimelineEntry[] = [
		{
			kind: "thinking",
			content: "done",
			thinkingDone: true,
			thinkingDurationMs: 1000,
		},
	];

	const output = formatTimelineEntriesForTranscript(entries, {
		maxWidth: 100,
		compact: true,
		toolOutputExpanded: false,
		thinkingStartIndex: 2,
	});

	assert.match(output, /\[3\] Thought for 1\.0s/);
});
