package updater

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "v1.26.1", want: "1.26.1"},
		{in: "  1.26.1\n", want: "1.26.1"},
		{in: "", want: ""},
		{in: "dev", want: "dev"},
	}

	for _, tt := range tests {
		if got := NormalizeVersion(tt.in); got != tt.want {
			t.Fatalf("NormalizeVersion(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "same with v prefix", a: "v1.26.1", b: "1.26.1", want: 0},
		{name: "patch newer", a: "v1.26.2", b: "v1.26.1", want: 1},
		{name: "minor older", a: "v1.25.9", b: "v1.26.0", want: -1},
		{name: "major newer", a: "v2.0.0", b: "v1.99.99", want: 1},
		{name: "dev unknown", a: "dev", b: "v1.26.1", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Fatalf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestAssetNameForPlatform(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{goos: "darwin", goarch: "arm64", want: "slimebot-v1.26.2-darwin-arm64.tar.gz"},
		{goos: "linux", goarch: "amd64", want: "slimebot-v1.26.2-linux-amd64.tar.gz"},
		{goos: "windows", goarch: "amd64", want: "slimebot-v1.26.2-windows-amd64.zip"},
	}

	for _, tt := range tests {
		got := AssetNameForPlatform("v1.26.2", tt.goos, tt.goarch)
		if got != tt.want {
			t.Fatalf("AssetNameForPlatform = %q, want %q", got, tt.want)
		}
	}
}

func TestParseLatestRelease(t *testing.T) {
	raw := []byte(`{
		"tag_name": "v1.26.2",
		"name": "SlimeBot v1.26.2",
		"body": "更新说明",
		"html_url": "https://github.com/natsuz0ra/SlimeBot/releases/tag/v1.26.2",
		"published_at": "2026-05-17T01:02:03Z",
		"prerelease": false,
		"assets": [
			{"name":"slimebot-v1.26.2-darwin-arm64.tar.gz","browser_download_url":"https://example.test/slimebot.tar.gz"}
		]
	}`)

	release, err := ParseLatestRelease(raw)
	if err != nil {
		t.Fatalf("ParseLatestRelease failed: %v", err)
	}
	if release.TagName != "v1.26.2" || release.Name != "SlimeBot v1.26.2" {
		t.Fatalf("unexpected release: %+v", release)
	}
	if release.Assets[0].DownloadURL != "https://example.test/slimebot.tar.gz" {
		t.Fatalf("unexpected asset url: %+v", release.Assets[0])
	}
}

func TestStatusStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStatusStore(filepath.Join(dir, "update-status.json"))
	want := JobStatus{
		Phase:      PhaseDownloading,
		Current:    "v1.26.1",
		Target:     "v1.26.2",
		Message:    "Downloading",
		UpdatedAt:  time.Date(2026, 5, 17, 1, 2, 3, 0, time.UTC),
		ManualHint: "slimebot update --version v1.26.2 --yes",
	}

	if err := store.Write(context.Background(), want); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	got, err := store.Read(context.Background())
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if got.Phase != want.Phase || got.Target != want.Target || got.ManualHint != want.ManualHint {
		t.Fatalf("unexpected status: %+v", got)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "update-status.json"))
	if err != nil {
		t.Fatalf("read raw status: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("status is not JSON: %v", err)
	}
}
