package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"slimebot/internal/domain"
	schedulesvc "slimebot/internal/services/schedule"
)

func TestScheduleToolCreateAndList(t *testing.T) {
	store := schedulesvc.NewMemoryStore()
	service := schedulesvc.NewService(store, nil, schedulesvc.Options{Now: func() time.Time {
		return time.Date(2026, 5, 23, 9, 0, 0, 0, time.Local)
	}})
	ctx := WithScheduleService(context.Background(), service)

	tool := &scheduleTool{}
	result, err := tool.Execute(ctx, "create", map[string]any{
		"name":             "每日摘要",
		"prompt":           "总结项目状态",
		"session_id":       "session-1",
		"schedule_kind":    "interval",
		"interval_minutes": float64(60),
	})
	if err != nil {
		t.Fatalf("Execute create failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("create returned error: %s", result.Error)
	}
	var created domain.ScheduledTask
	if err := json.Unmarshal([]byte(result.Output), &created); err != nil {
		t.Fatalf("decode create output failed: %v", err)
	}
	if created.Name != "每日摘要" {
		t.Fatalf("name = %q, want 每日摘要", created.Name)
	}

	result, err = tool.Execute(ctx, "list", map[string]any{})
	if err != nil {
		t.Fatalf("Execute list failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("list returned error: %s", result.Error)
	}
	var listed []domain.ScheduledTask
	if err := json.Unmarshal([]byte(result.Output), &listed); err != nil {
		t.Fatalf("decode list output failed: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("listed tasks = %+v, want created task", listed)
	}
}

func TestScheduleToolCreateDefaultsToCurrentSession(t *testing.T) {
	store := schedulesvc.NewMemoryStore()
	service := schedulesvc.NewService(store, nil, schedulesvc.Options{Now: func() time.Time {
		return time.Date(2026, 5, 23, 9, 0, 0, 0, time.Local)
	}})
	ctx := WithCurrentSessionID(WithScheduleService(context.Background(), service), "current-session")

	tool := &scheduleTool{}
	result, err := tool.Execute(ctx, "create", map[string]any{
		"name":             "提醒",
		"prompt":           "提醒我喝水",
		"schedule_kind":    "interval",
		"interval_minutes": float64(60),
	})
	if err != nil {
		t.Fatalf("Execute create failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("create returned error: %s", result.Error)
	}
	var created domain.ScheduledTask
	if err := json.Unmarshal([]byte(result.Output), &created); err != nil {
		t.Fatalf("decode create output failed: %v", err)
	}
	if created.SessionID != "current-session" {
		t.Fatalf("session ID = %q, want current-session", created.SessionID)
	}
}
