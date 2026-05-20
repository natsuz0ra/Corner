package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"slimebot/internal/constants"
	sandboxpolicy "slimebot/internal/sandbox"
)

func TestGrepSearchContentModeFiltersAndPaginates(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("slime in text\n"), 0o644); err != nil {
		t.Fatalf("write alpha: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "beta.go"), []byte("package beta\nconst name = \"slime\"\nconst other = \"slime\"\n"), 0o644); err != nil {
		t.Fatalf("write beta: %v", err)
	}

	res, err := (&grepTool{}).search(context.Background(), map[string]any{
		"path":             dir,
		"pattern":          "slime",
		"glob":             "*.go",
		"output_mode":      "content",
		"line_numbers":     true,
		"head_limit":       1,
		"offset":           1,
		"case_insensitive": true,
	})
	if err != nil {
		t.Fatalf("grep failed: %v", err)
	}
	if strings.Contains(res.Output, "alpha.txt") {
		t.Fatalf("glob should exclude alpha.txt:\n%s", res.Output)
	}
	if !strings.Contains(res.Output, "beta.go") || !strings.Contains(res.Output, "L3: const other = \"slime\"") {
		t.Fatalf("expected paginated beta.go content match, got:\n%s", res.Output)
	}
	if !strings.Contains(res.Output, "pagination") || !strings.Contains(res.Output, "offset: 1") {
		t.Fatalf("expected pagination hint, got:\n%s", res.Output)
	}
}

func TestGrepSearchFilesAndCountModes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "one.go"), []byte("needle\nneedle\n"), 0o644); err != nil {
		t.Fatalf("write one: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "two.go"), []byte("needle\n"), 0o644); err != nil {
		t.Fatalf("write two: %v", err)
	}

	files, err := (&grepTool{}).search(context.Background(), map[string]any{
		"path":        dir,
		"pattern":     "needle",
		"output_mode": "files_with_matches",
		"head_limit":  1,
	})
	if err != nil {
		t.Fatalf("grep files failed: %v", err)
	}
	if !strings.Contains(files.Output, "Found 1 file") || !strings.Contains(files.Output, ".go") {
		t.Fatalf("expected one file result, got:\n%s", files.Output)
	}

	count, err := (&grepTool{}).search(context.Background(), map[string]any{
		"path":        dir,
		"pattern":     "needle",
		"output_mode": "count",
	})
	if err != nil {
		t.Fatalf("grep count failed: %v", err)
	}
	if !strings.Contains(count.Output, "Found 3 total occurrences across 2 files") {
		t.Fatalf("expected count summary, got:\n%s", count.Output)
	}
}

func TestGlobFindSortsLimitsAndTruncates(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.go")
	newPath := filepath.Join(dir, "new.go")
	if err := os.WriteFile(oldPath, []byte("package old\n"), 0o644); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if err := os.WriteFile(newPath, []byte("package new\n"), 0o644); err != nil {
		t.Fatalf("write new: %v", err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	newTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes old: %v", err)
	}
	if err := os.Chtimes(newPath, newTime, newTime); err != nil {
		t.Fatalf("chtimes new: %v", err)
	}

	res, err := (&globTool{}).find(context.Background(), map[string]any{
		"path":    dir,
		"pattern": "*.go",
		"limit":   1,
	})
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if !strings.Contains(res.Output, "old.go") || strings.Contains(res.Output, "new.go") {
		t.Fatalf("expected oldest modified file only, got:\n%s", res.Output)
	}
	if !strings.Contains(res.Output, "truncated=true") {
		t.Fatalf("expected truncation marker, got:\n%s", res.Output)
	}
}

func TestGlobFindReturnsPartialResultsWhenRipgrepHitsUnreadablePaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script fake rg")
	}

	exeDir := t.TempDir()
	vendor := filepath.Join(exeDir, "vendor", "ripgrep", runtime.GOOS+"-"+runtime.GOARCH)
	if err := os.MkdirAll(vendor, 0o755); err != nil {
		t.Fatalf("mkdir vendor: %v", err)
	}
	fakeRG := filepath.Join(vendor, ripgrepExecutableName())
	script := "#!/bin/sh\nprintf 'Downloads/test_text_2_副本.txt\\n'\nprintf 'rg: Library: Operation not permitted (os error 1)\\n' >&2\nexit 2\n"
	if err := os.WriteFile(fakeRG, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake rg: %v", err)
	}
	restore := setRipgrepTestOverrides(ripgrepTestOverrides{
		lookPath: func(file string) (string, error) {
			return "", exec.ErrNotFound
		},
		executable: func() (string, error) {
			return filepath.Join(exeDir, "slimebot"), nil
		},
	})
	defer restore()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "Downloads"), 0o755); err != nil {
		t.Fatalf("mkdir downloads: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Downloads", "test_text_2_副本.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write match: %v", err)
	}

	res, err := (&globTool{}).find(context.Background(), map[string]any{
		"path":    dir,
		"pattern": "**/*test_text_2_副本*",
	})
	if err != nil {
		t.Fatalf("glob should return partial results instead of failing: %v", err)
	}
	if !strings.Contains(res.Output, "Downloads/test_text_2_副本.txt") {
		t.Fatalf("expected partial match in output, got:\n%s", res.Output)
	}
}

