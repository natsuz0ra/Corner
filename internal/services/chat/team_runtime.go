package chat

import (
	"context"
	"sync"

	"slimebot/internal/domain"
	teamsvc "slimebot/internal/services/team"
)

type teamRuntime struct {
	mu        sync.Mutex
	service   *teamsvc.Service
	sessionID string
	requestID string
	run       *domain.TeamRun
}

func newTeamRuntime(service *teamsvc.Service, sessionID, requestID string) *teamRuntime {
	if service == nil {
		return nil
	}
	return &teamRuntime{service: service, sessionID: sessionID, requestID: requestID}
}

func (r *teamRuntime) reserveMember(ctx context.Context, toolCallID, title, task, modelConfigID string, callbacks AgentCallbacks) (*domain.TeamMemberRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.run == nil {
		run, err := r.service.EnsureTeam(ctx, r.sessionID, r.requestID)
		if err != nil {
			return nil, err
		}
		r.run = run
		if callbacks.OnTeamStart != nil {
			if err := callbacks.OnTeamStart(*run); err != nil {
				return nil, err
			}
		}
	}
	member, err := r.service.EnsureMember(ctx, r.run.ID, toolCallID, title, task, modelConfigID)
	if err != nil {
		return nil, err
	}
	if callbacks.OnTeamMemberQueued != nil && member.Status == domain.TeamMemberRunStatusQueued {
		if err := callbacks.OnTeamMemberQueued(*member); err != nil {
			return nil, err
		}
	}
	return member, nil
}

func (r *teamRuntime) startMember(ctx context.Context, memberRunID, subagentRunID, modelConfigID string) (*domain.TeamMemberRun, error) {
	return r.service.StartMember(ctx, memberRunID, subagentRunID, modelConfigID)
}

func (r *teamRuntime) finishMember(ctx context.Context, memberRunID string, result teamsvc.MemberResult) (*domain.TeamMemberRun, error) {
	return r.service.FinishMember(ctx, memberRunID, result)
}

func (r *teamRuntime) finalize(ctx context.Context, assistantMessageID string, canceled bool, callbacks AgentCallbacks) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.run == nil {
		return nil
	}
	if err := r.service.BindAssistantMessage(ctx, r.sessionID, r.requestID, assistantMessageID); err != nil {
		return err
	}
	reason := teamsvc.FinalizeReasonCompleted
	if canceled {
		reason = teamsvc.FinalizeReasonCanceled
	}
	run, err := r.service.Finalize(ctx, r.run.ID, reason)
	if err != nil {
		return err
	}
	r.run = run
	if callbacks.OnTeamDone != nil {
		return callbacks.OnTeamDone(*run)
	}
	return nil
}
