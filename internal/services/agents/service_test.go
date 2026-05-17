package agents

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceReadGlobalCreatesEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".slimebot", "AGENTS.md")
	svc := NewService(path)

	got, err := svc.ReadGlobal(context.Background())
	if err != nil {
		t.Fatalf("ReadGlobal failed: %v", err)
	}
	if got.Content != "" {
		t.Fatalf("content = %q, want empty", got.Content)
	}
	if got.Path != path {
		t.Fatalf("path = %q, want %q", got.Path, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected AGENTS.md to be created: %v", err)
	}
	if info.IsDir() {
		t.Fatal("AGENTS.md should be a file")
	}
}

func TestServiceUpdateGlobalPersistsContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	svc := NewService(path)

	if err := svc.UpdateGlobal(context.Background(), "请使用中文\n"); err != nil {
		t.Fatalf("UpdateGlobal failed: %v", err)
	}
	got, err := svc.ReadGlobal(context.Background())
	if err != nil {
		t.Fatalf("ReadGlobal failed: %v", err)
	}
	if got.Content != "请使用中文\n" {
		t.Fatalf("content = %q", got.Content)
	}
}

func TestReadProjectInstructionsConcatenatesRootToWorkingDir(t *testing.T) {
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), []byte("root doc"), 0o644); err != nil {
		t.Fatalf("write root agents failed: %v", err)
	}
	nested := filepath.Join(repo, "apps", "cli")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "AGENTS.md"), []byte("cli doc"), 0o644); err != nil {
		t.Fatalf("write nested agents failed: %v", err)
	}
	svc := NewService(filepath.Join(t.TempDir(), "AGENTS.md"))

	got, err := svc.ReadProject(context.Background(), nested)
	if err != nil {
		t.Fatalf("ReadProject failed: %v", err)
	}
	if got != "root doc\n\ncli doc" {
		t.Fatalf("project instructions = %q", got)
	}
}

func TestReadProjectInstructionsWithoutGitOnlyReadsWorkingDir(t *testing.T) {
	parent := t.TempDir()
	if err := os.WriteFile(filepath.Join(parent, "AGENTS.md"), []byte("parent doc"), 0o644); err != nil {
		t.Fatalf("write parent agents failed: %v", err)
	}
	nested := filepath.Join(parent, "child")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "AGENTS.md"), []byte("child doc"), 0o644); err != nil {
		t.Fatalf("write child agents failed: %v", err)
	}
	svc := NewService(filepath.Join(t.TempDir(), "AGENTS.md"))

	got, err := svc.ReadProject(context.Background(), nested)
	if err != nil {
		t.Fatalf("ReadProject failed: %v", err)
	}
	if strings.Contains(got, "parent doc") || got != "child doc" {
		t.Fatalf("project instructions = %q", got)
	}
}
