import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

test("standalone CLI defaults to the configured server port", () => {
  const source = readFileSync(new URL("../index.tsx", import.meta.url), "utf8");

  assert.match(source, /http:\/\/127\.0\.0\.1:6247/);
  assert.doesNotMatch(source, /http:\/\/127\.0\.0\.1:8080/);
});
