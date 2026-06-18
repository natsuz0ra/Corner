import assert from "node:assert/strict";
import test from "node:test";
import {
	formatMCPToolParameters,
	getMCPToolsStatusLine,
	truncateMCPToolDescription,
} from "./MCPToolsView";

test("MCP tools view formats required parameter summaries", () => {
	assert.equal(formatMCPToolParameters({ parameterCount: 3, requiredParameters: ["query"] }), "3 params · required query");
	assert.equal(formatMCPToolParameters({ parameterCount: 0, requiredParameters: [] }), "0 params");
	assert.equal(formatMCPToolParameters({ parameterCount: 1, requiredParameters: null as unknown as string[] }), "1 param");
});

test("MCP tools view formats loaded, disabled and error status lines", () => {
	assert.equal(getMCPToolsStatusLine({ status: "loaded", toolCount: 2, error: "" }), "2 tools loaded");
	assert.equal(getMCPToolsStatusLine({ status: "disabled", toolCount: 0, error: "" }), "disabled · tools not loaded");
	assert.equal(getMCPToolsStatusLine({ status: "error", toolCount: 0, error: "connect failed" }), "tools load failed: connect failed");
});

test("MCP tools view truncates long descriptions", () => {
	assert.equal(truncateMCPToolDescription("abcdef", 10), "abcdef");
	assert.equal(truncateMCPToolDescription("abcdefghijklmnopqrstuvwxyz", 12), "abcdefghijk…");
});
