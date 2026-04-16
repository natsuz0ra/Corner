package platforms

import (
	"context"
	"time"

	chatsvc "slimebot/internal/services/chat"
)

// ApprovalBroker registers, waits on, and resolves tool approvals from platform callbacks.
type ApprovalBroker interface {
	Register(toolCallID string, chatID string, ttl time.Duration) (string, string, error)
	Wait(ctx context.Context, toolCallID string) (*chatsvc.ApprovalResponse, error)
	ResolveByCallback(chatID string, callbackData string) (bool, error)
	Remove(toolCallID string)
}
