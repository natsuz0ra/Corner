import assert from "node:assert/strict";
import test from "node:test";
import { buildSandboxMenuItems, normalizeCliSandboxMode } from "./sandboxSettings.js";

test("normalizeCliSandboxMode keeps all supported CLI sandbox modes", () => {
  assert.equal(normalizeCliSandboxMode("read-only"), "read-only");
  assert.equal(normalizeCliSandboxMode("workspace-write"), "workspace-write");
  assert.equal(normalizeCliSandboxMode("danger-full-access"), "danger-full-access");
});

test("normalizeCliSandboxMode falls back to workspace-write", () => {
  assert.equal(normalizeCliSandboxMode("unknown"), "workspace-write");
  assert.equal(normalizeCliSandboxMode(undefined), "workspace-write");
});

test("buildSandboxMenuItems includes workspace-write and network toggle", () => {
  const items = buildSandboxMenuItems({
    defaultModel: "",
    sandboxMode: "workspace-write",
    sandboxNetworkEnabled: true,
  });

  assert.deepEqual(items.map((item) => item.title), [
    "Read Only",
    "Workspace Write",
    "Danger Full Access",
    "Network Access: On",
  ]);
  assert.match(items[1]!.desc, /^current · /);
  assert.deepEqual(items[1]!.data, { type: "mode", mode: "workspace-write" });
  assert.deepEqual(items[3]!.data, { type: "network", enabled: false });
});
