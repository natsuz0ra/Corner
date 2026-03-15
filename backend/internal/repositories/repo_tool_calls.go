package repositories

import (
	"encoding/json"
	"errors"
	"time"

	"slimebot/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

func (r *Repository) UpsertToolCallStart(input ToolCallStartRecordInput) error {
	paramsJSONBytes, err := json.Marshal(input.Params)
	if err != nil {
		return err
	}
	paramsJSON := string(paramsJSONBytes)

	startedAt := input.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}

	var existing models.ToolCallRecord
	query := r.db.Where("session_id = ? AND request_id = ? AND tool_call_id = ?", input.SessionID, input.RequestID, input.ToolCallID).First(&existing)
	if query.Error == nil {
		return r.db.Model(&models.ToolCallRecord{}).
			Where("id = ?", existing.ID).
			Updates(map[string]any{
				"tool_name":            input.ToolName,
				"command":              input.Command,
				"params_json":          paramsJSON,
				"status":               input.Status,
				"requires_approval":    input.RequiresApproval,
				"started_at":           startedAt,
				"finished_at":          nil,
				"output":               "",
				"error":                "",
				"assistant_message_id": nil,
				"updated_at":           time.Now(),
			}).
			Error
	}
	if query.Error != nil && !errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return query.Error
	}

	record := models.ToolCallRecord{
		ID:               uuid.NewString(),
		SessionID:        input.SessionID,
		RequestID:        input.RequestID,
		ToolCallID:       input.ToolCallID,
		ToolName:         input.ToolName,
		Command:          input.Command,
		ParamsJSON:       paramsJSON,
		Status:           input.Status,
		RequiresApproval: input.RequiresApproval,
		StartedAt:        startedAt,
	}
	return r.db.Create(&record).Error
}

func (r *Repository) UpdateToolCallResult(input ToolCallResultRecordInput) error {
	updates := map[string]any{
		"status":      input.Status,
		"output":      input.Output,
		"error":       input.Error,
		"updated_at":  time.Now(),
		"finished_at": input.FinishedAt,
	}
	if input.FinishedAt.IsZero() {
		updates["finished_at"] = time.Now()
	}

	return r.db.Model(&models.ToolCallRecord{}).
		Where("session_id = ? AND request_id = ? AND tool_call_id = ?", input.SessionID, input.RequestID, input.ToolCallID).
		Updates(updates).
		Error
}

func (r *Repository) BindToolCallsToAssistantMessage(sessionID, requestID, assistantMessageID string) error {
	return r.db.Model(&models.ToolCallRecord{}).
		Where("session_id = ? AND request_id = ?", sessionID, requestID).
		Updates(map[string]any{
			"assistant_message_id": assistantMessageID,
			"updated_at":           time.Now(),
		}).
		Error
}

func (r *Repository) ListSessionToolCallRecords(sessionID string) ([]models.ToolCallRecord, error) {
	var records []models.ToolCallRecord
	err := r.db.
		Where("session_id = ?", sessionID).
		Order("started_at asc").
		Order("created_at asc").
		Find(&records).
		Error
	return records, err
}
