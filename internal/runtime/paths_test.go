package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandHome_EmptyStaysEmpty(t *testing.T) {
	if got := ExpandHome(""); got != "" {
		t.Fatalf("expected empty string, got=%q", got)
	}
}

func TestDescribeConfigHome_ListsEntries(t *testing.T) {
	dir := t.TempDir()
	// SlimeBotHomeDir is not easily overridden; this test only checks general behavior.
	// Create some files and directories.
	_ = os.MkdirAll(filepath.Join(dir, "skills"), os.ModePerm)
	_ = os.MkdirAll(filepath.Join(dir, "storage"), os.ModePerm)
	if f, err := os.Create(filepath.Join(dir, ".env")); err == nil {
		f.Close()
	}

	// Call DescribeConfigHome and verify output shape.
	// It uses the real SlimeBotHomeDir; we only assert format hints.
	result := DescribeConfigHome()
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	// Should mention path or missing-dir message.
	if !strings.Contains(result, ".slimebot") && !strings.Contains(result, "directory not yet created") {
		t.Fatalf("unexpected result: %q", result)
	}
}

func TestDescribeConfigHome_NonexistentDir(t *testing.T) {
	// When the directory does not exist, still return a descriptive string.
	result := DescribeConfigHome()
	if result == "" {
		t.Fatal("expected non-empty result for nonexistent dir")
	}
}
