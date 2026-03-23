package memory

import (
	"context"
	"strings"
	"testing"
	"time"

	"slimebot/internal/domain"
	"slimebot/internal/repositories"
)

func TestParseMemoryFacts_RejectsInvalidPayload(t *testing.T) {
	if _, err := parseMemoryFacts("not-json"); err == nil {
		t.Fatal("expected parser error")
	}
}

func TestMemoryServiceSyncFactsAsync_EventuallyPersistsStructuredFact(t *testing.T) {
	repo := newTestRepo(t)
	svc := NewMemoryService(repo, nil)

	session, err := repo.CreateSession(context.Background(), "s")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	if _, err := addMessage(t, repo, session.ID, "user", "以后默认中文回复"); err != nil {
		t.Fatalf("add message failed: %v", err)
	}
	if _, err := addMessage(t, repo, session.ID, "assistant", "收到"); err != nil {
		t.Fatalf("add message failed: %v", err)
	}

	svc.SyncFactsAsync(session.ID, `{"facts":[{"memory_type":"preference","subject":"user","predicate":"reply_language","value":"zh-cn","summary":"用户偏好中文回复","confidence":0.96}]}`)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		fact, getErr := repo.FindActiveMemoryFact(context.Background(), session.ID, domain.MemoryTypePreference, "user", "reply_language")
		if getErr != nil {
			t.Fatalf("find fact failed: %v", getErr)
		}
		if fact != nil {
			if fact.Value != "zh-cn" {
				t.Fatalf("unexpected fact value: %s", fact.Value)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected structured fact persisted")
}

func TestMemoryServiceSyncFactsAsync_StalesConflictingPreference(t *testing.T) {
	repo := newTestRepo(t)
	svc := NewMemoryService(repo, nil)
	session, err := repo.CreateSession(context.Background(), "s")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	if _, err := repo.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      session.ID,
		MemoryType:     domain.MemoryTypePreference,
		Subject:        "user",
		Predicate:      "theme",
		Value:          "dark",
		Summary:        "用户偏好深色主题",
		Confidence:     0.89,
		SourceStartSeq: 1,
		SourceEndSeq:   2,
		LastSeenAt:     time.Now(),
	}); err != nil {
		t.Fatalf("seed fact failed: %v", err)
	}
	if _, err := addMessage(t, repo, session.ID, "user", "现在改成浅色主题"); err != nil {
		t.Fatalf("add message failed: %v", err)
	}

	svc.SyncFactsAsync(session.ID, `{"facts":[{"memory_type":"preference","subject":"user","predicate":"theme","value":"light","summary":"用户改为偏好浅色主题","confidence":0.97}]}`)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		active, getErr := repo.FindActiveMemoryFact(context.Background(), session.ID, domain.MemoryTypePreference, "user", "theme")
		if getErr != nil {
			t.Fatalf("find active failed: %v", getErr)
		}
		if active != nil && active.Value == "light" {
			rows, err := repo.GetSessionMemoriesBySessionIDs([]string{session.ID})
			if err != nil {
				t.Fatalf("list facts failed: %v", err)
			}
			staleSeen := false
			for _, row := range rows {
				if row.Predicate == "theme" && row.Value == "dark" && row.Status == domain.MemoryStatusStale {
					staleSeen = true
				}
			}
			if !staleSeen {
				t.Fatal("expected old preference to become stale")
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected preference conflict resolved")
}

func TestMemoryServiceBuildSessionMemoryContextForPrompt_GroupsByType(t *testing.T) {
	repo := newTestRepo(t)
	svc := NewMemoryService(repo, nil)
	session, err := repo.CreateSession(context.Background(), "s")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	seedFact(t, repo, session.ID, domain.MemoryTypeConstraint, "project", "budget", "low_cost", "项目要求极致省钱", 0.99, nil)
	seedFact(t, repo, session.ID, domain.MemoryTypeTask, "project", "current_task", "memory_refactor", "当前任务是重构记忆系统", 0.92, nil)
	seedFact(t, repo, session.ID, domain.MemoryTypePreference, "user", "reply_language", "zh-cn", "用户偏好中文回复", 0.95, nil)

	ctxText := svc.BuildSessionMemoryContextForPrompt(context.Background(), session.ID, nil)
	if !strings.Contains(ctxText, "<constraints>") {
		t.Fatalf("expected constraints group, got %q", ctxText)
	}
	if !strings.Contains(ctxText, "<active_tasks>") {
		t.Fatalf("expected active_tasks group, got %q", ctxText)
	}
	if !strings.Contains(ctxText, "<preferences>") {
		t.Fatalf("expected preferences group, got %q", ctxText)
	}
}

func TestMemoryServiceQueryForAgent_OutputIncludesTypeAndStatus(t *testing.T) {
	repo := newTestRepo(t)
	svc := NewMemoryService(repo, nil)
	seedFact(t, repo, "s2", domain.MemoryTypeConstraint, "project", "budget", "low_cost", "项目要求极致省钱并控制 token 成本", 0.98, nil)

	result, err := svc.QueryForAgent(context.Background(), "s1", "预算 省钱", 3)
	if err != nil {
		t.Fatalf("query for agent failed: %v", err)
	}
	if !strings.Contains(result.Output, "type=constraint") {
		t.Fatalf("expected type in output: %q", result.Output)
	}
	if !strings.Contains(result.Output, "status=active") {
		t.Fatalf("expected status in output: %q", result.Output)
	}
}

func seedFact(t *testing.T, repo *repositories.Repository, sessionID, memoryType, subject, predicate, value, summary string, confidence float64, expiresAt *time.Time) {
	t.Helper()
	if _, err := repo.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      sessionID,
		MemoryType:     memoryType,
		Subject:        subject,
		Predicate:      predicate,
		Value:          value,
		Summary:        summary,
		Confidence:     confidence,
		SourceStartSeq: 1,
		SourceEndSeq:   1,
		LastSeenAt:     time.Now(),
		ExpiresAt:      expiresAt,
	}); err != nil {
		t.Fatalf("seed fact failed: %v", err)
	}
}

func addMessage(_ *testing.T, repo *repositories.Repository, sessionID, role, content string) (*domain.Message, error) {
	return repo.AddMessageWithInput(context.Background(), domain.AddMessageInput{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
	})
}

func newTestRepo(t *testing.T) *repositories.Repository {
	t.Helper()
	return repositories.New(repositories.NewSQLiteDBTest(t, "services_test"))
}
