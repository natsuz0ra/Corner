package tools

import (
	"context"
	"strings"

	"slimebot/internal/constants"
)

// CommandParam describes one parameter for a tool command.
type CommandParam struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
	Schema      any    `json:"schema,omitempty"`
}

// Command describes one subcommand supported by a tool.
type Command struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Params      []CommandParam `json:"params,omitempty"`
}

// ExecuteResult is the outcome of a tool command.
type ExecuteResult struct {
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Metadata any    `json:"metadata,omitempty"`
}

// Tool is the interface every built-in tool implements.
// Add a new tool by implementing it in this package and calling Register from init().
type Tool interface {
	// Name returns the stable tool id (e.g. "exec", "http_request").
	Name() string
	// Description returns a short capability summary.
	Description() string
	// Commands lists supported subcommands.
	Commands() []Command
	Execute(ctx context.Context, command string, params map[string]any) (*ExecuteResult, error)
}

func IsStableNameTool(name string) bool {
	switch name {
	case constants.ActivateSkillTool, constants.RunSubagentTool:
		return true
	default:
		return false
	}
}

// IsPlanModeAllowedFunction reports whether a model-facing tool function is
// allowed while the agent is in read-only planning mode.
func IsPlanModeAllowedFunction(funcName string) bool {
	funcName = strings.TrimSpace(funcName)
	if funcName == "" || funcName == "todo__update" {
		return false
	}
	switch funcName {
	case constants.PlanStartTool, constants.PlanCompleteTool, constants.RunSubagentTool, "todo_update":
		return true
	}
	toolName, _, ok := ParseFunctionName(funcName)
	if !ok {
		return false
	}
	return IsPlanModeAllowedCommand(toolName)
}

func IsPlanModeAllowedCommand(toolName string) bool {
	switch strings.TrimSpace(toolName) {
	case "web_search", "web_extract", "file_read", "search_files", "skills", "todo", "plan_complete":
		return true
	default:
		return false
	}
}

func ParseFunctionName(funcName string) (toolName, command string, ok bool) {
	parts := strings.SplitN(funcName, "__", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
