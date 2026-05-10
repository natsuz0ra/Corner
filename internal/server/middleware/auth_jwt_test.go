package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"slimebot/internal/auth"
	"slimebot/internal/constants"
)

func TestRequireJWTMarksCLISurfaceForCLIToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/settings", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-CLI-Token", "secret")
	rec := httptest.NewRecorder()
	var got string

	mw := RequireJWT(nil, "secret")
	mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = constants.ClientSurfaceFromContext(r.Context())
	})).ServeHTTP(rec, req)

	if got != constants.ClientSurfaceCLI {
		t.Fatalf("client surface = %q, want cli", got)
	}
}

func TestRequireJWTMarksWebSurfaceForJWT(t *testing.T) {
	manager, err := auth.NewTokenManager("test-secret", 60)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	token, err := manager.Generate("alice")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "http://example.test/api/settings?token="+token, nil)
	rec := httptest.NewRecorder()
	var got string

	mw := RequireJWT(manager, "secret")
	mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = constants.ClientSurfaceFromContext(r.Context())
	})).ServeHTTP(rec, req)

	if got != constants.ClientSurfaceWeb {
		t.Fatalf("client surface = %q, want web", got)
	}
}
