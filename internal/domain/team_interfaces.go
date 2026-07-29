package domain

import "context"

// TeamStore persists Agent Team lifecycle data independently from execution logic.
type TeamStore interface {
	GetOrCreateTeamRun(ctx context.Context, input CreateTeamRunInput) (*TeamRun, error)
	GetTeamRunByID(ctx context.Context, id string) (*TeamRun, error)
	CreateTeamMemberRun(ctx context.Context, input CreateTeamMemberRunInput) (*TeamMemberRun, error)
	ListTeamMemberRuns(ctx context.Context, teamRunID string) ([]TeamMemberRun, error)
	UpdateTeamMemberRun(ctx context.Context, input UpdateTeamMemberRunInput) error
	UpdateTeamRun(ctx context.Context, input UpdateTeamRunInput) error
	BindTeamRunToAssistantMessage(ctx context.Context, sessionID, requestID, assistantMessageID string) error
	InterruptActiveTeamRuns(ctx context.Context) error
	ListSessionTeamRunsByAssistantMessageIDs(ctx context.Context, sessionID string, messageIDs []string) ([]TeamRun, error)
	ListSessionTeamMemberRunsByAssistantMessageIDs(ctx context.Context, sessionID string, messageIDs []string) ([]TeamMemberRun, error)
}
