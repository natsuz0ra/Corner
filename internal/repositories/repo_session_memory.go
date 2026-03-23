package repositories

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"slimebot/internal/constants"
	"slimebot/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repository) CreateMemoryFact(input domain.MemoryFactCreateInput) (*domain.MemoryFact, error) {
	fact, err := buildMemoryFact(input)
	if err != nil {
		return nil, err
	}
	if err := r.db.Create(fact).Error; err != nil {
		return nil, err
	}
	return fact, nil
}

func (r *Repository) UpdateMemoryFact(input domain.MemoryFactUpdateInput) (*domain.MemoryFact, error) {
	id := strings.TrimSpace(input.ID)
	sessionID := strings.TrimSpace(input.SessionID)
	if id == "" || sessionID == "" {
		return nil, fmt.Errorf("id and session_id required")
	}
	updates := map[string]any{
		"value":            strings.TrimSpace(input.Value),
		"summary":          strings.TrimSpace(input.Summary),
		"confidence":       clampConfidence(input.Confidence),
		"source_start_seq": input.SourceStartSeq,
		"source_end_seq":   input.SourceEndSeq,
		"last_seen_at":     normalizeFactTime(input.LastSeenAt),
		"updated_at":       time.Now(),
		"version":          gorm.Expr("version + 1"),
	}
	if input.ExpiresAt != nil {
		updates["expires_at"] = input.ExpiresAt
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		updates["status"] = status
	}
	res := r.db.Model(&domain.MemoryFact{}).
		Where("id = ? AND session_id = ?", id, sessionID).
		Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var updated domain.MemoryFact
	if err := r.db.Where("id = ?", id).First(&updated).Error; err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *Repository) FindActiveMemoryFact(ctx context.Context, sessionID, memoryType, subject, predicate string) (*domain.MemoryFact, error) {
	var fact domain.MemoryFact
	err := r.dbWithContext(ctx).
		Where("session_id = ? AND memory_type = ? AND subject = ? AND predicate = ? AND status = ?",
			strings.TrimSpace(sessionID),
			strings.TrimSpace(memoryType),
			strings.TrimSpace(subject),
			strings.TrimSpace(predicate),
			domain.MemoryStatusActive,
		).
		Order("updated_at DESC").
		First(&fact).Error
	if err != nil {
		if isRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &fact, nil
}

func (r *Repository) MarkMemoryFactStale(id, sessionID string) error {
	id = strings.TrimSpace(id)
	sessionID = strings.TrimSpace(sessionID)
	if id == "" || sessionID == "" {
		return fmt.Errorf("id and session_id required")
	}
	now := time.Now()
	res := r.db.Model(&domain.MemoryFact{}).
		Where("id = ? AND session_id = ?", id, sessionID).
		Updates(map[string]any{
			"status":     domain.MemoryStatusStale,
			"updated_at": now,
			"version":    gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) ListMemoryFactsForPrompt(ctx context.Context, sessionID string, limit int, now time.Time) ([]domain.MemoryFact, error) {
	if limit <= 0 {
		return []domain.MemoryFact{}, nil
	}
	query := r.dbWithContext(ctx).
		Where("session_id = ? AND status = ?", strings.TrimSpace(sessionID), domain.MemoryStatusActive).
		Where("(expires_at IS NULL OR expires_at > ?)", normalizeNow(now)).
		Order("updated_at DESC").
		Limit(limit)
	var rows []domain.MemoryFact
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(rows, func(i, j int) bool {
		pi := promptPriority(rows[i].MemoryType)
		pj := promptPriority(rows[j].MemoryType)
		if pi == pj {
			if rows[i].Confidence == rows[j].Confidence {
				return rows[i].UpdatedAt.After(rows[j].UpdatedAt)
			}
			return rows[i].Confidence > rows[j].Confidence
		}
		return pi < pj
	})
	return rows, nil
}

func (r *Repository) SearchMemoryFacts(input domain.MemoryFactSearchInput) ([]domain.MemoryFactSearchHit, error) {
	terms := normalizeKeywords(splitQueryTerms(input.Query))
	if len(terms) == 0 || input.Limit <= 0 {
		return []domain.MemoryFactSearchHit{}, nil
	}
	now := normalizeNow(input.Now)
	candidateLimit := input.Limit * 20
	if candidateLimit < constants.DefaultMemoryCandidateLimit {
		candidateLimit = constants.DefaultMemoryCandidateLimit
	}
	if candidateLimit > constants.MaxMemoryCandidateLimit {
		candidateLimit = constants.MaxMemoryCandidateLimit
	}

	var candidates []domain.MemoryFact
	match := buildFTSMatchQuery(terms)
	if r.ftsSessionMemoriesTableExists() && match != "" {
		raw := `SELECT mf.id FROM memory_facts mf
INNER JOIN memory_facts_fts ON memory_facts_fts.rowid = mf.rowid
WHERE memory_facts_fts MATCH ? AND mf.status = ? AND (mf.expires_at IS NULL OR mf.expires_at > ?)`
		args := []any{match, domain.MemoryStatusActive, now}
		if sid := strings.TrimSpace(input.ExcludeSession); sid != "" {
			raw += ` AND mf.session_id <> ?`
			args = append(args, sid)
		}
		if len(input.MemoryTypes) > 0 {
			raw += ` AND mf.memory_type IN ?`
			args = append(args, normalizeKeywords(input.MemoryTypes))
		}
		raw += ` LIMIT ?`
		args = append(args, candidateLimit)
		var ids []struct{ ID string }
		if err := r.db.Raw(raw, args...).Scan(&ids).Error; err == nil && len(ids) > 0 {
			foundIDs := make([]string, 0, len(ids))
			for _, row := range ids {
				if v := strings.TrimSpace(row.ID); v != "" {
					foundIDs = append(foundIDs, v)
				}
			}
			if len(foundIDs) > 0 {
				if err := r.db.Where("id IN ?", foundIDs).Find(&candidates).Error; err != nil {
					return nil, err
				}
			}
		}
	}
	if len(candidates) == 0 {
		query := r.db.Model(&domain.MemoryFact{}).
			Where("status = ?", domain.MemoryStatusActive).
			Where("(expires_at IS NULL OR expires_at > ?)", now).
			Order("updated_at DESC").
			Limit(candidateLimit)
		if sid := strings.TrimSpace(input.ExcludeSession); sid != "" {
			query = query.Where("session_id <> ?", sid)
		}
		if len(input.MemoryTypes) > 0 {
			query = query.Where("memory_type IN ?", normalizeKeywords(input.MemoryTypes))
		}
		orParts := make([]string, 0, len(terms)*4)
		orArgs := make([]any, 0, len(terms)*4)
		for _, term := range terms {
			like := "%" + term + "%"
			orParts = append(orParts, "summary LIKE ?", "value LIKE ?", "predicate LIKE ?", "subject LIKE ?")
			orArgs = append(orArgs, like, like, like, like)
		}
		if len(orParts) > 0 {
			query = query.Where("("+strings.Join(orParts, " OR ")+")", orArgs...)
		}
		if err := query.Find(&candidates).Error; err != nil {
			return nil, err
		}
	}

	hits := make([]domain.MemoryFactSearchHit, 0, len(candidates))
	for _, fact := range candidates {
		matched := matchFactKeywords(terms, fact)
		if len(matched) == 0 {
			continue
		}
		hits = append(hits, domain.MemoryFactSearchHit{
			Fact:            fact,
			MatchedKeywords: matched,
			Score:           scoreFactHit(fact, matched, now),
		})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].Fact.UpdatedAt.After(hits[j].Fact.UpdatedAt)
		}
		return hits[i].Score > hits[j].Score
	})
	if len(hits) > input.Limit {
		hits = hits[:input.Limit]
	}
	return hits, nil
}

