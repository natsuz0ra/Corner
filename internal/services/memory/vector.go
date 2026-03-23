package memory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"slimebot/internal/constants"
	"slimebot/internal/domain"
)

func retrieveMemoriesByVectorImpl(m *MemoryService, ctx context.Context, query string, excludeSessionID string, limit int) ([]domain.MemoryFactSearchHit, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []domain.MemoryFactSearchHit{}, nil
	}
	queryKeywords := m.TokenizeKeywords(q)
	queryVector, err := m.embedding.Embed(ctx, q)
	if err != nil {
		slog.Warn("memory_vector_query_embedding_failed", "keyword_count", len(queryKeywords), "exclude_session", strings.TrimSpace(excludeSessionID), "err", err)
		return nil, err
	}

	searchLimit := limit
	if m.vectorTopK > searchLimit {
		searchLimit = m.vectorTopK
	}
	vectorHits, err := m.vectorStore.SearchSimilarSessionIDs(ctx, queryVector, searchLimit, excludeSessionID)
	if err != nil {
		return nil, err
	}
	if len(vectorHits) == 0 {
		return []domain.MemoryFactSearchHit{}, nil
	}

	memoryIDs := make([]string, 0, len(vectorHits))
	for _, hit := range vectorHits {
		if id := strings.TrimSpace(hit.MemoryID); id != "" {
			memoryIDs = append(memoryIDs, id)
		}
	}
	if len(memoryIDs) == 0 {
		return []domain.MemoryFactSearchHit{}, nil
	}
	memories, err := m.store.GetSessionMemoriesByIDs(memoryIDs)
	if err != nil {
		return nil, err
	}
	memoryByID := make(map[string]domain.SessionMemory, len(memories))
	for _, item := range memories {
		memoryByID[item.ID] = item
	}
	results := make([]domain.MemoryFactSearchHit, 0, len(vectorHits))
	for _, hit := range vectorHits {
		memory, ok := memoryByID[strings.TrimSpace(hit.MemoryID)]
		if !ok || memory.Status != domain.MemoryStatusActive {
			continue
		}
		if memory.ExpiresAt != nil && !memory.ExpiresAt.After(timeNow()) {
			continue
		}
		matched := intersectKeywordSlicesImpl(queryKeywords, factTerms(domain.MemoryFact(memory)))
		results = append(results, domain.MemoryFactSearchHit{
			Fact:            domain.MemoryFact(memory),
			MatchedKeywords: matched,
			Score:           hit.Score + memory.Confidence*10,
		})
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func upsertMemoryVector(m *MemoryService, ctx context.Context, memoryID, sessionID, summary, value, memoryType string) error {
	vector, err := m.embedding.Embed(ctx, strings.TrimSpace(summary+"\n"+value))
	if err != nil {
		return err
	}
	payload := map[string]any{
		"summary":     summary,
		"value":       value,
		"memory_type": memoryType,
	}
	return m.vectorStore.UpsertSessionMemoryVector(ctx, domain.MemoryVectorUpsertInput{
		MemoryID:  memoryID,
		SessionID: sessionID,
		Vector:    vector,
		Payload:   payload,
	})
}

func (m *MemoryService) retrieveMemoriesByVector(ctx context.Context, query string, excludeSessionID string, limit int) ([]domain.MemoryFactSearchHit, error) {
	return retrieveMemoriesByVectorImpl(m, ctx, query, excludeSessionID, limit)
}

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

func buildSearchQuery(history []domain.Message, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	var parts []string
	runeCount := 0
	for i := len(history) - 1; i >= 0 && runeCount < maxRunes; i-- {
		text := strings.TrimSpace(history[i].Content)
		if text == "" {
			continue
		}
		textRunes := []rune(text)
		if runeCount+len(textRunes) > maxRunes {
			textRunes = textRunes[:maxRunes-runeCount]
			text = string(textRunes)
		}
		parts = append([]string{text}, parts...)
		runeCount += len(textRunes)
	}
	return strings.Join(parts, "\n")
}

func (m *MemoryService) buildSessionMemoryContextForPrompt(ctx context.Context, sessionID string, _ []domain.Message) string {
	sid := strings.TrimSpace(sessionID)
	if sid == "" {
		return ""
	}
	facts, err := m.store.ListMemoryFactsForPrompt(ctx, sid, promptFactMaxCount, timeNow())
	if err != nil || len(facts) == 0 {
		return ""
	}
	return formatFactsForPrompt(facts, constants.MemoryContextMaxRunes)
}

func formatFactsForPrompt(facts []domain.MemoryFact, maxRunes int) string {
	if maxRunes <= 0 || len(facts) == 0 {
		return ""
	}
	groups := map[string][]domain.MemoryFact{
		"constraints":     {},
		"active_tasks":    {},
		"preferences":     {},
		"project_context": {},
	}
	for _, fact := range facts {
		switch fact.MemoryType {
		case domain.MemoryTypeConstraint:
			groups["constraints"] = append(groups["constraints"], fact)
		case domain.MemoryTypeTask:
			groups["active_tasks"] = append(groups["active_tasks"], fact)
		case domain.MemoryTypePreference, domain.MemoryTypeProfile:
			groups["preferences"] = append(groups["preferences"], fact)
		default:
			groups["project_context"] = append(groups["project_context"], fact)
		}
	}
	order := []string{"constraints", "active_tasks", "preferences", "project_context"}
	var b strings.Builder
	used := 0
	for _, group := range order {
		items := groups[group]
		if len(items) == 0 {
			continue
		}
		openTag := "<" + group + ">\n"
		closeTag := "</" + group + ">\n"
		if used+len([]rune(openTag))+len([]rune(closeTag))+promptGroupOverhead > maxRunes {
			break
		}
		b.WriteString(openTag)
		used += len([]rune(openTag))
		for _, fact := range items {
			entry := fmt.Sprintf("  <memory id=\"%s\" type=\"%s\" subject=\"%s\" predicate=\"%s\" confidence=\"%.2f\">%s</memory>\n",
				fact.ID, fact.MemoryType, fact.Subject, fact.Predicate, fact.Confidence, strings.TrimSpace(fact.Summary))
			entryRunes := len([]rune(entry))
			if used+entryRunes+len([]rune(closeTag)) > maxRunes {
				break
			}
			b.WriteString(entry)
			used += entryRunes
		}
		b.WriteString(closeTag)
		used += len([]rune(closeTag))
	}
	return strings.TrimSpace(b.String())
}

func timeNow() time.Time {
	return time.Now()
}

func factTerms(fact domain.MemoryFact) []string {
	return normalizeFactTerms(strings.Join([]string{
		fact.MemoryType,
		fact.Subject,
		fact.Predicate,
		fact.Value,
		fact.Summary,
	}, " "))
}

func normalizeFactTerms(text string) []string {
	parts := strings.Fields(strings.NewReplacer("\n", " ", "\t", " ", ",", " ", "，", " ").Replace(text))
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return out
}
