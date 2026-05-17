package mcp

import (
	"context"
	"strings"
	"testing"

	"slimebot/internal/constants"
	"slimebot/internal/domain"
)

type fakeMCPClient struct {
	tools []Tool
}

func (f *fakeMCPClient) ListTools(context.Context) ([]Tool, error) {
	return f.tools, nil
}

func (f *fakeMCPClient) CallTool(context.Context, string, map[string]any) (*CallResult, error) {
	return &CallResult{}, nil
}

func (f *fakeMCPClient) Close() error {
	return nil
}

func TestBuildMCPFuncName_LengthBounded(t *testing.T) {
	serverAlias := "mcp_" + strings.Repeat("a", 60)
	toolName := strings.Repeat("very_long_tool_name_", 5)

	name := BuildFuncName(serverAlias, toolName)
	if len(name) > constants.MCPFuncNameMaxLen {
		t.Fatalf("expected len <= %d, got %d: %s", constants.MCPFuncNameMaxLen, len(name), name)
	}
	if !strings.Contains(name, "__") {
		t.Fatalf("expected tool function name to contain separator '__': %s", name)
	}
}

func TestBuildMCPFuncName_StableAndDifferent(t *testing.T) {
	serverAlias := "mcp_server_123"

	nameA1 := BuildFuncName(serverAlias, "tool_alpha")
	nameA2 := BuildFuncName(serverAlias, "tool_alpha")
	if nameA1 != nameA2 {
		t.Fatalf("expected stable func name, got %s vs %s", nameA1, nameA2)
	}

	nameB := BuildFuncName(serverAlias, "tool_beta")
	if nameA1 == nameB {
		t.Fatalf("expected different tools to have different func names, both got %s", nameA1)
	}
}

func TestLoadToolsMetadataIncludesInternalAndDisplayNames(t *testing.T) {
	manager := NewManager()
	manager.clients["config-1"] = &managedClient{
		configID: "config-1",
		raw:      `{"command":"fake"}`,
		alias:    "mcp_config_1",
		client: &fakeMCPClient{tools: []Tool{{
			Name:        "search_repositories",
			Description: "Search repositories",
			InputSchema: map[string]any{"type": "object"},
		}}},
	}

	metas, defs, err := manager.LoadTools(context.Background(), []domain.MCPConfig{{
		ID:        "config-1",
		Name:      "github",
		Config:    `{"command":"fake"}`,
		IsEnabled: true,
	}})
	if err != nil {
		t.Fatalf("LoadTools failed: %v", err)
	}
	if len(metas) != 1 || len(defs) != 1 {
		t.Fatalf("expected one meta and one def, got metas=%d defs=%d", len(metas), len(defs))
	}

	meta := metas[0]
	if meta.FuncName == "" || meta.FuncName != BuildFuncName("mcp_config_1", "search_repositories") {
		t.Fatalf("unexpected FuncName: %#v", meta)
	}
	if meta.ServerAlias != "mcp_config_1" {
		t.Fatalf("ServerAlias = %q, want mcp_config_1", meta.ServerAlias)
	}
	if meta.ServerName != "github" {
		t.Fatalf("ServerName = %q, want github", meta.ServerName)
	}
	if meta.ToolName != "search_repositories" {
		t.Fatalf("ToolName = %q, want search_repositories", meta.ToolName)
	}
}
