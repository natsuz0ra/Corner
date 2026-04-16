package repositories

import (
	"context"
	"testing"
	"time"

	"slimebot/internal/domain"
)

func TestUpsertToolCallStart_UpdatesExistingRow(t *testing.T) {
	repo := New(NewSQLiteDBTest(t, "repo_tool_calls_upsert"))
	session, err := repo.CreateSession(context.Background(), "s")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	first := domain.ToolCallStartRecordInput{
		SessionID:        session.ID,
		RequestID:        "r1",
		ToolCallID:       "tc1",
		ToolName:         "exec",
		Command:          "run",
		Params:           map[string]string{"cmd": "echo 1"},
		Status:           "pending",
		RequiresApproval: true,
		StartedAt:        time.Now().Add(-1 * time.Minute),
	}
	if err := repo.UpsertToolCallStart(context.Background(), first); err != nil {
		t.Fatalf("first upsert failed: %v", err)
	}

	second := first
	second.ToolName = "web_search"
	second.Command = "search"
	second.Status = "executing"
	second.RequiresApproval = false
	second.Params = map[string]string{"q": "golang"}
	second.StartedAt = time.Now()
	if err := repo.UpsertToolCallStart(context.Background(), second); err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}

	var rows []domain.ToolCallRecord
	if err := repo.db.Where("session_id = ?", session.ID).Order("started_at asc").Find(&rows).Error; err != nil {
		t.Fatalf("list tool call records failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].ToolName != "web_search" || rows[0].Command != "search" || rows[0].Status != "executing" {
		t.Fatalf("row was not updated: %+v", rows[0])
	}
}

func TestUpsertToolCallStart_PersistsParentAndSubagentRun(t *testing.T) {
	repo := New(NewSQLiteDBTest(t, "repo_tool_calls_parent"))
	session, err := repo.CreateSession(context.Background(), "s2")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	parentID := "parent-tc"
	in := domain.ToolCallStartRecordInput{
		SessionID:        session.ID,
		RequestID:        "r2",
		ToolCallID:       "child-tc",
		ToolName:         "web_search",
		Command:          "search",
		Params:           map[string]string{"query": "x"},
		Status:           "executing",
		RequiresApproval: false,
		StartedAt:        time.Now(),
		ParentToolCallID: parentID,
		SubagentRunID:    "run-uuid-1",
	}
	if err := repo.UpsertToolCallStart(context.Background(), in); err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	var row domain.ToolCallRecord
	if err := repo.db.Where("session_id = ? AND tool_call_id = ?", session.ID, "child-tc").First(&row).Error; err != nil {
		t.Fatalf("load row: %v", err)
	}
	if row.ParentToolCallID != parentID {
		t.Fatalf("parent_tool_call_id: got %q want %q", row.ParentToolCallID, parentID)
	}
	if row.SubagentRunID != "run-uuid-1" {
		t.Fatalf("subagent_run_id: got %q", row.SubagentRunID)
	}

	in2 := in
	in2.ToolName = "exec"
	in2.ParentToolCallID = parentID + "-updated"
	in2.SubagentRunID = "run-uuid-2"
	if err := repo.UpsertToolCallStart(context.Background(), in2); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if err := repo.db.Where("session_id = ? AND tool_call_id = ?", session.ID, "child-tc").First(&row).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if row.ParentToolCallID != parentID+"-updated" || row.SubagentRunID != "run-uuid-2" {
		t.Fatalf("conflict update lost parent/subagent: %+v", row)
	}
}
