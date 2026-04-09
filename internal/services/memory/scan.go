package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

const (
	frontmatterDelimiter = "---"
	frontmatterMaxLines  = 30
	maxMemoryFiles       = 200
)

// memoryFrontmatter is the YAML frontmatter shape.
type memoryFrontmatter struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Type        MemoryType `yaml:"type"`
	SessionID   string     `yaml:"session_id"`
	Created     time.Time  `yaml:"created"`
	Updated     time.Time  `yaml:"updated"`
}

// parseMemoryFile parses a single memory file.
func parseMemoryFile(filePath string) (*MemoryEntry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filePath, err)
	}

	fm, content, err := parseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse frontmatter %s: %w", filePath, err)
	}

	entry := &MemoryEntry{
		Name:        fm.Name,
		Description: fm.Description,
		Type:        fm.Type,
		SessionID:   fm.SessionID,
		Created:     fm.Created,
		Updated:     fm.Updated,
		Content:     strings.TrimSpace(content),
		FilePath:    filePath,
	}
	// Restore slug from filename so Slug() stays stable vs recomputing from name.
	entry.SetSlug(strings.TrimSuffix(filepath.Base(filePath), ".md"))
	return entry, nil
}

// parseFrontmatter extracts YAML frontmatter and body from markdown.
func parseFrontmatter(raw string) (*memoryFrontmatter, string, error) {
	raw = strings.TrimSpace(raw)

	if !strings.HasPrefix(raw, frontmatterDelimiter) {
		return nil, "", fmt.Errorf("missing opening frontmatter delimiter")
	}

	// Find closing --- delimiter.
	afterFirst := raw[len(frontmatterDelimiter):]
	closeIdx := strings.Index(afterFirst, "\n"+frontmatterDelimiter)
	if closeIdx < 0 {
		return nil, "", fmt.Errorf("missing closing frontmatter delimiter")
	}

	fmContent := strings.TrimSpace(afterFirst[:closeIdx])
	bodyStart := closeIdx + len("\n"+frontmatterDelimiter)
	body := ""
	if bodyStart < len(afterFirst) {
		body = afterFirst[bodyStart:]
	}

	var fm memoryFrontmatter
	if err := yaml.Unmarshal([]byte(fmContent), &fm); err != nil {
		return nil, "", fmt.Errorf("unmarshal yaml: %w", err)
	}

	// Validate memory type.
	if _, err := ParseMemoryType(string(fm.Type)); err != nil {
		return nil, "", err
	}

	return &fm, body, nil
}

// parseFrontmatterOnly parses frontmatter only (fast scan, no body).
func parseFrontmatterOnly(raw string) (*memoryFrontmatter, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, frontmatterDelimiter) {
		return nil, fmt.Errorf("missing opening frontmatter delimiter")
	}

	afterFirst := raw[len(frontmatterDelimiter):]
	closeIdx := strings.Index(afterFirst, "\n"+frontmatterDelimiter)
	if closeIdx < 0 {
		return nil, fmt.Errorf("missing closing frontmatter delimiter")
	}

	fmContent := strings.TrimSpace(afterFirst[:closeIdx])

	var fm memoryFrontmatter
	if err := yaml.Unmarshal([]byte(fmContent), &fm); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	return &fm, nil
}

// scanMemoryDir scans the directory with frontmatter only (no body).
func scanMemoryDir(baseDir string) ([]*MemoryEntry, error) {
	return scanDir(baseDir, false)
}

// scanMemoryDirFull scans the directory with full content.
func scanMemoryDirFull(baseDir string) ([]*MemoryEntry, error) {
	return scanDir(baseDir, true)
}

func scanDir(baseDir string, fullContent bool) ([]*MemoryEntry, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read dir %s: %w", baseDir, err)
	}

	// Collect .md files (exclude MEMORY.md).
	var mdFiles []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == entrypointFileName {
			continue
		}
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		mdFiles = append(mdFiles, e)
	}

	// Sort by mtime, newest first.
	sort.Slice(mdFiles, func(i, j int) bool {
		fi, _ := mdFiles[i].Info()
		fj, _ := mdFiles[j].Info()
		return fi.ModTime().After(fj.ModTime())
	})

	// Cap file count.
	if len(mdFiles) > maxMemoryFiles {
		mdFiles = mdFiles[:maxMemoryFiles]
	}

	var results []*MemoryEntry
	for _, f := range mdFiles {
		filePath := filepath.Join(baseDir, f.Name())

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		if fullContent {
			entry, parseErr := parseMemoryFile(filePath)
			if parseErr != nil {
				continue
			}
			results = append(results, entry)
		} else {
			// Read full file but only frontmatter is used in this branch.
			content := string(data)
			fm, parseErr := parseFrontmatterOnly(content)
			if parseErr != nil {
				continue
			}
			entry := &MemoryEntry{
				Name:        fm.Name,
				Description: fm.Description,
				Type:        fm.Type,
				SessionID:   fm.SessionID,
				Created:     fm.Created,
				Updated:     fm.Updated,
				FilePath:    filePath,
			}
			entry.SetSlug(strings.TrimSuffix(f.Name(), ".md"))
			results = append(results, entry)
		}
	}

	return results, nil
}
