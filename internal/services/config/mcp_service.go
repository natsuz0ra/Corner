package config

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"slimebot/internal/domain"
	"slimebot/internal/mcp"
)

var ErrMCPConfigNotFound = errors.New("MCP config not found")

type MCPConfigInput struct {
	Name      string
	Config    string
	IsEnabled bool
}

type MCPToolLoadStatus string

const (
	MCPToolStatusLoaded   MCPToolLoadStatus = "loaded"
	MCPToolStatusError    MCPToolLoadStatus = "error"
	MCPToolStatusDisabled MCPToolLoadStatus = "disabled"
)

type MCPToolParameterSummary struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type MCPToolItem struct {
	Name               string                    `json:"name"`
	FunctionName       string                    `json:"functionName"`
	Description        string                    `json:"description"`
	ParameterCount     int                       `json:"parameterCount"`
	RequiredParameters []string                  `json:"requiredParameters"`
	Parameters         []MCPToolParameterSummary `json:"parameters"`
	InputSchema        map[string]any            `json:"inputSchema"`
}

type MCPToolListResponse struct {
	ConfigID  string            `json:"configId"`
	Name      string            `json:"name"`
	IsEnabled bool              `json:"isEnabled"`
	Status    MCPToolLoadStatus `json:"status"`
	ToolCount int               `json:"toolCount"`
	LoadedAt  string            `json:"loadedAt"`
	Tools     []MCPToolItem     `json:"tools"`
	Error     string            `json:"error"`
}

type mcpToolsLoader interface {
	ListToolsForConfig(ctx context.Context, item domain.MCPConfig) ([]mcp.ToolMeta, []mcp.Tool, error)
}

type MCPConfigService struct {
	store domain.MCPConfigStore
	tools mcpToolsLoader
}

func NewMCPConfigService(store domain.MCPConfigStore) *MCPConfigService {
	return &MCPConfigService{store: store}
}

func NewMCPConfigServiceWithTools(store domain.MCPConfigStore, tools mcpToolsLoader) *MCPConfigService {
	return &MCPConfigService{store: store, tools: tools}
}

func (s *MCPConfigService) List(ctx context.Context) ([]domain.MCPConfig, error) {
	return s.store.ListMCPConfigs(ctx)
}

func (s *MCPConfigService) ValidateConfig(raw string) error {
	_, err := mcp.ParseAndValidateConfig(strings.TrimSpace(raw))
	return err
}

func (s *MCPConfigService) Create(ctx context.Context, input MCPConfigInput) (*domain.MCPConfig, error) {
	return s.store.CreateMCPConfig(ctx, domain.MCPConfig{
		Name:      strings.TrimSpace(input.Name),
		Config:    strings.TrimSpace(input.Config),
		IsEnabled: input.IsEnabled,
	})
}

func (s *MCPConfigService) Update(ctx context.Context, id string, input MCPConfigInput) error {
	return s.store.UpdateMCPConfig(ctx, id, domain.MCPConfig{
		Name:      strings.TrimSpace(input.Name),
		Config:    strings.TrimSpace(input.Config),
		IsEnabled: input.IsEnabled,
	})
}

func (s *MCPConfigService) Delete(ctx context.Context, id string) error {
	return s.store.DeleteMCPConfig(ctx, id)
}

func (s *MCPConfigService) GetTools(ctx context.Context, id string) (*MCPToolListResponse, error) {
	item, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := &MCPToolListResponse{
		ConfigID:  item.ID,
		Name:      item.Name,
		IsEnabled: item.IsEnabled,
		Tools:     []MCPToolItem{},
	}
	if !item.IsEnabled {
		resp.Status = MCPToolStatusDisabled
		return resp, nil
	}
	if s.tools == nil {
		return nil, fmt.Errorf("MCP tools loader is not configured")
	}

	metas, tools, err := s.tools.ListToolsForConfig(ctx, item)
	if err != nil {
		resp.Status = MCPToolStatusError
		resp.Error = err.Error()
		return resp, nil
	}
	resp.Status = MCPToolStatusLoaded
	resp.LoadedAt = time.Now().UTC().Format(time.RFC3339)
	resp.Tools = buildMCPToolItems(metas, tools)
	resp.ToolCount = len(resp.Tools)
	return resp, nil
}

func (s *MCPConfigService) findByID(ctx context.Context, id string) (domain.MCPConfig, error) {
	items, err := s.store.ListMCPConfigs(ctx)
	if err != nil {
		return domain.MCPConfig{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.MCPConfig{}, ErrMCPConfigNotFound
}

func buildMCPToolItems(metas []mcp.ToolMeta, tools []mcp.Tool) []MCPToolItem {
	funcNames := make(map[string]string, len(metas))
	for _, meta := range metas {
		funcNames[meta.ToolName] = meta.FuncName
	}
	items := make([]MCPToolItem, 0, len(tools))
	for _, tool := range tools {
		required := schemaRequiredParameters(tool.InputSchema)
		params := schemaParameterSummaries(tool.InputSchema, required)
		items = append(items, MCPToolItem{
			Name:               tool.Name,
			FunctionName:       funcNames[tool.Name],
			Description:        strings.TrimSpace(tool.Description),
			ParameterCount:     len(params),
			RequiredParameters: required,
			Parameters:         params,
			InputSchema:        tool.InputSchema,
		})
	}
	return items
}

func schemaRequiredParameters(schema map[string]any) []string {
	values, ok := schema["required"].([]any)
	if !ok {
		if stringsValues, ok := schema["required"].([]string); ok {
			out := append([]string(nil), stringsValues...)
			sort.Strings(out)
			return out
		}
		return []string{}
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if name, ok := value.(string); ok && strings.TrimSpace(name) != "" {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func schemaParameterSummaries(schema map[string]any, required []string) []MCPToolParameterSummary {
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return []MCPToolParameterSummary{}
	}
	requiredSet := make(map[string]bool, len(required))
	for _, name := range required {
		requiredSet[name] = true
	}
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]MCPToolParameterSummary, 0, len(names))
	for _, name := range names {
		prop, _ := properties[name].(map[string]any)
		out = append(out, MCPToolParameterSummary{
			Name:        name,
			Type:        schemaString(prop, "type"),
			Required:    requiredSet[name],
			Description: schemaString(prop, "description"),
		})
	}
	return out
}

func schemaString(prop map[string]any, key string) string {
	if prop == nil {
		return ""
	}
	value, _ := prop[key].(string)
	return strings.TrimSpace(value)
}
