import assert from "node:assert/strict";
import test from "node:test";
import { formatUpdateHint, formatUpdateSummaryLines } from "./UpdateView";

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
	assert.equal(formatUpdateHint(true, false), "C check again | U update | Esc return");
	assert.equal(formatUpdateHint(false, false), "C check again | Esc return");
	assert.equal(formatUpdateHint(true, true), "Updating... | Esc return");
});
