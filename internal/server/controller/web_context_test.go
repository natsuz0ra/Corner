package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestChiContextParamDecodesEscapedPathSegment(t *testing.T) {
	var got string
	r := chi.NewRouter()
	r.Patch("/api/skills/{id}/enabled", func(w http.ResponseWriter, req *http.Request) {
		got = NewChiContext(w, req).Param("id")
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/skills/codex%3Aglobal%3Asequential-thinking/enabled", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d body=%s", resp.Code, resp.Body.String())
	}
	if got != "codex:global:sequential-thinking" {
		t.Fatalf("expected decoded path param, got %q", got)
	}
}
