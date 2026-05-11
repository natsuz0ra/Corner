package tools

import (
	"context"
	"fmt"
	"strings"
)

type skillsTool struct{}

func init() {
	Register(&skillsTool{})
}

func (s *skillsTool) Name() string { return "skills" }

func (s *skillsTool) Description() string {
	return "List and view available SlimeBot skills without activating them."
}

func (s *skillsTool) Commands() []Command {
	return []Command{
		{
			Name:        "list",
			Description: "List installed and discovered skills with id, name, enabled state, source, and description.",
		},
		{
			Name:        "view",
			Description: "View one enabled skill's SKILL.md content by name or id.",
			Params: []CommandParam{
				{Name: "name", Required: true, Description: "Skill name or id.", Example: "imagegen"},
			},
		},
	}
}

func (s *skillsTool) Execute(ctx context.Context, command string, params map[string]any) (*ExecuteResult, error) {
	switch command {
	case "list":
		return s.list(ctx, params)
	case "view":
		return s.view(ctx, params)
	default:
		return nil, fmt.Errorf("skills tool does not support command: %s", command)
	}
}

func (s *skillsTool) list(ctx context.Context, _ map[string]any) (*ExecuteResult, error) {
	runtime, ok := skillRuntimeFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("skill runtime is unavailable")
	}
	items, err := runtime.ListSkills()
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return &ExecuteResult{Output: "No skills found."}, nil
	}
	var out strings.Builder
	out.WriteString(fmt.Sprintf("Skills: %d\n", len(items)))
	for _, item := range items {
		state := "disabled"
		if item.Enabled {
			state = "enabled"
		}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = item.Name
		}
		out.WriteString(fmt.Sprintf("- %s (%s, %s): %s\n", id, state, item.SourceLabel, item.Description))
	}
	return &ExecuteResult{Output: strings.TrimSpace(out.String())}, nil
}

func (s *skillsTool) view(ctx context.Context, params map[string]any) (*ExecuteResult, error) {
	runtime, ok := skillRuntimeFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("skill runtime is unavailable")
	}
	name := paramStringTrim(params, "name")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	content, _, err := runtime.ActivateSkill(name, map[string]struct{}{})
	if err != nil {
		return nil, err
	}
	return &ExecuteResult{Output: content}, nil
}
