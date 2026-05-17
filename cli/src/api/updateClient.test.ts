import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

test("API client exposes update endpoints", () => {
	const source = readFileSync(new URL("./client.ts", import.meta.url), "utf8");

	assert.match(source, /\/api\/update\/check/);
	assert.match(source, /\/api\/update\/job/);
	assert.match(source, /\/api\/update\/apply/);
});
