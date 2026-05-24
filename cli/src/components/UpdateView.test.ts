import assert from "node:assert/strict";
import test from "node:test";
import { stripAnsi } from "../utils/terminal";
import { formatUpdateHeaderLine, formatUpdateHint, formatUpdateProgressLine, formatUpdateReleaseNotesLines, formatUpdateSummaryLines, getInitialUpdateJobForView } from "./UpdateView";

test("formatUpdateHeaderLine shows available update target and apply mode", () => {
	const line = formatUpdateHeaderLine({
		current: "v1.26.1",
		latest: "v1.26.2",
		updateAvailable: true,
		canApply: true,
	});

	assert.equal(line, "Update  v1.26.1 -> v1.26.2  auto");
});

test("formatUpdateSummaryLines shows compact update status and release url", () => {
	const lines = formatUpdateSummaryLines({
		check: {
			current: "v1.26.1",
			latest: "v1.26.2",
			updateAvailable: true,
			canApply: true,
			releaseUrl: "https://example.test/release",
			releaseNotes: "更新说明",
			reason: "",
			manualHint: "",
		},
		job: null,
		loading: false,
		confirming: false,
	});

	assert.ok(lines.some((line) => line.includes("Status    waiting for confirmation")));
	assert.ok(lines.some((line) => line.includes("Release   https://example.test/release")));
});

test("formatUpdateHint switches update shortcut by availability", () => {
	assert.equal(formatUpdateHint(true, false, false), "C recheck · U update · Esc close");
	assert.equal(formatUpdateHint(true, false, true), "Y confirm · N cancel · Esc close");
	assert.equal(formatUpdateHint(false, false), "C recheck · Esc close");
	assert.equal(formatUpdateHint(true, true), "Updating... · Esc close");
});

test("formatUpdateProgressLine renders known and unknown download size", () => {
	assert.match(formatUpdateProgressLine({ downloadedBytes: 512, totalBytes: 1024, progressPercent: 50 }), /50%/);
	assert.match(formatUpdateProgressLine({ downloadedBytes: 512, totalBytes: 0, progressPercent: 0 }), /512 B downloaded/);
});

test("formatUpdateSummaryLines leaves markdown release notes out of metadata lines", () => {
	const lines = formatUpdateSummaryLines({
		check: {
			current: "dev",
			latest: "v1.26.0",
			updateAvailable: true,
			canApply: false,
			releaseUrl: "https://example.test/release",
			releaseNotes: "## Highlights\n- **A** item",
			reason: "",
			manualHint: "slimebot update --version v1.26.0",
		},
		job: null,
		loading: false,
		confirming: false,
	});

	assert.equal(lines.some((line) => line.includes("## Highlights")), false);
	assert.equal(lines.some((line) => line.includes("- **A** item")), false);
	assert.ok(lines.some((line) => line.includes("Manual    slimebot update --version v1.26.0")));
});

test("formatUpdateReleaseNotesLines keeps Notes title separate from content", () => {
	const lines = formatUpdateReleaseNotesLines("## Highlights\n- **A** item", 80).map(stripAnsi);

	assert.equal(lines[0], "Notes");
	assert.notEqual(lines[1], undefined);
	assert.ok(lines[1].includes("Highlights"));
	assert.equal(lines[0].includes("Highlights"), false);
});

test("formatUpdateReleaseNotesLines renders markdown release notes for terminal output", () => {
	const lines = formatUpdateReleaseNotesLines("## Highlights\n- **A** item", 80).map(stripAnsi);

	assert.ok(lines.some((line) => line.includes("Highlights")));
	assert.ok(lines.some((line) => line.includes("- A item")));
	assert.equal(lines.some((line) => line.includes("## Highlights")), false);
	assert.equal(lines.some((line) => line.includes("**A**")), false);
});

test("formatUpdateReleaseNotesLines keeps complete release notes", () => {
	const notes = Array.from({ length: 12 }, (_, index) => `- Item ${index + 1}`).join("\n");
	const lines = formatUpdateReleaseNotesLines(notes, 80).map(stripAnsi);

	assert.ok(lines.some((line) => line.includes("Item 1")));
	assert.ok(lines.some((line) => line.includes("Item 12")));
	assert.equal(lines.some((line) => line === "..."), false);
});

test("getInitialUpdateJobForView ignores historical terminal update jobs", () => {
	assert.equal(getInitialUpdateJobForView({ phase: "succeeded", message: "Update installed successfully" }), null);
	assert.equal(getInitialUpdateJobForView({ phase: "failed", message: "Update failed" }), null);
	assert.equal(getInitialUpdateJobForView({ phase: "idle" }), null);

	const downloading = { phase: "downloading" as const, downloadedBytes: 512, totalBytes: 1024 };
	assert.equal(getInitialUpdateJobForView(downloading), downloading);
});
