import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.tsx"],
  format: ["esm"],
  target: "node20",
  bundle: true,
  outDir: "dist",
  clean: true,
  // ESM bundles cannot carry shebang; handle entry/bootstrap in app code.
  banner: {},
});
