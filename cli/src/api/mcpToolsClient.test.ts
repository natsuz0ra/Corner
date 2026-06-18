import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

test("API client exposes MCP tools endpoint", () => {
	const source = readFileSync(new URL("./client.ts", import.meta.url), "utf8");

	assert.match(source, /\/api\/mcp-configs\/\$\{id\}\/tools/);
});
