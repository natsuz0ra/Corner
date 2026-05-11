package tools

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestWebSearchToolSearchSuccess(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
			"query":"Who is Leo Messi?",
			"answer":"Messi is an Argentine footballer.",
			"results":[
				{"title":"Britannica","url":"https://www.britannica.com/facts/Lionel-Messi","content":"summary text","score":0.88}
			]
		}`)),
			Request: r,
		}, nil
	})}

	tool := newWebSearchTool("https://example.test", client, func() string { return "test-key" })
	result, err := tool.Execute(context.Background(), "search", map[string]any{"query": "Who is Leo Messi?"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Error != "" {
		t.Fatalf("expected empty result error, got %s", result.Error)
	}
	if !strings.Contains(result.Output, "Messi is an Argentine footballer.") {
		t.Fatalf("expected output to contain answer, got %s", result.Output)
	}
	if !strings.Contains(result.Output, "https://www.britannica.com/facts/Lionel-Messi") {
		t.Fatalf("expected output to contain source URL, got %s", result.Output)
	}
}

func TestWebSearchToolMissingAPIKey(t *testing.T) {
	tool := newWebSearchTool("https://api.tavily.com", &http.Client{}, func() string { return "" })
	_, err := tool.Execute(context.Background(), "search", map[string]any{"query": "test"})
	if err == nil {
		t.Fatal("expected error when API key is empty")
	}
	if !strings.Contains(err.Error(), "WEB_SEARCH_API_KEY") {
		t.Fatalf("expected env key hint in error, got %v", err)
	}
}

func TestWebSearchToolUpstreamError(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"detail":{"error":"Invalid topic. Must be 'general' or 'news'."}}`)),
			Request:    r,
		}, nil
	})}

	tool := newWebSearchTool("https://example.test", client, func() string { return "test-key" })
	result, err := tool.Execute(context.Background(), "search", map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !strings.Contains(result.Error, "Invalid topic") {
		t.Fatalf("expected upstream detail error, got %s", result.Error)
	}
}
