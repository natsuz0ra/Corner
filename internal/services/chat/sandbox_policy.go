package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"slimebot/internal/constants"
	sandboxpolicy "slimebot/internal/sandbox"
)

func (s *ChatService) resolveSandboxPolicy(ctx context.Context) (*sandboxpolicy.Policy, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve sandbox cwd: %w", err)
	}
	mode := string(sandboxpolicy.ModeWorkspaceWrite)
	writableRoots := []string{}
	networkEnabled := true
	allowedDomains := []string{}
	if s.settingsStore != nil {
		if raw, err := s.settingsStore.GetSetting(ctx, constants.SettingSandboxMode); err == nil && strings.TrimSpace(raw) != "" {
			mode = strings.TrimSpace(raw)
		}
		if raw, err := s.settingsStore.GetSetting(ctx, constants.SettingSandboxWritableRoots); err == nil {
			writableRoots = parseStringSliceSetting(raw)
		}
		if raw, err := s.settingsStore.GetSetting(ctx, constants.SettingSandboxNetworkEnabled); err == nil && strings.TrimSpace(raw) != "" {
			switch strings.ToLower(strings.TrimSpace(raw)) {
			case "true", "1", "yes", "on":
				networkEnabled = true
			case "false", "0", "no", "off":
				networkEnabled = false
			default:
				return nil, fmt.Errorf("invalid sandbox network setting: %s", raw)
			}
		}
		if raw, err := s.settingsStore.GetSetting(ctx, constants.SettingSandboxNetworkAllowedDomains); err == nil {
			allowedDomains = parseStringSliceSetting(raw)
		}
	}
	return sandboxpolicy.NewPolicy(sandboxpolicy.Config{
		Mode:          sandboxpolicy.Mode(mode),
		CWD:           cwd,
		WritableRoots: writableRoots,
		Network: sandboxpolicy.NetworkPolicy{
			Enabled:        networkEnabled,
			AllowedDomains: allowedDomains,
		},
	})
}

func parseStringSliceSetting(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return []string{}
	}
	clean := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	return clean
}
