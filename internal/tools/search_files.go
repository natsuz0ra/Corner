package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	searchFilesDefaultMaxMatches        = 50
	searchFilesMaxMatches               = 100
	searchFilesDefaultMaxMatchesPerFile = 5
	searchFilesMaxLineRunes             = 220
)

type searchFileMatch struct {
	line int
	text string
}

type searchFileGroup struct {
	path    string
	total   int
	matches []searchFileMatch
}

type searchFilesTool struct{}

func init() {
	Register(&searchFilesTool{})
}

func (s *searchFilesTool) Name() string { return "search_files" }

func (s *searchFilesTool) Description() string {
	return "Search local filenames and UTF-8 text file contents with sandbox read checks."
}

func (s *searchFilesTool) Commands() []Command {
	return []Command{{
		Name:        "search",
		Description: "Search under path for matching filenames or text content. Returns path:line matches.",
		Params: []CommandParam{
			{Name: "path", Required: false, Description: "Directory or file to search. Defaults to current working directory.", Example: "/path/to/repo"},
			{Name: "query", Required: true, Description: "Text to search for in file names and UTF-8 file contents.", Example: "func BuildToolDefs"},
			{Name: "pattern", Required: false, Description: "Optional glob matched against file basename or relative path.", Example: "*.go"},
			{Name: "max_matches", Required: false, Description: "Maximum matches to return. Default 50, max 100.", Example: "50"},
			{Name: "max_matches_per_file", Required: false, Description: "Maximum matches to show per file. Default 5.", Example: "5"},
		},
	}}
}

func (s *searchFilesTool) Execute(ctx context.Context, command string, params map[string]any) (*ExecuteResult, error) {
	switch command {
	case "search":
		return s.search(ctx, params)
	default:
		return nil, fmt.Errorf("search_files tool does not support command: %s", command)
	}
}

func (s *searchFilesTool) search(ctx context.Context, params map[string]any) (*ExecuteResult, error) {
	query := paramStringTrim(params, "query")
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	rootRaw := paramStringTrim(params, "path")
	if rootRaw == "" {
		rootRaw = "."
	}
	root, err := resolveFilePath(rootRaw)
	if err != nil {
		return nil, err
	}
	if err := checkSandboxRead(ctx, root); err != nil {
		return nil, err
	}
	maxMatches, _, err := paramInt(params, "max_matches")
	if err != nil {
		return nil, err
	}
	if maxMatches <= 0 {
		maxMatches = searchFilesDefaultMaxMatches
	}
	if maxMatches > searchFilesMaxMatches {
		maxMatches = searchFilesMaxMatches
	}
	maxMatchesPerFile, _, err := paramInt(params, "max_matches_per_file")
	if err != nil {
		return nil, err
	}
	if maxMatchesPerFile <= 0 {
		maxMatchesPerFile = searchFilesDefaultMaxMatchesPerFile
	}
	pattern := paramStringTrim(params, "pattern")

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("failed to stat search path: %w", err)
	}

	groupsByPath := make(map[string]*searchFileGroup)
	var groupOrder []string
	shown := 0
	total := 0
	visit := func(path string, info os.FileInfo) error {
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "web/dist" {
				return filepath.SkipDir
			}
			return checkSandboxRead(ctx, path)
		}
		if isBlockedDevicePath(path) {
			return nil
		}
		if err := checkSandboxRead(ctx, path); err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			rel = filepath.Base(path)
		}
		if pattern != "" && !globMatches(pattern, rel, filepath.Base(path)) {
			return nil
		}
		group := func() *searchFileGroup {
			existing := groupsByPath[path]
			if existing != nil {
				return existing
			}
			next := &searchFileGroup{path: path}
			groupsByPath[path] = next
			groupOrder = append(groupOrder, path)
			return next
		}
		addMatch := func(line int, text string) {
			g := group()
			g.total++
			total++
			if shown >= maxMatches || len(g.matches) >= maxMatchesPerFile {
				return
			}
			g.matches = append(g.matches, searchFileMatch{line: line, text: truncateSearchFileLine(strings.TrimSpace(text))})
			shown++
		}
		if strings.Contains(filepath.Base(path), query) {
			addMatch(0, "filename match")
		}
		if info.Size() > fileReadMaxSizeBytes {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text, err := validateTextBytes(data)
		if err != nil {
			return nil
		}
		for i, line := range splitTextLines(text) {
			if strings.Contains(line, query) {
				addMatch(i+1, line)
			}
		}
		return nil
	}

	if info.IsDir() {
		err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			return visit(path, info)
		})
	} else {
		err = visit(root, info)
	}
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &ExecuteResult{Output: "No matches found."}, nil
	}
	return &ExecuteResult{Output: formatSearchFileGroups(groupOrder, groupsByPath, shown, total)}, nil
}

func globMatches(pattern, rel, base string) bool {
	if ok, _ := filepath.Match(pattern, base); ok {
		return true
	}
	ok, _ := filepath.Match(filepath.ToSlash(pattern), filepath.ToSlash(rel))
	return ok
}

func formatSearchFileGroups(order []string, groups map[string]*searchFileGroup, shown int, total int) string {
	truncated := shown < total
	var out strings.Builder
	out.WriteString(fmt.Sprintf("Matches: %d shown, %d found, files=%d, truncated=%t\n", shown, total, len(order), truncated))
	for _, path := range order {
		group := groups[path]
		if group == nil || len(group.matches) == 0 {
			continue
		}
		if len(group.matches) == group.total {
			out.WriteString(fmt.Sprintf("%s (%d matches)\n", group.path, group.total))
		} else {
			out.WriteString(fmt.Sprintf("%s (%d of %d matches)\n", group.path, len(group.matches), group.total))
		}
		for _, match := range group.matches {
			if match.line <= 0 {
				out.WriteString("  filename match\n")
				continue
			}
			out.WriteString(fmt.Sprintf("  L%d: %s\n", match.line, match.text))
		}
	}
	if truncated {
		out.WriteString("Output truncated. Refine path/pattern/query or raise max_matches/max_matches_per_file for more results.\n")
	}
	return strings.TrimSpace(out.String())
}

func truncateSearchFileLine(line string) string {
	runes := []rune(line)
	if len(runes) <= searchFilesMaxLineRunes {
		return line
	}
	return string(runes[:searchFilesMaxLineRunes]) + "..."
}
