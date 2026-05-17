package updater

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type StatusStore struct {
	path string
	mu   sync.Mutex
}

func NewStatusStore(path string) *StatusStore {
	return &StatusStore{path: path}
}

func (s *StatusStore) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func (s *StatusStore) Read(ctx context.Context) (JobStatus, error) {
	if err := ctx.Err(); err != nil {
		return JobStatus{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return JobStatus{Phase: PhaseIdle}, nil
	}
	if err != nil {
		return JobStatus{}, err
	}
	var status JobStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return JobStatus{}, err
	}
	if status.Phase == "" {
		status.Phase = PhaseIdle
	}
	return status, nil
}

func (s *StatusStore) Write(ctx context.Context, status JobStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if status.UpdatedAt.IsZero() {
		status.UpdatedAt = time.Now().UTC()
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
