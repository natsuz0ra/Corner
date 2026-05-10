package chat

import (
	"context"
	llmsvc "slimebot/internal/services/llm"
)

// ApprovalRequest is sent to the client for tool-call approval.
type ApprovalRequest struct {
	ToolCallID       string         `json:"toolCallId"`
	ToolName         string         `json:"toolName"`
	Command          string         `json:"command"`
	Params           map[string]any `json:"params"`
	RequiresApproval bool           `json:"requiresApproval"`
	ReviewStatus     string         `json:"reviewStatus,omitempty"`
	ReviewRisk       string         `json:"reviewRisk,omitempty"`
	ReviewReason     string         `json:"reviewReason,omitempty"`
	Preamble         string         `json:"preamble,omitempty"`
	ParentToolCallID string         `json:"parentToolCallId,omitempty"`
	SubagentRunID    string         `json:"subagentRunId,omitempty"`
}

// ApprovalResponse is the client's approval decision.
type ApprovalResponse struct {
	ToolCallID string `json:"toolCallId"`
	Approved   bool   `json:"approved"`
	Answers    string `json:"answers,omitempty"` // JSON-encoded answers for ask_questions tool
}

type ApprovalReviewDecision string

const (
	ApprovalReviewDecisionApprove ApprovalReviewDecision = "approve"
	ApprovalReviewDecisionAskUser ApprovalReviewDecision = "ask_user"

	ApprovalReviewStatusReviewing ApprovalReviewStatus = "reviewing"
	ApprovalReviewStatusApproved  ApprovalReviewStatus = "approved"
	ApprovalReviewStatusNeedsUser ApprovalReviewStatus = "needs_user"
)

type ApprovalReviewStatus string

type ApprovalReviewRequest struct {
	ToolCallID string
	ToolName   string
	Command    string
	Params     map[string]any
	Preamble   string
}

type ApprovalReviewResult struct {
	Decision ApprovalReviewDecision `json:"decision"`
	Risk     string                 `json:"risk"`
	Reason   string                 `json:"reason"`
}

type ApprovalReviewEvent struct {
	ToolCallID       string `json:"toolCallId"`
	ToolName         string `json:"toolName"`
	Command          string `json:"command"`
	ReviewStatus     string `json:"reviewStatus"`
	ReviewRisk       string `json:"reviewRisk,omitempty"`
	ReviewReason     string `json:"reviewReason,omitempty"`
	ParentToolCallID string `json:"parentToolCallId,omitempty"`
	SubagentRunID    string `json:"subagentRunId,omitempty"`
}

// ToolCallResult is pushed to the client after tool execution.
type ToolCallResult struct {
	ToolCallID       string `json:"toolCallId"`
	ToolName         string `json:"toolName"`
	Command          string `json:"command"`
	RequiresApproval bool   `json:"requiresApproval"`
	Status           string `json:"status"`
	Output           string `json:"output"`
	Error            string `json:"error"`
	Metadata         any    `json:"metadata,omitempty"`
	ParentToolCallID string `json:"parentToolCallId,omitempty"`
	SubagentRunID    string `json:"subagentRunId,omitempty"`
}

type TodoItem struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type TodoUpdate struct {
	Items []TodoItem `json:"items"`
	Note  string     `json:"note,omitempty"`
}

type ThinkingEventMeta struct {
	ParentToolCallID string
	SubagentRunID    string
}

type ContextUsage struct {
	SessionID        string `json:"sessionId"`
	ModelConfigID    string `json:"modelConfigId"`
	UsedTokens       int    `json:"usedTokens"`
	TotalTokens      int    `json:"totalTokens"`
	UsedPercent      int    `json:"usedPercent"`
	AvailablePercent int    `json:"availablePercent"`
	IsCompacted      bool   `json:"isCompacted"`
	CompactedAt      string `json:"compactedAt,omitempty"`
}

// AgentCallbacks wires the agent loop to the outside world (streaming, approval, results).
type AgentCallbacks struct {
	OnChunk                func(chunk string) error
	OnContextUsage         func(usage ContextUsage) error
	OnContextCompacted     func(usage ContextUsage) error
	OnToolCallStart        func(req ApprovalRequest) error
	OnToolApprovalReview   func(event ApprovalReviewEvent) error
	OnToolApprovalRequired func(req ApprovalRequest) error
	WaitApproval           func(ctx context.Context, toolCallID string) (*ApprovalResponse, error)
	OnToolCallResult       func(result ToolCallResult) error
	OnSubagentStart        func(parentToolCallID, runID, title, task string) error
	OnSubagentChunk        func(parentToolCallID, runID, chunk string) error
	OnSubagentDone         func(parentToolCallID, runID string, runErr error) error
	OnThinkingStart        func(meta ThinkingEventMeta) error
	OnThinkingChunk        func(chunk string, meta ThinkingEventMeta) error
	OnThinkingDone         func(meta ThinkingEventMeta) error
	OnTodoUpdate           func(update TodoUpdate) error
	OnPlanStart            func() error
	OnPlanChunk            func(chunk string) error
	OnPlanBody             func(planBody string) error
	OnTitleGenerated       func(sessionID, title string)
}

// AgentLoopOptions configures nested agent execution.
type AgentLoopOptions struct {
	Depth           int
	ApprovalMode    string
	PlanMode        bool
	PlanStarted     *bool
	PlanComplete    *bool
	SubagentModelID string
	LatestUsage     *llmsvc.TokenUsage
	OnProviderUsage func(usage llmsvc.TokenUsage) error
}
