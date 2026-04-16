package chat

import (
	"context"
	"testing"

	"slimebot/internal/constants"
	llmsvc "slimebot/internal/services/llm"
)

func TestBuildRuntimeToolDefs_IncludesRunSubagentAtDepth0Only(t *testing.T) {
	ctx := context.Background()
	agent := NewAgentService(nil, nil, nil, nil)

	defs0, _, err := agent.buildRuntimeToolDefs(ctx, nil, 0)
	if err != nil {
		t.Fatalf("depth 0: %v", err)
	}
	if !containsToolName(defs0, constants.RunSubagentTool) {
		t.Fatal("expected run_subagent in defs at depth 0")
	}

	defs1, _, err := agent.buildRuntimeToolDefs(ctx, nil, 1)
	if err != nil {
		t.Fatalf("depth 1: %v", err)
	}
	if containsToolName(defs1, constants.RunSubagentTool) {
		t.Fatal("run_subagent must not appear at depth > 0")
	}
}

func containsToolName(defs []llmsvc.ToolDef, name string) bool {
	for _, d := range defs {
		if d.Name == name {
			return true
		}
	}
	return false
}
