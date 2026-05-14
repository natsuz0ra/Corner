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
	if !strings.Contains(res.Output, "beta.go (1 matches)") || !strings.Contains(res.Output, "L2: const name = \"slime\"") {
		t.Fatalf("expected beta.go line match, got:\n%s", res.Output)
	}
}

func TestSearchFilesGroupsAndTruncatesOutput(t *testing.T) {
	dir := t.TempDir()
	longLine := "hit " + strings.Repeat("x", 260)
	content := strings.Join([]string{
		"hit one",
		"hit two",
		"hit three",
		"hit four",
		longLine,
		"hit six",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(dir, "many.txt"), []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "other.txt"), []byte("hit other\n"), 0o644); err != nil {
		t.Fatalf("write other: %v", err)
	}

	res, err := (&searchFilesTool{}).search(context.Background(), map[string]any{
		"path":                 dir,
		"query":                "hit",
		"max_matches":          "4",
		"max_matches_per_file": "3",
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if !strings.Contains(res.Output, "Matches: 4 shown, 7 found, files=2, truncated=true") {
		t.Fatalf("expected grouped truncated header, got:\n%s", res.Output)
	}
	if !strings.Contains(res.Output, "many.txt (3 of 6 matches)") {
		t.Fatalf("expected per-file truncation summary, got:\n%s", res.Output)
	}
	if strings.Contains(res.Output, "hit four") {
		t.Fatalf("per-file limit should hide fourth many.txt match:\n%s", res.Output)
	}
	if !strings.Contains(res.Output, "other.txt (1 matches)") {
		t.Fatalf("expected second file match, got:\n%s", res.Output)
	}
	if !strings.Contains(res.Output, "Refine path/pattern/query") {
		t.Fatalf("expected refinement hint, got:\n%s", res.Output)
	}
}

func TestSearchFilesTruncatesLongLines(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "long.txt"), []byte("needle "+strings.Repeat("x", 260)+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	res, err := (&searchFilesTool{}).search(context.Background(), map[string]any{
		"path":  dir,
		"query": "needle",
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if !strings.Contains(res.Output, "...") {
		t.Fatalf("expected long line truncation, got:\n%s", res.Output)
	}
	if strings.Contains(res.Output, strings.Repeat("x", 240)) {
		t.Fatalf("long line was not truncated enough:\n%s", res.Output)
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
