package mcp

import "strings"

// parseTools maps tools/list JSON into internal Tool values.
func parseTools(result map[string]any) []Tool {
	toolItems, _ := result["tools"].([]any)
	tools := make([]Tool, 0, len(toolItems))
	for _, item := range toolItems {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := obj["name"].(string)
		if strings.TrimSpace(name) == "" {
			continue
		}
		description, _ := obj["description"].(string)
		inputSchema, _ := obj["inputSchema"].(map[string]any)
		tools = append(tools, Tool{
			Name:        name,
			Description: description,
			InputSchema: inputSchema,
		})
	}
	return tools
}

// parseCallResult maps tools/call result JSON into CallResult.
func parseCallResult(result map[string]any) *CallResult {
	var out strings.Builder
	if contents, ok := result["content"].([]any); ok {
		for _, c := range contents {
			item, ok := c.(map[string]any)
			if !ok {
				continue
			}
			text, _ := item["text"].(string)
			if text == "" {
				continue
			}
			if out.Len() > 0 {
				out.WriteString("\n")
			}
			out.WriteString(text)
		}
	}

	callErr := ""
	if isError, _ := result["isError"].(bool); isError {
		callErr = out.String()
	}
	return &CallResult{
		Output: out.String(),
		Error:  callErr,
	}
}
