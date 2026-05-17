import assert from "node:assert/strict";
import test from "node:test";
import React from "react";
import { Banner } from "./Banner";

function collectText(node: unknown): string {
  if (typeof node === "string" || typeof node === "number") return String(node);
  if (!React.isValidElement(node)) return "";
  const props = node.props as { children?: unknown };
  const children = props.children;
  if (Array.isArray(children)) return children.map(collectText).join("");
  return collectText(children);
}

test("Banner omits update marker when no update is available", () => {
  const text = collectText(
    Banner({ version: "1.26.1", modelName: "gpt", cwd: "/repo" }),
  );

  assert.match(text, /1\.26\.1/);
  assert.doesNotMatch(text, /\[new\]/);
});

test("Banner shows a compact update marker without the latest version", () => {
  const text = collectText(
    Banner({
      version: "1.26.1",
      modelName: "gpt",
      cwd: "/repo",
      updateAvailable: true,
    }),
  );

  assert.match(text, /1\.26\.1 \[new\]/);
  assert.doesNotMatch(text, /v1\.26\.2/);
});
