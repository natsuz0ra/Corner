package repositories

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"slimebot/backend/internal/models"
)

func TestUpsertSessionMemoryIfNewer_MonotonicBySourceMessageCount(t *testing.T) {
	repo := newSessionMemoryRepo(t)

	updated, err := repo.UpsertSessionMemoryIfNewer(SessionMemoryUpsertInput{
		SessionID:          "s1",
		Summary:            "summary-v1",
		Keywords:           []string{"v1"},
		SourceMessageCount: 5,
	})
	if err != nil {
		t.Fatalf("initial upsert failed: %v", err)
	}
	if !updated {
		t.Fatal("expected initial upsert to update")
	}

	updated, err = repo.UpsertSessionMemoryIfNewer(SessionMemoryUpsertInput{
		SessionID:          "s1",
		Summary:            "summary-stale",
		Keywords:           []string{"stale"},
		SourceMessageCount: 3,
	})
	if err != nil {
		t.Fatalf("stale upsert failed: %v", err)
	}
	if updated {
		t.Fatal("expected stale upsert to be ignored")
	}

	item, err := repo.GetSessionMemory("s1")
	if err != nil {
		t.Fatalf("get memory failed: %v", err)
	}
	if item == nil {
		t.Fatal("expected existing memory row")
	}
	if item.Summary != "summary-v1" {
		t.Fatalf("stale write should not override summary, got=%s", item.Summary)
	}
	if item.SourceMessageCount != 5 {
		t.Fatalf("stale write should keep source_message_count=5, got=%d", item.SourceMessageCount)
	}

	updated, err = repo.UpsertSessionMemoryIfNewer(SessionMemoryUpsertInput{
		SessionID:          "s1",
		Summary:            "summary-v2",
		Keywords:           []string{"v2"},
		SourceMessageCount: 8,
	})
	if err != nil {
		t.Fatalf("newer upsert failed: %v", err)
	}
	if !updated {
		t.Fatal("expected newer upsert to update")
	}

	item, err = repo.GetSessionMemory("s1")
	if err != nil {
		t.Fatalf("get memory failed: %v", err)
	}
	if item.Summary != "summary-v2" {
		t.Fatalf("expected updated summary-v2, got=%s", item.Summary)
	}
	if item.SourceMessageCount != 8 {
		t.Fatalf("expected source_message_count=8, got=%d", item.SourceMessageCount)
	}
}

func newSessionMemoryRepo(t *testing.T) *Repository {
	t.Helper()

	dsn := fmt.Sprintf("file:memory_repo_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Session{},
		&models.Message{},
		&models.SessionMemory{},
		&models.ToolCallRecord{},
		&models.AppSetting{},
		&models.LLMConfig{},
		&models.MCPConfig{},
		&models.Skill{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return New(db)
}