func (r *Repository) GetSessionMemory(ctx context.Context, sessionID string) (*domain.SessionMemory, error) {
	facts, err := r.ListMemoryFactsForPrompt(ctx, sessionID, 1, time.Now())
	if err != nil || len(facts) == 0 {
		return nil, err
	}
	fact := domain.SessionMemory(facts[0])
	return &fact, nil
}

func (r *Repository) GetSessionMemoriesBySessionIDs(sessionIDs []string) ([]domain.SessionMemory, error) {
	if len(sessionIDs) == 0 {
		return []domain.SessionMemory{}, nil
	}
	var rows []domain.MemoryFact
	if err := r.db.Where("session_id IN ?", sessionIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.SessionMemory, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.SessionMemory(row))
	}
	return result, nil
}

func (r *Repository) GetSessionMemoriesByIDs(ids []string) ([]domain.SessionMemory, error) {
	if len(ids) == 0 {
		return []domain.SessionMemory{}, nil
	}
	var rows []domain.MemoryFact
	if err := r.db.Where("id IN ? AND status = ?", ids, domain.MemoryStatusActive).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.SessionMemory, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.SessionMemory(row))
	}
	return result, nil
}

func (r *Repository) CountActiveSessionMemories(sessionID string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.MemoryFact{}).
		Where("session_id = ? AND status = ?", strings.TrimSpace(sessionID), domain.MemoryStatusActive).
		Count(&count).Error
	return count, err
}

