package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"
)

func TestStdioRequestUsesMCPJSONLines(t *testing.T) {
	stdin := &recordingWriteCloser{}
	client := &stdioClient{
		stdin:  stdin,
		stdout: io.NopCloser(strings.NewReader(`{"jsonrpc":"2.0","id":1,"result":{"tools":[]}}` + "\n")),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := client.request(ctx, "tools/list", map[string]any{})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if _, ok := result["tools"].([]any); !ok {
		t.Fatalf("expected tools array in result, got %#v", result)
	}

	line := stdin.String()
	if strings.HasPrefix(line, "Content-Length:") {
		t.Fatalf("stdio request used content-length framing: %q", line)
	}
	if !strings.HasSuffix(line, "\n") {
		t.Fatalf("stdio request must end with newline, got %q", line)
	}

	var req map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &req); err != nil {
		t.Fatalf("request was not a JSON line: %v", err)
	}
	if got := req["method"]; got != "tools/list" {
		t.Fatalf("method = %v, want tools/list", got)
	}
}

func TestStdioRequestSkipsNonRPCStdoutLines(t *testing.T) {
	stdin := &recordingWriteCloser{}
	client := &stdioClient{
		stdin: stdin,
		stdout: io.NopCloser(strings.NewReader(
			`$time=1 $scope=app $level=warn $msg="curl transfer error"` + "\n" +
				`{"jsonrpc":"2.0","id":1,"result":{"tools":[]}}` + "\n",
		)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := client.request(ctx, "tools/list", map[string]any{})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if _, ok := result["tools"].([]any); !ok {
		t.Fatalf("expected tools array in result, got %#v", result)
	}
}

type recordingWriteCloser struct {
	bytes.Buffer
}

func (r *recordingWriteCloser) Close() error {
	return nil
}
