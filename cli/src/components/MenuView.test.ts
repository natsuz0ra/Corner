import assert from "node:assert/strict";
import test from "node:test";
import {
	CLI_HINT_COLOR,
	MENU_ITEM_COLORS,
	MENU_TITLE_GAP_LINES,
	MENU_VISIBLE_LIMIT,
	formatMenuDescriptionLines,
	getVisibleMenuItems,
	truncateMenuDescription,
	truncateMenuTitle,
} from "./MenuView";
import type { MenuItem } from "../types";

function menuItems(count: number): MenuItem[] {
	return Array.from({ length: count }, (_, index) => ({
		title: `Session ${index + 1}`,
		desc: `Updated ${index + 1}`,
		data: { id: `session-${index + 1}` },
	}));
}

test("truncateMenuTitle truncates long titles with ellipsis", () => {
	const input = "12345678901234567890123456";
	assert.equal(truncateMenuTitle(input), "123456789012345678901234…");
});

test("truncateMenuDescription limits description to 80 characters", () => {
	const input = "a".repeat(100);
	const output = truncateMenuDescription(input);

	assert.equal(output.length, 80);
	assert.ok(output.endsWith("…"));
});

test("formatMenuDescriptionLines wraps by terminal width", () => {
	const lines = formatMenuDescriptionLines(
		"Use when user asks to run a Python script locally to write files.",
		24,
	);

	assert.ok(lines.length > 1);
	assert.ok(lines.every((line) => line.length <= 22));
});

test("menu spacing leaves a blank line after the title", () => {
	assert.equal(MENU_TITLE_GAP_LINES, 1);
});

test("menu palette gives active and inactive items distinct colors", () => {
	assert.notEqual(MENU_ITEM_COLORS.activeTitle, MENU_ITEM_COLORS.inactiveTitle);
	assert.notEqual(
		MENU_ITEM_COLORS.activeCursor,
		MENU_ITEM_COLORS.inactiveCursor,
	);
});

test("menu hint uses the shared CLI hint color", () => {
	assert.equal(MENU_ITEM_COLORS.hint, CLI_HINT_COLOR);
});

test("getVisibleMenuItems returns all items below the visible limit", () => {
	const visible = getVisibleMenuItems(menuItems(3), 0, MENU_VISIBLE_LIMIT);

	assert.deepEqual(
		visible.items.map((item) => item.title),
		["Session 1", "Session 2", "Session 3"],
	);
	assert.equal(visible.startIndex, 0);
});

test("getVisibleMenuItems scrolls down to keep the sixth item visible", () => {
	const visible = getVisibleMenuItems(menuItems(8), 5, MENU_VISIBLE_LIMIT);

	assert.deepEqual(
		visible.items.map((item) => item.title),
		["Session 2", "Session 3", "Session 4", "Session 5", "Session 6"],
	);
	assert.equal(visible.startIndex, 1);
});

test("getVisibleMenuItems anchors the final window near the end", () => {
	const visible = getVisibleMenuItems(menuItems(8), 7, MENU_VISIBLE_LIMIT);

	assert.deepEqual(
		visible.items.map((item) => item.title),
		["Session 4", "Session 5", "Session 6", "Session 7", "Session 8"],
	);
	assert.equal(visible.startIndex, 3);
});
