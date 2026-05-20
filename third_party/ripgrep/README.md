# Bundled ripgrep

SlimeBot copies these platform-specific ripgrep binaries into release packages
as `bin/vendor/ripgrep/<platform>/...`. At runtime, grep/glob tools use them
as a fallback when `rg` is not available on `PATH`.

Expected source layout:

```text
third_party/ripgrep/darwin-arm64/rg
third_party/ripgrep/darwin-amd64/rg
third_party/ripgrep/linux-amd64/rg
third_party/ripgrep/linux-arm64/rg
third_party/ripgrep/windows-amd64/rg.exe
```

Release packaging fails if the target platform binary is missing. Use official
ripgrep release artifacts and keep the executable filename as `rg` or `rg.exe`.
The current bundled binaries are from ripgrep `15.1.0`; Linux arm64 uses the
official `aarch64-unknown-linux-gnu` artifact.
