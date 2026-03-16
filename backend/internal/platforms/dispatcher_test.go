package platforms

import (
	"context"
	"testing"

	"slimebot/backend/internal/models"
	"slimebot/backend/internal/services"
)

type mockPlatformChatService struct{}

func (m *mockPlatformChatService) EnsureMessagePlatformSession() (*models.Session, error) {
	return &models.Session{ID: services.MessagePlatformSessionID, Name: services.MessagePlatformSessionName}, nil
}

func (m *mockPlatformChatService) ResolvePlatformModel() (string, error) {
	return "mock-model-id", nil
}

func (m *mockPlatformChatService) HandleChatStream(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ string,
	callbacks services.AgentCallbacks,
) (*services.ChatStreamResult, error) {
	_ = callbacks.OnToolCallStart(services.ApprovalRequest{
		ToolCallID:       "tc_1",
		ToolName:         "exec",
		Command:          "run",
		Params:           map[string]string{"cmd": "echo hi"},
		RequiresApproval: true,
	})
	approval, _ := callbacks.WaitApproval(context.Background(), "tc_1")
	status := "completed"
	errText := ""
	if approval == nil || !approval.Approved {
		status = "rejected"
		errText = "用户拒绝执行"
	}
	_ = callbacks.OnToolCallResult(services.ToolCallResult{
		ToolCallID:       "tc_1",
		ToolName:         "exec",
		Command:          "run",
		RequiresApproval: true,
		Status:           status,
		Error:            errText,
	})
	return &services.ChatStreamResult{Answer: "hello-from-mock"}, nil
}

type mockSender struct {
	items []string
}

func (m *mockSender) SendText(_ string, text string) error {
	m.items = append(m.items, text)
	return nil
}

func TestDispatcherHandleInbound_SendsToolSummaryAndFinalAnswer(t *testing.T) {
	dispatcher := NewDispatcher(&mockPlatformChatService{})
	sender := &mockSender{}

	err := dispatcher.HandleInbound(context.Background(), InboundMessage{
		Platform: "telegram",
		ChatID:   "10001",
		Text:     "hello",
	}, sender)
	if err != nil {
		t.Fatalf("handle inbound failed: %v", err)
	}
	if len(sender.items) < 3 {
		t.Fatalf("expected at least 3 messages(tool start, approval fallback, final), got=%d", len(sender.items))
	}
	last := sender.items[len(sender.items)-1]
	if last != "hello-from-mock" {
		t.Fatalf("expected final answer message, got=%s", last)
	}
}
