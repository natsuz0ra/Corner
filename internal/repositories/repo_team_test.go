package repositories

import (
	"context"
	"sync"
	"testing"
	"time"

	"slimebot/internal/domain"
)

func TestRepositoryGetOrCreateTeamRunConcurrent(t *testing.T) {
	repo := New(NewSQLiteDBTest(t, "team_create"))
	ctx := context.Background()

	const workers = 8
	ids := make(chan string, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			run, err := repo.GetOrCreateTeamRun(ctx, domain.CreateTeamRunInput{
				SessionID:   "session-1",
				RequestID:   "request-1",
				MaxMembers:  8,
				MaxParallel: 4,
			})
			if err != nil {
				errs <- err
				return
			}
			ids <- run.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		t.Fatalf("GetOrCreateTeamRun failed: %v", err)
	}
	firstID := ""
	for id := range ids {
		if firstID == "" {
			firstID = id
		}
		if id != firstID {
			t.Fatalf("expected one team id, got %q and %q", firstID, id)
		}
	}
	var count int64
	if err := repo.db.Model(&domain.TeamRun{}).Count(&count).Error; err != nil {
		t.Fatalf("count team runs: %v", err)
	}
	if count != 1 {
		t.Fatalf("team run count = %d, want 1", count)
	}
}

