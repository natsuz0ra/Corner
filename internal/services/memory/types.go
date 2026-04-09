package memory

import (
	"fmt"
	"strings"
	"time"
)

// MemoryType classifies memories; aligned with Claude Code's four types.
type MemoryType string

const (
	MemoryTypeUser      MemoryType = "user"      // user profile, role, goals, preferences
	MemoryTypeFeedback  MemoryType = "feedback"  // working-style guidance (do / don't)
	MemoryTypeProject   MemoryType = "project"   // project context, goals, progress
	MemoryTypeReference MemoryType = "reference" // pointers to external systems
)

var validMemoryTypes = map[MemoryType]bool{
	MemoryTypeUser:      true,
	MemoryTypeFeedback:  true,
	MemoryTypeProject:   true,
	MemoryTypeReference: true,
}

func ParseMemoryType(s string) (MemoryType, error) {
	t := MemoryType(strings.ToLower(strings.TrimSpace(s)))
	if !validMemoryTypes[t] {
		return "", fmt.Errorf("invalid memory type: %q (valid: user, feedback, project, reference)", s)
	}
	return t, nil
}

// MemoryEntry is one memory record, stored as one .md file on disk.
type MemoryEntry struct {
	Name        string     `yaml:"name"`             // display name; also basis for filename
	Description string     `yaml:"description"`      // one-line summary for MEMORY.md index
	Type        MemoryType `yaml:"type"`             // memory type
	SessionID   string     `yaml:"session_id"`       // owning session; empty means global
	Created     time.Time  `yaml:"created"`          // creation time
	Updated     time.Time  `yaml:"updated"`          // last update time
	Content     string     `yaml:"-" json:"content"` // body after YAML frontmatter
	FilePath    string     `yaml:"-" json:"-"`       // absolute path (filled at runtime)
	slug        string     `yaml:"-" json:"-"`       // cached slug to avoid repeated time.Now()
}

// Slug returns a filename-safe identifier.
// Result is cached so repeated calls stay stable.
func (e *MemoryEntry) Slug() string {
	if e.slug != "" {
		return e.slug
	}
	e.slug = Slugify(e.Name)
	if e.slug == "" {
		e.slug = fmt.Sprintf("memory_%d", time.Now().UnixNano())
	}
	return e.slug
}

// SetSlug forces the slug (e.g. when restoring from filename).
func (e *MemoryEntry) SetSlug(s string) {
	e.slug = s
}

// Slugify turns a name into a filename-safe ASCII identifier.
func Slugify(name string) string {
	s := strings.TrimSpace(name)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// FileName returns the filename including extension.
func (e *MemoryEntry) FileName() string {
	return e.Slug() + ".md"
}

// BleveDocument is the bleve index document shape.
type BleveDocument struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	SessionID   string    `json:"session_id"`
	Content     string    `json:"content"`
	Updated     time.Time `json:"updated"`
}

// ToBleveDocument converts this entry to a bleve document.
func (e *MemoryEntry) ToBleveDocument() BleveDocument {
	return BleveDocument{
		Name:        e.Name,
		Description: e.Description,
		Type:        string(e.Type),
		SessionID:   e.SessionID,
		Content:     e.Content,
		Updated:     e.Updated,
	}
}
