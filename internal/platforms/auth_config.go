package platforms

import (
	"encoding/json"
	"fmt"
	"strings"

	"slimebot/internal/constants"
)

type telegramAuthConfig struct {
	BotToken string `json:"botToken"`
}

// ValidateAuthConfig validates platform auth JSON before persistence.
func ValidateAuthConfig(platform string, raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("auth config is empty")
	}
	var asObject map[string]any
	if err := json.Unmarshal([]byte(trimmed), &asObject); err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(platform), constants.TelegramPlatformName) {
		if strings.TrimSpace(ParseTelegramBotToken(trimmed)) == "" {
			return fmt.Errorf("telegram botToken is required")
		}
	}
	return nil
}

// ParseTelegramBotToken reads the bot token from Telegram auth JSON; empty on failure.
func ParseTelegramBotToken(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	var cfg telegramAuthConfig
	if err := json.Unmarshal([]byte(trimmed), &cfg); err != nil {
		return ""
	}
	return strings.TrimSpace(cfg.BotToken)
}
