package repositories

import (
	"context"
	"errors"
	"fmt"
	"slimebot/internal/apperrors"
	"slimebot/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repository) CreateScheduledTask(ctx context.Context, task *domain.ScheduledTask) error {
	if task.ID == "" {
		task.ID = uuid.NewString()
	}
	return r.dbWithContext(ctx).Create(task).Error
}

func (r *Repository) GetScheduledTask(ctx context.Context, id string) (*domain.ScheduledTask, error) {
	var task domain.ScheduledTask
	err := r.dbWithContext(ctx).First(&task, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("scheduled task %s: %w", id, apperrors.ErrNotFound)
	}
	return &task, err
}

func (r *Repository) ListScheduledTasks(ctx context.Context, includeInactive bool) ([]domain.ScheduledTask, error) {
	var tasks []domain.ScheduledTask
	q := r.dbWithContext(ctx).Order("created_at desc")
	if !includeInactive {
		q = q.Where("status NOT IN ?", []string{domain.ScheduledTaskStatusPaused, domain.ScheduledTaskStatusCompleted})
	}
	return tasks, q.Find(&tasks).Error
}

func (r *Repository) ListDueScheduledTasks(ctx context.Context, now any) ([]domain.ScheduledTask, error) {
	var tasks []domain.ScheduledTask
	err := r.dbWithContext(ctx).
		Where("status = ? AND next_run_at IS NOT NULL AND next_run_at <= ?", domain.ScheduledTaskStatusScheduled, now).
		Order("next_run_at asc").
		Find(&tasks).Error
	return tasks, err
}

func (r *Repository) UpdateScheduledTask(ctx context.Context, id string, updates map[string]any) error {
	return r.dbWithContext(ctx).Model(&domain.ScheduledTask{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) DeleteScheduledTask(ctx context.Context, id string) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ?", id).Delete(&domain.ScheduledTaskRun{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&domain.ScheduledTask{}).Error
	})
}

func (r *Repository) CreateScheduledTaskRun(ctx context.Context, run *domain.ScheduledTaskRun) error {
	if run.ID == "" {
		run.ID = uuid.NewString()
	}
	return r.dbWithContext(ctx).Create(run).Error
}
