package team_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"slimebot/internal/domain"
	"slimebot/internal/repositories"
	teamsvc "slimebot/internal/services/team"
)

func newTeamService(t *testing.T, options teamsvc.Options) (*teamsvc.Service, *repositories.Repository) {
	t.Helper()
	repo := repositories.New(repositories.NewSQLiteDBTest(t, "team_service_"+t.Name()))
	return teamsvc.NewService(repo, options), repo
}

func TestServiceUsesDefaultLimitsAndEnforcesMemberBudget(t *testing.T) {
	service, repo := newTeamService(t, teamsvc.Options{})
	ctx := context.Background()
	run, err := service.EnsureTeam(ctx, "session-1", "request-1")
	if err != nil {
		t.Fatalf("EnsureTeam failed: %v", err)
	}
	if run.MaxMembers != 8 || run.MaxParallel != 4 {
		t.Fatalf("limits = %d/%d", run.MaxMembers, run.MaxParallel)
	}
	for index := 0; index < 8; index++ {
		toolCallID := fmt.Sprintf("tool-%d", index)
		member, err := service.EnsureMember(ctx, run.ID, toolCallID, "Worker", "Task", "model-1")
		if err != nil {
			t.Fatalf("EnsureMember(%d) failed: %v", index, err)
		}
		if member.Status != domain.TeamMemberRunStatusQueued {
			t.Fatalf("member status = %q", member.Status)
		}
	}
	first, err := service.EnsureMember(ctx, run.ID, "tool-0", "Changed", "Changed", "model-2")
	if err != nil {
		t.Fatalf("idempotent EnsureMember failed: %v", err)
	}
	if first.ToolCallID != "tool-0" {
		t.Fatalf("idempotent member = %#v", first)
	}
	if _, err := service.EnsureMember(ctx, run.ID, "tool-8", "Overflow", "Task", "model-1"); !errors.Is(err, teamsvc.ErrMemberLimitExceeded) {
		t.Fatalf("overflow error = %v", err)
	}
	members, err := repo.ListTeamMemberRuns(ctx, run.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(members) != 8 {
		t.Fatalf("member count = %d", len(members))
	}
}

func TestServiceMemberLifecycleRejectsInvalidTransition(t *testing.T) {
	service, _ := newTeamService(t, teamsvc.Options{})
	ctx := context.Background()
	run, _ := service.EnsureTeam(ctx, "session-1", "request-1")
	member, _ := service.EnsureMember(ctx, run.ID, "tool-1", "Worker", "Task", "model-1")

	running, err := service.StartMember(ctx, member.ID, "subagent-1", "model-2")
	if err != nil {
		t.Fatalf("StartMember failed: %v", err)
	}
	if running.Status != domain.TeamMemberRunStatusRunning || running.SubagentRunID != "subagent-1" || running.StartedAt == nil {
		t.Fatalf("running member = %#v", running)
	}
	finished, err := service.FinishMember(ctx, member.ID, teamsvc.MemberResult{Answer: "done"})
	if err != nil {
		t.Fatalf("FinishMember failed: %v", err)
	}
	if finished.Status != domain.TeamMemberRunStatusSucceeded || finished.Answer != "done" || finished.FinishedAt == nil {
		t.Fatalf("finished member = %#v", finished)
	}
	if _, err := service.StartMember(ctx, member.ID, "subagent-2", "model-2"); !errors.Is(err, teamsvc.ErrInvalidTransition) {
		t.Fatalf("restart error = %v", err)
	}
}

func TestServiceFinalizeAggregatesMemberStatuses(t *testing.T) {
	tests := []struct {
		name     string
		results  []teamsvc.MemberResult
		expected string
	}{
		{name: "all success", results: []teamsvc.MemberResult{{Answer: "a"}, {Answer: "b"}}, expected: domain.TeamRunStatusSucceeded},
		{name: "partial failure", results: []teamsvc.MemberResult{{Answer: "a"}, {Err: errors.New("boom")}}, expected: domain.TeamRunStatusPartialFailed},
		{name: "all failure", results: []teamsvc.MemberResult{{Err: errors.New("a")}, {Err: errors.New("b")}}, expected: domain.TeamRunStatusFailed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, _ := newTeamService(t, teamsvc.Options{})
			ctx := context.Background()
			run, _ := service.EnsureTeam(ctx, "session-1", "request-1")
			for index, result := range test.results {
				member, _ := service.EnsureMember(ctx, run.ID, fmt.Sprintf("tool-%d", index), "Worker", "Task", "model")
				if _, err := service.StartMember(ctx, member.ID, fmt.Sprintf("sub-%d", index), "model"); err != nil {
					t.Fatalf("start member: %v", err)
				}
				if _, err := service.FinishMember(ctx, member.ID, result); err != nil {
					t.Fatalf("finish member: %v", err)
				}
			}
			final, err := service.Finalize(ctx, run.ID, teamsvc.FinalizeReasonCompleted)
			if err != nil {
				t.Fatalf("Finalize failed: %v", err)
			}
			if final.Status != test.expected || final.FinishedAt == nil {
				t.Fatalf("final team = %#v", final)
			}
		})
	}
}

func TestServiceCancelFinalizesActiveMembersAndTeam(t *testing.T) {
	service, repo := newTeamService(t, teamsvc.Options{})
	ctx := context.Background()
	run, _ := service.EnsureTeam(ctx, "session-1", "request-1")
	queued, _ := service.EnsureMember(ctx, run.ID, "tool-queued", "Queued", "Task", "model")
	running, _ := service.EnsureMember(ctx, run.ID, "tool-running", "Running", "Task", "model")
	if _, err := service.StartMember(ctx, running.ID, "sub-running", "model"); err != nil {
		t.Fatalf("start member: %v", err)
	}

	final, err := service.Finalize(ctx, run.ID, teamsvc.FinalizeReasonCanceled)
	if err != nil {
		t.Fatalf("Finalize canceled failed: %v", err)
	}
	if final.Status != domain.TeamRunStatusCanceled {
		t.Fatalf("team status = %q", final.Status)
	}
	members, err := repo.ListTeamMemberRuns(ctx, run.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	for _, member := range members {
		if member.ID == queued.ID || member.ID == running.ID {
			if member.Status != domain.TeamMemberRunStatusCanceled || member.FinishedAt == nil {
				t.Fatalf("canceled member = %#v", member)
			}
		}
	}
}

func TestServiceRecoverInterruptedDelegatesToStore(t *testing.T) {
	service, repo := newTeamService(t, teamsvc.Options{})
	ctx := context.Background()
	run, _ := service.EnsureTeam(ctx, "session-1", "request-1")
	member, _ := service.EnsureMember(ctx, run.ID, "tool-1", "Worker", "Task", "model")
	if _, err := service.StartMember(ctx, member.ID, "sub-1", "model"); err != nil {
		t.Fatalf("start member: %v", err)
	}
	if err := service.RecoverInterrupted(ctx); err != nil {
		t.Fatalf("RecoverInterrupted failed: %v", err)
	}
	gotRun, _ := repo.GetTeamRunByID(ctx, run.ID)
	gotMember, _ := repo.GetTeamMemberRunByID(ctx, member.ID)
	if gotRun.Status != domain.TeamRunStatusInterrupted || gotMember.Status != domain.TeamMemberRunStatusInterrupted {
		t.Fatalf("recovered states = %q/%q", gotRun.Status, gotMember.Status)
	}
}
