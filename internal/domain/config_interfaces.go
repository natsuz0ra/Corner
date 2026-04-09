package domain

import "context"

// LLMConfigStore persists LLM model configs.
type LLMConfigStore interface {
	ListLLMConfigs(ctx context.Context) ([]LLMConfig, error)
	CreateLLMConfig(item LLMConfig) (*LLMConfig, error)
	DeleteLLMConfig(id string) error
}

// MCPConfigStore persists MCP server configs.
type MCPConfigStore interface {
	ListMCPConfigs() ([]MCPConfig, error)
	CreateMCPConfig(item MCPConfig) (*MCPConfig, error)
	UpdateMCPConfig(id string, item MCPConfig) error
	DeleteMCPConfig(id string) error
}

// MessagePlatformConfigStore persists message platform configs.
type MessagePlatformConfigStore interface {
	ListMessagePlatformConfigs() ([]MessagePlatformConfig, error)
	CreateMessagePlatformConfig(item MessagePlatformConfig) (*MessagePlatformConfig, error)
	UpdateMessagePlatformConfig(id string, item MessagePlatformConfig) error
	DeleteMessagePlatformConfig(id string) error
}
