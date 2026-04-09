package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"slimebot/internal/logging"
	"strings"
	"time"

	"slimebot/internal/constants"
	"slimebot/internal/domain"
)

// MemoryService wraps FileMemoryStore and exposes a single interface for callers.
// Kept compatible with chat/agent services.
type MemoryService struct {
	store *FileMemoryStore
}

// MemorySearchHit is one hit in a memory search result.
type MemorySearchHit struct {
	Kind      string
	ID        string
	Title     string
	Summary   string
	Score     float64
	CreatedAt time.Time
}

// MemoryQueryResult is the full memory search result.
type MemoryQueryResult struct {
	Query  string
	Hits   []MemorySearchHit
	Output string
}

// NewMemoryService creates the memory service. baseDir is usually ~/.slimebot/memory/.
func NewMemoryService(baseDir string) (*MemoryService, error) {
	store, err := NewFileMemoryStore(baseDir)
	if err != nil {
		return nil, fmt.Errorf("create file memory store: %w", err)
	}
	return &MemoryService{store: store}, nil
}

// Shutdown closes the service.
func (m *MemoryService) Shutdown(ctx context.Context) error {
	if m == nil || m.store == nil {
		return nil
	}
	return m.store.Close()
}

// BuildMemoryContext builds memory context to inject into the chat prompt.
func (m *MemoryService) BuildMemoryContext(ctx context.Context, sessionID string, history []domain.Message) string {
	if m == nil || m.store == nil {
		return ""
	}
	return m.buildMemoryContext(ctx, sessionID, history)
}

// BuildSessionMemoryContextForPrompt is an alias for the legacy API.
func (m *MemoryService) BuildSessionMemoryContextForPrompt(ctx context.Context, sessionID string, history []domain.Message) string {
	return m.BuildMemoryContext(ctx, sessionID, history)
}

// BuildRecentHistory returns recent history messages (legacy API).
func (m *MemoryService) BuildRecentHistory(sessionID string, limit int) ([]domain.Message, error) {
	// Current design does not persist history here; return empty.
	return nil, nil
}

// QueryForAgent searches memories for Agent tool calls.
func (m *MemoryService) QueryForAgent(ctx context.Context, sessionID string, query string, topK int) (MemoryQueryResult, error) {
	result := MemoryQueryResult{Query: strings.TrimSpace(query)}
	if result.Query == "" {
		return result, fmt.Errorf("memory_query query cannot be empty")
	}
	if topK <= 0 {
		topK = constants.MemoryToolDefaultTopK
	}

	entries, err := m.store.Search(result.Query, topK)
	if err != nil {
		return result, fmt.Errorf("search memory: %w", err)
	}

	for _, entry := range entries {
		result.Hits = append(result.Hits, MemorySearchHit{
			Kind:      string(entry.Type),
			ID:        entry.Slug(),
			Title:     entry.Name,
			Summary:   truncateContent(entry.Content, 200),
			Score:     1.0, // bleve already ranks results
			CreatedAt: entry.Created,
		})
	}

	result.Output = buildMemoryQueryOutput(result.Query, nil, result.Hits)
	return result, nil
}

// EnqueueTurnMemory processes the model's memory payload.
// Parses JSON, deduplicates, then writes file-backed memories.
func (m *MemoryService) EnqueueTurnMemory(sessionID, assistantMessageID, rawMemoryPayload string) {
	if m == nil || m.store == nil {
		return
	}
	payload := strings.TrimSpace(rawMemoryPayload)
	if payload == "" {
		return
	}
	logging.Info("memory_process_start", "session", sessionID, "payload_len", len(payload))

	// Try to parse as MemoryEntry JSON.
	entry, err := parseMemoryPayload(payload)
	if err != nil {
		logging.Warn("memory_payload_parse_failed", "error", err)
		return
	}

	// Set session ID.
	entry.SessionID = sessionID

	// Dedup: search for highly similar existing memories.
	duplicates, _ := m.store.Search(entry.Name+" "+entry.Description, 3)
	for _, dup := range duplicates {
		if dup.Slug() == entry.Slug() {
			// Same slug: update path (Save merges internally).
			break
		}
		// Exact duplicate name or description; skip.
		if dup.Name == entry.Name || dup.Description == entry.Description {
			logging.Info("memory_duplicate_skipped", "name", entry.Name, "existing", dup.Name)
			return
		}
	}

	if err := m.store.Save(entry); err != nil {
		logging.Warn("memory_save_failed", "name", entry.Name, "error", err)
	}
}

// ReadEntrypoint reads MEMORY.md content.
func (m *MemoryService) ReadEntrypoint() string {
	if m == nil || m.store == nil {
		return ""
	}
	return m.store.ReadEntrypoint()
}

// Store returns the underlying FileMemoryStore (tests or advanced use).
func (m *MemoryService) Store() *FileMemoryStore {
	return m.store
}

// Consolidate runs one consolidation pass: merge fragments and remove redundancy.
func (m *MemoryService) Consolidate() (merged int, deleted int, err error) {
	if m == nil || m.store == nil {
		return 0, 0, nil
	}
	return NewConsolidator(m.store).Run()
}