func (r *Repository) ListActiveSessionMemories(ctx context.Context, sessionID string) ([]domain.SessionMemory, error) {
	rows, err := r.ListMemoryFactsForPrompt(ctx, sessionID, constants.MaxMemoryCandidateLimit, time.Now())
	if err != nil {
		return nil, err
	}
	result := make([]domain.SessionMemory, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.SessionMemory(row))
	}
	return result, nil
}

func (r *Repository) ListRecentActiveSessionMemories(sessionID string, limit int) ([]domain.SessionMemory, error) {
	rows, err := r.ListMemoryFactsForPrompt(context.Background(), sessionID, limit, time.Now())
	if err != nil {
		return nil, err
	}
	result := make([]domain.SessionMemory, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.SessionMemory(row))
	}
	return result, nil
}

func (r *Repository) CreateSessionMemory(input domain.SessionMemoryCreateInput) (*domain.SessionMemory, error) {
	fact, err := r.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      input.SessionID,
		MemoryType:     domain.MemoryTypeProject,
		Subject:        "session",
		Predicate:      "summary",
		Value:          strings.TrimSpace(input.Summary),
		Summary:        strings.TrimSpace(input.Summary),
		Confidence:     0.7,
		SourceStartSeq: int64(input.SourceMessageCount),
		SourceEndSeq:   int64(input.SourceMessageCount),
		LastSeenAt:     time.Now(),
	})
	if err != nil {
		return nil, err
	}
	item := domain.SessionMemory(*fact)
	return &item, nil
}

func (r *Repository) UpdateSessionMemoryContent(id, sessionID, summary string, _ []string, sourceMessageCount int) error {
	_, err := r.UpdateMemoryFact(domain.MemoryFactUpdateInput{
		ID:             id,
		SessionID:      sessionID,
		Value:          summary,
		Summary:        summary,
		Confidence:     0.7,
		SourceStartSeq: int64(sourceMessageCount),
		SourceEndSeq:   int64(sourceMessageCount),
		LastSeenAt:     time.Now(),
		Status:         domain.MemoryStatusActive,
	})
	return err
}

