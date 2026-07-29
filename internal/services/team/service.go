package team

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"slimebot/internal/constants"
	"slimebot/internal/domain"
)

var (
	ErrMemberLimitExceeded = errors.New("agent team member limit exceeded")
	ErrInvalidTransition   = errors.New("invalid agent team state transition")
)

type Options struct {
	MaxMembers  int
	MaxParallel int
}

type Service struct {
	store       domain.TeamStore
	maxMembers  int
	maxParallel int
}

func NewService(store domain.TeamStore, options Options) *Service {
	maxMembers := options.MaxMembers
	if maxMembers <= 0 {
		maxMembers = constants.MaxTeamMembers
	}
	maxParallel := options.MaxParallel
	if maxParallel <= 0 {
		maxParallel = constants.MaxParallelToolCalls
	}
	return &Service{store: store, maxMembers: maxMembers, maxParallel: maxParallel}
}

func (s *Service) EnsureTeam(ctx context.Context, sessionID, requestID string) (*domain.TeamRun, error) {
	return s.store.GetOrCreateTeamRun(ctx, domain.CreateTeamRunInput{
		SessionID:   strings.TrimSpace(sessionID),
		RequestID:   strings.TrimSpace(requestID),
		MaxMembers:  s.maxMembers,
		MaxParallel: s.maxParallel,
	})
}

func (s *Service) EnsureMember(ctx context.Context, teamRunID, toolCallID, title, task, modelConfigID string) (*domain.TeamMemberRun, error) {
	members, err := s.store.ListTeamMemberRuns(ctx, teamRunID)
	if err != nil {
		return nil, err
	}
	trimmedToolCallID := strings.TrimSpace(toolCallID)
	for index := range members {
		if members[index].ToolCallID == trimmedToolCallID {
			return &members[index], nil
		}
	}
	run, err := s.store.GetTeamRunByID(ctx, teamRunID)
	if err != nil {
		return nil, err
	}
	if len(members) >= run.MaxMembers {
		return nil, fmt.Errorf("%w: maximum %d members per request", ErrMemberLimitExceeded, run.MaxMembers)
	}
	return s.store.CreateTeamMemberRun(ctx, domain.CreateTeamMemberRunInput{
		TeamRunID:     teamRunID,
		ToolCallID:    trimmedToolCallID,
		Title:         strings.TrimSpace(title),
		Task:          strings.TrimSpace(task),
		ModelConfigID: strings.TrimSpace(modelConfigID),
	})
}

func (s *Service) StartMember(ctx context.Context, memberRunID, subagentRunID, modelConfigID string) (*domain.TeamMemberRun, error) {
	member, err := s.store.GetTeamMemberRunByID(ctx, memberRunID)
	if err != nil {
		return nil, err
	}
	if member.Status != domain.TeamMemberRunStatusQueued {
		return nil, fmt.Errorf("%w: member %s is %s", ErrInvalidTransition, member.ID, member.Status)
	}
	now := time.Now()
	if err := s.store.UpdateTeamMemberRun(ctx, domain.UpdateTeamMemberRunInput{
		ID: member.ID, Status: domain.TeamMemberRunStatusRunning, SubagentRunID: subagentRunID, ModelConfigID: modelConfigID, StartedAt: &now,
	}); err != nil {
		return nil, err
	}
	return s.store.GetTeamMemberRunByID(ctx, member.ID)
}

type MemberResult struct {
	Answer      string
	Err         error
	Canceled    bool
	Interrupted bool
}

func (s *Service) FinishMember(ctx context.Context, memberRunID string, result MemberResult) (*domain.TeamMemberRun, error) {
	member, err := s.store.GetTeamMemberRunByID(ctx, memberRunID)
	if err != nil {
		return nil, err
	}
	if member.Status != domain.TeamMemberRunStatusQueued && member.Status != domain.TeamMemberRunStatusRunning {
		return nil, fmt.Errorf("%w: member %s is %s", ErrInvalidTransition, member.ID, member.Status)
	}
	status := domain.TeamMemberRunStatusSucceeded
	errorText := ""
	switch {
	case result.Interrupted:
		status = domain.TeamMemberRunStatusInterrupted
	case result.Canceled:
		status = domain.TeamMemberRunStatusCanceled
	case result.Err != nil:
		status = domain.TeamMemberRunStatusFailed
		errorText = result.Err.Error()
	}
	now := time.Now()
	if err := s.store.UpdateTeamMemberRun(ctx, domain.UpdateTeamMemberRunInput{
		ID: member.ID, Status: status, Answer: strings.TrimSpace(result.Answer), Error: errorText, FinishedAt: &now,
	}); err != nil {
		return nil, err
	}
	return s.store.GetTeamMemberRunByID(ctx, member.ID)
}

type FinalizeReason string

const (
	FinalizeReasonCompleted FinalizeReason = "completed"
	FinalizeReasonCanceled  FinalizeReason = "canceled"
)

func (s *Service) Finalize(ctx context.Context, teamRunID string, reason FinalizeReason) (*domain.TeamRun, error) {
	run, err := s.store.GetTeamRunByID(ctx, teamRunID)
	if err != nil {
		return nil, err
	}
	if run.Status != domain.TeamRunStatusRunning {
		return run, nil
	}
	members, err := s.store.ListTeamMemberRuns(ctx, teamRunID)
	if err != nil {
		return nil, err
	}
	if reason == FinalizeReasonCanceled {
		for index := range members {
			if members[index].Status != domain.TeamMemberRunStatusQueued && members[index].Status != domain.TeamMemberRunStatusRunning {
				continue
			}
			if _, err := s.FinishMember(ctx, members[index].ID, MemberResult{Canceled: true}); err != nil {
				return nil, err
			}
		}
		return s.finishTeam(ctx, run.ID, domain.TeamRunStatusCanceled, "")
	}

	succeeded := 0
	nonSucceeded := 0
	lastError := ""
	for index := range members {
		member := members[index]
		if member.Status == domain.TeamMemberRunStatusQueued || member.Status == domain.TeamMemberRunStatusRunning {
			unfinishedErr := errors.New("member did not reach a terminal state")
			updated, finishErr := s.FinishMember(ctx, member.ID, MemberResult{Err: unfinishedErr})
			if finishErr != nil {
				return nil, finishErr
			}
			member = *updated
		}
		if member.Status == domain.TeamMemberRunStatusSucceeded {
			succeeded++
		} else {
			nonSucceeded++
			if member.Error != "" {
				lastError = member.Error
			}
		}
	}
	status := domain.TeamRunStatusFailed
	switch {
	case succeeded > 0 && nonSucceeded == 0:
		status = domain.TeamRunStatusSucceeded
	case succeeded > 0:
		status = domain.TeamRunStatusPartialFailed
	}
	return s.finishTeam(ctx, run.ID, status, lastError)
}

func (s *Service) finishTeam(ctx context.Context, teamRunID, status, lastError string) (*domain.TeamRun, error) {
	now := time.Now()
	if err := s.store.UpdateTeamRun(ctx, domain.UpdateTeamRunInput{
		ID: teamRunID, Status: status, LastError: lastError, FinishedAt: &now,
	}); err != nil {
		return nil, err
	}
	return s.store.GetTeamRunByID(ctx, teamRunID)
}

func (s *Service) BindAssistantMessage(ctx context.Context, sessionID, requestID, assistantMessageID string) error {
	return s.store.BindTeamRunToAssistantMessage(ctx, sessionID, requestID, assistantMessageID)
}

func (s *Service) RecoverInterrupted(ctx context.Context) error {
	return s.store.InterruptActiveTeamRuns(ctx)
}
