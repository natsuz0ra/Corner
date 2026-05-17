package updater

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func AssetNameForPlatform(version, goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("slimebot-%s-%s-%s%s", strings.TrimSpace(version), goos, goarch, ext)
}

func ParseLatestRelease(data []byte) (Release, error) {
	var raw struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		Body        string `json:"body"`
		HTMLURL     string `json:"html_url"`
		PublishedAt string `json:"published_at"`
		Prerelease  bool   `json:"prerelease"`
		Assets      []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return Release{}, err
	}
	if strings.TrimSpace(raw.TagName) == "" {
		return Release{}, fmt.Errorf("release tag_name is empty")
	}
	publishedAt := time.Time{}
	if strings.TrimSpace(raw.PublishedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, raw.PublishedAt)
		if err != nil {
			return Release{}, fmt.Errorf("parse published_at: %w", err)
		}
		publishedAt = parsed
	}
	assets := make([]Asset, 0, len(raw.Assets))
	for _, item := range raw.Assets {
		assets = append(assets, Asset{
			Name:        item.Name,
			DownloadURL: item.DownloadURL,
		})
	}
	return Release{
		TagName:     raw.TagName,
		Name:        raw.Name,
		Body:        raw.Body,
		HTMLURL:     raw.HTMLURL,
		PublishedAt: publishedAt,
		Prerelease:  raw.Prerelease,
		Assets:      assets,
	}, nil
}

func FindAsset(release Release, name string) (Asset, bool) {
	for _, asset := range release.Assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return Asset{}, false
}