// buildMemoryContext builds context from MEMORY.md index plus related memories.
// Uses conversation history as the search query and injects bleve-ranked hits
// instead of dumping all recent memories. Similar to Claude Code findRelevantMemories.
func (m *MemoryService) buildMemoryContext(ctx context.Context, sessionID string, history []domain.Message) string {
	// Read MEMORY.md index.
	entrypoint := m.store.ReadEntrypoint()
	if strings.TrimSpace(entrypoint) == "" {
		return ""
	}

	// Layer 1: always inject the manifest (MEMORY.md index).
	var b strings.Builder
	b.WriteString("<memory_index>\n")
	b.WriteString(entrypoint)
	b.WriteString("\n</memory_index>\n")

	// Layer 2: retrieve full content of session-scoped memories via history query.
	query := extractSearchQuery(history, 3)
	if query == "" {
		return b.String()
	}

	entries, err := m.store.SearchBySession(sessionID, query, constants.MemoryContextTopK)
	if err != nil || len(entries) == 0 {
		// Session-scoped search failed or empty; do not fall back to global (session isolation).
		return b.String()
	}

	if len(entries) == 0 {
		return b.String()
	}

	b.WriteString("<relevant_memories>\n")
	for _, entry := range entries {
		freshness := freshnessLabel(entry.Updated)
		if freshness != "" {
			b.WriteString(fmt.Sprintf("## %s (%s) %s\n", entry.Name, entry.Type, freshness))
		} else {
			b.WriteString(fmt.Sprintf("## %s (%s)\n", entry.Name, entry.Type))
		}
		// Inject full content, not description only.
		if strings.TrimSpace(entry.Content) != "" {
			b.WriteString(entry.Content)
		} else {
			b.WriteString(entry.Description)
		}
		b.WriteString("\n\n")
	}
	b.WriteString("</relevant_memories>")

	return b.String()
}

// extractSearchQuery builds search text from the last few turns.
// Concatenates user message text from the last lastN relevant turns.
func extractSearchQuery(history []domain.Message, lastN int) string {
	start := len(history) - lastN*2
	if start < 0 {
		start = 0
	}
	var parts []string
	for i := start; i < len(history); i++ {
		if history[i].Role == "user" {
			content := strings.TrimSpace(history[i].Content)
			if content != "" {
				parts = append(parts, content)
			}
		}
	}
	query := strings.Join(parts, " ")
	if len(query) > 200 {
		query = query[:200]
	}
	return query
}

// parseMemoryPayload parses model JSON payload into a MemoryEntry.
func parseMemoryPayload(raw string) (*MemoryEntry, error) {
	// Strip optional markdown code fence wrapper.
	cleaned := strings.TrimSpace(raw)
	if strings.HasPrefix(cleaned, "```") {
		// Strip opening ```json or ``` marker.
		firstNewline := strings.Index(cleaned, "\n")
		if firstNewline > 0 {
			cleaned = cleaned[firstNewline+1:]
		} else {
			cleaned = cleaned[3:]
		}
		// Strip closing ```.
		if idx := strings.LastIndex(cleaned, "```"); idx >= 0 {
			cleaned = cleaned[:idx]
		}
		cleaned = strings.TrimSpace(cleaned)
	}

	// Parse as standard JSON.
	type payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Content     string `json:"content"`
	}

	var p payload
	if err := json.Unmarshal([]byte(cleaned), &p); err != nil {
		return nil, fmt.Errorf("unmarshal memory payload: %w", err)
	}

	if p.Name == "" {
		return nil, fmt.Errorf("empty name in memory payload")
	}

	memType, err := ParseMemoryType(p.Type)
	if err != nil {
		memType = MemoryTypeProject // default type
	}

	return &MemoryEntry{
		Name:        p.Name,
		Description: p.Description,
		Type:        memType,
		Content:     p.Content,
	}, nil
}

// freshnessLabel returns a freshness tag from update time (see Claude Code memoryAgeDays).
func freshnessLabel(updated time.Time) string {
	days := int(time.Since(updated).Hours() / 24)
	switch {
	case days <= 1:
		return ""
	case days <= 7:
		return fmt.Sprintf("[%d days ago]", days)
	case days <= 30:
		return fmt.Sprintf("[%d days ago, may be stale]", days)
	default:
		return fmt.Sprintf("[%d days ago, verify before use]", days)
	}
}

// truncateContent truncates content to maxRunes runes.
func truncateContent(s string, maxRunes int) string {
	if len([]rune(s)) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes]) + "..."
}

// buildMemoryQueryOutput formats search hits as XML for the tool output.
func buildMemoryQueryOutput(query string, keywords []string, hits []MemorySearchHit) string {
	var b strings.Builder
	b.WriteString("<memory_query_result>\n")
	b.WriteString("query: ")
	b.WriteString(strings.TrimSpace(query))
	b.WriteString("\nkeywords: ")
	if len(keywords) == 0 {
		b.WriteString("(none)")
	} else {
		b.WriteString(strings.Join(keywords, ", "))
	}
	b.WriteString(fmt.Sprintf("\nhit_count: %d\n", len(hits)))
	if len(hits) == 0 {
		b.WriteString("No related memories found.\n</memory_query_result>")
		return b.String()
	}
	for idx, item := range hits {
		b.WriteString(fmt.Sprintf("- [%d] %s | %s | %.2f | %s\n", idx+1, item.Kind, item.Title, item.Score, item.CreatedAt.Format(time.RFC3339)))
		b.WriteString("  ")
		b.WriteString(strings.TrimSpace(item.Summary))
		b.WriteString("\n")
	}
	b.WriteString("</memory_query_result>")
	return strings.TrimSpace(b.String())
}
