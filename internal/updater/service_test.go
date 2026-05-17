package updater

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServiceCheckReportsAvailableUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/example/SlimeBot/releases/latest" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name":"v1.26.2",
			"name":"SlimeBot v1.26.2",
			"body":"更新说明",
			"html_url":"https://github.com/example/SlimeBot/releases/tag/v1.26.2",
			"published_at":"2026-05-17T01:02:03Z",
			"assets":[
				{"name":"slimebot-v1.26.2-darwin-arm64.tar.gz","browser_download_url":"https://example.test/slimebot.tar.gz"}
			]
		}`))
	}))
	defer server.Close()

	service := NewService(ServiceOptions{
		CurrentVersion: "v1.26.1",
		Repo:           "example/SlimeBot",
		APIBaseURL:     server.URL,
		HTTPClient:     server.Client(),
		GOOS:           "darwin",
		GOARCH:         "arm64",
		StatusPath:     t.TempDir() + "/update-status.json",
	})

	got, err := service.Check(context.Background(), true)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !got.UpdateAvailable || !got.CanApply {
		t.Fatalf("expected applicable update, got %+v", got)
	}
	if got.Current != "v1.26.1" || got.Latest != "v1.26.2" {
		t.Fatalf("unexpected versions: %+v", got)
	}
	if got.AssetName != "slimebot-v1.26.2-darwin-arm64.tar.gz" {
		t.Fatalf("unexpected asset name: %+v", got)
	}
}

func TestFetchReleaseByTagUsesTargetEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/example/SlimeBot/releases/tags/v1.26.1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name":"v1.26.1",
			"name":"SlimeBot v1.26.1",
			"assets":[
				{"name":"slimebot-v1.26.1-linux-amd64.tar.gz","browser_download_url":"https://example.test/slimebot.tar.gz"}
			]
		}`))
	}))
	defer server.Close()

	service := NewService(ServiceOptions{
		Repo:       "example/SlimeBot",
		APIBaseURL: server.URL,
		HTTPClient: server.Client(),
		StatusPath: t.TempDir() + "/update-status.json",
	})

	release, err := service.fetchReleaseByTag(context.Background(), "v1.26.1")
	if err != nil {
		t.Fatalf("fetchReleaseByTag failed: %v", err)
	}
	if release.TagName != "v1.26.1" {
		t.Fatalf("unexpected release: %+v", release)
	}
}

func TestServiceApplyStartsHelperWithTarget(t *testing.T) {
	var helper HelperOptions
	service := NewService(ServiceOptions{
		CurrentVersion: "v1.26.1",
		Repo:           "example/SlimeBot",
		StatusPath:     t.TempDir() + "/update-status.json",
		HelperStarter: func(_ context.Context, opts HelperOptions) error {
			helper = opts
			return nil
		},
	})

	status, err := service.Apply(context.Background(), ApplyRequest{TargetVersion: "v1.26.2"})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if status.Phase != PhaseChecking || status.Target != "v1.26.2" {
		t.Fatalf("unexpected status: %+v", status)
	}
	if helper.TargetVersion != "v1.26.2" || helper.Repo != "example/SlimeBot" {
		t.Fatalf("unexpected helper opts: %+v", helper)
	}
	stored, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if stored.Phase != PhaseChecking || stored.Target != "v1.26.2" {
		t.Fatalf("unexpected stored status: %+v", stored)
	}
}

func TestServiceApplyRejectsDevWithoutExplicitTarget(t *testing.T) {
	service := NewService(ServiceOptions{
		CurrentVersion: "dev",
		StatusPath:     t.TempDir() + "/update-status.json",
		HelperStarter: func(_ context.Context, _ HelperOptions) error {
			return errors.New("helper should not start")
		},
	})

	if _, err := service.Apply(context.Background(), ApplyRequest{}); err == nil {
		t.Fatal("expected dev auto update to fail without target")
	}
}

func TestServiceCheckKeepsDevVersionManual(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name":"v1.26.2",
			"name":"SlimeBot v1.26.2",
			"html_url":"https://github.com/example/SlimeBot/releases/tag/v1.26.2",
			"assets":[
				{"name":"slimebot-v1.26.2-linux-amd64.tar.gz","browser_download_url":"https://example.test/slimebot.tar.gz"}
			]
		}`))
	}))
	defer server.Close()

	service := NewService(ServiceOptions{
		CurrentVersion: "dev",
		Repo:           "example/SlimeBot",
		APIBaseURL:     server.URL,
		HTTPClient:     server.Client(),
		GOOS:           "linux",
		GOARCH:         "amd64",
		StatusPath:     t.TempDir() + "/update-status.json",
	})

	got, err := service.Check(context.Background(), true)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !got.UpdateAvailable || got.CanApply {
		t.Fatalf("expected manual-only update for dev build, got %+v", got)
	}
	if !strings.Contains(got.Reason, "dev") {
		t.Fatalf("expected dev reason, got %q", got.Reason)
	}
	if got.ManualHint != "slimebot update --version v1.26.2 --yes" {
		t.Fatalf("unexpected manual hint: %q", got.ManualHint)
	}
}
