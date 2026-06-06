package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"slimebot/internal/domain"
	configsvc "slimebot/internal/services/config"
)

type mcpConfigServiceStub struct {
	tools *configsvc.MCPToolListResponse
}

func (s *mcpConfigServiceStub) List(context.Context) ([]domain.MCPConfig, error) {
	return nil, nil
}

func (s *mcpConfigServiceStub) ValidateConfig(string) error {
	return nil
}

func (s *mcpConfigServiceStub) Create(context.Context, configsvc.MCPConfigInput) (*domain.MCPConfig, error) {
	return nil, nil
}

func (s *mcpConfigServiceStub) Update(context.Context, string, configsvc.MCPConfigInput) error {
	return nil
}

func (s *mcpConfigServiceStub) Delete(context.Context, string) error {
	return nil
}

func (s *mcpConfigServiceStub) GetTools(context.Context, string) (*configsvc.MCPToolListResponse, error) {
	return s.tools, nil
}

func TestGetMCPConfigToolsReturnsToolList(t *testing.T) {
	controller := NewHTTPController(nil, nil, nil, nil, &mcpConfigServiceStub{tools: &configsvc.MCPToolListResponse{
		ConfigID:  "mcp-1",
		Name:      "github",
		IsEnabled: true,
		Status:    configsvc.MCPToolStatusLoaded,
		ToolCount: 1,
		Tools: []configsvc.MCPToolItem{{
			Name:               "search",
			FunctionName:       "mcp_1__search",
			Description:        "Search repositories",
			ParameterCount:     1,
			RequiredParameters: []string{"query"},
		}},
	}}, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/mcp-configs/mcp-1/tools", nil)
	resp := httptest.NewRecorder()
	controller.GetMCPConfigTools(NewChiContext(resp, req))

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != string(configsvc.MCPToolStatusLoaded) || body["toolCount"].(float64) != 1 {
		t.Fatalf("unexpected response: %#v", body)
	}
}
