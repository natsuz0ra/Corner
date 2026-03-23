package repositories

import (
	"context"
	"testing"
	"time"

	"slimebot/internal/domain"
)

func TestCreateMemoryFact_ReusesSameIdentityAsUpdate(t *testing.T) {
	repo := newSessionMemoryRepo(t)

	first, err := repo.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      "s1",
		MemoryType:     domain.MemoryTypePreference,
		Subject:        "user",
		Predicate:      "reply_language",
		Value:          "zh-cn",
		Summary:        "用户偏好中文回复",
		Confidence:     0.91,
		SourceStartSeq: 1,
		SourceEndSeq:   4,
		LastSeenAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("create fact failed: %v", err)
	}

	updated, err := repo.UpdateMemoryFact(domain.MemoryFactUpdateInput{
		ID:             first.ID,
		SessionID:      "s1",
		Value:          "zh-cn",
		Summary:        "用户持续偏好中文回复",
		Confidence:     0.97,
		SourceStartSeq: 1,
		SourceEndSeq:   8,
		LastSeenAt:     time.Now().Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("update fact failed: %v", err)
	}
	if updated.Version != first.Version+1 {
		t.Fatalf("expected version increment, got %d -> %d", first.Version, updated.Version)
	}

	found, err := repo.FindActiveMemoryFact(context.Background(), "s1", domain.MemoryTypePreference, "user", "reply_language")
	if err != nil {
		t.Fatalf("find active fact failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected active fact")
	}
	if found.ID != first.ID {
		t.Fatalf("expected same fact id, got %s want %s", found.ID, first.ID)
	}
	if found.Summary != "用户持续偏好中文回复" {
		t.Fatalf("unexpected summary: %s", found.Summary)
	}
}

func TestMarkMemoryFactStale_DeactivatesPreviousConflict(t *testing.T) {
	repo := newSessionMemoryRepo(t)

	oldFact, err := repo.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      "s1",
		MemoryType:     domain.MemoryTypePreference,
		Subject:        "user",
		Predicate:      "theme",
		Value:          "dark",
		Summary:        "用户偏好深色主题",
		Confidence:     0.88,
		SourceStartSeq: 1,
		SourceEndSeq:   2,
		LastSeenAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("create old fact failed: %v", err)
	}

	if err := repo.MarkMemoryFactStale(oldFact.ID, "s1"); err != nil {
		t.Fatalf("mark stale failed: %v", err)
	}

	newFact, err := repo.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      "s1",
		MemoryType:     domain.MemoryTypePreference,
		Subject:        "user",
		Predicate:      "theme",
		Value:          "light",
		Summary:        "用户改为偏好浅色主题",
		Confidence:     0.95,
		SourceStartSeq: 3,
		SourceEndSeq:   6,
		LastSeenAt:     time.Now().Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("create new fact failed: %v", err)
	}

	active, err := repo.FindActiveMemoryFact(context.Background(), "s1", domain.MemoryTypePreference, "user", "theme")
	if err != nil {
		t.Fatalf("find active fact failed: %v", err)
	}
	if active == nil || active.ID != newFact.ID {
		t.Fatalf("expected latest active fact, got %#v", active)
	}

	all, err := repo.ListMemoryFactsForPrompt(context.Background(), "s1", 10, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("list facts for prompt failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected only one active prompt fact, got %d", len(all))
	}
	if all[0].Value != "light" {
		t.Fatalf("unexpected active value: %s", all[0].Value)
	}
}

func TestListMemoryFactsForPrompt_ExcludesExpiredTasks(t *testing.T) {
	repo := newSessionMemoryRepo(t)
	now := time.Now()

	_, err := repo.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      "s1",
		MemoryType:     domain.MemoryTypeTask,
		Subject:        "project",
		Predicate:      "current_task",
		Value:          "cleanup",
		Summary:        "当前任务是清理历史脚本",
		Confidence:     0.87,
		SourceStartSeq: 1,
		SourceEndSeq:   2,
		LastSeenAt:     now.Add(-2 * time.Hour),
		ExpiresAt:      ptrTime(now.Add(-time.Minute)),
	})
	if err != nil {
		t.Fatalf("create expired task failed: %v", err)
	}

	validFact, err := repo.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      "s1",
		MemoryType:     domain.MemoryTypeConstraint,
		Subject:        "project",
		Predicate:      "budget",
		Value:          "low_cost",
		Summary:        "项目要求极致省钱",
		Confidence:     0.99,
		SourceStartSeq: 3,
		SourceEndSeq:   4,
		LastSeenAt:     now,
	})
	if err != nil {
		t.Fatalf("create valid fact failed: %v", err)
	}

	all, err := repo.ListMemoryFactsForPrompt(context.Background(), "s1", 10, now)
	if err != nil {
		t.Fatalf("list prompt facts failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected only one non-expired fact, got %d", len(all))
	}
	if all[0].ID != validFact.ID {
		t.Fatalf("unexpected remaining fact %s", all[0].ID)
	}
}

func TestSearchMemoryFacts_PrioritizesActiveHighConfidence(t *testing.T) {
	repo := newSessionMemoryRepo(t)
	now := time.Now()

	active, err := repo.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      "s2",
		MemoryType:     domain.MemoryTypeConstraint,
		Subject:        "project",
		Predicate:      "budget",
		Value:          "low_cost",
		Summary:        "项目要求极致省钱并控制 token 成本",
		Confidence:     0.98,
		SourceStartSeq: 1,
		SourceEndSeq:   2,
		LastSeenAt:     now,
	})
	if err != nil {
		t.Fatalf("create active fact failed: %v", err)
	}

	stale, err := repo.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      "s3",
		MemoryType:     domain.MemoryTypeConstraint,
		Subject:        "project",
		Predicate:      "budget",
		Value:          "high_cost",
		Summary:        "旧项目接受高成本预算",
		Confidence:     0.50,
		SourceStartSeq: 1,
		SourceEndSeq:   1,
		LastSeenAt:     now.Add(-24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create stale seed failed: %v", err)
	}
	if err := repo.MarkMemoryFactStale(stale.ID, "s3"); err != nil {
		t.Fatalf("mark stale failed: %v", err)
	}

	hits, err := repo.SearchMemoryFacts(domain.MemoryFactSearchInput{
		Query:          "预算 成本 省钱",
		Limit:          5,
		Now:            now,
		ExcludeSession: "s1",
	})
	if err != nil {
		t.Fatalf("search facts failed: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected hits")
	}
	if hits[0].Fact.ID != active.ID {
		t.Fatalf("expected active fact ranked first, got %s", hits[0].Fact.ID)
	}
	if hits[0].Fact.Status != domain.MemoryStatusActive {
		t.Fatalf("unexpected status: %s", hits[0].Fact.Status)
	}
}

func ptrTime(v time.Time) *time.Time {
	return &v
}

func newSessionMemoryRepo(t *testing.T) *Repository {
	t.Helper()
	return New(NewSQLiteDBTest(t, "memory_repo"))
}
