package mcp

import "context"

// Tool is one tool exposed by an MCP server.
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]any
}

// CallResult is the outcome of one MCP tools/call.
type CallResult struct {
	Output string
	Error  string
}

// Client is the minimal MCP transport contract.
type Client interface {
	// ListTools returns tools currently advertised by the server.
	ListTools(ctx context.Context) ([]Tool, error)
	// CallTool invokes a tool by name and returns a normalized result.
	CallTool(ctx context.Context, name string, arguments map[string]any) (*CallResult, error)
	// Close releases subprocess or connection resources.
	Close() error
}
