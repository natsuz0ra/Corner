import { renderToString } from "ink";
import React from "react";
import { Timeline } from "../components/Timeline.js";
import type { TimelineEntry, ToolCallStatus } from "../types.js";

export const LIVE_ASSISTANT_TAIL_LINES = 12;

export interface TranscriptRenderOptions {
	maxWidth: number;
	compact: boolean;
	toolOutputExpanded: boolean;
	thinkingStartIndex?: number;
	firstAssistantContinuation?: boolean;
}

export interface LiveAssistantSplit {
	flushText: string;
	nextFlushedLength: number;
	activeText: string;
	activeContinuation: boolean;
	flushContinuation: boolean;
}

const STABLE_TOOL_STATUSES = new Set<ToolCallStatus>([
	"completed",
	"error",
	"rejected",
]);

function isStableTimelineEntry(entry: TimelineEntry): boolean {
	if (entry.kind === "thinking") {
		return entry.thinkingDone === true;
	}
	if (entry.kind === "tool") {
		return STABLE_TOOL_STATUSES.has(
			(entry.status || "completed") as ToolCallStatus,
		);
	}
	return true;
}

export function countThinkingEntries(
	entries: readonly TimelineEntry[],
): number {
	return entries.reduce(
		(count, entry) => count + (entry.kind === "thinking" ? 1 : 0),
		0,
	);
}

export function findStableTimelineFlushCount(
	entries: readonly TimelineEntry[],
): number {
	let boundary = entries.length;
	for (let index = 0; index < entries.length; index += 1) {
		if (!isStableTimelineEntry(entries[index])) {
			boundary = index;
			break;
		}
	}

	for (let index = boundary; index < entries.length; index += 1) {
		const entry = entries[index];
		if (
			entry.kind !== "tool" ||
			isStableTimelineEntry(entry) ||
			!entry.parentToolCallId
		) {
			continue;
		}
		const parentIndex = entries.findIndex(
			(candidate) =>
				candidate.kind === "tool" &&
				candidate.toolCallId === entry.parentToolCallId,
		);
		if (parentIndex >= 0 && parentIndex < boundary) {
			boundary = parentIndex;
		}
	}

	return Math.max(0, boundary);
}

function findNewlineTailBoundary(content: string, tailLines: number): number {
	const newlineIndexes: number[] = [];
	for (let index = 0; index < content.length; index += 1) {
		if (content[index] === "\n") {
			newlineIndexes.push(index);
		}
	}

	const lineCount = content.endsWith("\n")
		? newlineIndexes.length
		: newlineIndexes.length + (content.length > 0 ? 1 : 0);
	if (lineCount <= tailLines) {
		return 0;
	}

	const boundaryLine = lineCount - tailLines;
	const boundaryNewlineIndex = newlineIndexes[boundaryLine - 1];
	return boundaryNewlineIndex === undefined ? 0 : boundaryNewlineIndex + 1;
}

export function splitLiveAssistantForTranscript(
	liveAssistant: string,
	flushedLength: number,
	tailLines = LIVE_ASSISTANT_TAIL_LINES,
): LiveAssistantSplit {
	const safeFlushedLength = Math.max(
		0,
		Math.min(flushedLength, liveAssistant.length),
	);
	const safeTailLines = Math.max(1, Math.floor(tailLines));
	const boundary = Math.max(
		safeFlushedLength,
		findNewlineTailBoundary(liveAssistant, safeTailLines),
	);
	return {
		flushText: liveAssistant.slice(safeFlushedLength, boundary),
		nextFlushedLength: boundary,
		activeText: liveAssistant.slice(boundary),
		activeContinuation: boundary > 0,
		flushContinuation: safeFlushedLength > 0,
	};
}

export function formatTimelineEntriesForTranscript(
	entries: readonly TimelineEntry[],
	options: TranscriptRenderOptions,
): string {
	if (entries.length === 0) {
		return "";
	}

	return renderToString(
		React.createElement(Timeline, {
			entries: [...entries],
			blinkOn: true,
			streaming: false,
			assistantWaiting: false,
			liveAssistant: "",
			maxWidth: options.maxWidth,
			compact: options.compact,
			toolOutputExpanded: options.toolOutputExpanded,
			thinkingEntryIndex: 0,
			planGenerating: false,
			planReceived: false,
			thinkingStartIndex: options.thinkingStartIndex ?? 0,
			firstAssistantContinuation: options.firstAssistantContinuation ?? false,
		}),
		{ columns: options.maxWidth },
	);
}

export function formatLiveAssistantForTranscript(
	content: string,
	options: TranscriptRenderOptions & { continuation?: boolean },
): string {
	if (content.length === 0) {
		return "";
	}
	return formatTimelineEntriesForTranscript([{ kind: "assistant", content }], {
		...options,
		firstAssistantContinuation: options.continuation ?? false,
	});
}
