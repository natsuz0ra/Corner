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

func TestScheduleToolUpdateIntervalMinutesWithoutScheduleKind(t *testing.T) {
	store := schedulesvc.NewMemoryStore()
	service := schedulesvc.NewService(store, nil, schedulesvc.Options{Now: func() time.Time {
		return time.Date(2026, 5, 23, 9, 0, 0, 0, time.Local)
	}})
	ctx := WithScheduleService(context.Background(), service)
	tool := &scheduleTool{}
	created := mustCreateScheduledTask(t, tool, ctx, map[string]any{
		"name":             "轮询",
		"prompt":           "总结状态",
		"session_id":       "session-1",
		"schedule_kind":    "interval",
		"interval_minutes": float64(30),
	})

	result, err := tool.Execute(ctx, "update", map[string]any{
		"task_id":          created.ID,
		"interval_minutes": float64(120),
	})
	if err != nil {
		t.Fatalf("Execute update failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("update returned error: %s", result.Error)
	}
	var updated domain.ScheduledTask
	if err := json.Unmarshal([]byte(result.Output), &updated); err != nil {
		t.Fatalf("decode update output failed: %v", err)
	}
	if updated.IntervalMinutes != 120 {
		t.Fatalf("interval minutes = %d, want 120", updated.IntervalMinutes)
	}
}

func TestScheduleToolUpdateOneShotRunAtWithoutScheduleKind(t *testing.T) {
	store := schedulesvc.NewMemoryStore()
	service := schedulesvc.NewService(store, nil, schedulesvc.Options{Now: func() time.Time {
		return time.Date(2026, 5, 23, 9, 0, 0, 0, time.Local)
	}})
	ctx := WithScheduleService(context.Background(), service)
	tool := &scheduleTool{}
	created := mustCreateScheduledTask(t, tool, ctx, map[string]any{
		"name":          "提醒",
		"prompt":        "提醒我喝水",
		"session_id":    "session-1",
		"schedule_kind": "once",
		"run_at":        "2026-05-23T10:00:00+08:00",
	})

	result, err := tool.Execute(ctx, "update", map[string]any{
		"task_id": created.ID,
		"run_at":  "2026-05-23T11:00:00+08:00",
	})
	if err != nil {
		t.Fatalf("Execute update failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("update returned error: %s", result.Error)
	}
	var updated domain.ScheduledTask
	if err := json.Unmarshal([]byte(result.Output), &updated); err != nil {
		t.Fatalf("decode update output failed: %v", err)
	}
	want, _ := time.Parse(time.RFC3339, "2026-05-23T11:00:00+08:00")
	if updated.RunAt == nil || !updated.RunAt.Equal(want) {
		t.Fatalf("run_at = %s, want %s", updated.RunAt, want)
	}
}

func TestScheduleToolUpdateCronExprWithoutScheduleKind(t *testing.T) {
	store := schedulesvc.NewMemoryStore()
	service := schedulesvc.NewService(store, nil, schedulesvc.Options{Now: func() time.Time {
		return time.Date(2026, 5, 23, 9, 0, 0, 0, time.Local)
	}})
	ctx := WithScheduleService(context.Background(), service)
	tool := &scheduleTool{}
	created := mustCreateScheduledTask(t, tool, ctx, map[string]any{
		"name":          "工作日摘要",
		"prompt":        "总结状态",
		"session_id":    "session-1",
		"schedule_kind": "cron",
		"cron_expr":     "0 9 * * 1-5",
	})

	result, err := tool.Execute(ctx, "update", map[string]any{
		"task_id":   created.ID,
		"cron_expr": "30 18 * * 1-5",
	})
	if err != nil {
		t.Fatalf("Execute update failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("update returned error: %s", result.Error)
	}
	var updated domain.ScheduledTask
	if err := json.Unmarshal([]byte(result.Output), &updated); err != nil {
		t.Fatalf("decode update output failed: %v", err)
	}
	if updated.CronExpr != "30 18 * * 1-5" {
		t.Fatalf("cron_expr = %q, want 30 18 * * 1-5", updated.CronExpr)
	}
}

func TestScheduleToolUpdateScheduleKindRequiresFields(t *testing.T) {
	store := schedulesvc.NewMemoryStore()
	service := schedulesvc.NewService(store, nil, schedulesvc.Options{Now: func() time.Time {
		return time.Date(2026, 5, 23, 9, 0, 0, 0, time.Local)
	}})
	ctx := WithScheduleService(context.Background(), service)
	tool := &scheduleTool{}
	created := mustCreateScheduledTask(t, tool, ctx, map[string]any{
		"name":             "轮询",
		"prompt":           "总结状态",
		"session_id":       "session-1",
		"schedule_kind":    "interval",
		"interval_minutes": float64(30),
	})

	result, err := tool.Execute(ctx, "update", map[string]any{
		"task_id":       created.ID,
		"schedule_kind": "once",
	})
	if err != nil {
		t.Fatalf("Execute update failed: %v", err)
	}
	if result.Error == "" {
		t.Fatal("update should require run_at when changing to once schedule")
	}
}

func mustCreateScheduledTask(t *testing.T, tool *scheduleTool, ctx context.Context, params map[string]any) domain.ScheduledTask {
	t.Helper()
	result, err := tool.Execute(ctx, "create", params)
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
	return created
}
