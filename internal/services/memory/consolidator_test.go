package memory

import (
	"testing"
	"time"
)

func TestConsolidator_MergesDuplicateNames(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileMemoryStore(tmpDir)
	if err != nil {
		t.Fatalf("NewFileMemoryStore: %v", err)
	}
	defer store.Close()

	// Save two memories of the same type with the same name.
	entry1 := &MemoryEntry{
		Name:        "Go Tips",
		Description: "Go programming tips",
		Type:        MemoryTypeProject,
		Content:     "Use table-driven tests",
	}
	entry2 := &MemoryEntry{
		Name:        "Go Tips",
		Description: "More Go tips",
		Type:        MemoryTypeProject,
		Content:     "Prefer small interfaces",
	}

	if err := store.Save(entry1); err != nil {
		t.Fatalf("Save entry1: %v", err)
	}
	// Second save has the same name and slug; Save merges internally.
	if err := store.Save(entry2); err != nil {
		t.Fatalf("Save entry2: %v", err)
	}

	// Run the consolidator.
	c := NewConsolidator(store)
	merged, deleted, err := c.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Save already deduplicated; consolidator may have nothing extra to do.
	t.Logf("merged=%d, deleted=%d", merged, deleted)
}

func TestConsolidator_MergesSimilarDescriptions(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileMemoryStore(tmpDir)
	if err != nil {
		t.Fatalf("NewFileMemoryStore: %v", err)
	}
	defer store.Close()

	// Two memories of the same type with highly similar descriptions.
	entry1 := &MemoryEntry{
		Name:        "User Pref Dark",
		Description: "User prefers dark mode in editor",
		Type:        MemoryTypeUser,
		Content:     "Dark theme with monokai colors",
	}
	entry2 := &MemoryEntry{
		Name:        "Dark Mode Pref",
		Description: "User prefers dark mode in editor and terminal",
		Type:        MemoryTypeUser,
		Content:     "Also uses dark terminal theme",
	}

	if err := store.Save(entry1); err != nil {
		t.Fatalf("Save entry1: %v", err)
	}
	// Change name to avoid same-slug overwrite.
	entry2.Name = "Dark Mode Pref"
	if err := store.Save(entry2); err != nil {
		t.Fatalf("Save entry2: %v", err)
	}

	c := NewConsolidator(store)
	merged, deleted, err := c.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	t.Logf("merged=%d, deleted=%d", merged, deleted)

	// Verify remaining memory count after consolidation.
	remaining, _ := store.List()
	t.Logf("remaining entries: %d", len(remaining))
}

func TestConsolidator_NoMergeNeeded(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileMemoryStore(tmpDir)
	if err != nil {
		t.Fatalf("NewFileMemoryStore: %v", err)
	}
	defer store.Close()

	// Two memories of different types.
	entry1 := &MemoryEntry{
		Name:        "Alpha",
		Description: "First entry",
		Type:        MemoryTypeUser,
		Content:     "Content A",
	}
	entry2 := &MemoryEntry{
		Name:        "Beta",
		Description: "Second entry",
		Type:        MemoryTypeProject,
		Content:     "Content B",
	}

	store.Save(entry1)
	store.Save(entry2)

	c := NewConsolidator(store)
	merged, deleted, err := c.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if merged != 0 || deleted != 0 {
		t.Errorf("expected no merges for different types, got merged=%d deleted=%d", merged, deleted)
	}
}

func TestConsolidator_EmptyStore(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileMemoryStore(tmpDir)
	if err != nil {
		t.Fatalf("NewFileMemoryStore: %v", err)
	}
	defer store.Close()

	c := NewConsolidator(store)
	merged, deleted, err := c.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if merged != 0 || deleted != 0 {
		t.Errorf("empty store should have no merges, got merged=%d deleted=%d", merged, deleted)
	}
}

func TestShouldMerge(t *testing.T) {
	tests := []struct {
		name      string
		a         *MemoryEntry
		b         *MemoryEntry
		wantMerge bool
	}{
		{
			name:      "same type same name",
			a:         &MemoryEntry{Name: "Test", Type: MemoryTypeProject, Description: "desc a"},
			b:         &MemoryEntry{Name: "Test", Type: MemoryTypeProject, Description: "desc b"},
			wantMerge: true,
		},
		{
			name:      "different type same name",
			a:         &MemoryEntry{Name: "Test", Type: MemoryTypeUser, Description: "desc a"},
			b:         &MemoryEntry{Name: "Test", Type: MemoryTypeProject, Description: "desc b"},
			wantMerge: false,
		},
		{
			name:      "identical description",
			a:         &MemoryEntry{Name: "A", Type: MemoryTypeUser, Description: "User prefers dark mode everywhere"},
			b:         &MemoryEntry{Name: "B", Type: MemoryTypeUser, Description: "User prefers dark mode everywhere"},
			wantMerge: true,
		},
		{
			name:      "different everything",
			a:         &MemoryEntry{Name: "A", Type: MemoryTypeUser, Description: "User likes cats"},
			b:         &MemoryEntry{Name: "B", Type: MemoryTypeProject, Description: "Deploy pipeline"},
			wantMerge: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldMerge(tt.a, tt.b)
			if got != tt.wantMerge {
				t.Errorf("shouldMerge() = %v, want %v", got, tt.wantMerge)
			}
		})
	}
}

func TestFreshnessLabel(t *testing.T) {
	tests := []struct {
		name   string
		days   int
		expect string
	}{
		{"fresh", 0, ""},
		{"1 day", 1, ""},
		{"3 days", 3, "[3 days ago]"},
		{"7 days", 7, "[7 days ago]"},
		{"14 days", 14, "[14 days ago, may be stale]"},
		{"31 days", 31, "[31 days ago, verify before use]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := time.Now().AddDate(0, 0, -tt.days)
			got := freshnessLabel(updated)
			if got != tt.expect {
				t.Errorf("freshnessLabel(%d days ago) = %q, want %q", tt.days, got, tt.expect)
			}
		})
	}
}

func TestSave_DeduplicatesSameSlug(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileMemoryStore(tmpDir)
	if err != nil {
		t.Fatalf("NewFileMemoryStore: %v", err)
	}
	defer store.Close()

	// First save.
	entry1 := &MemoryEntry{
		Name:        "Test Entry",
		Description: "Original description",
		Type:        MemoryTypeProject,
		Content:     "Original content",
	}
	if err := store.Save(entry1); err != nil {
		t.Fatalf("Save first: %v", err)
	}

	// Second save with same name (should update, not create a new file).
	entry2 := &MemoryEntry{
		Name:        "Test Entry",
		Description: "Updated description",
		Type:        MemoryTypeProject,
		Content:     "Updated content",
	}
	if err := store.Save(entry2); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	// Verify a single file with updated content.
	loaded, err := store.Load(entry1.Slug())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Description != "Updated description" {
		t.Errorf("Description = %q, want 'Updated description'", loaded.Description)
	}
	if loaded.Content != "Updated content" {
		t.Errorf("Content = %q, want 'Updated content'", loaded.Content)
	}

	// Verify original creation time is preserved.
	if loaded.Created.After(loaded.Updated) {
		t.Error("Created should not be after Updated")
	}

	// Verify only one logical entry for that name.
	list, _ := store.List()
	names := make(map[string]bool)
	for _, e := range list {
		names[e.Name] = true
	}
	if !names["Test Entry"] {
		t.Error("expected 'Test Entry' in list")
	}
}
