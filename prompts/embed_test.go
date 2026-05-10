package prompts

import (
	"strings"
	"testing"
)

func TestSystemPrompt_FollowsCurrentUserConversationLanguage(t *testing.T) {
	prompt := SystemPrompt()
	for _, want := range []string{
		"primary language of the user's latest message",
		"current turn's user conversation language",
		"Do not privilege any specific language",
		"clear natural language",
		"user's language",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("system prompt should constrain response language with %q", want)
		}
	}
	for _, forbidden := range []string{
		"Default to Simplified Chinese",
		"Only default to English",
		"English latest message",
		"Chinese latest message",
		"Prefer English keywords",
		"Prefer Chinese keywords",
	} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("system prompt should not contain fallback language bias with %q", forbidden)
		}
	}
}

func TestSystemPrompt_PlacesLanguageBehaviorBeforeInstructionPriority(t *testing.T) {
	prompt := SystemPrompt()
	languageRule := "primary language of the user's latest message"
	instructionPriority := "## 1. Instruction Priority"
	languageIndex := strings.Index(prompt, languageRule)
	priorityIndex := strings.Index(prompt, instructionPriority)
	if languageIndex < 0 {
		t.Fatalf("system prompt missing language rule %q", languageRule)
	}
	if priorityIndex < 0 {
		t.Fatalf("system prompt missing %q", instructionPriority)
	}
	if languageIndex > priorityIndex {
		t.Fatalf("language behavior should appear before instruction priority")
	}
}

func TestSystemPrompt_EncouragesBoundedSubagentDelegation(t *testing.T) {
	prompt := SystemPrompt()
	for _, want := range []string{
		"independent, bounded",
		"Prefer completing small or direct tasks yourself",
		"write the sub-agent `task` and `context` with the expected deliverable",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("system prompt missing %q", want)
		}
	}
	if strings.Contains(prompt, "tool-heavy work") {
		t.Fatal("system prompt should not encourage tool-heavy subagent delegation")
	}
}

func TestSystemPrompt_PrefersFileReadMultiRangeSingleCall(t *testing.T) {
	prompt := SystemPrompt()
	for _, want := range []string{
		"multiple non-contiguous lines/ranges",
		"prefer one call with `requests[].ranges[]`",
		"Keep `offset/limit` for simple single-range reads",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("system prompt missing %q", want)
		}
	}
}
