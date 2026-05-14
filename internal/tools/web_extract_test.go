package tools

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestWebExtractExtractsReadableHTML(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`<html><head><title>Demo Page</title><script>ignore()</script></head><body><h1>Hello</h1><p>Readable text.</p><style>body{}</style></body></html>`)),
			Request:    req,
		}, nil
	})}

	res, err := (&webExtractTool{client: client}).extract(context.Background(), map[string]any{"url": "https://example.test/page"})
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if !strings.Contains(res.Output, "Title: Demo Page") || !strings.Contains(res.Output, "Hello Readable text.") {
		t.Fatalf("unexpected extraction:\n%s", res.Output)
	}
	if strings.Contains(res.Output, "ignore()") || strings.Contains(res.Output, "body{}") {
		t.Fatalf("script/style content should be removed:\n%s", res.Output)
	}
}

func TestWebExtractRejectsNonHTTPURL(t *testing.T) {
	_, err := (&webExtractTool{client: http.DefaultClient}).extract(context.Background(), map[string]any{"url": "file:///tmp/a.html"})
	if err == nil || !strings.Contains(err.Error(), "http or https") {
		t.Fatalf("expected URL scheme rejection, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
