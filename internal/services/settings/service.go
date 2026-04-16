package settings

import (
	"context"
	"os"
	"slimebot/internal/constants"
	"slimebot/internal/domain"
	"slimebot/internal/runtime"
	"strings"
)

// AppSettings is the settings DTO exposed to the frontend.
type AppSettings struct {
	Language                    string
	DefaultModel                string
	MessagePlatformDefaultModel string
	WebSearchAPIKey             string
	ApprovalMode                string
}

// UpdateSettingsInput is the domain input for partial settings updates.
type UpdateSettingsInput struct {
	Language                    string
	DefaultModel                string
	MessagePlatformDefaultModel string
	WebSearchAPIKey             string
	ApprovalMode                string
}

type SettingsService struct {
	store domain.SettingsStore
}

func NewSettingsService(store domain.SettingsStore) *SettingsService {
	return &SettingsService{store: store}
}

// Get loads settings and fills defaults for a stable API surface.
func (s *SettingsService) Get() (*AppSettings, error) {
	language, err := s.store.GetSetting(context.Background(), constants.SettingLanguage)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(language) == "" {
		language = "zh-CN"
	}
	defaultModel, err := s.store.GetSetting(context.Background(), constants.SettingDefaultModel)
	if err != nil {
		return nil, err
	}
	messagePlatformDefaultModel, err := s.store.GetSetting(context.Background(), constants.SettingMessagePlatformDefaultModel)
	if err != nil {
		return nil, err
	}
	webSearchAPIKey, err := runtime.ReadEnvValue(constants.SettingWebSearchAPIKey)
	if err != nil {
		return nil, err
	}
	approvalMode, err := s.store.GetSetting(context.Background(), constants.SettingApprovalMode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(approvalMode) == "" {
		approvalMode = constants.ApprovalModeStandard
	}
	return &AppSettings{
		Language:                    language,
		DefaultModel:                defaultModel,
		MessagePlatformDefaultModel: messagePlatformDefaultModel,
		WebSearchAPIKey:             webSearchAPIKey,
		ApprovalMode:                approvalMode,
	}, nil
}

// Update applies only fields that are explicitly set in the request.
func (s *SettingsService) Update(input UpdateSettingsInput) error {
	if strings.TrimSpace(input.Language) != "" {
		if err := s.store.SetSetting(context.Background(), constants.SettingLanguage, input.Language); err != nil {
			return err
		}
	}
	if strings.TrimSpace(input.DefaultModel) != "" {
		if err := s.store.SetSetting(context.Background(), constants.SettingDefaultModel, input.DefaultModel); err != nil {
			return err
		}
	}
	if strings.TrimSpace(input.MessagePlatformDefaultModel) != "" {
		if err := s.store.SetSetting(context.Background(), constants.SettingMessagePlatformDefaultModel, input.MessagePlatformDefaultModel); err != nil {
			return err
		}
	}
	if strings.TrimSpace(input.WebSearchAPIKey) != "" {
		if err := runtime.UpsertEnvValue(constants.SettingWebSearchAPIKey, input.WebSearchAPIKey); err != nil {
			return err
		}
		if err := os.Setenv(constants.SettingWebSearchAPIKey, input.WebSearchAPIKey); err != nil {
			return err
		}
	}
	if strings.TrimSpace(input.ApprovalMode) != "" {
		if err := s.store.SetSetting(context.Background(), constants.SettingApprovalMode, input.ApprovalMode); err != nil {
			return err
		}
	}
	return nil
}
