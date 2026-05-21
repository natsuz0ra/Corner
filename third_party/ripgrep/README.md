# Bundled ripgrep cache

SlimeBot release packages include platform-specific ripgrep binaries at
`bin/vendor/ripgrep/<platform>/...`. At runtime, grep/glob tools use them as a
fallback when `rg` is not available on `PATH`.

The source repository does not track these binaries. During packaging,
`scripts/package-release.sh` downloads the required official ripgrep release
artifacts into this local cache when a platform binary is missing.

Expected local cache layout:

```text
third_party/ripgrep/darwin-arm64/rg
third_party/ripgrep/darwin-amd64/rg
third_party/ripgrep/linux-amd64/rg
third_party/ripgrep/linux-arm64/rg
third_party/ripgrep/windows-amd64/rg.exe
```

The default bundled version is ripgrep `15.1.0`; Linux arm64 uses the official
`aarch64-unknown-linux-gnu` artifact.

Packaging options:

- `RIPGREP_VERSION`: ripgrep version to download. Defaults to `15.1.0`.
- `RIPGREP_VENDOR_DIR`: local binary cache directory. Defaults to this
  directory.
- `RIPGREP_DOWNLOAD_BASE_URL`: release download base URL. Defaults to the
  official GitHub release URL for `RIPGREP_VERSION`; set this to a mirror when
  needed.

For offline packaging, pre-download official ripgrep release artifacts, extract
`rg` or `rg.exe`, and place the executable in the matching cache directory
shown above. Packaging validates that the expected executable exists and, on
Unix-like platforms, marks it executable before copying it into release
packages.
