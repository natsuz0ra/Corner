import assert from "node:assert/strict";
import test from "node:test";
import { stripAnsi } from "../utils/terminal";
import { formatUpdateHint, formatUpdateProgressLine, formatUpdateReleaseNotesLines, formatUpdateSummaryLines } from "./UpdateView";

test("formatUpdateSummaryLines shows available update and release url", () => {
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
	});

	assert.ok(lines.some((line) => line.includes("v1.26.1 -> v1.26.2")));
	assert.ok(lines.some((line) => line.includes("https://example.test/release")));
});

test("formatUpdateHint switches update shortcut by availability", () => {
	assert.equal(formatUpdateHint(true, false, false), "C check again | U update | Esc return");
	assert.equal(formatUpdateHint(true, false, true), "Y confirm | N cancel | Esc return");
	assert.equal(formatUpdateHint(false, false), "C check again | Esc return");
	assert.equal(formatUpdateHint(true, true), "Updating... | Esc return");
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
	});

	assert.equal(lines.some((line) => line.includes("## Highlights")), false);
	assert.equal(lines.some((line) => line.includes("- **A** item")), false);
	assert.ok(lines.some((line) => line.includes("Manual: slimebot update --version v1.26.0")));
});

test("formatUpdateReleaseNotesLines renders markdown release notes for terminal output", () => {
	const lines = formatUpdateReleaseNotesLines("## Highlights\n- **A** item", 80).map(stripAnsi);

	assert.ok(lines.some((line) => line.includes("Highlights")));
	assert.ok(lines.some((line) => line.includes("- A item")));
	assert.equal(lines.some((line) => line.includes("## Highlights")), false);
	assert.equal(lines.some((line) => line.includes("**A**")), false);
});
