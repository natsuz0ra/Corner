package tools

import (
	"context"
	"strings"
	"testing"

	skillsvc "slimebot/internal/services/skill"
)

func TestActivateSkillToolActivatesSkillFromRuntimeContext(t *testing.T) {
	root := t.TempDir()
	writeToolSkill(t, root, "alpha", "Alpha skill", "Alpha body")
	runtime := skillsvc.NewSkillRuntimeService(skillsvc.NewFileSystemSkillStore(root), root)
	activated := map[string]struct{}{}
	ctx := WithActivatedSkills(WithSkillRuntime(context.Background(), runtime), activated)

	result, err := (&activateSkillTool{}).Execute(ctx, "activate", map[string]any{"name": "alpha"})
	if err != nil {
		t.Fatalf("activate failed: %v", err)
	}
	if !strings.Contains(result.Output, "Alpha body") {
		t.Fatalf("unexpected activation output:\n%s", result.Output)
	}
	if _, ok := activated["alpha"]; !ok {
		t.Fatalf("expected activated skill state to include alpha")
	}
}

func TestActivateSkillToolRequiresRuntimeContext(t *testing.T) {
	_, err := (&activateSkillTool{}).Execute(context.Background(), "activate", map[string]any{"name": "alpha"})
	if err == nil || !strings.Contains(err.Error(), "skill runtime is unavailable") {
		t.Fatalf("expected missing runtime error, got %v", err)
	}
}

func TestRunSubagentToolDelegatesToContextRunner(t *testing.T) {
	runner := &captureSubagentRunner{}
	ctx := WithSubagentRunner(context.Background(), runner)

	result, err := (&runSubagentTool{}).Execute(ctx, "run", map[string]any{
		"title":   "Inspect",
		"task":    "inspect the code",
		"context": "repo background",
	})
	if err != nil {
		t.Fatalf("run_subagent failed: %v", err)
	}
	if result.Output != "subagent result" {
		t.Fatalf("unexpected output: %q", result.Output)
	}
	if runner.request.Title != "Inspect" || runner.request.Task != "inspect the code" || runner.request.Context != "repo background" {
		t.Fatalf("unexpected runner request: %+v", runner.request)
	}
}

func TestRunSubagentToolRequiresRunnerAndTask(t *testing.T) {
	_, err := (&runSubagentTool{}).Execute(context.Background(), "run", map[string]any{"task": "inspect"})
	if err == nil || !strings.Contains(err.Error(), "subagent runner is unavailable") {
		t.Fatalf("expected missing runner error, got %v", err)
	}

	ctx := WithSubagentRunner(context.Background(), &captureSubagentRunner{})
	_, err = (&runSubagentTool{}).Execute(ctx, "run", map[string]any{"title": "Inspect"})
	if err == nil || !strings.Contains(err.Error(), "task is required") {
		t.Fatalf("expected missing task error, got %v", err)
	}
}

type captureSubagentRunner struct {
	request SubagentRunRequest
}

func (r *captureSubagentRunner) RunSubagent(ctx context.Context, request SubagentRunRequest) (*ExecuteResult, error) {
	_ = ctx
	r.request = request
	return &ExecuteResult{Output: "subagent result"}, nil
}
