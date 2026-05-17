import React from "react";
import { Box, Text } from "ink";
import type { UpdateCheckResult, UpdateJobStatus } from "../types.js";

interface UpdateViewProps {
	check: UpdateCheckResult | null;
	job: UpdateJobStatus | null;
	loading: boolean;
	applying: boolean;
}

interface SummaryInput {
	check: Partial<UpdateCheckResult> | null;
	job: Partial<UpdateJobStatus> | null;
}

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
		const notes = (check.releaseNotes || "").trim().split(/\r?\n/).filter(Boolean).slice(0, 4);
		for (const line of notes) {
			lines.push(`- ${line}`);
		}
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

export function formatUpdateHint(canApply: boolean, active: boolean): string {
	if (active) return "Updating... | Esc return";
	return canApply ? "C check again | U update | Esc return" : "C check again | Esc return";
}

export function UpdateView({ check, job, loading, applying }: UpdateViewProps): React.ReactElement {
	const active = applying || isUpdateActive(job);
	const canApply = Boolean(check?.canApply) && !active && !loading;
	const titleColor = check?.updateAvailable ? "#facc15" : "#67e8f9";
	const hint = formatUpdateHint(canApply, active);
	const lines = formatUpdateSummaryLines({ check, job });

	return (
		<Box flexDirection="column">
			<Text bold color={titleColor}>Update Center</Text>
			<Text color="#64748b">{loading ? "Checking for updates..." : hint}</Text>
			<Text> </Text>
			{lines.map((line, index) => (
				<Text key={`${index}-${line}`} color={line.startsWith("Error:") ? "#f87171" : "#cbd5e1"}>
					{line}
				</Text>
			))}
			{check?.updateAvailable && check.canApply && !active && (
				<>
					<Text> </Text>
					<Text color="#94a3b8">Press U to start the detached updater. The current TUI exits after the helper starts.</Text>
				</>
			)}
		</Box>
	);
}
