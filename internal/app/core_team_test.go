package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"slimebot/internal/config"
	"slimebot/internal/domain"
	"slimebot/internal/repositories"
)

func TestNewCoreWiresTeamServiceAndRecoversInterruptedRuns(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "storage", "data.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatalf("create storage directory: %v", err)
	}
	db, err := repositories.NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("open seed database: %v", err)
	}
	repo := repositories.New(db)
	ctx := context.Background()
	run, err := repo.GetOrCreateTeamRun(ctx, domain.CreateTeamRunInput{
		SessionID: "session-1", RequestID: "request-1", MaxMembers: 8, MaxParallel: 4,
	})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	member, err := repo.CreateTeamMemberRun(ctx, domain.CreateTeamMemberRunInput{
		TeamRunID: run.ID, ToolCallID: "tool-1", Title: "Worker", Task: "Task",
	})
	if err != nil {
		t.Fatalf("create member: %v", err)
	}
	if err := repo.Close(); err != nil {
		t.Fatalf("close seed database: %v", err)
	}

	core, err := NewCore(config.Config{
		DBPath: dbPath, SkillsRoot: filepath.Join(root, "skills"), ChatUploadRoot: filepath.Join(root, "uploads"), DefaultContextSize: 1_000_000,
	})
	if err != nil {
		t.Fatalf("NewCore failed: %v", err)
	}
	defer core.Close(context.Background())
	if core.TeamService == nil {
		t.Fatal("TeamService is nil")
	}
	gotRun, err := core.Repo.GetTeamRunByID(ctx, run.ID)
	if err != nil {
		t.Fatalf("get recovered team: %v", err)
	}
	gotMember, err := core.Repo.GetTeamMemberRunByID(ctx, member.ID)
	if err != nil {
		t.Fatalf("get recovered member: %v", err)
	}
	if gotRun.Status != domain.TeamRunStatusInterrupted || gotMember.Status != domain.TeamMemberRunStatusInterrupted {
		t.Fatalf("recovered statuses = %q/%q", gotRun.Status, gotMember.Status)
	}
}
