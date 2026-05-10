package settings

import (
	"context"
	"fmt"
	"os"
	"slimebot/internal/constants"
	"slimebot/internal/domain"
	"slimebot/internal/runtime"
	"strings"
)

// AppSettings is the settings DTO exposed to the frontend.
type AppSettings struct {
	Language                     string
	DefaultModel                 string
	MessagePlatformDefaultModel  string
	MessagePlatformThinkingLevel string
	MessagePlatformApprovalMode  string
	WebSearchAPIKey              string
	ApprovalMode                 string
	ThinkingLevel                string
}

// UpdateSettingsInput is the domain input for partial settings updates.
type UpdateSettingsInput struct {
	Language                     *string
	DefaultModel                 *string
	MessagePlatformDefaultModel  *string
	MessagePlatformThinkingLevel *string
	MessagePlatformApprovalMode  *string
	WebSearchAPIKey              *string
	ApprovalMode                 *string
	ThinkingLevel                *string
}

type SettingsService struct {
	store domain.SettingsStore
}

func NewSettingsService(store domain.SettingsStore) *SettingsService {
	return &SettingsService{store: store}
}

// Get loads settings and fills defaults for a stable API surface.
func (s *SettingsService) Get(ctx context.Context) (*AppSettings, error) {
	language, err := s.store.GetSetting(ctx, constants.SettingLanguage)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(language) == "" {
		language = "zh-CN"
	}
	defaultModel, err := s.store.GetSetting(ctx, constants.SettingDefaultModel)
	if err != nil {
		return nil, err
	}
	messagePlatformDefaultModel, err := s.store.GetSetting(ctx, constants.SettingMessagePlatformDefaultModel)
	if err != nil {
		return nil, err
	}
	messagePlatformThinkingLevel, err := s.store.GetSetting(ctx, constants.SettingMessagePlatformThinkingLevel)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(messagePlatformThinkingLevel) == "" {
		messagePlatformThinkingLevel = "off"
	}
	messagePlatformApprovalMode, err := s.store.GetSetting(ctx, constants.SettingMessagePlatformApprovalMode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(messagePlatformApprovalMode) == "" {
		messagePlatformApprovalMode = constants.ApprovalModeStandard
	}
	webSearchAPIKey, err := runtime.ReadEnvValue(constants.SettingWebSearchAPIKey)
	if err != nil {
		return nil, err
	}
	approvalMode, err := s.store.GetSetting(ctx, constants.SettingApprovalMode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(approvalMode) == "" {
		approvalMode = constants.ApprovalModeStandard
	}
	thinkingLevel, err := s.store.GetSetting(ctx, constants.SettingThinkingLevel)
	if err != nil {
		return nil, err
	}
	return &AppSettings{
		Language:                     language,
		DefaultModel:                 defaultModel,
		MessagePlatformDefaultModel:  messagePlatformDefaultModel,
		MessagePlatformThinkingLevel: messagePlatformThinkingLevel,
		MessagePlatformApprovalMode:  messagePlatformApprovalMode,
		WebSearchAPIKey:              webSearchAPIKey,
		ApprovalMode:                 approvalMode,
		ThinkingLevel:                thinkingLevel,
	}, nil
}

// Update applies only fields that are explicitly set in the request.
func (s *SettingsService) Update(ctx context.Context, input UpdateSettingsInput) error {
	if input.Language != nil && strings.TrimSpace(*input.Language) != "" {
		if err := s.store.SetSetting(ctx, constants.SettingLanguage, *input.Language); err != nil {
			return err
		}
	}
	if input.DefaultModel != nil && strings.TrimSpace(*input.DefaultModel) != "" {
		if err := s.store.SetSetting(ctx, constants.SettingDefaultModel, *input.DefaultModel); err != nil {
			return err
		}
	}
	if input.MessagePlatformDefaultModel != nil {
		if err := s.store.SetSetting(ctx, constants.SettingMessagePlatformDefaultModel, *input.MessagePlatformDefaultModel); err != nil {
			return err
		}
	}
	if input.MessagePlatformThinkingLevel != nil {
		thinkingLevel := strings.TrimSpace(*input.MessagePlatformThinkingLevel)
		if thinkingLevel == "" {
			thinkingLevel = "off"
		}
		if err := s.store.SetSetting(ctx, constants.SettingMessagePlatformThinkingLevel, thinkingLevel); err != nil {
			return err
		}
	}
	if input.MessagePlatformApprovalMode != nil {
		approvalMode := strings.TrimSpace(*input.MessagePlatformApprovalMode)
		if !isValidApprovalMode(approvalMode) {
			return fmt.Errorf("invalid message platform approval mode: %s", approvalMode)
		}
		if err := s.store.SetSetting(ctx, constants.SettingMessagePlatformApprovalMode, approvalMode); err != nil {
			return err
		}
	}
	if input.WebSearchAPIKey != nil && strings.TrimSpace(*input.WebSearchAPIKey) != "" {
		if err := runtime.UpsertEnvValue(constants.SettingWebSearchAPIKey, *input.WebSearchAPIKey); err != nil {
			return err
		}
		if err := os.Setenv(constants.SettingWebSearchAPIKey, *input.WebSearchAPIKey); err != nil {
			return err
		}
	}
	if input.ApprovalMode != nil && strings.TrimSpace(*input.ApprovalMode) != "" {
		approvalMode := strings.TrimSpace(*input.ApprovalMode)
		if !isValidApprovalMode(approvalMode) {
			return fmt.Errorf("invalid approval mode: %s", approvalMode)
		}
		if err := s.store.SetSetting(ctx, constants.SettingApprovalMode, approvalMode); err != nil {
			return err
		}
	}
	if input.ThinkingLevel != nil && strings.TrimSpace(*input.ThinkingLevel) != "" {
		if err := s.store.SetSetting(ctx, constants.SettingThinkingLevel, *input.ThinkingLevel); err != nil {
			return err
		}
	}
	return nil
}

func isValidApprovalMode(mode string) bool {
	switch strings.TrimSpace(mode) {
	case constants.ApprovalModeStandard, constants.ApprovalModeAutoReview, constants.ApprovalModeAuto:
		return true
	default:
		return false
	}
}
