import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.tsx"],
  format: ["esm"],
  target: "node20",
  bundle: true,
  external: ["@slimebot/color-diff-native"],
  noExternal: [/^(?!@slimebot\/color-diff-native$).+/],
  shims: true,
  outDir: "dist",
  clean: true,
  // ESM bundles cannot carry shebang; handle entry/bootstrap in app code.
  banner: {
    js: 'import { createRequire as __slimebotCreateRequire } from "node:module"; const require = __slimebotCreateRequire(import.meta.url);',
  },
});
