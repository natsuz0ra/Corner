package settings

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type memorySettingsStore struct {
	values map[string]string
}

func (m *memorySettingsStore) GetSetting(_ context.Context, key string) (string, error) {
	return m.values[key], nil
}

func (m *memorySettingsStore) SetSetting(_ context.Context, key, value string) error {
	m.values[key] = value
	return nil
}

func TestSettingsService_GetIncludesWebSearchAPIKey(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("WEB_SEARCH_API_KEY=test-key\n"), 0o644); err != nil {
		t.Fatalf("write env failed: %v", err)
	}
	store := &memorySettingsStore{values: map[string]string{"language": "en-US", "defaultModel": "gpt", "messagePlatformDefaultModel": "mp"}}
	svc := NewSettingsService(store)

	got, err := svc.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.WebSearchAPIKey != "test-key" {
		t.Fatalf("expected web search key, got %q", got.WebSearchAPIKey)
	}
}

func TestSettingsService_UpdateWritesWebSearchAPIKeyToEnv(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("WEB_SEARCH_API_KEY=old\n"), 0o644); err != nil {
		t.Fatalf("write env failed: %v", err)
	}
	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	if err := svc.Update(UpdateSettingsInput{WebSearchAPIKey: "new-key"}); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	raw, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("read env failed: %v", err)
	}
	if string(raw) != "WEB_SEARCH_API_KEY=new-key\n" {
		t.Fatalf("unexpected env content: %q", string(raw))
	}
	if got := os.Getenv("WEB_SEARCH_API_KEY"); got != "new-key" {
		t.Fatalf("expected process env updated, got %q", got)
	}
}

func TestSettingsService_UpdatePreservesOtherSettingsStoreWrites(t *testing.T) {
	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	if err := svc.Update(UpdateSettingsInput{Language: "en-US", DefaultModel: "gpt-4.1", MessagePlatformDefaultModel: "gpt-4.1-mini"}); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if store.values["language"] != "en-US" || store.values["defaultModel"] != "gpt-4.1" || store.values["messagePlatformDefaultModel"] != "gpt-4.1-mini" {
		t.Fatalf("unexpected settings store values: %#v", store.values)
	}
}

func TestSettingsService_GetReturnsEnvErrors(t *testing.T) {
	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	_, err := svc.Get()
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}
