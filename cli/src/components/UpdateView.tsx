import React from "react";
import { Box, Text } from "ink";
import type { UpdateCheckResult, UpdateJobStatus } from "../types.js";
import { renderMarkdownLines } from "../utils/markdownRenderer.js";

interface UpdateViewProps {
	check: UpdateCheckResult | null;
	job: UpdateJobStatus | null;
	loading: boolean;
	applying: boolean;
	confirming?: boolean;
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
		const progress = formatUpdateProgressLine(job);
		if (progress) lines.push(progress);
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

export function formatUpdateProgressLine(job: Partial<UpdateJobStatus> | null): string {
	if (!job || (job.phase && job.phase !== "downloading")) return "";
	const downloaded = Math.max(0, Math.trunc(job.downloadedBytes || 0));
	const total = Math.max(0, Math.trunc(job.totalBytes || 0));
	if (total > 0) {
		const percent = Math.min(100, Math.max(0, Math.trunc(job.progressPercent ?? downloaded * 100 / total)));
		return `Download: ${renderProgressBar(percent, 18)} ${percent}% (${formatBytes(downloaded)}/${formatBytes(total)})`;
	}
	if (downloaded > 0) return `Download: ${formatBytes(downloaded)} downloaded`;
	return "Download: in progress";
}

export function formatUpdateHint(canApply: boolean, active: boolean, confirming = false): string {
	if (active) return "Updating... | Esc return";
	if (confirming) return "Y confirm | N cancel | Esc return";
	return canApply ? "C check again | U update | Esc return" : "C check again | Esc return";
}

export function UpdateView({ check, job, loading, applying, confirming = false, columns = 80 }: UpdateViewProps): React.ReactElement {
	const active = applying || isUpdateActive(job);
	const canApply = Boolean(check?.canApply) && !active && !loading;
	const titleColor = check?.updateAvailable ? "#facc15" : "#67e8f9";
	const hint = formatUpdateHint(canApply, active, confirming);
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
			{confirming && (
				<>
					<Text> </Text>
					<Text color="#facc15">Confirm update? SlimeBot may restart while installing.</Text>
				</>
			)}
			{check?.updateAvailable && check.canApply && !active && !confirming && (
				<>
					<Text> </Text>
					<Text color="#94a3b8">Press U to review and confirm the update.</Text>
				</>
			)}
			{active && job?.phase && (job.phase === "installing" || job.phase === "restarting") && (
				<Text color="#94a3b8">The service may restart and disconnect this TUI.</Text>
			)}
			<Text> </Text>
			<Text color="#64748b">{loading ? "Checking for updates..." : hint}</Text>
		</Box>
	);
}

function renderProgressBar(percent: number, width: number): string {
	const filled = Math.max(0, Math.min(width, Math.trunc(percent * width / 100)));
	return `[${"=".repeat(filled)}${"-".repeat(width - filled)}]`;
}

function formatBytes(value: number): string {
	if (value < 1024) return `${value} B`;
	if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KiB`;
	return `${(value / 1024 / 1024).toFixed(1)} MiB`;
}
