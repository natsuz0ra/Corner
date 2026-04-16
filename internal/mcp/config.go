package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ServerConfig describes how to connect to one MCP server.
type ServerConfig struct {
	Transport      string            `json:"transport"`
	Command        string            `json:"command"`
	Args           []string          `json:"args"`
	URL            string            `json:"url"`
	Headers        map[string]string `json:"headers"`
	Timeout        int               `json:"timeout"`
	SSEReadTimeout int               `json:"sse_read_timeout"`
}

// ParseAndValidateConfig parses and validates MCP JSON config.
func ParseAndValidateConfig(raw string) (*ServerConfig, error) {
	content := strings.TrimSpace(raw)
	if content == "" {
		return nil, fmt.Errorf("config is required.")
	}

	var cfg ServerConfig
	if err := json.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("config JSON is invalid: %w", err)
	}

	transport := strings.TrimSpace(cfg.Transport)
	if transport == "" {
		// Historical default: stdio when transport is omitted.
		transport = "stdio"
		cfg.Transport = transport
	}

	// Required fields differ per transport; validate accordingly.
	switch transport {
	case "stdio":
		if strings.TrimSpace(cfg.Command) == "" {
			return nil, fmt.Errorf("stdio config requires command.")
		}
	case "streamable_http", "sse":
		if strings.TrimSpace(cfg.URL) == "" {
			return nil, fmt.Errorf("%s config requires url.", transport)
		}
	default:
		return nil, fmt.Errorf("unsupported transport: %s", transport)
	}

	return &cfg, nil
}
