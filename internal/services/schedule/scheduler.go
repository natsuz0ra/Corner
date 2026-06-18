package schedule

import (
	"context"
	"slimebot/internal/logging"
	"time"
)

// Scheduler periodically asks the service to run due tasks.
type Scheduler struct {
	service  *Service
	interval time.Duration
}

func NewScheduler(service *Service, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Scheduler{service: service, interval: interval}
}

func (s *Scheduler) Start(ctx context.Context) {
	if s == nil || s.service == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		if err := s.service.RestoreRunningTasks(ctx); err != nil {
			logging.Warn("scheduled_tasks_restore_running_failed", "err", err)
		}
		s.tick(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.tick(ctx)
			}
		}
	}()
}

func (s *Scheduler) tick(ctx context.Context) {
	if err := s.service.RunDue(ctx); err != nil {
		logging.Warn("scheduled_tasks_tick_failed", "err", err)
	}
}
