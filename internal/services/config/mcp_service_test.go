package config

import (
	"context"
	"errors"
	"testing"

	"slimebot/internal/domain"
	"slimebot/internal/mcp"
)

type mcpStoreStub struct {
	items []domain.MCPConfig
}

func (s *mcpStoreStub) ListMCPConfigs(context.Context) ([]domain.MCPConfig, error) {
	return s.items, nil
}

func (s *mcpStoreStub) CreateMCPConfig(context.Context, domain.MCPConfig) (*domain.MCPConfig, error) {
	return nil, nil
}

func (s *mcpStoreStub) UpdateMCPConfig(context.Context, string, domain.MCPConfig) error {
	return nil
}

func (s *mcpStoreStub) DeleteMCPConfig(context.Context, string) error {
	return nil
}

type mcpToolsLoaderStub struct {
	metas []mcp.ToolMeta
	tools []mcp.Tool
	err   error
	calls int
}

func (s *mcpToolsLoaderStub) ListToolsForConfig(ctx context.Context, item domain.MCPConfig) ([]mcp.ToolMeta, []mcp.Tool, error) {
	s.calls++
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.metas, s.tools, nil
}

func TestGetToolsReturnsDisabledWithoutLoading(t *testing.T) {
	loader := &mcpToolsLoaderStub{}
	service := NewMCPConfigServiceWithTools(&mcpStoreStub{items: []domain.MCPConfig{{
		ID:        "mcp-1",
		Name:      "filesystem",
		IsEnabled: false,
	}}}, loader)

	got, err := service.GetTools(context.Background(), "mcp-1")
	if err != nil {
		t.Fatalf("GetTools failed: %v", err)
	}
	if got.Status != MCPToolStatusDisabled || got.ToolCount != 0 || len(got.Tools) != 0 {
		t.Fatalf("unexpected disabled response: %#v", got)
	}
	if loader.calls != 0 {
		t.Fatalf("disabled config must not load tools, calls=%d", loader.calls)
	}
}

func TestGetToolsSummarizesParameters(t *testing.T) {
	service := NewMCPConfigServiceWithTools(&mcpStoreStub{items: []domain.MCPConfig{{
		ID:        "mcp-1",
		Name:      "github",
		Config:    `{"command":"fake"}`,
		IsEnabled: true,
	}}}, &mcpToolsLoaderStub{
		metas: []mcp.ToolMeta{{FuncName: "mcp_1__search", ToolName: "search"}},
		tools: []mcp.Tool{{
			Name:        "search",
			Description: "Search repositories",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []any{"query"},
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search query",
					},
					"limit": map[string]any{
						"type": "number",
					},
				},
			},
		}},
	})

	got, err := service.GetTools(context.Background(), "mcp-1")
	if err != nil {
		t.Fatalf("GetTools failed: %v", err)
	}
	if got.Status != MCPToolStatusLoaded || got.ToolCount != 1 {
		t.Fatalf("unexpected loaded response: %#v", got)
	}
	tool := got.Tools[0]
	if tool.FunctionName != "mcp_1__search" || tool.ParameterCount != 2 {
		t.Fatalf("unexpected tool summary: %#v", tool)
	}
	if len(tool.RequiredParameters) != 1 || tool.RequiredParameters[0] != "query" {
		t.Fatalf("unexpected required params: %#v", tool.RequiredParameters)
	}
	if len(tool.Parameters) != 2 {
		t.Fatalf("unexpected parameter summaries: %#v", tool.Parameters)
	}
	var queryParam MCPToolParameterSummary
	for _, param := range tool.Parameters {
		if param.Name == "query" {
			queryParam = param
		}
	}
	if !queryParam.Required || queryParam.Type != "string" || queryParam.Description != "Search query" {
		t.Fatalf("unexpected query parameter summary: %#v", queryParam)
	}
}

func TestGetToolsUsesEmptyArraysForMissingSchemaCollections(t *testing.T) {
	service := NewMCPConfigServiceWithTools(&mcpStoreStub{items: []domain.MCPConfig{{
		ID:        "mcp-1",
		Name:      "github",
		Config:    `{"command":"fake"}`,
		IsEnabled: true,
	}}}, &mcpToolsLoaderStub{
		metas: []mcp.ToolMeta{{FuncName: "mcp_1__ping", ToolName: "ping"}},
		tools: []mcp.Tool{{
			Name:        "ping",
			Description: "Ping",
			InputSchema: map[string]any{"type": "object"},
		}},
	})

	got, err := service.GetTools(context.Background(), "mcp-1")
	if err != nil {
		t.Fatalf("GetTools failed: %v", err)
	}
	tool := got.Tools[0]
	if tool.RequiredParameters == nil {
		t.Fatal("requiredParameters must be an empty array, not null")
	}
	if tool.Parameters == nil {
		t.Fatal("parameters must be an empty array, not null")
	}
}

func TestGetToolsReturnsErrorStatusWhenLoaderFails(t *testing.T) {
	service := NewMCPConfigServiceWithTools(&mcpStoreStub{items: []domain.MCPConfig{{
		ID:        "mcp-1",
		Name:      "github",
		Config:    `{"command":"fake"}`,
		IsEnabled: true,
	}}}, &mcpToolsLoaderStub{err: errors.New("connect failed")})

	got, err := service.GetTools(context.Background(), "mcp-1")
	if err != nil {
		t.Fatalf("GetTools should return error status instead of handler error: %v", err)
	}
	if got.Status != MCPToolStatusError || got.Error == "" {
		t.Fatalf("unexpected error response: %#v", got)
	}
}
