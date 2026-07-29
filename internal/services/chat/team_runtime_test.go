package chat

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"slimebot/internal/constants"
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

func TestHandleChatStreamReservesQueuedMembersBeforeParallelSlots(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	session, _ := repo.CreateSession(ctx, "queued team")
	model, _ := repo.CreateLLMConfig(ctx, domain.LLMConfig{
		Name: "fake", Provider: llmsvc.ProviderOpenAI, BaseURL: "http://fake", APIKey: "key", Model: "fake-model",
	})
	provider := newQueuedSubagentProvider(5)
	service := NewChatService(repo, nil, llmsvc.NewFactory(provider), nil, nil)
	service.SetTeamService(teamsvc.NewService(repo, teamsvc.Options{}))

	teamStarted := make(chan domain.TeamRun, 1)
	done := make(chan error, 1)
	go func() {
		_, runErr := service.HandleChatStream(ctx, session.ID, "request-queued", "delegate five subagents", "", model.ID, nil, "off", false, "", "", AgentCallbacks{
			OnTeamStart: func(run domain.TeamRun) error {
				teamStarted <- run
				return nil
			},
		})
		done <- runErr
	}()
	defer func() {
		provider.releaseAll()
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}()

	var run domain.TeamRun
	select {
	case run = <-teamStarted:
	case <-time.After(time.Second):
		t.Fatal("team did not start")
	}
	for started := 0; started < constants.MaxParallelToolCalls; started++ {
		select {
		case <-provider.started:
		case <-time.After(time.Second):
			t.Fatalf("only %d subagents entered running slots", started)
		}
	}

	members, err := repo.ListTeamMemberRuns(ctx, run.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(members) != 5 {
		t.Fatalf("expected all five members to be reserved before a slot is released, got %d: %#v", len(members), members)
	}
	var queued, running int
	for _, member := range members {
		switch member.Status {
		case domain.TeamMemberRunStatusQueued:
			queued++
		case domain.TeamMemberRunStatusRunning:
			running++
		}
	}
	if queued != 1 || running != constants.MaxParallelToolCalls {
		t.Fatalf("member states queued=%d running=%d: %#v", queued, running, members)
	}
}

func TestHandleChatStreamRejectsNinthMemberWithoutCancelingExistingMembers(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	session, _ := repo.CreateSession(ctx, "team budget")
	model, _ := repo.CreateLLMConfig(ctx, domain.LLMConfig{
		Name: "fake", Provider: llmsvc.ProviderOpenAI, BaseURL: "http://fake", APIKey: "key", Model: "fake-model",
	})
	provider := newQueuedSubagentProvider(constants.MaxTeamMembers + 1)
	service := NewChatService(repo, nil, llmsvc.NewFactory(provider), nil, nil)
	service.SetTeamService(teamsvc.NewService(repo, teamsvc.Options{}))

	var finalTeam domain.TeamRun
	done := make(chan error, 1)
	go func() {
		_, runErr := service.HandleChatStream(ctx, session.ID, "request-budget", "delegate nine subagents", "", model.ID, nil, "off", false, "", "", AgentCallbacks{
			OnTeamDone: func(run domain.TeamRun) error {
				finalTeam = run
				return nil
			},
		})
		done <- runErr
	}()

	for started := 0; started < constants.MaxParallelToolCalls; started++ {
		select {
		case <-provider.started:
		case <-time.After(time.Second):
			provider.releaseAll()
			t.Fatalf("only %d subagents entered running slots", started)
		}
	}
	provider.releaseAll()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("HandleChatStream failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("chat did not finish")
	}

	members, err := repo.ListTeamMemberRuns(ctx, finalTeam.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(members) != constants.MaxTeamMembers {
		t.Fatalf("members = %d, want %d", len(members), constants.MaxTeamMembers)
	}
	for _, member := range members {
		if member.Status != domain.TeamMemberRunStatusSucceeded {
			t.Fatalf("existing member did not continue: %#v", member)
		}
	}
	if finalTeam.Status != domain.TeamRunStatusSucceeded {
		t.Fatalf("team status = %q", finalTeam.Status)
	}
	var foundBudgetError bool
	for _, output := range provider.parentToolOutputs() {
		if strings.Contains(output, teamsvc.ErrMemberLimitExceeded.Error()) {
			foundBudgetError = true
			break
		}
	}
	if !foundBudgetError {
		t.Fatalf("parent tool outputs do not contain member budget error: %v", provider.parentToolOutputs())
	}
}

type queuedSubagentProvider struct {
	memberCount int
	started     chan string
	release     chan struct{}
	releaseOnce sync.Once
	mu          sync.Mutex
	parentMsgs  []llmsvc.ChatMessage
}

func (p *queuedSubagentProvider) parentToolOutputs() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var outputs []string
	for _, message := range p.parentMsgs {
		if message.Role == "tool" {
			outputs = append(outputs, message.Content)
		}
	}
	return outputs
}

func newQueuedSubagentProvider(memberCount int) *queuedSubagentProvider {
	return &queuedSubagentProvider{
		memberCount: memberCount,
		started:     make(chan string, memberCount),
		release:     make(chan struct{}),
	}
}

func (p *queuedSubagentProvider) releaseAll() {
	p.releaseOnce.Do(func() { close(p.release) })
}

func (p *queuedSubagentProvider) StreamChatWithTools(
	ctx context.Context,
	_ llmsvc.ModelRuntimeConfig,
	messages []llmsvc.ChatMessage,
	_ []llmsvc.ToolDef,
	callbacks llmsvc.StreamCallbacks,
) (*llmsvc.StreamResult, error) {
	for index := 0; index < p.memberCount; index++ {
		task := fmt.Sprintf("queued-task-%d", index)
		if containsUserMessageText(messages, task) {
			p.started <- task
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-p.release:
			}
			if callbacks.OnChunk != nil {
				if err := callbacks.OnChunk("answer " + task); err != nil {
					return nil, err
				}
			}
			return &llmsvc.StreamResult{Type: llmsvc.StreamResultText}, nil
		}
	}
	if hasToolMessages(messages) {
		p.mu.Lock()
		p.parentMsgs = append([]llmsvc.ChatMessage{}, messages...)
		p.mu.Unlock()
		if callbacks.OnChunk != nil {
			if err := callbacks.OnChunk("parent done"); err != nil {
				return nil, err
			}
		}
		return &llmsvc.StreamResult{Type: llmsvc.StreamResultText}, nil
	}

	toolCalls := make([]llmsvc.ToolCallInfo, 0, p.memberCount)
	for index := 0; index < p.memberCount; index++ {
		toolCalls = append(toolCalls, llmsvc.ToolCallInfo{
			ID:        fmt.Sprintf("queued-call-%d", index),
			Name:      constants.RunSubagentTool,
			Arguments: fmt.Sprintf(`{"title":"Member %d","task":"queued-task-%d"}`, index, index),
		})
	}
	return &llmsvc.StreamResult{
		Type:             llmsvc.StreamResultToolCalls,
		ToolCalls:        toolCalls,
		AssistantMessage: llmsvc.ChatMessage{Role: "assistant", ToolCalls: toolCalls},
	}, nil
}

func containsUserMessageText(messages []llmsvc.ChatMessage, text string) bool {
	for _, message := range messages {
		if message.Role == "user" && strings.Contains(message.Content, text) {
			return true
		}
	}
	return false
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
