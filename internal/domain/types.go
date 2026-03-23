package domain

import (
	"context"
	"time"
)

type SessionMemoryUpsertInput struct {
	SessionID          string
	Summary            string
	Keywords           []string
	SourceMessageCount int
}

type SessionMemoryCreateInput struct {
	SessionID          string
	Summary            string
	Keywords           []string
	SourceMessageCount int
}

type SessionMemorySearchHit struct {
	Memory          SessionMemory
	MatchedKeywords []string
	Score           float64
}

const (
	MemoryTypePreference = "preference"
	MemoryTypeConstraint = "constraint"
	MemoryTypeTask       = "task"
	MemoryTypeProfile    = "profile"
	MemoryTypeProject    = "project"

	MemoryStatusActive   = "active"
	MemoryStatusStale    = "stale"
	MemoryStatusArchived = "archived"
)

type MemoryFactCreateInput struct {
	SessionID      string
	MemoryType     string
	Subject        string
	Predicate      string
	Value          string
	Summary        string
	Confidence     float64
	SourceStartSeq int64
	SourceEndSeq   int64
	LastSeenAt     time.Time
	ExpiresAt      *time.Time
}

type MemoryFactUpdateInput struct {
	ID             string
	SessionID      string
	Value          string
	Summary        string
	Confidence     float64
	SourceStartSeq int64
	SourceEndSeq   int64
	LastSeenAt     time.Time
	ExpiresAt      *time.Time
	Status         string
}

type MemoryFactSearchInput struct {
	Query          string
	MemoryTypes    []string
	Limit          int
	ExcludeSession string
	Now            time.Time
}

type MemoryFactSearchHit struct {
	Fact            MemoryFact
	MatchedKeywords []string
	Score           float64
}

type AddMessageInput struct {
	SessionID         string
	Role              string
	Content           string
	IsInterrupted     bool
	IsStopPlaceholder bool
	Attachments       []MessageAttachment
}

type ToolCallStartRecordInput struct {
	SessionID        string
	RequestID        string
	ToolCallID       string
	ToolName         string
	Command          string
	Params           map[string]string
	Status           string
	RequiresApproval bool
	StartedAt        time.Time
}

type ToolCallResultRecordInput struct {
	SessionID  string
	RequestID  string
	ToolCallID string
	Status     string
	Output     string
	Error      string
	FinishedAt time.Time
}

type MemoryVectorStore interface {
	UpsertSessionMemoryVector(ctx context.Context, input MemoryVectorUpsertInput) error
	SearchSimilarSessionIDs(ctx context.Context, queryVector []float32, limit int, excludeSessionID string) ([]MemoryVectorSearchHit, error)
	SearchMemoriesInSession(ctx context.Context, queryVector []float32, sessionID string, limit int) ([]MemoryVectorSearchHit, error)
	DeleteMemoryVector(ctx context.Context, memoryID string) error
}

type MemoryVectorUpsertInput struct {
	MemoryID  string
	SessionID string
	Vector    []float32
	Payload   map[string]any
}

type MemoryVectorSearchHit struct {
	SessionID string
	MemoryID  string
	Score     float64
}
