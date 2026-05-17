import React from "react";
import { Box, Text } from "ink";
import type { UpdateCheckResult, UpdateJobStatus } from "../types.js";
import { renderMarkdownLines } from "../utils/markdownRenderer.js";

interface UpdateViewProps {
	check: UpdateCheckResult | null;
	job: UpdateJobStatus | null;
	loading: boolean;
	applying: boolean;
	columns?: number;
}

interface SummaryInput {
	check: Partial<UpdateCheckResult> | null;
	job: Partial<UpdateJobStatus> | null;
}

const MAX_RELEASE_NOTE_LINES = 8;

export function isUpdateActive(job: Partial<UpdateJobStatus> | null): boolean {
	const phase = job?.phase;
	return phase === "checking" || phase === "downloading" || phase === "installing" || phase === "restarting";
}

export function formatUpdateSummaryLines({ check, job }: SummaryInput): string[] {
	const lines: string[] = [];
	if (check) {
		const current = check.current || "(unknown)";
		const latest = check.latest || "(unknown)";
		if (check.updateAvailable) {
			lines.push(`Version: ${current} -> ${latest}`);
		} else {
			lines.push(`Version: ${current} (latest)`);
		}
		if (check.releaseUrl) lines.push(`Release: ${check.releaseUrl}`);
		if (check.reason) lines.push(`Note: ${check.reason}`);
		if (check.manualHint && !check.canApply) lines.push(`Manual: ${check.manualHint}`);
	} else {
		lines.push("No update check loaded.");
	}
	if (job && job.phase && job.phase !== "idle") {
		lines.push(`Job: ${job.phase}${job.message ? ` · ${job.message}` : ""}`);
		if (job.error) lines.push(`Error: ${job.error}`);
	}
	return lines;
}

export function formatUpdateReleaseNotesLines(releaseNotes = "", columns = 80): string[] {
	const notes = releaseNotes.trim();
	if (!notes) return [];
	const rendered = renderMarkdownLines(notes, Math.max(20, columns), true);
	if (rendered.length <= MAX_RELEASE_NOTE_LINES) return rendered;
	return [...rendered.slice(0, MAX_RELEASE_NOTE_LINES), "..."];
}

export function formatUpdateHint(canApply: boolean, active: boolean): string {
	if (active) return "Updating... | Esc return";
	return canApply ? "C check again | U update | Esc return" : "C check again | Esc return";
}

export function UpdateView({ check, job, loading, applying, columns = 80 }: UpdateViewProps): React.ReactElement {
	const active = applying || isUpdateActive(job);
	const canApply = Boolean(check?.canApply) && !active && !loading;
	const titleColor = check?.updateAvailable ? "#facc15" : "#67e8f9";
	const hint = formatUpdateHint(canApply, active);
	const lines = formatUpdateSummaryLines({ check, job });
	const releaseNoteLines = formatUpdateReleaseNotesLines(check?.releaseNotes || "", columns);

	return (
		<Box flexDirection="column">
			<Text bold color={titleColor}>Update Center</Text>
			<Text> </Text>
			{lines.map((line, index) => (
				<Text key={`${index}-${line}`} color={line.startsWith("Error:") ? "#f87171" : "#cbd5e1"}>
					{line}
				</Text>
			))}
			{releaseNoteLines.length > 0 && (
				<>
					<Text> </Text>
					{releaseNoteLines.map((line, index) => (
						<Text key={`note-${index}-${line}`} color="#cbd5e1">
							{line}
						</Text>
					))}
				</>
			)}
			{check?.updateAvailable && check.canApply && !active && (
				<>
					<Text> </Text>
					<Text color="#94a3b8">Press U to start the detached updater. The current TUI exits after the helper starts.</Text>
				</>
			)}
			<Text> </Text>
			<Text color="#64748b">{loading ? "Checking for updates..." : hint}</Text>
		</Box>
	);
}
