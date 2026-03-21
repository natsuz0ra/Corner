package memory

import (
	"context"
	"log/slog"
	"strings"

	"slimebot/internal/domain"
)

// retrieveMemoriesByVectorImpl 关键词拼查询句嵌入 -> Qdrant 相似 session -> 反查 DB 记忆并与查询词算关键词交集。
func retrieveMemoriesByVectorImpl(m *MemoryService, ctx context.Context, keywords []string, excludeSessionID string, limit int) ([]domain.SessionMemorySearchHit, error) {
	if len(keywords) == 0 {
		return []domain.SessionMemorySearchHit{}, nil
	}
	query := strings.Join(keywords, " ")
	queryVector, err := m.embedding.Embed(ctx, query)
	if err != nil {
		slog.Warn("memory_vector_query_embedding_failed",
			"keyword_count", len(keywords),
			"exclude_session", strings.TrimSpace(excludeSessionID),
			"err", err,
		)
		return nil, err
	}
	slog.Info("memory_vector_query_embedding_succeeded",
		"keyword_count", len(keywords),
		"vector_dim", len(queryVector),
		"exclude_session", strings.TrimSpace(excludeSessionID),
	)

	searchLimit := limit
	if m.vectorTopK > searchLimit {
		searchLimit = m.vectorTopK
	}
	vectorHits, err := m.vectorStore.SearchSimilarSessionIDs(ctx, queryVector, searchLimit, excludeSessionID)
	if err != nil {
		slog.Warn("memory_vector_query_failed",
			"keyword_count", len(keywords),
			"search_limit", searchLimit,
			"exclude_session", strings.TrimSpace(excludeSessionID),
			"err", err,
		)
		return nil, err
	}
	if len(vectorHits) == 0 {
		slog.Info("memory_vector_query_no_hit",
			"keyword_count", len(keywords),
			"search_limit", searchLimit,
			"exclude_session", strings.TrimSpace(excludeSessionID),
		)
		return []domain.SessionMemorySearchHit{}, nil
	}
	slog.Info("memory_vector_query_hit",
		"keyword_count", len(keywords),
		"search_limit", searchLimit,
		"raw_hit_count", len(vectorHits),
		"exclude_session", strings.TrimSpace(excludeSessionID),
	)

	sessionIDs := make([]string, 0, len(vectorHits))
	for _, hit := range vectorHits {
		sessionIDs = append(sessionIDs, hit.SessionID)
	}
	memories, err := m.store.GetSessionMemoriesBySessionIDs(sessionIDs)
	if err != nil {
		return nil, err
	}
	memoryBySessionID := make(map[string]domain.SessionMemory, len(memories))
	for _, item := range memories {
		memoryBySessionID[item.SessionID] = item
	}

	results := make([]domain.SessionMemorySearchHit, 0, len(vectorHits))
	for _, hit := range vectorHits {
		memory, ok := memoryBySessionID[hit.SessionID]
		if !ok {
			continue
		}
		matched := intersectKeywordSlicesImpl(keywords, m.TokenizeKeywords(memory.KeywordsText+" "+memory.Summary))
		results = append(results, domain.SessionMemorySearchHit{
			Memory:          memory,
			MatchedKeywords: matched,
			Score:           hit.Score,
		})
		if len(results) >= limit {
			break
		}
	}
	slog.Info("memory_vector_query_resolved",
		"keyword_count", len(keywords),
		"resolved_hit_count", len(results),
		"limit", limit,
	)
	return results, nil
}

// upsertSessionMemoryVectorImpl 对摘要做嵌入并 Upsert 到向量库，payload 含 summary 与关键词等。
func upsertSessionMemoryVectorImpl(m *MemoryService, ctx context.Context, sessionID string, summary string, keywords []string, messageCount int) error {
	vector, err := m.embedding.Embed(ctx, summary)
	if err != nil {
		slog.Warn("memory_vector_generate_failed",
			"session", strings.TrimSpace(sessionID),
			"summary_len", len(strings.TrimSpace(summary)),
			"keyword_count", len(keywords),
			"err", err,
		)
		return err
	}
	slog.Info("memory_vector_generate_succeeded",
		"session", strings.TrimSpace(sessionID),
		"summary_len", len(strings.TrimSpace(summary)),
		"keyword_count", len(keywords),
		"vector_dim", len(vector),
	)
	payload := map[string]any{
		"summary":              summary,
		"source_message_count": messageCount,
	}
	if len(keywords) > 0 {
		keywordValues := make([]any, 0, len(keywords))
		for _, item := range keywords {
			keywordValues = append(keywordValues, item)
		}
		payload["keywords"] = keywordValues
	}
	if err := m.vectorStore.UpsertSessionMemoryVector(ctx, domain.MemoryVectorUpsertInput{
		SessionID: sessionID,
		Vector:    vector,
		Payload:   payload,
	}); err != nil {
		slog.Warn("memory_vector_upsert_failed",
			"session", strings.TrimSpace(sessionID),
			"vector_dim", len(vector),
			"err", err,
		)
		return err
	}
	slog.Info("memory_vector_upsert_succeeded",
		"session", strings.TrimSpace(sessionID),
		"vector_dim", len(vector),
	)
	return nil
}

// intersectKeywordSlicesImpl 返回左右关键词列表的交集（用于向量命中后的可解释 matched 字段）。
func intersectKeywordSlicesImpl(left []string, right []string) []string {
	if len(left) == 0 || len(right) == 0 {
		return []string{}
	}
	rightSet := make(map[string]struct{}, len(right))
	for _, item := range right {
		rightSet[item] = struct{}{}
	}
	result := make([]string, 0, len(left))
	for _, item := range left {
		if _, ok := rightSet[item]; ok {
			result = append(result, item)
		}
	}
	return result
}