func TestRunRipgrepPartialModeKeepsInvalidPatternErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script fake rg")
	}

	exeDir := t.TempDir()
	vendor := filepath.Join(exeDir, "vendor", "ripgrep", runtime.GOOS+"-"+runtime.GOARCH)
	if err := os.MkdirAll(vendor, 0o755); err != nil {
		t.Fatalf("mkdir vendor: %v", err)
	}
	fakeRG := filepath.Join(vendor, ripgrepExecutableName())
	script := "#!/bin/sh\nprintf \"rg: error parsing glob '[': unclosed character class; missing ']'\\n\" >&2\nexit 2\n"
	if err := os.WriteFile(fakeRG, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake rg: %v", err)
	}
	restore := setRipgrepTestOverrides(ripgrepTestOverrides{
		lookPath: func(file string) (string, error) {
			return "", exec.ErrNotFound
		},
		executable: func() (string, error) {
			return filepath.Join(exeDir, "slimebot"), nil
		},
	})
	defer restore()

	_, err := runRipgrepWithOptions(context.Background(), []string{"--files"}, t.TempDir(), ripgrepRunOptions{allowPartialOutput: true})
	if err == nil || !strings.Contains(err.Error(), "error parsing glob") {
		t.Fatalf("expected invalid pattern error to be preserved, got %v", err)
	}
}

func TestGrepAndGlobRejectSandboxDeniedPath(t *testing.T) {
	dir := t.TempDir()
	policy, err := sandboxpolicy.NewPolicy(sandboxpolicy.Config{
		Mode:      sandboxpolicy.ModeWorkspaceWrite,
		CWD:       dir,
		DenyPaths: []string{dir},
	})
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	ctx := sandboxpolicy.WithPolicy(context.Background(), policy)

	if _, err := (&grepTool{}).search(ctx, map[string]any{"path": dir, "pattern": "x"}); err == nil || !strings.Contains(err.Error(), "sandbox read denied") {
		t.Fatalf("expected grep sandbox read denial, got %v", err)
	}
	if _, err := (&globTool{}).find(ctx, map[string]any{"path": dir, "pattern": "*.go"}); err == nil || !strings.Contains(err.Error(), "sandbox read denied") {
		t.Fatalf("expected glob sandbox read denial, got %v", err)
	}
}

func TestGrepAndGlobRequirePathInWebSurface(t *testing.T) {
	ctx := constants.WithClientSurface(context.Background(), constants.ClientSurfaceWeb)

	if _, err := (&grepTool{}).search(ctx, map[string]any{"pattern": "x"}); err == nil || !strings.Contains(err.Error(), "path is required") {
		t.Fatalf("expected grep path requirement in web mode, got %v", err)
	}
	if _, err := (&globTool{}).find(ctx, map[string]any{"pattern": "*.go"}); err == nil || !strings.Contains(err.Error(), "path is required") {
		t.Fatalf("expected glob path requirement in web mode, got %v", err)
	}
}