func (r *Repository) SoftDeleteSessionMemory(id, sessionID string) error {
	id = strings.TrimSpace(id)
	sessionID = strings.TrimSpace(sessionID)
	res := r.db.Model(&domain.MemoryFact{}).
		Where("id = ? AND session_id = ?", id, sessionID).
		Updates(map[string]any{
			"status":     domain.MemoryStatusArchived,
			"updated_at": time.Now(),
			"version":    gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) UpsertSessionMemoryIfNewer(input domain.SessionMemoryUpsertInput) (bool, error) {
	existing, err := r.FindActiveMemoryFact(context.Background(), input.SessionID, domain.MemoryTypeProject, "session", "summary")
	if err != nil {
		return false, err
	}
	if existing != nil && int64(input.SourceMessageCount) < existing.SourceEndSeq {
		return false, nil
	}
	if existing != nil {
		_, err = r.UpdateMemoryFact(domain.MemoryFactUpdateInput{
			ID:             existing.ID,
			SessionID:      input.SessionID,
			Value:          input.Summary,
			Summary:        input.Summary,
			Confidence:     0.7,
			SourceStartSeq: existing.SourceStartSeq,
			SourceEndSeq:   int64(input.SourceMessageCount),
			LastSeenAt:     time.Now(),
			Status:         domain.MemoryStatusActive,
		})
		return err == nil, err
	}
	_, err = r.CreateMemoryFact(domain.MemoryFactCreateInput{
		SessionID:      input.SessionID,
		MemoryType:     domain.MemoryTypeProject,
		Subject:        "session",
		Predicate:      "summary",
		Value:          input.Summary,
		Summary:        input.Summary,
		Confidence:     0.7,
		SourceStartSeq: int64(input.SourceMessageCount),
		SourceEndSeq:   int64(input.SourceMessageCount),
		LastSeenAt:     time.Now(),
	})
	return err == nil, err
}

func (r *Repository) ftsSessionMemoriesTableExists() bool {
	r.ftsOnce.Do(func() {
		var n int64
		_ = r.db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='memory_facts_fts'`).Scan(&n)
		r.ftsOK = n > 0
	})
	return r.ftsOK
}

func (r *Repository) SearchMemoriesByKeywords(keywords []string, limit int, excludeSessionID string) ([]domain.SessionMemorySearchHit, error) {
	hits, err := r.SearchMemoryFacts(domain.MemoryFactSearchInput{
		Query:          strings.Join(keywords, " "),
		Limit:          limit,
		ExcludeSession: excludeSessionID,
		Now:            time.Now(),
	})
	if err != nil {
		return nil, err
	}
	result := make([]domain.SessionMemorySearchHit, 0, len(hits))
	for _, hit := range hits {
		result = append(result, domain.SessionMemorySearchHit{
			Memory:          domain.SessionMemory(hit.Fact),
			MatchedKeywords: hit.MatchedKeywords,
			Score:           hit.Score,
		})
	}
	return result, nil
}

func (r *Repository) CountSessionMessages(sessionID string) (int64, error) {
	var total int64
	err := r.db.Model(&domain.Message{}).
		Where("session_id = ?", strings.TrimSpace(sessionID)).
		Count(&total).
		Error
	return total, err
}

func isRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func normalizeKeywords(keywords []string) []string {
	seen := make(map[string]struct{}, len(keywords))
	result := make([]string, 0, len(keywords))
	for _, item := range keywords {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func buildMemoryFact(input domain.MemoryFactCreateInput) (*domain.MemoryFact, error) {
	sessionID := strings.TrimSpace(input.SessionID)
	memoryType := strings.TrimSpace(input.MemoryType)
	subject := strings.TrimSpace(input.Subject)
	predicate := strings.TrimSpace(input.Predicate)
	value := strings.TrimSpace(input.Value)
	summary := strings.TrimSpace(input.Summary)
	if sessionID == "" || memoryType == "" || subject == "" || predicate == "" || value == "" || summary == "" {
		return nil, fmt.Errorf("memory fact requires session_id, type, subject, predicate, value and summary")
	}
	now := time.Now()
	return &domain.MemoryFact{
		ID:             uuid.NewString(),
		SessionID:      sessionID,
		MemoryType:     memoryType,
		Subject:        subject,
		Predicate:      predicate,
		Value:          value,
		Summary:        summary,
		Confidence:     clampConfidence(input.Confidence),
		Status:         domain.MemoryStatusActive,
		SourceStartSeq: input.SourceStartSeq,
		SourceEndSeq:   input.SourceEndSeq,
		LastSeenAt:     normalizeFactTime(input.LastSeenAt),
		ExpiresAt:      input.ExpiresAt,
		Version:        1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func splitQueryTerms(query string) []string {
	return strings.Fields(strings.NewReplacer("\n", " ", "\t", " ", ",", " ", "，", " ").Replace(query))
}

func factTerms(fact domain.MemoryFact) []string {
	return normalizeKeywords(splitQueryTerms(strings.Join([]string{
		fact.MemoryType,
		fact.Subject,
		fact.Predicate,
		fact.Value,
		fact.Summary,
	}, " ")))
}

func intersectKeywords(queries []string, candidate []string) []string {
	if len(queries) == 0 || len(candidate) == 0 {
		return []string{}
	}
	candidateSet := make(map[string]struct{}, len(candidate))
	for _, item := range candidate {
		candidateSet[item] = struct{}{}
	}
	result := make([]string, 0, len(queries))
	for _, query := range queries {
		if _, ok := candidateSet[query]; ok {
			result = append(result, query)
		}
	}
	return result
}

func buildFTSMatchQuery(keywords []string) string {
	parts := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		if kw == "" {
			continue
		}
		parts = append(parts, `"`+strings.ReplaceAll(kw, `"`, `""`)+`"`)
	}
	return strings.Join(parts, " OR ")
}

func promptPriority(memoryType string) int {
	switch strings.TrimSpace(memoryType) {
	case domain.MemoryTypeConstraint:
		return 0
	case domain.MemoryTypeTask:
		return 1
	case domain.MemoryTypePreference:
		return 2
	case domain.MemoryTypeProject:
		return 3
	default:
		return 4
	}
}

func clampConfidence(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func normalizeFactTime(v time.Time) time.Time {
	if v.IsZero() {
		return time.Now()
	}
	return v
}

func normalizeNow(v time.Time) time.Time {
	if v.IsZero() {
		return time.Now()
	}
	return v
}

func scoreFactHit(fact domain.MemoryFact, matched []string, now time.Time) float64 {
	score := float64(len(matched)) * 100
	if fact.Status == domain.MemoryStatusActive {
		score += 80
	}
	score += fact.Confidence * 100
	score += float64(20 - promptPriority(fact.MemoryType)*5)
	ageHours := normalizeNow(now).Sub(fact.LastSeenAt).Hours()
	switch {
	case ageHours <= 24:
		score += 30
	case ageHours <= 24*7:
		score += 20
	case ageHours <= 24*30:
		score += 10
	}
	return score
}

func matchFactKeywords(queries []string, fact domain.MemoryFact) []string {
	text := strings.ToLower(strings.Join([]string{
		fact.MemoryType,
		fact.Subject,
		fact.Predicate,
		fact.Value,
		fact.Summary,
	}, " "))
	result := make([]string, 0, len(queries))
	for _, query := range queries {
		query = strings.TrimSpace(strings.ToLower(query))
		if query == "" {
			continue
		}
		if strings.Contains(text, query) {
			result = append(result, query)
		}
	}
	return result
}
