import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { SUPPORTED_COMMANDS } from "../types.js";
import { completeCommand, getVisibleCommandHints, matchCommandHints, moveCommandHintCursor } from "./commands.js";

test("matchCommandHints returns all commands for slash input", () => {
  assert.deepEqual(
    matchCommandHints("/").map((hint) => hint.command),
    SUPPORTED_COMMANDS.map((hint) => hint.command),
  );
});

test("matchCommandHints returns matching command prefixes", () => {
  assert.deepEqual(
    matchCommandHints("/m").map((hint) => hint.command),
    ["/model", "/mcp"],
  );
});

test("matchCommandHints includes sandbox command", () => {
  assert.deepEqual(
    matchCommandHints("/s").map((hint) => hint.command),
    ["/session", "/subagent_model", "/sandbox", "/skills"],
  );
});

test("matchCommandHints includes update command", () => {
  assert.deepEqual(
    matchCommandHints("/u").map((hint) => hint.command),
    ["/update"],
  );
});

test("matchCommandHints ignores completed commands with trailing content", () => {
  assert.deepEqual(matchCommandHints("/model "), []);
  assert.deepEqual(matchCommandHints("/model abc"), []);
});

test("completeCommand fills the selected matching command", () => {
  assert.equal(completeCommand("/m", 1), "/mcp");
});

test("completeCommand safely clamps out-of-range selected indexes", () => {
  assert.equal(completeCommand("/m", 99), "/mcp");
  assert.equal(completeCommand("/m", -99), "/model");
});

test("completeCommand clamps slash completion across all commands", () => {
  assert.equal(completeCommand("/", 99), SUPPORTED_COMMANDS.at(-1)?.command);
  assert.equal(completeCommand("/", -99), SUPPORTED_COMMANDS[0]?.command);
});

test("moveCommandHintCursor wraps through command hints", () => {
  assert.equal(moveCommandHintCursor(0, -1, 2), 1);
  assert.equal(moveCommandHintCursor(1, 1, 2), 0);
  assert.equal(moveCommandHintCursor(0, 1, 0), 0);
});

test("getVisibleCommandHints shows the first page by default", () => {
  const visible = getVisibleCommandHints(SUPPORTED_COMMANDS, 0);

  assert.deepEqual(
    visible.hints.map((hint) => hint.command),
    SUPPORTED_COMMANDS.slice(0, 5).map((hint) => hint.command),
  );
  assert.equal(visible.startIndex, 0);
  assert.equal(visible.aboveCount, 0);
  assert.equal(visible.belowCount, SUPPORTED_COMMANDS.length - 5);
});

test("getVisibleCommandHints scrolls down to include the sixth item", () => {
  const visible = getVisibleCommandHints(SUPPORTED_COMMANDS, 5);

  assert.deepEqual(
    visible.hints.map((hint) => hint.command),
    SUPPORTED_COMMANDS.slice(1, 6).map((hint) => hint.command),
  );
  assert.equal(visible.startIndex, 1);
  assert.equal(visible.aboveCount, 1);
  assert.equal(visible.belowCount, SUPPORTED_COMMANDS.length - 6);
});

test("getVisibleCommandHints anchors the final page near the end", () => {
  const visible = getVisibleCommandHints(SUPPORTED_COMMANDS, SUPPORTED_COMMANDS.length - 1);

  assert.deepEqual(
    visible.hints.map((hint) => hint.command),
    SUPPORTED_COMMANDS.slice(-5).map((hint) => hint.command),
  );
  assert.equal(visible.startIndex, SUPPORTED_COMMANDS.length - 5);
  assert.equal(visible.aboveCount, SUPPORTED_COMMANDS.length - 5);
  assert.equal(visible.belowCount, 0);
});

test("CommandHints does not render hidden-count more prompts", () => {
  const source = readFileSync(new URL("../components/CommandHints.tsx", import.meta.url), "utf8");

  assert.doesNotMatch(source, /more above|more below/);
});
