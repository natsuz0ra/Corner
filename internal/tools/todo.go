package tools

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

var defaultTodoState = NewTodoState()

type TodoState struct {
	mu    sync.Mutex
	items []ToolTodoItem
	note  string
}

type ToolTodoItem struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

func NewTodoState() *TodoState {
	return &TodoState{}
}

type todoTool struct{}

func init() {
	Register(&todoTool{})
}

func (t *todoTool) Name() string { return "todo" }

func (t *todoTool) Description() string {
	return "Maintain a temporary todo list for the current agent session."
}

func (t *todoTool) Commands() []Command {
	return []Command{
		{Name: "list", Description: "List the current session todo items."},
		{
			Name:        "update",
			Description: "Replace the current session todo list. Use exactly one in_progress item while work remains.",
			Params: []CommandParam{
				{Name: "items", Required: true, Description: "Todo items: [{id,content,status}], status is pending|in_progress|completed.", Schema: map[string]any{"type": "array"}},
				{Name: "note", Required: false, Description: "Optional short note for this update."},
			},
		},
	}
}

func (t *todoTool) Execute(ctx context.Context, command string, params map[string]any) (*ExecuteResult, error) {
	switch command {
	case "list":
		return t.list(ctx, params)
	case "update":
		return t.update(ctx, params)
	default:
		return nil, fmt.Errorf("todo tool does not support command: %s", command)
	}
}

func (t *todoTool) list(ctx context.Context, _ map[string]any) (*ExecuteResult, error) {
	state := todoStateFromContext(ctx)
	state.mu.Lock()
	defer state.mu.Unlock()
	if len(state.items) == 0 {
		return &ExecuteResult{Output: "No todo items."}, nil
	}
	return &ExecuteResult{Output: formatToolTodosLocked(state)}, nil
}

func (t *todoTool) update(ctx context.Context, params map[string]any) (*ExecuteResult, error) {
	var items []ToolTodoItem
	if ok, err := decodeParamInto(params, "items", &items); err != nil {
		return nil, fmt.Errorf("invalid items: %w", err)
	} else if !ok || len(items) == 0 {
		return nil, fmt.Errorf("items must contain at least one todo")
	}
	inProgress := 0
	allCompleted := true
	seen := make(map[string]struct{}, len(items))
	for i := range items {
		items[i].ID = strings.TrimSpace(items[i].ID)
		items[i].Content = strings.TrimSpace(items[i].Content)
		items[i].Status = strings.TrimSpace(items[i].Status)
		if items[i].ID == "" {
			return nil, fmt.Errorf("items[%d].id is required", i)
		}
		if _, ok := seen[items[i].ID]; ok {
			return nil, fmt.Errorf("duplicate todo id: %s", items[i].ID)
		}
		seen[items[i].ID] = struct{}{}
		if items[i].Content == "" {
			return nil, fmt.Errorf("items[%d].content is required", i)
		}
		switch items[i].Status {
		case "pending":
			allCompleted = false
		case "in_progress":
			inProgress++
			allCompleted = false
		case "completed":
		default:
			return nil, fmt.Errorf("items[%d].status must be pending, in_progress, or completed", i)
		}
	}
	if !allCompleted && inProgress != 1 {
		return nil, fmt.Errorf("todo updates must have exactly one in_progress item unless all items are completed")
	}
	state := todoStateFromContext(ctx)
	state.mu.Lock()
	defer state.mu.Unlock()
	state.items = append([]ToolTodoItem(nil), items...)
	state.note = paramStringTrim(params, "note")
	return &ExecuteResult{Output: "Todo list updated.\n" + formatToolTodosLocked(state)}, nil
}

func formatToolTodosLocked(state *TodoState) string {
	var out strings.Builder
	for _, item := range state.items {
		out.WriteString(fmt.Sprintf("[%s] %s %s\n", item.Status, item.ID, item.Content))
	}
	if state.note != "" {
		out.WriteString("Note: " + state.note + "\n")
	}
	return strings.TrimSpace(out.String())
}
