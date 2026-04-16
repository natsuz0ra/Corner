package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type httpClient struct {
	url        string
	headers    map[string]string
	httpClient *http.Client
	id         int64
}

// newHTTPClient builds an HTTP transport client from server config.
func newHTTPClient(cfg *ServerConfig) Client {
	timeoutSec := cfg.Timeout
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	return &httpClient{
		url:     strings.TrimSpace(cfg.URL),
		headers: cfg.Headers,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
	}
}

func (c *httpClient) nextID() int64 {
	return atomic.AddInt64(&c.id, 1)
}

// postRPC sends a JSON-RPC request and returns the result as a map.
func (c *httpClient) postRPC(ctx context.Context, method string, params map[string]any) (map[string]any, error) {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextID(),
		"method":  method,
		"params":  params,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mcp http status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	contentType := resp.Header.Get("Content-Type")
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// If the response is SSE-shaped, extract JSON from the first data line.
	var jsonPayload []byte
	if strings.Contains(contentType, "text/event-stream") {
		jsonPayload, err = extractSSEData(raw)
		if err != nil {
			return nil, fmt.Errorf("SSE parse error (content-type=%s, raw=%q): %w", contentType, string(raw), err)
		}
	} else {
		jsonPayload = raw
	}

	var rpc map[string]any
	if err := json.Unmarshal(jsonPayload, &rpc); err != nil {
		return nil, fmt.Errorf("invalid JSON in MCP HTTP response (content-type=%s, body=%s): %w", contentType, string(jsonPayload), err)
	}
	if errObj, ok := rpc["error"]; ok && errObj != nil {
		return nil, fmt.Errorf("mcp rpc error: %v", errObj)
	}
	result, _ := rpc["result"].(map[string]any)
	return result, nil
}

// extractSSEData pulls the first data: JSON payload from raw SSE text.
func extractSSEData(raw []byte) ([]byte, error) {
	var dataLines []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data:"))
		}
		if len(dataLines) > 0 && line == "" {
			break
		}
	}
	if len(dataLines) == 0 {
		return nil, fmt.Errorf("no data event found in SSE stream")
	}
	return []byte(strings.TrimSpace(strings.Join(dataLines, ""))), nil
}

// ListTools fetches tools from the MCP server into the internal Tool shape.
func (c *httpClient) ListTools(ctx context.Context) ([]Tool, error) {
	result, err := c.postRPC(ctx, "tools/list", map[string]any{})
	if err != nil {
		return nil, err
	}
	return parseTools(result), nil
}

// CallTool invokes a tool and returns a normalized CallResult.
func (c *httpClient) CallTool(ctx context.Context, name string, arguments map[string]any) (*CallResult, error) {
	result, err := c.postRPC(ctx, "tools/call", map[string]any{
		"name":      name,
		"arguments": arguments,
	})
	if err != nil {
		return nil, err
	}
	return parseCallResult(result), nil
}

// Close is a no-op for HTTP; there is no persistent connection to tear down.
func (c *httpClient) Close() error { return nil }
