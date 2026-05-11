package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const searchFilesDefaultMaxMatches = 100

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
			{Name: "max_matches", Required: false, Description: "Maximum matches to return. Default 100.", Example: "50"},
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
	if maxMatches <= 0 || maxMatches > searchFilesDefaultMaxMatches {
		maxMatches = searchFilesDefaultMaxMatches
	}
	pattern := paramStringTrim(params, "pattern")

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("failed to stat search path: %w", err)
	}

	var out strings.Builder
	matches := 0
	visit := func(path string, info os.FileInfo) error {
		if matches >= maxMatches {
			return filepath.SkipAll
		}
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
		if strings.Contains(filepath.Base(path), query) {
			out.WriteString(fmt.Sprintf("%s: filename match\n", path))
			matches++
			if matches >= maxMatches {
				return nil
			}
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
				out.WriteString(fmt.Sprintf("%s:%d: %s\n", path, i+1, strings.TrimSpace(line)))
				matches++
				if matches >= maxMatches {
					return nil
				}
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
	if matches == 0 {
		return &ExecuteResult{Output: "No matches found."}, nil
	}
	return &ExecuteResult{Output: fmt.Sprintf("Matches: %d\n%s", matches, strings.TrimSpace(out.String()))}, nil
}

func globMatches(pattern, rel, base string) bool {
	if ok, _ := filepath.Match(pattern, base); ok {
		return true
	}
	ok, _ := filepath.Match(filepath.ToSlash(pattern), filepath.ToSlash(rel))
	return ok
}
