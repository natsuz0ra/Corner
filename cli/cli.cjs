#!/usr/bin/env node
// CJS wrapper to load the ESM entry (Node ESM output cannot use shebang in bundle).

// Dynamic import of ESM bundle.
async function main() {
  await import('./dist/index.js');
}

main().catch((error) => {
  console.error('Failed to start CLI:', error);
  process.exit(1);
});
