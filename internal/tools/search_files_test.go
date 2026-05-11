package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sandboxpolicy "slimebot/internal/sandbox"
)

func TestSearchFilesFindsContentAndNames(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("hello slime\nsecond line\n"), 0o644); err != nil {
		t.Fatalf("write alpha: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "beta.go"), []byte("package beta\nconst name = \"slime\"\n"), 0o644); err != nil {
		t.Fatalf("write beta: %v", err)
	}

	res, err := (&searchFilesTool{}).search(context.Background(), map[string]any{
		"path":    dir,
		"query":   "slime",
		"pattern": "*.go",
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if strings.Contains(res.Output, "alpha.txt") {
		t.Fatalf("glob should exclude alpha.txt:\n%s", res.Output)
	}
	if !strings.Contains(res.Output, "beta.go:2") || !strings.Contains(res.Output, "slime") {
		t.Fatalf("expected beta.go line match, got:\n%s", res.Output)
	}
}

func TestSearchFilesRejectsSandboxDeniedPath(t *testing.T) {
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

	_, err = (&searchFilesTool{}).search(ctx, map[string]any{"path": dir, "query": "x"})
	if err == nil || !strings.Contains(err.Error(), "sandbox read denied") {
		t.Fatalf("expected sandbox read denial, got %v", err)
	}
}
