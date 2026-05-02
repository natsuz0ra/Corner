package repositories

import (
	"context"
	"testing"

	"slimebot/internal/domain"
)

func TestSessionContextSummaryLifecycle(t *testing.T) {
	repo := New(NewSQLiteDBTest(t, "context_summary"))
	ctx := context.Background()
	session, err := repo.CreateSession(ctx, "summary")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if _, err := repo.GetSessionContextSummary(ctx, session.ID, ""); err == nil {
		t.Fatal("expected missing summary to return an error")
	}

	first := &domain.SessionContextSummary{
		SessionID:               session.ID,
		ModelConfigID:           "",
		Summary:                 "first summary",
		SummarizedUntilSeq:      4,
		PreCompactTokenEstimate: 100,
	}
	if err := repo.UpsertSessionContextSummary(ctx, first); err != nil {
		t.Fatalf("UpsertSessionContextSummary first failed: %v", err)
	}
	got, err := repo.GetSessionContextSummary(ctx, session.ID, "")
	if err != nil {
		t.Fatalf("GetSessionContextSummary failed: %v", err)
	}
	if got.Summary != "first summary" || got.SummarizedUntilSeq != 4 {
		t.Fatalf("unexpected first summary: %+v", got)
	}

	second := &domain.SessionContextSummary{
		SessionID:               session.ID,
		ModelConfigID:           "",
		Summary:                 "second summary",
		SummarizedUntilSeq:      8,
		PreCompactTokenEstimate: 200,
	}
	if err := repo.UpsertSessionContextSummary(ctx, second); err != nil {
		t.Fatalf("UpsertSessionContextSummary second failed: %v", err)
	}
	got, err = repo.GetSessionContextSummary(ctx, session.ID, "")
	if err != nil {
		t.Fatalf("GetSessionContextSummary after update failed: %v", err)
	}
	if got.Summary != "second summary" || got.SummarizedUntilSeq != 8 {
		t.Fatalf("unexpected updated summary: %+v", got)
	}
}
