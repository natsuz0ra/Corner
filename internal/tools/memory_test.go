package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	memorysvc "slimebot/internal/services/memory"
)

func TestMemoryToolWritesTargetsAndReturnsJSONState(t *testing.T) {
	service := memorysvc.NewService(memorysvc.NewStore(t.TempDir()), memorysvc.NewStaticSettings(memorysvc.DefaultConfig()))
	ctx := WithMemoryService(context.Background(), service)
	tool := &memoryTool{}

	result, err := tool.Execute(ctx, "add", map[string]any{
		"target":  "user",
		"content": "用户偏好中文回复。",
	})
	if err != nil {
		t.Fatalf("Execute add failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected tool error: %s", result.Error)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(result.Output), &body); err != nil {
		t.Fatalf("decode output failed: %v\n%s", err, result.Output)
	}
	if body["success"] != true || body["target"] != "user" {
		t.Fatalf("unexpected output: %#v", body)
	}
	if body["entryCount"].(float64) != 1 {
		t.Fatalf("entryCount = %#v", body["entryCount"])
	}

	read, err := tool.Execute(ctx, "read", map[string]any{"target": "user"})
	if err != nil {
		t.Fatalf("Execute read failed: %v", err)
	}
	if !strings.Contains(read.Output, "用户偏好中文回复") {
		t.Fatalf("read output missing entry: %s", read.Output)
	}
}

func TestMemoryToolValidatesServiceTargetAndContent(t *testing.T) {
	tool := &memoryTool{}
	if result, err := tool.Execute(context.Background(), "add", map[string]any{"target": "memory", "content": "fact"}); err != nil {
		t.Fatalf("Execute returned Go error: %v", err)
	} else if !strings.Contains(result.Error, "memory service") {
		t.Fatalf("expected missing service tool error, got %#v", result)
	}

	service := memorysvc.NewService(memorysvc.NewStore(t.TempDir()), memorysvc.NewStaticSettings(memorysvc.DefaultConfig()))
	ctx := WithMemoryService(context.Background(), service)
	if result, err := tool.Execute(ctx, "add", map[string]any{"target": "project", "content": "fact"}); err != nil {
		t.Fatalf("Execute returned Go error: %v", err)
	} else if !strings.Contains(result.Error, "target") {
		t.Fatalf("expected invalid target error, got %#v", result)
	}
	if result, err := tool.Execute(ctx, "add", map[string]any{"target": "memory"}); err != nil {
		t.Fatalf("Execute returned Go error: %v", err)
	} else if !strings.Contains(result.Error, "content") {
		t.Fatalf("expected missing content error, got %#v", result)
	}
}

func TestMemoryToolIsNotAllowedInPlanMode(t *testing.T) {
	if IsPlanModeAllowedFunction("memory__add") || IsPlanModeAllowedFunction("memory__read") {
		t.Fatal("memory tool must not be exposed in plan mode")
	}
}
