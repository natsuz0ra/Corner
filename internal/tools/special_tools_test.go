package tools

import (
	"context"
	"strings"
	"testing"

	"slimebot/internal/constants"
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

func TestBuildPlanToolDefs(t *testing.T) {
	defs := BuildPlanToolDefs()
	if len(defs) != 2 {
		t.Fatalf("expected two plan tool defs, got %d: %#v", len(defs), defs)
	}
	if defs[0].Name != constants.PlanStartTool {
		t.Fatalf("expected first plan tool to be %s, got %s", constants.PlanStartTool, defs[0].Name)
	}
	if defs[1].Name != constants.PlanCompleteTool {
		t.Fatalf("expected second plan tool to be %s, got %s", constants.PlanCompleteTool, defs[1].Name)
	}

	startParams := defs[0].Parameters
	if startParams["type"] != "object" {
		t.Fatalf("expected plan_start object schema, got %#v", startParams["type"])
	}
	startProps, ok := startParams["properties"].(map[string]any)
	if !ok || len(startProps) != 0 {
		t.Fatalf("expected plan_start empty properties, got %#v", startParams["properties"])
	}

	completeParams := defs[1].Parameters
	completeProps, ok := completeParams["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected plan_complete properties, got %#v", completeParams["properties"])
	}
	titleProp, ok := completeProps["title"].(map[string]any)
	if !ok {
		t.Fatalf("expected plan_complete title property, got %#v", completeProps["title"])
	}
	if titleProp["type"] != "string" {
		t.Fatalf("expected title property to be string, got %#v", titleProp["type"])
	}
	if _, ok := completeParams["required"]; ok {
		t.Fatalf("plan_complete title should remain optional, got required=%#v", completeParams["required"])
	}
}

func TestPlanToolsAllowedInPlanMode(t *testing.T) {
	if !IsPlanModeAllowedFunction(constants.PlanStartTool) {
		t.Fatalf("expected %s to be allowed in plan mode", constants.PlanStartTool)
	}
	if !IsPlanModeAllowedFunction(constants.PlanCompleteTool) {
		t.Fatalf("expected %s to be allowed in plan mode", constants.PlanCompleteTool)
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
