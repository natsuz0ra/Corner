package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	ripgrepDefaultTimeout = 20 * time.Second
	ripgrepMaxOutputBytes = 1024 * 1024
)

var (
	ripgrepPathMu sync.RWMutex
	ripgrepPath   = "rg"
)

func setRipgrepPathForTest(path string) func() {
	ripgrepPathMu.Lock()
	old := ripgrepPath
	ripgrepPath = path
	ripgrepPathMu.Unlock()
	return func() {
		ripgrepPathMu.Lock()
		ripgrepPath = old
		ripgrepPathMu.Unlock()
	}
}

func currentRipgrepPath() string {
	ripgrepPathMu.RLock()
	defer ripgrepPathMu.RUnlock()
	return ripgrepPath
}

type ripgrepResult struct {
	lines     []string
	truncated bool
}

func runRipgrep(ctx context.Context, args []string, target string) (ripgrepResult, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ripgrepDefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, currentRipgrepPath(), append(append([]string{}, args...), target)...)
	var stdout limitedBuffer
	var stderr limitedBuffer
	stdout.limit = ripgrepMaxOutputBytes
	stderr.limit = 64 * 1024
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if timeoutCtx.Err() == context.DeadlineExceeded {
		return ripgrepResult{}, fmt.Errorf("ripgrep search timed out after %s; try a more specific path or pattern", ripgrepDefaultTimeout)
	}
	if errors.Is(err, exec.ErrNotFound) {
		return ripgrepResult{}, fmt.Errorf("ripgrep (rg) was not found; install ripgrep or make sure rg is available in PATH")
	}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return ripgrepResult{}, nil
		}
		if strings.TrimSpace(stderr.String()) != "" {
			return ripgrepResult{}, fmt.Errorf("ripgrep failed: %s", strings.TrimSpace(stderr.String()))
		}
		return ripgrepResult{}, fmt.Errorf("ripgrep failed: %w", err)
	}

	lines := splitRipgrepLines(stdout.String())
	return ripgrepResult{lines: lines, truncated: stdout.truncated}, nil
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
