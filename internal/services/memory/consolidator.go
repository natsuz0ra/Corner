package memory

import (
	"fmt"
	"slimebot/internal/logging"
	"strings"
)

// Consolidator merges fragmented memories and removes redundancy periodically.
// Similar in spirit to Claude Code's autoDream service.
type Consolidator struct {
	store *FileMemoryStore
}

// NewConsolidator creates a consolidator.
func NewConsolidator(store *FileMemoryStore) *Consolidator {
	return &Consolidator{store: store}
}

// Run performs one consolidation pass: scan all memories and merge same-type fragments.
// Returns (merge count, delete count, error).
func (c *Consolidator) Run() (merged int, deleted int, err error) {
	entries, err := c.store.Scan()
	if err != nil {
		return 0, 0, fmt.Errorf("scan memories for consolidation: %w", err)
	}

	if len(entries) < 2 {
		return 0, 0, nil
	}

	// Group by memory type.
	grouped := make(map[MemoryType][]*MemoryEntry)
	for _, e := range entries {
		grouped[e.Type] = append(grouped[e.Type], e)
	}

	var toDelete []string
	var toCreate []*MemoryEntry

	for _, group := range grouped {
		mergedSet := make(map[string]bool)
		for i := 0; i < len(group); i++ {
			if mergedSet[group[i].Slug()] {
				continue
			}
			for j := i + 1; j < len(group); j++ {
				if mergedSet[group[j].Slug()] {
					continue
				}
				if shouldMerge(group[i], group[j]) {
					merged := mergeEntries(group[i], group[j])
					toCreate = append(toCreate, merged)
					toDelete = append(toDelete, group[i].Slug(), group[j].Slug())
					mergedSet[group[i].Slug()] = true
					mergedSet[group[j].Slug()] = true
					break
				}
			}
		}
	}

	// Delete old entries first.
	for _, slug := range toDelete {
		if delErr := c.store.Delete(slug); delErr != nil {
			logging.Warn("consolidator_delete_failed", "slug", slug, "error", delErr)
		}
	}

	// Then save merged entries.
	for _, entry := range toCreate {
		if saveErr := c.store.Save(entry); saveErr != nil {
			logging.Warn("consolidator_save_failed", "name", entry.Name, "error", saveErr)
		}
	}

	merged = len(toCreate)
	deleted = len(toDelete)
	logging.Info("consolidator_completed", "merged", merged, "deleted", deleted)
	return merged, deleted, nil
}

// shouldMerge reports whether two memories should merge.
// Same type and (same name or highly overlapping descriptions).
func shouldMerge(a, b *MemoryEntry) bool {
	if a.Type != b.Type {
		return false
	}
	if a.Name == b.Name {
		return true
	}
	return isSimilarDescription(a.Description, b.Description)
}

// isSimilarDescription is a simple similarity check for descriptions.
// One string contains the other and the shorter/longer length ratio exceeds 80%.
func isSimilarDescription(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	shorter, longer := a, b
	if len(a) > len(b) {
		shorter, longer = b, a
	}
	if strings.Contains(longer, shorter) && float64(len(shorter))/float64(len(longer)) > 0.8 {
		return true
	}
	return false
}

// mergeEntries merges two memories into one new entry.
func mergeEntries(a, b *MemoryEntry) *MemoryEntry {
	var content strings.Builder
	content.WriteString(a.Content)
	if strings.TrimSpace(b.Content) != "" {
		content.WriteString("\n\n---\n\n")
		content.WriteString(b.Content)
	}

	created := a.Created
	if b.Created.Before(created) {
		created = b.Created
	}

	return &MemoryEntry{
		Name:        pickBetterName(a.Name, b.Name),
		Description: pickLonger(a.Description, b.Description),
		Type:        a.Type,
		Content:     content.String(),
		Created:     created,
	}
}

// pickBetterName prefers the longer name as a simple heuristic.
func pickBetterName(a, b string) string {
	if len(a) >= len(b) {
		return a
	}
	return b
}

// pickLonger returns the longer of two strings.
func pickLonger(a, b string) string {
	if len(a) >= len(b) {
		return a
	}
	return b
}
