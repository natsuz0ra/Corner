package config

import (
	"context"
	"testing"

	"slimebot/internal/domain"
)

type llmConfigStoreStub struct {
	created domain.LLMConfig
}

func (s *llmConfigStoreStub) ListLLMConfigs(context.Context) ([]domain.LLMConfig, error) {
	return []domain.LLMConfig{s.created}, nil
}

func (s *llmConfigStoreStub) CreateLLMConfig(_ context.Context, item domain.LLMConfig) (*domain.LLMConfig, error) {
	s.created = item
	return &item, nil
}

func (s *llmConfigStoreStub) DeleteLLMConfig(context.Context, string) error {
	return nil
}

func TestLLMConfigService_DefaultsContextSize(t *testing.T) {
	store := &llmConfigStoreStub{}
	svc := NewLLMConfigService(store, 1_000_000)

	item, err := svc.Create(context.Background(), LLMConfigCreateInput{
		Name:    "test",
		BaseURL: "http://fake",
		APIKey:  "key",
		Model:   "fake-model",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if item.ContextSize != 1_000_000 {
		t.Fatalf("expected default context size, got %d", item.ContextSize)
	}
}

func TestLLMConfigService_UsesPositiveContextSize(t *testing.T) {
	store := &llmConfigStoreStub{}
	svc := NewLLMConfigService(store, 1_000_000)

	item, err := svc.Create(context.Background(), LLMConfigCreateInput{
		Name:        "test",
		BaseURL:     "http://fake",
		APIKey:      "key",
		Model:       "fake-model",
		ContextSize: 2048,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if item.ContextSize != 2048 {
		t.Fatalf("expected explicit context size, got %d", item.ContextSize)
	}
}
