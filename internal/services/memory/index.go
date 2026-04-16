package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// rebuildEntrypoint rebuilds the MEMORY.md index file.
// Scans all .md memories and writes a list ordered by update time.
func rebuildEntrypoint(baseDir string, maxLines int) error {
	entries, err := scanMemoryDir(baseDir)
	if err != nil {
		return fmt.Errorf("scan for entrypoint: %w", err)
	}

	var b strings.Builder
	b.WriteString("# Memory Index\n\n")

	for _, entry := range entries {
		line := fmt.Sprintf("- [%s](%s) — %s\n", entry.Name, filepath.Base(entry.FilePath), entry.Description)
		b.WriteString(line)
	}

	content := b.String()

	// Truncate to max line count.
	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		content = strings.Join(lines, "\n")
	}

	entrypointPath := filepath.Join(baseDir, entrypointFileName)
	if err := os.WriteFile(entrypointPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write entrypoint: %w", err)
	}

	return nil
}
