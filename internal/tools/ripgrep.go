package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	ripgrepDefaultTimeout = 20 * time.Second
	ripgrepMaxOutputBytes = 1024 * 1024
)

var (
	ripgrepOverrideMu sync.RWMutex
	ripgrepLookPath   = exec.LookPath
	ripgrepExecutable = os.Executable
	ripgrepRepoRoot   = ""
)

type ripgrepTestOverrides struct {
	lookPath   func(string) (string, error)
	executable func() (string, error)
	repoRoot   string
}

func setRipgrepTestOverrides(overrides ripgrepTestOverrides) func() {
	ripgrepOverrideMu.Lock()
	oldLookPath := ripgrepLookPath
	oldExecutable := ripgrepExecutable
	oldRepoRoot := ripgrepRepoRoot
	if overrides.lookPath != nil {
		ripgrepLookPath = overrides.lookPath
	}
	if overrides.executable != nil {
		ripgrepExecutable = overrides.executable
	}
	ripgrepRepoRoot = overrides.repoRoot
	ripgrepOverrideMu.Unlock()
	return func() {
		ripgrepOverrideMu.Lock()
		ripgrepLookPath = oldLookPath
		ripgrepExecutable = oldExecutable
		ripgrepRepoRoot = oldRepoRoot
		ripgrepOverrideMu.Unlock()
	}
}

func resolveRipgrepCommand() (string, error) {
	ripgrepOverrideMu.RLock()
	lookPath := ripgrepLookPath
	executable := ripgrepExecutable
	repoRoot := ripgrepRepoRoot
	ripgrepOverrideMu.RUnlock()

	if _, err := lookPath("rg"); err == nil {
		return "rg", nil
	}

	candidates := bundledRipgrepCandidates(executable, repoRoot)
	for _, candidate := range candidates {
		if isExecutableFile(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("ripgrep (rg) was not found; tried system PATH and bundled vendor paths (%s). Install ripgrep or make sure the release package includes vendor/ripgrep/%s/%s",
		strings.Join(candidates, ", "),
		ripgrepPlatformDir(),
		ripgrepExecutableName(),
	)
}

func bundledRipgrepCandidates(executable func() (string, error), repoRoot string) []string {
	platformDir := ripgrepPlatformDir()
	name := ripgrepExecutableName()
	candidates := make([]string, 0, 2)

	if exePath, err := executable(); err == nil && exePath != "" {
		candidates = append(candidates, filepath.Join(filepath.Dir(exePath), "vendor", "ripgrep", platformDir, name))
	}
	if repoRoot != "" {
		candidates = append(candidates, filepath.Join(repoRoot, "third_party", "ripgrep", platformDir, name))
	} else if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "third_party", "ripgrep", platformDir, name))
	}

	return candidates
}

func ripgrepPlatformDir() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

func ripgrepExecutableName() string {
	if runtime.GOOS == "windows" {
		return "rg.exe"
	}
	return "rg"
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0o111 != 0
}

type ripgrepResult struct {
	lines     []string
	truncated bool
}

type ripgrepRunOptions struct {
	allowPartialOutput bool
}

func runRipgrep(ctx context.Context, args []string, target string) (ripgrepResult, error) {
	return runRipgrepWithOptions(ctx, args, target, ripgrepRunOptions{})
}

func runRipgrepWithOptions(ctx context.Context, args []string, target string, opts ripgrepRunOptions) (ripgrepResult, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ripgrepDefaultTimeout)
	defer cancel()

	rgPath, err := resolveRipgrepCommand()
	if err != nil {
		return ripgrepResult{}, err
	}

	cmd := exec.CommandContext(timeoutCtx, rgPath, append(append([]string{}, args...), target)...)
	var stdout limitedBuffer
	var stderr limitedBuffer
	stdout.limit = ripgrepMaxOutputBytes
	stderr.limit = 64 * 1024
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if timeoutCtx.Err() == context.DeadlineExceeded {
		return ripgrepResult{}, fmt.Errorf("ripgrep search timed out after %s; try a more specific path or pattern", ripgrepDefaultTimeout)
	}
	if errors.Is(err, exec.ErrNotFound) {
		return ripgrepResult{}, fmt.Errorf("ripgrep (rg) was not found after resolving command %q; install ripgrep or make sure the release package includes bundled vendor ripgrep", rgPath)
	}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return ripgrepResult{}, nil
		}
		lines := splitRipgrepLines(stdout.String())
		stderrText := strings.TrimSpace(stderr.String())
		if opts.allowPartialOutput && errors.As(err, &exitErr) && exitErr.ExitCode() == 2 && (len(lines) > 0 || stderrText == "" || isIgnorableRipgrepTraversalError(stderrText)) {
			return ripgrepResult{lines: lines, truncated: stdout.truncated || stderr.truncated}, nil
		}
		if stderrText != "" {
			return ripgrepResult{}, fmt.Errorf("ripgrep failed: %s", stderrText)
		}
		return ripgrepResult{}, fmt.Errorf("ripgrep failed: %w", err)
	}

	lines := splitRipgrepLines(stdout.String())
	return ripgrepResult{lines: lines, truncated: stdout.truncated}, nil
}

func isIgnorableRipgrepTraversalError(stderr string) bool {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "IO error for operation on") {
			return false
		}
		if !strings.Contains(line, "Operation not permitted") && !strings.Contains(line, "Permission denied") {
			return false
		}
	}
	return strings.TrimSpace(stderr) != ""
}

type limitedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		return len(p), nil
	}
	remaining := b.limit - b.buf.Len()
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}
	_, _ = b.buf.Write(p)
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}

func splitRipgrepLines(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, "\n")
	lines := make([]string, 0, len(parts))
	for _, line := range parts {
		line = strings.TrimSuffix(line, "\r")
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
