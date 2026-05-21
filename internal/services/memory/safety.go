package memory

import (
	"regexp"
	"strings"
)

var invisibleMemoryRunes = regexp.MustCompile(`[\x{200B}-\x{200F}\x{202A}-\x{202E}\x{2060}-\x{206F}\x{FEFF}]`)

var unsafeMemoryPatterns = []string{
	"ignore previous instructions",
	"ignore all previous instructions",
	"disregard previous instructions",
	"forget previous instructions",
	"you are now",
	"act as system",
	"system prompt",
	"developer message",
	"reveal secrets",
	"reveal secret",
	"read secrets",
	"read secret",
	"api key",
	"private key",
	"cat ~/.ssh",
	"cat /etc/passwd",
	"curl ",
	"wget ",
	"nc -",
	"bash -i",
	"sh -i",
	"ssh-rsa",
}

func scanEntrySafety(content string) error {
	if invisibleMemoryRunes.MatchString(content) {
		return ErrUnsafeMemory
	}
	lower := strings.ToLower(content)
	for _, pattern := range unsafeMemoryPatterns {
		if strings.Contains(lower, pattern) {
			return ErrUnsafeMemory
		}
	}
	return nil
}
