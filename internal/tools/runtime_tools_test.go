package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	skillsvc "slimebot/internal/services/skill"
)

func TestSkillsToolListsAndViewsSkills(t *testing.T) {
	root := t.TempDir()
	writeToolSkill(t, root, "alpha", "Alpha skill", "Alpha body")
	runtime := skillsvc.NewSkillRuntimeService(skillsvc.NewFileSystemSkillStore(root), root)
	ctx := WithSkillRuntime(context.Background(), runtime)

	list, err := (&skillsTool{}).list(ctx, nil)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(list.Output, "alpha") || !strings.Contains(list.Output, "Alpha skill") {
		t.Fatalf("unexpected list:\n%s", list.Output)
	}

	view, err := (&skillsTool{}).view(ctx, map[string]any{"name": "alpha"})
	if err != nil {
		t.Fatalf("view failed: %v", err)
	}
	if !strings.Contains(view.Output, "Alpha body") {
		t.Fatalf("unexpected view:\n%s", view.Output)
	}
}

func TestTodoToolUpdatesAndListsSessionState(t *testing.T) {
	state := NewTodoState()
	ctx := WithTodoState(context.Background(), state)
	tool := &todoTool{}

	if _, err := tool.update(ctx, map[string]any{"items": []map[string]any{
		{"id": "a", "content": "first", "status": "in_progress"},
		{"id": "b", "content": "second", "status": "pending"},
	}}); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	res, err := tool.list(ctx, nil)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(res.Output, "[in_progress] a first") || !strings.Contains(res.Output, "[pending] b second") {
		t.Fatalf("unexpected todo list:\n%s", res.Output)
	}
}

func TestProcessToolReportsAndStopsManagedProcesses(t *testing.T) {
	manager := NewProcessManager()
	id := manager.Start("test command", func(ctx context.Context) processRunResult {
		<-ctx.Done()
		return processRunResult{ExitCode: -1, Error: "canceled"}
	})
	ctx := WithProcessManager(context.Background(), manager)
	tool := &processTool{}

	status, err := tool.status(ctx, map[string]any{"process_id": id})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(status.Output, id) || !strings.Contains(status.Output, "running") {
		t.Fatalf("unexpected status:\n%s", status.Output)
	}

	if _, err := tool.stop(ctx, map[string]any{"process_id": id}); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	status, err = tool.status(ctx, map[string]any{"process_id": id})
	if err != nil {
		t.Fatalf("status after stop failed: %v", err)
	}
	if !strings.Contains(status.Output, "completed") && !strings.Contains(status.Output, "stopped") {
		t.Fatalf("expected stopped/completed process, got:\n%s", status.Output)
	}
}

func writeToolSkill(t *testing.T, root, name, desc, body string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir skill: %v", err)
	}
	content := "---\nname: " + name + "\ndescription: " + desc + "\n---\n\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
}
