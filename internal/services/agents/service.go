package agents

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

const AgentsFileName = "AGENTS.md"

type File struct {
	Content string
	Path    string
}

type Service struct {
	globalPath string
}

func NewService(globalPath string) *Service {
	return &Service{globalPath: globalPath}
}

func (s *Service) ReadGlobal(_ context.Context) (File, error) {
	if err := s.ensureGlobalFile(); err != nil {
		return File{}, err
	}
	content, err := os.ReadFile(s.globalPath)
	if err != nil {
		return File{}, err
	}
	return File{Content: string(content), Path: s.globalPath}, nil
}

func (s *Service) UpdateGlobal(_ context.Context, content string) error {
	if err := os.MkdirAll(filepath.Dir(s.globalPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.globalPath, []byte(content), 0o644)
}

func (s *Service) ReadProject(_ context.Context, workingDir string) (string, error) {
	dir := strings.TrimSpace(workingDir)
	if dir == "" {
		return "", nil
	}
	absDir, err := filepath.Abs(dir)
	if err == nil {
		dir = absDir
	}
	root := findProjectRoot(dir)
	dirs := []string{dir}
	if root != "" {
		dirs = dirsFromRootToDir(root, dir)
	}

	parts := make([]string, 0, len(dirs))
	for _, current := range dirs {
		path := filepath.Join(current, AgentsFileName)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		if info.IsDir() || !info.Mode().IsRegular() {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		if trimmed := strings.TrimSpace(string(content)); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, "\n\n"), nil
}

func (s *Service) ensureGlobalFile() error {
	if err := os.MkdirAll(filepath.Dir(s.globalPath), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(s.globalPath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}

func findProjectRoot(dir string) string {
	for current := filepath.Clean(dir); ; current = filepath.Dir(current) {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
	}
}

func dirsFromRootToDir(root, dir string) []string {
	var reversed []string
	for current := filepath.Clean(dir); ; current = filepath.Dir(current) {
		reversed = append(reversed, current)
		if current == root {
			break
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	dirs := make([]string, 0, len(reversed))
	for i := len(reversed) - 1; i >= 0; i-- {
		dirs = append(dirs, reversed[i])
	}
	return dirs
}