func TestRepositoryTeamMemberOrderAndUpdate(t *testing.T) {
	repo := New(NewSQLiteDBTest(t, "team_members"))
	ctx := context.Background()
	teamRun, err := repo.GetOrCreateTeamRun(ctx, domain.CreateTeamRunInput{
		SessionID: "session-1", RequestID: "request-1", MaxMembers: 8, MaxParallel: 4,
	})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	createdAt := time.Now().Add(-time.Minute)
	first, err := repo.CreateTeamMemberRun(ctx, domain.CreateTeamMemberRunInput{
		ID: "member-a", TeamRunID: teamRun.ID, ToolCallID: "tool-a", Title: "Research", Task: "Inspect", CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("create first member: %v", err)
	}
	_, err = repo.CreateTeamMemberRun(ctx, domain.CreateTeamMemberRunInput{
		ID: "member-b", TeamRunID: teamRun.ID, ToolCallID: "tool-b", Title: "Review", Task: "Check", CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("create second member: %v", err)
	}
	now := time.Now()
	if err := repo.UpdateTeamMemberRun(ctx, domain.UpdateTeamMemberRunInput{
		ID: first.ID, Status: domain.TeamMemberRunStatusRunning, SubagentRunID: "sub-1", ModelConfigID: "model-1", StartedAt: &now,
	}); err != nil {
		t.Fatalf("start member: %v", err)
	}

	members, err := repo.ListTeamMemberRuns(ctx, teamRun.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(members) != 2 || members[0].ID != "member-a" || members[1].ID != "member-b" {
		t.Fatalf("member order = %#v", members)
	}
	if members[0].Status != domain.TeamMemberRunStatusRunning || members[0].SubagentRunID != "sub-1" {
		t.Fatalf("updated member = %#v", members[0])
	}
}

func TestRepositoryBindsAndListsTeamHistory(t *testing.T) {
	repo := New(NewSQLiteDBTest(t, "team_history"))
	ctx := context.Background()
	teamRun, err := repo.GetOrCreateTeamRun(ctx, domain.CreateTeamRunInput{
		SessionID: "session-1", RequestID: "request-1", MaxMembers: 8, MaxParallel: 4,
	})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if _, err := repo.CreateTeamMemberRun(ctx, domain.CreateTeamMemberRunInput{
		ID: "member-1", TeamRunID: teamRun.ID, ToolCallID: "tool-1", Title: "Research", Task: "Inspect",
	}); err != nil {
		t.Fatalf("create member: %v", err)
	}
	if err := repo.BindTeamRunToAssistantMessage(ctx, "session-1", "request-1", "message-1"); err != nil {
		t.Fatalf("bind team: %v", err)
	}

	teams, err := repo.ListSessionTeamRunsByAssistantMessageIDs(ctx, "session-1", []string{"message-1"})
	if err != nil {
		t.Fatalf("list team history: %v", err)
	}
	members, err := repo.ListSessionTeamMemberRunsByAssistantMessageIDs(ctx, "session-1", []string{"message-1"})
	if err != nil {
		t.Fatalf("list member history: %v", err)
	}
	if len(teams) != 1 || teams[0].ID != teamRun.ID {
		t.Fatalf("team history = %#v", teams)
	}
	if len(members) != 1 || members[0].ID != "member-1" {
		t.Fatalf("member history = %#v", members)
	}
}

func TestRepositoryInterruptActiveTeamRunsKeepsTerminalRows(t *testing.T) {
	repo := New(NewSQLiteDBTest(t, "team_interrupt"))
	ctx := context.Background()
	active, err := repo.GetOrCreateTeamRun(ctx, domain.CreateTeamRunInput{
		SessionID: "session-1", RequestID: "request-active", MaxMembers: 8, MaxParallel: 4,
	})
	if err != nil {
		t.Fatalf("create active team: %v", err)
	}
	terminal, err := repo.GetOrCreateTeamRun(ctx, domain.CreateTeamRunInput{
		SessionID: "session-1", RequestID: "request-done", MaxMembers: 8, MaxParallel: 4,
	})
	if err != nil {
		t.Fatalf("create terminal team: %v", err)
	}
	finishedAt := time.Now()
	if err := repo.UpdateTeamRun(ctx, domain.UpdateTeamRunInput{
		ID: terminal.ID, Status: domain.TeamRunStatusSucceeded, FinishedAt: &finishedAt,
	}); err != nil {
		t.Fatalf("finish terminal team: %v", err)
	}
	member, err := repo.CreateTeamMemberRun(ctx, domain.CreateTeamMemberRunInput{
		ID: "member-active", TeamRunID: active.ID, ToolCallID: "tool-active", Title: "Active", Task: "Work",
	})
	if err != nil {
		t.Fatalf("create active member: %v", err)
	}
	if err := repo.InterruptActiveTeamRuns(ctx); err != nil {
		t.Fatalf("interrupt active runs: %v", err)
	}

	gotActive, err := repo.GetTeamRunByID(ctx, active.ID)
	if err != nil {
		t.Fatalf("get active team: %v", err)
	}
	gotTerminal, err := repo.GetTeamRunByID(ctx, terminal.ID)
	if err != nil {
		t.Fatalf("get terminal team: %v", err)
	}
	members, err := repo.ListTeamMemberRuns(ctx, active.ID)
	if err != nil {
		t.Fatalf("list active members: %v", err)
	}
	if gotActive.Status != domain.TeamRunStatusInterrupted {
		t.Fatalf("active status = %q", gotActive.Status)
	}
	if gotTerminal.Status != domain.TeamRunStatusSucceeded {
		t.Fatalf("terminal status = %q", gotTerminal.Status)
	}
	if len(members) != 1 || members[0].ID != member.ID || members[0].Status != domain.TeamMemberRunStatusInterrupted {
		t.Fatalf("interrupted members = %#v", members)
	}
}

func TestRepositoryDeleteSessionRemovesTeamRunsAndMembers(t *testing.T) {
	repo := New(NewSQLiteDBTest(t, "team_delete_session"))
	ctx := context.Background()
	session, err := repo.CreateSession(ctx, "Team session")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	teamRun, err := repo.GetOrCreateTeamRun(ctx, domain.CreateTeamRunInput{
		SessionID: session.ID, RequestID: "request-1", MaxMembers: 8, MaxParallel: 4,
	})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if _, err := repo.CreateTeamMemberRun(ctx, domain.CreateTeamMemberRunInput{
		TeamRunID: teamRun.ID, ToolCallID: "tool-1", Title: "Research", Task: "Inspect",
	}); err != nil {
		t.Fatalf("create member: %v", err)
	}

	if err := repo.DeleteSession(ctx, session.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	var teamCount, memberCount int64
	if err := repo.db.Model(&domain.TeamRun{}).Count(&teamCount).Error; err != nil {
		t.Fatalf("count teams: %v", err)
	}
	if err := repo.db.Model(&domain.TeamMemberRun{}).Count(&memberCount).Error; err != nil {
		t.Fatalf("count members: %v", err)
	}
	if teamCount != 0 || memberCount != 0 {
		t.Fatalf("remaining teams=%d members=%d", teamCount, memberCount)
	}
}

func TestRepositoryEditMessagePrunesBoundTeamRunsAndMembers(t *testing.T) {
	repo := New(NewSQLiteDBTest(t, "team_edit_prune"))
	ctx := context.Background()
	session, err := repo.CreateSession(ctx, "Team session")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	userMessage, err := repo.AddMessageWithInput(ctx, domain.AddMessageInput{
		SessionID: session.ID, Role: "user", Content: "before",
	})
	if err != nil {
		t.Fatalf("create user message: %v", err)
	}
	assistantMessage, err := repo.AddMessageWithInput(ctx, domain.AddMessageInput{
		SessionID: session.ID, Role: "assistant", Content: "answer",
	})
	if err != nil {
		t.Fatalf("create assistant message: %v", err)
	}
	teamRun, err := repo.GetOrCreateTeamRun(ctx, domain.CreateTeamRunInput{
		SessionID: session.ID, RequestID: "request-1", MaxMembers: 8, MaxParallel: 4,
	})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if _, err := repo.CreateTeamMemberRun(ctx, domain.CreateTeamMemberRunInput{
		TeamRunID: teamRun.ID, ToolCallID: "tool-1", Title: "Research", Task: "Inspect",
	}); err != nil {
		t.Fatalf("create member: %v", err)
	}
	if err := repo.BindTeamRunToAssistantMessage(ctx, session.ID, "request-1", assistantMessage.ID); err != nil {
		t.Fatalf("bind team: %v", err)
	}

	if _, err := repo.UpdateUserMessageAndPruneAfter(ctx, session.ID, userMessage.ID, "after"); err != nil {
		t.Fatalf("edit user message: %v", err)
	}
	var teamCount, memberCount int64
	if err := repo.db.Model(&domain.TeamRun{}).Count(&teamCount).Error; err != nil {
		t.Fatalf("count teams: %v", err)
	}
	if err := repo.db.Model(&domain.TeamMemberRun{}).Count(&memberCount).Error; err != nil {
		t.Fatalf("count members: %v", err)
	}
	if teamCount != 0 || memberCount != 0 {
		t.Fatalf("remaining teams=%d members=%d", teamCount, memberCount)
	}
}
