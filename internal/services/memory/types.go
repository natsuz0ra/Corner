package memory

import (
	"errors"
	"strings"
)

type Target string

const (
	TargetMemory Target = "memory"
	TargetUser   Target = "user"
)

const entrySeparator = "\n§\n"

var (
	ErrInvalidTarget        = errors.New("invalid memory target")
	ErrEmptyEntry           = errors.New("memory entry is empty")
	ErrDuplicateEntry       = errors.New("memory entry already exists")
	ErrEntryNotFound        = errors.New("memory entry not found")
	ErrAmbiguousMatch       = errors.New("memory entry match is ambiguous")
	ErrMemoryLimitExceeded  = errors.New("memory character limit exceeded")
	ErrUnsafeMemory         = errors.New("memory entry failed safety scan")
	ErrMemoryDisabled       = errors.New("memory is disabled")
	ErrUserProfileDisabled  = errors.New("user profile memory is disabled")
	ErrInvalidMemorySetting = errors.New("invalid memory setting")
)

type Config struct {
	Enabled            bool
	UserProfileEnabled bool
	MemoryCharLimit    int
	UserCharLimit      int
	NudgeInterval      int
}

type TargetState struct {
	Target     Target   `json:"target"`
	Entries    []string `json:"entries"`
	UsageChars int      `json:"usageChars"`
	CharLimit  int      `json:"charLimit"`
	EntryCount int      `json:"entryCount"`
	Enabled    bool     `json:"enabled"`
}

type Snapshot struct {
	Memory                   TargetState `json:"memory"`
	User                     TargetState `json:"user"`
	MemoryEnabled            bool        `json:"memoryEnabled"`
	MemoryUserProfileEnabled bool        `json:"memoryUserProfileEnabled"`
	MemoryNudgeInterval      int         `json:"memoryNudgeInterval"`
	MemoryDirectory          string      `json:"memoryDirectory"`
}

func DefaultConfig() Config {
	return Config{
		Enabled:            true,
		UserProfileEnabled: true,
		MemoryCharLimit:    2200,
		UserCharLimit:      1375,
		NudgeInterval:      10,
	}
}

func NormalizeTarget(value string) (Target, error) {
	switch Target(strings.ToLower(strings.TrimSpace(value))) {
	case TargetMemory:
		return TargetMemory, nil
	case TargetUser:
		return TargetUser, nil
	default:
		return "", ErrInvalidTarget
	}
}

func targetLimit(cfg Config, target Target) int {
	if target == TargetUser {
		return cfg.UserCharLimit
	}
	return cfg.MemoryCharLimit
}

func targetEnabled(cfg Config, target Target) bool {
	if !cfg.Enabled {
		return false
	}
	if target == TargetUser {
		return cfg.UserProfileEnabled
	}
	return true
}
