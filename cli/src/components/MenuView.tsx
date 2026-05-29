/**
 * MenuView — shared menu for session / model / skills / mcp / help.
 */

import { Box, Text, useStdout } from "ink";
import type React from "react";
import type { MenuItem } from "../types.js";
import { wrapText } from "../utils/format.js";
import { CLI_ACCENT_COLOR } from "../utils/terminal.js";

interface MenuViewProps {
	title: string;
	items: MenuItem[];
	cursor: number;
	hint: string;
	maxVisibleItems?: number;
}

const MAX_MENU_TITLE_LENGTH = 25;
const MAX_MENU_DESC_LENGTH = 80;

export const MENU_TITLE_GAP_LINES = 1;
export const MENU_VISIBLE_LIMIT = 5;
export const CLI_HINT_COLOR = "#64748b";
export const MENU_ITEM_COLORS = {
	title: CLI_ACCENT_COLOR,
	activeCursor: CLI_ACCENT_COLOR,
	inactiveCursor: "#64748b",
	activeTitle: "#f8fafc",
	inactiveTitle: "#cbd5e1",
	description: "#94a3b8",
	empty: "#94a3b8",
	hint: CLI_HINT_COLOR,
} as const;

export function truncateMenuTitle(
	title: string,
	maxLen = MAX_MENU_TITLE_LENGTH,
): string {
	const normalized = (title ?? "").trim();
	if (normalized.length <= maxLen) return normalized;
	if (maxLen <= 1) return "…";
	return `${normalized.slice(0, maxLen - 1)}…`;
}

export function truncateMenuDescription(
	desc: string,
	maxLen = MAX_MENU_DESC_LENGTH,
): string {
	const normalized = (desc ?? "").replace(/\s+/g, " ").trim();
	if (!normalized) return "(No description)";
	if (normalized.length <= maxLen) return normalized;
	if (maxLen <= 1) return "…";
	return `${normalized.slice(0, maxLen - 1)}…`;
}

export function formatMenuDescriptionLines(
	desc: string,
	terminalWidth: number,
): string[] {
	const text = truncateMenuDescription(desc);
	const lineWidth = Math.max(
		10,
		Math.min(MAX_MENU_DESC_LENGTH, terminalWidth - 2),
	);
	return wrapText(text, lineWidth).split("\n");
}

function menuItemKey(item: MenuItem): string {
	const data = item.data;
	if (
		data &&
		typeof data === "object" &&
		"id" in data &&
		typeof data.id === "string"
	) {
		return data.id;
	}
	return `${item.title}:${item.desc}`;
}

function clampMenuCursor(cursor: number, length: number): number {
	if (length <= 0) return 0;
	return Math.max(0, Math.min(length - 1, cursor));
}

export function getVisibleMenuItems(
	items: MenuItem[],
	cursor: number,
	maxVisible = items.length,
): { items: MenuItem[]; startIndex: number } {
	if (items.length === 0 || maxVisible <= 0) {
		return { items: [], startIndex: 0 };
	}

	const visibleCount = Math.min(maxVisible, items.length);
	const selected = clampMenuCursor(cursor, items.length);
	const maxStart = items.length - visibleCount;
	const startIndex = Math.max(
		0,
		Math.min(selected - visibleCount + 1, maxStart),
	);

	return {
		items: items.slice(startIndex, startIndex + visibleCount),
		startIndex,
	};
}

export function MenuView({
	title,
	items,
	cursor,
	hint,
	maxVisibleItems,
}: MenuViewProps): React.ReactElement {
	const { stdout } = useStdout();
	const terminalWidth = Math.max(20, stdout?.columns || 80);
	const visible = getVisibleMenuItems(
		items,
		cursor,
		maxVisibleItems ?? items.length,
	);

	return (
		<Box flexDirection="column">
			<Text bold color={MENU_ITEM_COLORS.title}>
				{title}
			</Text>
			{MENU_TITLE_GAP_LINES > 0 && <Text> </Text>}
			{items.length === 0 ? (
				<Text color={MENU_ITEM_COLORS.empty}>(empty)</Text>
			) : (
				visible.items.map((item, i) => {
					const absoluteIndex = visible.startIndex + i;
					const selected = absoluteIndex === cursor;
					return (
						<Box key={menuItemKey(item)} flexDirection="column">
							<Text>
								<Text
									color={
										selected
											? MENU_ITEM_COLORS.activeCursor
											: MENU_ITEM_COLORS.inactiveCursor
									}
								>
									{selected ? "\u276F" : " "}
								</Text>
								<Text> </Text>
								<Text
									bold={selected}
									color={
										selected
											? MENU_ITEM_COLORS.activeTitle
											: MENU_ITEM_COLORS.inactiveTitle
									}
								>
									{truncateMenuTitle(item.title)}
								</Text>
							</Text>
							{formatMenuDescriptionLines(item.desc, terminalWidth).map(
								(line) => (
									<Text
										key={`${item.title}-desc-${line}`}
										color={MENU_ITEM_COLORS.description}
									>
										{`  ${line}`}
									</Text>
								),
							)}
						</Box>
					);
				})
			)}
			{hint && (
				<Box flexDirection="column">
					<Text> </Text>
					<Text color={MENU_ITEM_COLORS.hint}>{hint}</Text>
				</Box>
			)}
		</Box>
	);
}
