package chat

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"slimebot/internal/domain"
	llmsvc "slimebot/internal/services/llm"
	teamsvc "slimebot/internal/services/team"
)

func TestHandleChatStreamPersistsTeamAndPublishesStableIdentity(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	session, err := repo.CreateSession(ctx, "team")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	model, err := repo.CreateLLMConfig(ctx, domain.LLMConfig{
		Name: "fake", Provider: llmsvc.ProviderOpenAI, BaseURL: "http://fake", APIKey: "key", Model: "fake-model",
	})
	if err != nil {
		t.Fatalf("create model: %v", err)
	}
	provider := newParallelSubagentProvider(false)
	service := NewChatService(repo, nil, llmsvc.NewFactory(provider), nil, nil)
	service.SetTeamService(teamsvc.NewService(repo, teamsvc.Options{}))

	var (
		events       []string
		teamRunID    string
		memberRunIDs = map[string]bool{}
		mu           sync.Mutex
	)
	done := make(chan error, 1)
	go func() {
		_, runErr := service.HandleChatStream(ctx, session.ID, "request-team", "delegate parallel subagents", "", model.ID, nil, "off", false, "", "", AgentCallbacks{
			OnChunk: func(string) error { return nil },
			OnTeamStart: func(run domain.TeamRun) error {
				mu.Lock()
				defer mu.Unlock()
				events = append(events, "team_start")
				teamRunID = run.ID
				return nil
			},
			OnSubagentStart: func(meta AgentEventMeta, title, task string) error {
				mu.Lock()
				defer mu.Unlock()
				events = append(events, "member_start")
				if meta.TeamRunID != teamRunID {
					return fmt.Errorf("member team id %q does not match %q", meta.TeamRunID, teamRunID)
				}
				if meta.MemberRunID == "" || meta.SubagentRunID == "" || meta.ParentToolCallID == "" || title == "" || task == "" {
					return fmt.Errorf("incomplete member event: %#v", meta)
				}
				memberRunIDs[meta.MemberRunID] = true
				return nil
			},
			OnTeamDone: func(run domain.TeamRun) error {
				mu.Lock()
				defer mu.Unlock()
				events = append(events, "team_done")
				if run.ID != teamRunID || run.Status != domain.TeamRunStatusSucceeded {
					return fmt.Errorf("unexpected final team: %#v", run)
				}
				return nil
			},
		})
		done <- runErr
	}()

	seen := map[string]bool{}
	for len(seen) < 2 {
		select {
		case task := <-provider.started:
			seen[task] = true
		case err := <-done:
			t.Fatalf("chat finished before members started: %v", err)
		case <-time.After(time.Second):
			t.Fatalf("members did not start: %v", seen)
		}
	}
	close(provider.release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("HandleChatStream failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("chat did not finish")
	}

	mu.Lock()
	if len(events) != 4 || events[0] != "team_start" || events[len(events)-1] != "team_done" {
		t.Fatalf("event order = %v", events)
	}
	if len(memberRunIDs) != 2 {
		t.Fatalf("member ids = %v", memberRunIDs)
	}
	mu.Unlock()

	messages, _, err := repo.ListSessionMessagesPage(ctx, session.ID, 10, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	var assistantID string
	for _, message := range messages {
		if message.Role == "assistant" {
			assistantID = message.ID
		}
	}
	teams, err := repo.ListSessionTeamRunsByAssistantMessageIDs(ctx, session.ID, []string{assistantID})
	if err != nil {
		t.Fatalf("list team history: %v", err)
	}
	members, err := repo.ListSessionTeamMemberRunsByAssistantMessageIDs(ctx, session.ID, []string{assistantID})
	if err != nil {
		t.Fatalf("list member history: %v", err)
	}
	if len(teams) != 1 || teams[0].Status != domain.TeamRunStatusSucceeded || len(members) != 2 {
		t.Fatalf("team history = %#v / %#v", teams, members)
	}
}

func TestHandleChatStreamDoesNotCreateTeamForDirectAnswer(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	session, _ := repo.CreateSession(ctx, "direct")
	model, _ := repo.CreateLLMConfig(ctx, domain.LLMConfig{
		Name: "fake", Provider: llmsvc.ProviderOpenAI, BaseURL: "http://fake", APIKey: "key", Model: "fake-model",
	})
	service := NewChatService(repo, nil, llmsvc.NewFactory(&captureMessagesProvider{}), nil, nil)
	service.SetTeamService(teamsvc.NewService(repo, teamsvc.Options{}))
	if _, err := service.HandleChatStream(ctx, session.ID, "request-direct", "answer directly", "", model.ID, nil, "off", false, "", "", AgentCallbacks{}); err != nil {
		t.Fatalf("HandleChatStream failed: %v", err)
	}
	messages, _, err := repo.ListSessionMessagesPage(ctx, session.ID, 10, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	messageIDs := make([]string, 0, len(messages))
	for _, message := range messages {
		messageIDs = append(messageIDs, message.ID)
	}
	teams, err := repo.ListSessionTeamRunsByAssistantMessageIDs(ctx, session.ID, messageIDs)
	if err != nil {
		t.Fatalf("list teams: %v", err)
	}
	if len(teams) != 0 {
		t.Fatalf("teams = %#v", teams)
	}
}

func TestHandleChatStreamCancellationMarksTeamAndMemberCanceled(t *testing.T) {
	repo := newTestRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	session, _ := repo.CreateSession(context.Background(), "cancel team")
	model, _ := repo.CreateLLMConfig(context.Background(), domain.LLMConfig{
		Name: "fake", Provider: llmsvc.ProviderOpenAI, BaseURL: "http://fake", APIKey: "key", Model: "fake-model",
	})
	provider := &cancelSubagentProvider{started: make(chan struct{})}
	service := NewChatService(repo, nil, llmsvc.NewFactory(provider), nil, nil)
	service.SetTeamService(teamsvc.NewService(repo, teamsvc.Options{}))

	var finalTeam domain.TeamRun
	done := make(chan error, 1)
	go func() {
		result, runErr := service.HandleChatStream(ctx, session.ID, "request-cancel", "delegate cancellable subagent", "", model.ID, nil, "off", false, "", "", AgentCallbacks{
			OnTeamDone: func(run domain.TeamRun) error {
				finalTeam = run
				return nil
			},
		})
		if runErr == nil && (result == nil || !result.IsInterrupted) {
			runErr = fmt.Errorf("result is not interrupted: %#v", result)
		}
		done <- runErr
	}()
	select {
	case <-provider.started:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("subagent did not start")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("HandleChatStream failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled chat did not finish")
	}
	if finalTeam.Status != domain.TeamRunStatusCanceled {
		t.Fatalf("team status = %q", finalTeam.Status)
	}
	members, err := repo.ListTeamMemberRuns(context.Background(), finalTeam.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(members) != 1 || members[0].Status != domain.TeamMemberRunStatusCanceled {
		t.Fatalf("canceled members = %#v", members)
	}
}