func TestRipgrepMissingBinaryReturnsActionableError(t *testing.T) {
	restore := setRipgrepTestOverrides(ripgrepTestOverrides{
		lookPath: func(file string) (string, error) {
			return "", exec.ErrNotFound
		},
		executable: func() (string, error) {
			return filepath.Join(t.TempDir(), "slimebot"), nil
		},
		repoRoot: t.TempDir(),
	})
	defer restore()

	dir := t.TempDir()
	_, err := (&grepTool{}).search(context.Background(), map[string]any{"path": dir, "pattern": "x"})
	if err == nil || !strings.Contains(err.Error(), "ripgrep (rg) was not found") || !strings.Contains(err.Error(), "bundled vendor") {
		t.Fatalf("expected missing rg error, got %v", err)
	}
}

func TestResolveRipgrepCommandPrefersSystemPath(t *testing.T) {
	restore := setRipgrepTestOverrides(ripgrepTestOverrides{
		lookPath: func(file string) (string, error) {
			if file != "rg" {
				t.Fatalf("lookPath file = %q, want rg", file)
			}
			return "/usr/local/bin/rg", nil
		},
		executable: func() (string, error) {
			t.Fatal("executable should not be checked when system rg exists")
			return "", nil
		},
		repoRoot: t.TempDir(),
	})
	defer restore()

	cmd, err := resolveRipgrepCommand()
	if err != nil {
		t.Fatalf("resolve ripgrep: %v", err)
	}
	if cmd != "rg" {
		t.Fatalf("command = %q, want rg", cmd)
	}
}

func TestResolveRipgrepCommandUsesBundledVendorWhenSystemMissing(t *testing.T) {
	exeDir := t.TempDir()
	vendor := filepath.Join(exeDir, "vendor", "ripgrep", runtime.GOOS+"-"+runtime.GOARCH)
	if err := os.MkdirAll(vendor, 0o755); err != nil {
		t.Fatalf("mkdir vendor: %v", err)
	}
	vendorRG := filepath.Join(vendor, ripgrepExecutableName())
	if err := os.WriteFile(vendorRG, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write vendor rg: %v", err)
	}

	restore := setRipgrepTestOverrides(ripgrepTestOverrides{
		lookPath: func(file string) (string, error) {
			return "", exec.ErrNotFound
		},
		executable: func() (string, error) {
			return filepath.Join(exeDir, "slimebot"), nil
		},
		repoRoot: t.TempDir(),
	})
	defer restore()

	cmd, err := resolveRipgrepCommand()
	if err != nil {
		t.Fatalf("resolve ripgrep: %v", err)
	}
	if cmd != vendorRG {
		t.Fatalf("command = %q, want %q", cmd, vendorRG)
	}
}

func TestResolveRipgrepCommandUsesThirdPartyFallbackInDevelopment(t *testing.T) {
	repoRoot := t.TempDir()
	thirdParty := filepath.Join(repoRoot, "third_party", "ripgrep", runtime.GOOS+"-"+runtime.GOARCH)
	if err := os.MkdirAll(thirdParty, 0o755); err != nil {
		t.Fatalf("mkdir third_party ripgrep: %v", err)
	}
	thirdPartyRG := filepath.Join(thirdParty, ripgrepExecutableName())
	if err := os.WriteFile(thirdPartyRG, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write third_party rg: %v", err)
	}

	restore := setRipgrepTestOverrides(ripgrepTestOverrides{
		lookPath: func(file string) (string, error) {
			return "", exec.ErrNotFound
		},
		executable: func() (string, error) {
			return filepath.Join(t.TempDir(), "slimebot"), nil
		},
		repoRoot: repoRoot,
	})
	defer restore()

	cmd, err := resolveRipgrepCommand()
	if err != nil {
		t.Fatalf("resolve ripgrep: %v", err)
	}
	if cmd != thirdPartyRG {
		t.Fatalf("command = %q, want %q", cmd, thirdPartyRG)
	}
}
