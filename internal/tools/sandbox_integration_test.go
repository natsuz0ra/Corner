package tools

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	sandboxpolicy "slimebot/internal/sandbox"
)

func TestFileWriteRejectsPathOutsideWorkspaceSandbox(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	policy, err := sandboxpolicy.NewPolicy(sandboxpolicy.Config{Mode: sandboxpolicy.ModeWorkspaceWrite, CWD: workspace})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}
	ctx := sandboxpolicy.WithPolicy(context.Background(), policy)

	_, err = (&fileWriteTool{}).write(ctx, map[string]any{
		"file_path": filepath.Join(outside, "escape.txt"),
		"content":   "nope\n",
	})
	if err == nil || !strings.Contains(err.Error(), "sandbox") {
		t.Fatalf("expected sandbox write rejection, got %v", err)
	}
}

func TestFileWriteAllowsWorkspaceSandboxPath(t *testing.T) {
	workspace := t.TempDir()
	policy, err := sandboxpolicy.NewPolicy(sandboxpolicy.Config{Mode: sandboxpolicy.ModeWorkspaceWrite, CWD: workspace})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}
	ctx := sandboxpolicy.WithPolicy(context.Background(), policy)

	_, err = (&fileWriteTool{}).write(ctx, map[string]any{
		"file_path": filepath.Join(workspace, "ok.txt"),
		"content":   "ok\n",
	})
	if err != nil {
		t.Fatalf("workspace write should be allowed: %v", err)
	}
}

func TestHTTPRequestRejectsNetworkDisabledBySandbox(t *testing.T) {
	policy, err := sandboxpolicy.NewPolicy(sandboxpolicy.Config{
		Mode:    sandboxpolicy.ModeWorkspaceWrite,
		CWD:     t.TempDir(),
		Network: sandboxpolicy.NetworkPolicy{Enabled: false},
	})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}
	ctx := sandboxpolicy.WithPolicy(context.Background(), policy)

	_, err = (&httpRequestTool{client: http.DefaultClient}).request(ctx, map[string]any{
		"method": "GET",
		"url":    "https://example.com",
	})
	if err == nil || !strings.Contains(err.Error(), "sandbox") {
		t.Fatalf("expected sandbox network rejection, got %v", err)
	}
}

func TestExecRequiredApprovalRejectedWithoutSandboxEscalationGrant(t *testing.T) {
	workspace := t.TempDir()
	policy, err := sandboxpolicy.NewPolicy(sandboxpolicy.Config{Mode: sandboxpolicy.ModeWorkspaceWrite, CWD: workspace})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}
	ctx := sandboxpolicy.WithPolicy(context.Background(), policy)

	_, err = (&execTool{}).run(ctx, map[string]any{
		"command":             "go version",
		"description":         "Check Go version",
		"sandbox_permissions": "required_approval",
		"working_directory":   workspace,
	})
	if err == nil || !strings.Contains(err.Error(), "requires approval") {
		t.Fatalf("expected required approval sandbox rejection, got %v", err)
	}
}
