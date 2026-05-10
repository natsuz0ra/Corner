package settings

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slimebot/internal/constants"
	"slimebot/internal/runtime"
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
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	envPath := filepath.Join(runtime.SlimeBotHomeDir(), "config.cfg")
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		t.Fatalf("mkdir env dir failed: %v", err)
	}
	if err := os.WriteFile(envPath, []byte("WEB_SEARCH_API_KEY=test-key\n"), 0o644); err != nil {
		t.Fatalf("write env failed: %v", err)
	}
	store := &memorySettingsStore{values: map[string]string{"language": "en-US", "defaultModel": "gpt", "messagePlatformDefaultModel": "mp"}}
	svc := NewSettingsService(store)

	got, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.WebSearchAPIKey != "test-key" {
		t.Fatalf("expected web search key, got %q", got.WebSearchAPIKey)
	}
}

func TestSettingsService_UpdatePreservesOtherSettingsStoreWrites(t *testing.T) {
	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	if err := svc.Update(context.Background(), UpdateSettingsInput{
		Language:                    stringPtr("en-US"),
		DefaultModel:                stringPtr("gpt-4.1"),
		MessagePlatformDefaultModel: stringPtr("gpt-4.1-mini"),
	}); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if store.values["language"] != "en-US" || store.values["defaultModel"] != "gpt-4.1" || store.values["messagePlatformDefaultModel"] != "gpt-4.1-mini" {
		t.Fatalf("unexpected settings store values: %#v", store.values)
	}
}

func TestSettingsService_GetReturnsEnvErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	_, err := svc.Get(context.Background())
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}

func TestSettingsService_UpdateValidatesApprovalMode(t *testing.T) {
	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	if err := svc.Update(context.Background(), UpdateSettingsInput{ApprovalMode: stringPtr(constants.ApprovalModeAutoReview)}); err != nil {
		t.Fatalf("Update auto_review failed: %v", err)
	}
	if got := store.values[constants.SettingApprovalMode]; got != constants.ApprovalModeAutoReview {
		t.Fatalf("approvalMode = %q, want %q", got, constants.ApprovalModeAutoReview)
	}

	err := svc.Update(context.Background(), UpdateSettingsInput{ApprovalMode: stringPtr("danger")})
	if err == nil {
		t.Fatal("expected invalid approval mode error")
	}
	if got := store.values[constants.SettingApprovalMode]; got != constants.ApprovalModeAutoReview {
		t.Fatalf("invalid mode should not overwrite existing value, got %q", got)
	}
}

func TestSettingsService_GetIncludesMessagePlatformRuntimeDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	envPath := filepath.Join(runtime.SlimeBotHomeDir(), "config.cfg")
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		t.Fatalf("mkdir env dir failed: %v", err)
	}
	if err := os.WriteFile(envPath, nil, 0o644); err != nil {
		t.Fatalf("write env failed: %v", err)
	}

	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)
	got, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.MessagePlatformThinkingLevel != "off" {
		t.Fatalf("message platform thinking level = %q, want off", got.MessagePlatformThinkingLevel)
	}
	if got.MessagePlatformApprovalMode != constants.ApprovalModeStandard {
		t.Fatalf("message platform approval mode = %q, want %q", got.MessagePlatformApprovalMode, constants.ApprovalModeStandard)
	}
}

func TestSettingsService_GetIncludesSandboxDefaults(t *testing.T) {
	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	settings, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if settings.SandboxMode != "workspace-write" {
		t.Fatalf("sandbox mode = %q, want workspace-write", settings.SandboxMode)
	}
	if settings.SandboxNetworkEnabled != true {
		t.Fatal("sandbox network should default to enabled")
	}
}

func TestSettingsService_UpdateValidatesSandboxMode(t *testing.T) {
	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	if err := svc.Update(context.Background(), UpdateSettingsInput{SandboxMode: stringPtr("read-only")}); err != nil {
		t.Fatalf("Update read-only failed: %v", err)
	}
	if got := store.values[constants.SettingSandboxMode]; got != "read-only" {
		t.Fatalf("sandbox mode = %q, want read-only", got)
	}

	err := svc.Update(context.Background(), UpdateSettingsInput{SandboxMode: stringPtr("root")})
	if err == nil {
		t.Fatal("expected invalid sandbox mode error")
	}
	if got := store.values[constants.SettingSandboxMode]; got != "read-only" {
		t.Fatalf("invalid update should preserve sandbox mode, got %q", got)
	}
}

func TestSettingsService_UpdateMessagePlatformRuntimeSettings(t *testing.T) {
	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	err := svc.Update(context.Background(), UpdateSettingsInput{
		MessagePlatformDefaultModel:  stringPtr(""),
		MessagePlatformThinkingLevel: stringPtr("high"),
		MessagePlatformApprovalMode:  stringPtr(constants.ApprovalModeAutoReview),
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if got := store.values[constants.SettingMessagePlatformDefaultModel]; got != "" {
		t.Fatalf("message platform default model = %q, want empty", got)
	}
	if got := store.values[constants.SettingMessagePlatformThinkingLevel]; got != "high" {
		t.Fatalf("message platform thinking level = %q, want high", got)
	}
	if got := store.values[constants.SettingMessagePlatformApprovalMode]; got != constants.ApprovalModeAutoReview {
		t.Fatalf("message platform approval mode = %q, want %q", got, constants.ApprovalModeAutoReview)
	}
}

func TestSettingsService_UpdateValidatesMessagePlatformApprovalMode(t *testing.T) {
	store := &memorySettingsStore{values: map[string]string{}}
	svc := NewSettingsService(store)

	err := svc.Update(context.Background(), UpdateSettingsInput{MessagePlatformApprovalMode: stringPtr("danger")})
	if err == nil {
		t.Fatal("expected invalid message platform approval mode error")
	}
}

func stringPtr(value string) *string {
	return &value
}
