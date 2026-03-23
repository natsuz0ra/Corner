package memory

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"slimebot/internal/domain"
)

type memoryWorkerState struct {
	running           bool
	pending           bool
	lastRawSummary    string
	pendingRawSummary string
}

func updateSummaryAsyncImpl(m *MemoryService, sessionID string, rawSummary string) {
	sessionID = strings.TrimSpace(sessionID)
	rawSummary = strings.TrimSpace(rawSummary)
	if sessionID == "" || rawSummary == "" {
		return
	}

	m.workerMu.Lock()
	state := m.workers[sessionID]
	if state == nil {
		state = &memoryWorkerState{}
		m.workers[sessionID] = state
	}
	state.lastRawSummary = rawSummary
	if state.running {
		state.pending = true
		state.pendingRawSummary = rawSummary
		m.workerMu.Unlock()
		slog.Info("memory_facts_queued", "session", sessionID, "reason", "worker_running")
		return
	}
	state.running = true
	m.workerMu.Unlock()

	m.workerWg.Add(1)
	go func() {
		defer m.workerWg.Done()
		runSummaryWorkerImpl(m, sessionID)
	}()
}

func runSummaryWorkerImpl(m *MemoryService, sessionID string) {
	defer func() {
		m.workerMu.Lock()
		delete(m.workers, sessionID)
		m.workerMu.Unlock()
	}()

	for {
		select {
		case <-m.workerCtx.Done():
			return
		default:
		}
		m.workerMu.Lock()
		state := m.workers[sessionID]
		if state == nil {
			m.workerMu.Unlock()
			return
		}
		raw := state.lastRawSummary
		m.workerMu.Unlock()

		runSummaryOnceImpl(m, sessionID, raw)

		m.workerMu.Lock()
		state = m.workers[sessionID]
		if state == nil {
			m.workerMu.Unlock()
			return
		}
		if state.pending {
			state.pending = false
			state.lastRawSummary = state.pendingRawSummary
			state.pendingRawSummary = ""
			m.workerMu.Unlock()
			continue
		}
		m.workerMu.Unlock()
		return
	}
}

func runSummaryOnceImpl(m *MemoryService, sessionID string, rawSummary string) {
	startAt := time.Now()
	ctx := m.workerCtx
	if ctx == nil {
		ctx = context.Background()
	}

	facts, err := parseMemoryFacts(rawSummary)
	if err != nil {
		slog.Warn("memory_facts_parse_failed", "session", sessionID, "err", err)
		return
	}
	totalMessages, err := m.store.CountSessionMessages(sessionID)
	if err != nil {
		slog.Warn("memory_facts_skip", "session", sessionID, "reason", "count_failed", "err", err)
		return
	}

	for _, fact := range facts {
		if fact.Confidence < minFactConfidence {
			continue
		}
		if err := applyFactCandidate(ctx, m, sessionID, fact, totalMessages); err != nil {
			slog.Warn("memory_fact_apply_failed", "session", sessionID, "type", fact.MemoryType, "subject", fact.Subject, "predicate", fact.Predicate, "err", err)
		}
	}
	slog.Info("memory_facts_updated", "session", sessionID, "facts", len(facts), "total_messages", totalMessages, "cost_ms", time.Since(startAt).Milliseconds())
}

func applyFactCandidate(ctx context.Context, m *MemoryService, sessionID string, candidate memoryFactCandidate, totalMessages int64) error {
	existing, err := m.store.FindActiveMemoryFact(ctx, sessionID, candidate.MemoryType, candidate.Subject, candidate.Predicate)
	if err != nil {
		return err
	}
	expiresAt := candidateExpiresAt(candidate)
	if existing != nil {
		if strings.EqualFold(strings.TrimSpace(existing.Value), strings.TrimSpace(candidate.Value)) {
			updated, err := m.store.UpdateMemoryFact(domain.MemoryFactUpdateInput{
				ID:             existing.ID,
				SessionID:      sessionID,
				Value:          candidate.Value,
				Summary:        candidate.Summary,
				Confidence:     candidate.Confidence,
				SourceStartSeq: existing.SourceStartSeq,
				SourceEndSeq:   totalMessages,
				LastSeenAt:     time.Now(),
				ExpiresAt:      expiresAt,
				Status:         domain.MemoryStatusActive,
			})
			if err != nil {
				return err
			}
			if m.embedding != nil && m.vectorStore != nil && updated != nil {
				_ = upsertMemoryVector(m, ctx, updated.ID, sessionID, updated.Summary, updated.Value, updated.MemoryType)
			}
			return nil
		}
		if err := m.store.MarkMemoryFactStale(existing.ID, sessionID); err != nil {
			return err
		}
		if m.vectorStore != nil {
			_ = m.vectorStore.DeleteMemoryVector(ctx, existing.ID)
		}
	}
	created, err := m.store.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      sessionID,
		MemoryType:     candidate.MemoryType,
		Subject:        candidate.Subject,
		Predicate:      candidate.Predicate,
		Value:          candidate.Value,
		Summary:        candidate.Summary,
		Confidence:     candidate.Confidence,
		SourceStartSeq: totalMessages,
		SourceEndSeq:   totalMessages,
		LastSeenAt:     time.Now(),
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		return err
	}
	if m.embedding != nil && m.vectorStore != nil && created != nil {
		_ = upsertMemoryVector(m, ctx, created.ID, sessionID, created.Summary, created.Value, created.MemoryType)
	}
	return nil
}

func candidateExpiresAt(candidate memoryFactCandidate) *time.Time {
	if candidate.MemoryType != domain.MemoryTypeTask {
		return nil
	}
	if candidate.ExpiresIn == "" {
		v := time.Now().Add(defaultTaskTTL)
		return &v
	}
	d, err := time.ParseDuration(candidate.ExpiresIn)
	if err != nil {
		v := time.Now().Add(defaultTaskTTL)
		return &v
	}
	v := time.Now().Add(d)
	return &v
}
