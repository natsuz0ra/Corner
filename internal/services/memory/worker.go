package memory

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"slimebot/internal/constants"
	"slimebot/internal/domain"
)

// updateSummaryAsyncImpl 异步触发摘要更新：同会话若已在跑则只打 pending，保证串行。
func updateSummaryAsyncImpl(m *MemoryService, modelConfig ModelRuntimeConfig, sessionID string) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return
	}

	m.workerMu.Lock()
	state := m.workers[sessionID]
	if state == nil {
		state = &memoryWorkerState{}
		m.workers[sessionID] = state
	}
	if state.running {
		state.pending = true
		m.workerMu.Unlock()
		slog.Info("memory_summary_queued", "session", sessionID, "reason", "worker_running")
		return
	}
	state.running = true
	m.workerMu.Unlock()

	m.workerWg.Add(1)
	go func() {
		defer m.workerWg.Done()
		runSummaryWorkerImpl(m, modelConfig, sessionID)
	}()
}

// runSummaryWorkerImpl 循环执行 runSummaryOnceImpl，直至无 pending 或 workerCtx 取消；结束从 workers 表删除。
func runSummaryWorkerImpl(m *MemoryService, modelConfig ModelRuntimeConfig, sessionID string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("memory_summary_panic", "session", sessionID, "recovered", recovered)
		}
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
		runSummaryOnceImpl(m, modelConfig, sessionID)

		m.workerMu.Lock()
		state, ok := m.workers[sessionID]
		if !ok {
			m.workerMu.Unlock()
			return
		}
		if state.pending {
			state.pending = false
			m.workerMu.Unlock()
			slog.Info("memory_summary_worker_continue", "session", sessionID, "reason", "pending_trigger")
			continue
		}
		m.workerMu.Unlock()
		return
	}
}

// runSummaryOnceImpl 单次摘要：读消息数与近期消息 -> MergeSummary -> Upsert；可选写向量。
func runSummaryOnceImpl(m *MemoryService, modelConfig ModelRuntimeConfig, sessionID string) {
	startAt := time.Now()

	ctx := m.workerCtx
	if ctx == nil {
		ctx = context.Background()
	}
	totalMessages, err := m.store.CountSessionMessages(sessionID)
	if err != nil {
		slog.Warn("memory_summary_skip", "session", sessionID, "reason", "count_failed", "err", err)
		return
	}
	if totalMessages == 0 {
		return
	}

	recent, err := m.store.ListRecentSessionMessages(sessionID, constants.MemorySummaryRecentMessageSize)
	if err != nil {
		slog.Warn("memory_summary_skip", "session", sessionID, "reason", "recent_failed", "err", err)
		return
	}
	if len(recent) == 0 {
		return
	}

	existing, err := m.store.GetSessionMemory(sessionID)
	if err != nil {
		slog.Warn("memory_summary_skip", "session", sessionID, "reason", "get_existing_failed", "err", err)
		return
	}

	oldSummary := ""
	if existing != nil {
		oldSummary = existing.Summary
	}

	mergedSummary, attempts, summaryCost, err := m.MergeSummary(ctx, modelConfig, oldSummary, recent)
	if err != nil {
		slog.Warn("memory_summary_skip",
			"session", sessionID,
			"reason", "merge_failed",
			"attempts", attempts,
			"cost_ms", summaryCost.Milliseconds(),
			"timeout_ms", constants.MemorySummaryTimeout.Milliseconds(),
			"err_class", classifyMemoryError(err),
			"err", err,
		)
		return
	}
	keywords := m.TokenizeKeywords(mergedSummary + "\n" + flattenMessages(recent))
	updated, err := m.store.UpsertSessionMemoryIfNewer(domain.SessionMemoryUpsertInput{
		SessionID:          sessionID,
		Summary:            mergedSummary,
		Keywords:           keywords,
		SourceMessageCount: int(totalMessages),
	})
	if err != nil {
		slog.Warn("memory_summary_skip", "session", sessionID, "reason", "upsert_failed", "err", err)
		return
	}
	if !updated {
		slog.Info("memory_summary_skip", "session", sessionID, "reason", "stale_write", "source_message_count", totalMessages)
		return
	}

	if m.embedding != nil && m.vectorStore != nil {
		if err := m.upsertSessionMemoryVector(ctx, sessionID, mergedSummary, keywords, int(totalMessages)); err != nil {
		}
	}

	slog.Info("memory_summary_updated",
		"session", sessionID,
		"total_messages", totalMessages,
		"keywords", len(keywords),
		"attempts", attempts,
		"summary_cost_ms", summaryCost.Milliseconds(),
		"timeout_ms", constants.MemorySummaryTimeout.Milliseconds(),
		"total_cost_ms", time.Since(startAt).Milliseconds(),
	)
}
