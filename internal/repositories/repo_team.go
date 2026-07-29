package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"slimebot/internal/apperrors"
	"slimebot/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) GetOrCreateTeamRun(ctx context.Context, input domain.CreateTeamRunInput) (*domain.TeamRun, error) {
	startedAt := input.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}
	record := domain.TeamRun{
		ID:          uuid.NewString(),
		SessionID:   strings.TrimSpace(input.SessionID),
		RequestID:   strings.TrimSpace(input.RequestID),
		Status:      domain.TeamRunStatusRunning,
		MaxMembers:  input.MaxMembers,
		MaxParallel: input.MaxParallel,
		StartedAt:   startedAt,
	}
	for attempt := 0; attempt < 8; attempt++ {
		err := r.dbWithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_id"}, {Name: "request_id"}},
			DoNothing: true,
		}).Create(&record).Error
		if err == nil {
			break
		}
		if !isSQLiteBusyError(err) || attempt == 7 {
			return nil, err
		}
		time.Sleep(time.Duration(attempt+1) * 2 * time.Millisecond)
	}
	var stored domain.TeamRun
	for attempt := 0; attempt < 8; attempt++ {
		err := r.dbWithContext(ctx).
			Where("session_id = ? AND request_id = ?", record.SessionID, record.RequestID).
			Take(&stored).Error
		if err == nil {
			return &stored, nil
		}
		if !isSQLiteBusyError(err) || attempt == 7 {
			return nil, err
		}
		time.Sleep(time.Duration(attempt+1) * 2 * time.Millisecond)
	}
	return nil, gorm.ErrRecordNotFound
}

func isSQLiteBusyError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "database is locked") || strings.Contains(message, "database table is locked") || strings.Contains(message, "database is busy")
}

func (r *Repository) GetTeamRunByID(ctx context.Context, id string) (*domain.TeamRun, error) {
	var run domain.TeamRun
	err := r.dbWithContext(ctx).Where("id = ?", strings.TrimSpace(id)).Take(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return &run, err
}

func (r *Repository) CreateTeamMemberRun(ctx context.Context, input domain.CreateTeamMemberRunInput) (*domain.TeamMemberRun, error) {
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = uuid.NewString()
	}
	record := domain.TeamMemberRun{
		ID:            id,
		TeamRunID:     strings.TrimSpace(input.TeamRunID),
		ToolCallID:    strings.TrimSpace(input.ToolCallID),
		Title:         strings.TrimSpace(input.Title),
		Task:          strings.TrimSpace(input.Task),
		ModelConfigID: strings.TrimSpace(input.ModelConfigID),
		Status:        domain.TeamMemberRunStatusQueued,
		CreatedAt:     input.CreatedAt,
	}
	if err := r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_run_id"}, {Name: "tool_call_id"}},
		DoNothing: true,
	}).Create(&record).Error; err != nil {
		return nil, err
	}
	var stored domain.TeamMemberRun
	err := r.dbWithContext(ctx).
		Where("team_run_id = ? AND tool_call_id = ?", record.TeamRunID, record.ToolCallID).
		Take(&stored).Error
	return &stored, err
}

func (r *Repository) ListTeamMemberRuns(ctx context.Context, teamRunID string) ([]domain.TeamMemberRun, error) {
	var members []domain.TeamMemberRun
	err := r.dbWithContext(ctx).
		Where("team_run_id = ?", strings.TrimSpace(teamRunID)).
		Order("created_at asc").
		Order("id asc").
		Find(&members).Error
	return members, err
}

func (r *Repository) UpdateTeamMemberRun(ctx context.Context, input domain.UpdateTeamMemberRunInput) error {
	updates := map[string]any{
		"status":     input.Status,
		"answer":     input.Answer,
		"error":      input.Error,
		"updated_at": time.Now(),
	}
	if input.SubagentRunID != "" {
		updates["subagent_run_id"] = strings.TrimSpace(input.SubagentRunID)
	}
	if input.ModelConfigID != "" {
		updates["model_config_id"] = strings.TrimSpace(input.ModelConfigID)
	}
	if input.StartedAt != nil {
		updates["started_at"] = input.StartedAt
	}
	if input.FinishedAt != nil {
		updates["finished_at"] = input.FinishedAt
	}
	return r.dbWithContext(ctx).Model(&domain.TeamMemberRun{}).
		Where("id = ?", strings.TrimSpace(input.ID)).
		Updates(updates).Error
}

func (r *Repository) UpdateTeamRun(ctx context.Context, input domain.UpdateTeamRunInput) error {
	updates := map[string]any{
		"status":     input.Status,
		"last_error": input.LastError,
		"updated_at": time.Now(),
	}
	if input.FinishedAt != nil {
		updates["finished_at"] = input.FinishedAt
	}
	return r.dbWithContext(ctx).Model(&domain.TeamRun{}).
		Where("id = ?", strings.TrimSpace(input.ID)).
		Updates(updates).Error
}

func (r *Repository) BindTeamRunToAssistantMessage(ctx context.Context, sessionID, requestID, assistantMessageID string) error {
	return r.dbWithContext(ctx).Model(&domain.TeamRun{}).
		Where("session_id = ? AND request_id = ?", strings.TrimSpace(sessionID), strings.TrimSpace(requestID)).
		Updates(map[string]any{"assistant_message_id": strings.TrimSpace(assistantMessageID), "updated_at": time.Now()}).Error
}

func (r *Repository) InterruptActiveTeamRuns(ctx context.Context) error {
	now := time.Now()
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.TeamMemberRun{}).
			Where("status IN ?", []string{domain.TeamMemberRunStatusQueued, domain.TeamMemberRunStatusRunning}).
			Updates(map[string]any{"status": domain.TeamMemberRunStatusInterrupted, "finished_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		return tx.Model(&domain.TeamRun{}).
			Where("status = ?", domain.TeamRunStatusRunning).
			Updates(map[string]any{"status": domain.TeamRunStatusInterrupted, "finished_at": now, "updated_at": now}).Error
	})
}

func normalizedIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (r *Repository) ListSessionTeamRunsByAssistantMessageIDs(ctx context.Context, sessionID string, messageIDs []string) ([]domain.TeamRun, error) {
	ids := normalizedIDs(messageIDs)
	if len(ids) == 0 {
		return []domain.TeamRun{}, nil
	}
	var runs []domain.TeamRun
	err := r.dbWithContext(ctx).
		Where("session_id = ?", strings.TrimSpace(sessionID)).
		Where("assistant_message_id IN ?", ids).
		Order("started_at asc").
		Order("id asc").
		Find(&runs).Error
	return runs, err
}

func (r *Repository) ListSessionTeamMemberRunsByAssistantMessageIDs(ctx context.Context, sessionID string, messageIDs []string) ([]domain.TeamMemberRun, error) {
	ids := normalizedIDs(messageIDs)
	if len(ids) == 0 {
		return []domain.TeamMemberRun{}, nil
	}
	var members []domain.TeamMemberRun
	err := r.dbWithContext(ctx).
		Table("team_member_runs").
		Select("team_member_runs.*").
		Joins("JOIN team_runs ON team_runs.id = team_member_runs.team_run_id").
		Where("team_runs.session_id = ?", strings.TrimSpace(sessionID)).
		Where("team_runs.assistant_message_id IN ?", ids).
		Order("team_member_runs.created_at asc").
		Order("team_member_runs.id asc").
		Scan(&members).Error
	return members, err
}
