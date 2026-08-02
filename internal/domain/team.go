package domain

import "time"

const (
	TeamRunStatusRunning       = "running"
	TeamRunStatusSucceeded     = "succeeded"
	TeamRunStatusPartialFailed = "partial_failed"
	TeamRunStatusFailed        = "failed"
	TeamRunStatusCanceled      = "canceled"
	TeamRunStatusInterrupted   = "interrupted"

	TeamMemberRunStatusQueued      = "queued"
	TeamMemberRunStatusRunning     = "running"
	TeamMemberRunStatusSucceeded   = "succeeded"
	TeamMemberRunStatusFailed      = "failed"
	TeamMemberRunStatusCanceled    = "canceled"
	TeamMemberRunStatusInterrupted = "interrupted"
)

// TeamRun groups all subagent members created by one root chat request.
type TeamRun struct {
	ID                 string     `gorm:"primaryKey;size:36" json:"id"`
	SessionID          string     `gorm:"size:36;index;not null;uniqueIndex:idx_team_request,priority:1" json:"sessionId"`
	RequestID          string     `gorm:"size:36;index;not null;uniqueIndex:idx_team_request,priority:2" json:"requestId"`
	AssistantMessageID *string    `gorm:"size:36;index" json:"assistantMessageId,omitempty"`
	Status             string     `gorm:"size:32;index;not null" json:"status"`
	MaxMembers         int        `gorm:"not null" json:"maxMembers"`
	MaxParallel        int        `gorm:"not null" json:"maxParallel"`
	LastError          string     `gorm:"type:text" json:"lastError,omitempty"`
	StartedAt          time.Time  `gorm:"index;not null" json:"startedAt"`
	FinishedAt         *time.Time `gorm:"index" json:"finishedAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// TeamMemberRun records one run_subagent member without duplicating its nested event history.
type TeamMemberRun struct {
	ID            string     `gorm:"primaryKey;size:36" json:"id"`
	TeamRunID     string     `gorm:"size:36;index;not null;uniqueIndex:idx_team_member_tool,priority:1" json:"teamRunId"`
	ToolCallID    string     `gorm:"size:128;index;not null;uniqueIndex:idx_team_member_tool,priority:2" json:"toolCallId"`
	SubagentRunID string     `gorm:"size:128;index" json:"subagentRunId,omitempty"`
	Title         string     `gorm:"size:160;not null" json:"title"`
	Task          string     `gorm:"type:text;not null" json:"task"`
	ModelConfigID string     `gorm:"size:36;index" json:"modelConfigId,omitempty"`
	Status        string     `gorm:"size:32;index;not null" json:"status"`
	Answer        string     `gorm:"type:text" json:"answer,omitempty"`
	Error         string     `gorm:"type:text" json:"error,omitempty"`
	StartedAt     *time.Time `gorm:"index" json:"startedAt,omitempty"`
	FinishedAt    *time.Time `gorm:"index" json:"finishedAt,omitempty"`
	CreatedAt     time.Time  `gorm:"index" json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type CreateTeamRunInput struct {
	SessionID   string
	RequestID   string
	MaxMembers  int
	MaxParallel int
	StartedAt   time.Time
}

type CreateTeamMemberRunInput struct {
	ID            string
	TeamRunID     string
	ToolCallID    string
	Title         string
	Task          string
	ModelConfigID string
	CreatedAt     time.Time
}

type UpdateTeamMemberRunInput struct {
	ID            string
	Status        string
	SubagentRunID string
	ModelConfigID string
	Answer        string
	Error         string
	StartedAt     *time.Time
	FinishedAt    *time.Time
}

type UpdateTeamRunInput struct {
	ID         string
	Status     string
	LastError  string
	FinishedAt *time.Time
}
