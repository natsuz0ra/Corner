package schedule

import (
	"context"
	"testing"
	"time"

	"slimebot/internal/constants"
	"slimebot/internal/domain"
	"slimebot/internal/repositories"
)

func TestServiceCreateTaskComputesNextRunAndDefaults(t *testing.T) {
	db := repositories.NewSQLiteDBTest(t, "schedule_create")
	repo := repositories.New(db)
	svc := NewService(repo, nil, Options{Now: fixedNow})

	session, err := repo.CreateSession(context.Background(), "定时任务")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	task, err := svc.Create(context.Background(), CreateInput{
		Name:      "每日摘要",
		Prompt:    "每天总结项目状态",
		SessionID: session.ID,
		Schedule:  ScheduleSpec{Kind: ScheduleKindInterval, IntervalMinutes: 30},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if task.ID == "" {
		t.Fatal("task ID should be generated")
	}
	if task.SessionID != session.ID {
		t.Fatalf("session ID = %q, want %q", task.SessionID, session.ID)
	}
	if task.Status != domain.ScheduledTaskStatusScheduled {
		t.Fatalf("status = %q, want scheduled", task.Status)
	}
	if task.ApprovalMode != constants.ApprovalModeAuto {
		t.Fatalf("approval mode = %q, want auto", task.ApprovalMode)
	}
	want := fixedNow().Add(30 * time.Minute)
	if !task.NextRunAt.Equal(want) {
		t.Fatalf("next run = %s, want %s", task.NextRunAt, want)
	}
}

func TestServiceDueTasksFastForwardsStaleRecurringTasks(t *testing.T) {
	db := repositories.NewSQLiteDBTest(t, "schedule_due")
	repo := repositories.New(db)
	now := fixedNow()
	svc := NewService(repo, nil, Options{Now: func() time.Time { return now }})

	session, err := repo.CreateSession(context.Background(), "定时任务")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	task, err := svc.Create(context.Background(), CreateInput{
		Name:      "轮询",
		Prompt:    "检查一次状态",
		SessionID: session.ID,
		Schedule:  ScheduleSpec{Kind: ScheduleKindInterval, IntervalMinutes: 10},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	stale := now.Add(-45 * time.Minute)
	if err := repo.UpdateScheduledTask(context.Background(), task.ID, map[string]any{"next_run_at": stale}); err != nil {
		t.Fatalf("UpdateScheduledTask failed: %v", err)
	}

	due, err := svc.DueTasks(context.Background())
	if err != nil {
		t.Fatalf("DueTasks failed: %v", err)
	}
	if len(due) != 0 {
		t.Fatalf("due tasks = %d, want 0 after stale fast-forward", len(due))
	}
	got, err := repo.GetScheduledTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetScheduledTask failed: %v", err)
	}
	if !got.NextRunAt.After(now) {
		t.Fatalf("next run should be fast-forwarded after now, got %s", got.NextRunAt)
	}
}

func TestServiceRestoreRunningRecurringTask(t *testing.T) {
	db := repositories.NewSQLiteDBTest(t, "schedule_restore_recurring")
	repo := repositories.New(db)
	now := fixedNow()
	svc := NewService(repo, nil, Options{Now: func() time.Time { return now }})

	session, err := repo.CreateSession(context.Background(), "定时任务")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	task, err := svc.Create(context.Background(), CreateInput{
		Name:      "轮询",
		Prompt:    "检查一次状态",
		SessionID: session.ID,
		Schedule:  ScheduleSpec{Kind: ScheduleKindInterval, IntervalMinutes: 10},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	nextRun := now.Add(-time.Minute)
	if err := repo.UpdateScheduledTask(context.Background(), task.ID, map[string]any{
		"status":      domain.ScheduledTaskStatusRunning,
		"next_run_at": &nextRun,
	}); err != nil {
		t.Fatalf("UpdateScheduledTask failed: %v", err)
	}

	if err := svc.RestoreRunningTasks(context.Background()); err != nil {
		t.Fatalf("RestoreRunningTasks failed: %v", err)
	}

	got, err := repo.GetScheduledTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetScheduledTask failed: %v", err)
	}
	if got.Status != domain.ScheduledTaskStatusScheduled {
		t.Fatalf("status = %q, want scheduled", got.Status)
	}
	if got.NextRunAt == nil || !got.NextRunAt.Equal(nextRun) {
		t.Fatalf("next run = %s, want preserved %s", got.NextRunAt, nextRun)
	}
}

func TestServiceRestoreRunningOneShotTaskIsDueAgain(t *testing.T) {
	db := repositories.NewSQLiteDBTest(t, "schedule_restore_oneshot")
	repo := repositories.New(db)
	now := fixedNow()
	svc := NewService(repo, nil, Options{Now: func() time.Time { return now }})

	session, err := repo.CreateSession(context.Background(), "定时任务")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	task, err := svc.Create(context.Background(), CreateInput{
		Name:      "一次提醒",
		Prompt:    "提醒我喝水",
		SessionID: session.ID,
		Schedule:  ScheduleSpec{Kind: ScheduleKindOnce, RunAt: now.Add(-time.Minute)},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.UpdateScheduledTask(context.Background(), task.ID, map[string]any{
		"status": domain.ScheduledTaskStatusRunning,
	}); err != nil {
		t.Fatalf("UpdateScheduledTask failed: %v", err)
	}

	if err := svc.RestoreRunningTasks(context.Background()); err != nil {
		t.Fatalf("RestoreRunningTasks failed: %v", err)
	}
	due, err := svc.DueTasks(context.Background())
	if err != nil {
		t.Fatalf("DueTasks failed: %v", err)
	}
	if len(due) != 1 || due[0].ID != task.ID {
		t.Fatalf("due tasks = %+v, want restored one-shot task", due)
	}
}

func TestServiceMarkRunCompletesOneShotTask(t *testing.T) {
	db := repositories.NewSQLiteDBTest(t, "schedule_oneshot")
	repo := repositories.New(db)
	now := fixedNow()
	svc := NewService(repo, nil, Options{Now: func() time.Time { return now }})

	session, err := repo.CreateSession(context.Background(), "定时任务")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	task, err := svc.Create(context.Background(), CreateInput{
		Name:      "一次提醒",
		Prompt:    "提醒我喝水",
		SessionID: session.ID,
		Schedule:  ScheduleSpec{Kind: ScheduleKindOnce, RunAt: now.Add(time.Minute)},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := svc.MarkRunComplete(context.Background(), task.ID, RunResult{
		Success: true,
		Answer:  "完成",
	}); err != nil {
		t.Fatalf("MarkRunComplete failed: %v", err)
	}
	got, err := repo.GetScheduledTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetScheduledTask failed: %v", err)
	}
	if got.Status != domain.ScheduledTaskStatusCompleted {
		t.Fatalf("status = %q, want completed", got.Status)
	}
	if got.NextRunAt != nil {
		t.Fatalf("next run = %s, want nil", got.NextRunAt)
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 5, 23, 9, 0, 0, 0, time.Local)
}
