package anthropic

import (
	"encoding/json"
	"testing"

	llmsvc "slimebot/internal/services/llm"
)

func TestBuildAssistantBlocksPreservesThinkingBeforeToolUse(t *testing.T) {
	blocks := buildAssistantBlocks(llmsvc.ChatMessage{
		Role:    "assistant",
		Content: "I will inspect the file.",
		ThinkingBlocks: []llmsvc.ThinkingBlockInfo{{
			Thinking:  "Need to inspect before answering.",
			Signature: "sig-1",
		}},
		ToolCalls: []llmsvc.ToolCallInfo{{
			ID:        "toolu_1",
			Name:      "exec__run",
			Arguments: `{"command":"pwd"}`,
		}},
	})

	raw, err := json.Marshal(blocks)
	if err != nil {
		t.Fatalf("marshal blocks failed: %v", err)
	}

	var got []map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal blocks failed: %v\njson=%s", err, raw)
	}
	if len(got) != 3 {
		t.Fatalf("expected thinking, text, tool_use blocks, got %d: %s", len(got), raw)
	}
	if got[0]["type"] != "thinking" {
		t.Fatalf("expected first block to be thinking, got %v: %s", got[0]["type"], raw)
	}
	if got[0]["thinking"] != "Need to inspect before answering." {
		t.Fatalf("thinking content was not preserved: %#v", got[0])
	}
	if got[0]["signature"] != "sig-1" {
		t.Fatalf("thinking signature was not preserved: %#v", got[0])
	}
	if got[1]["type"] != "text" {
		t.Fatalf("expected second block to be text, got %v: %s", got[1]["type"], raw)
	}
	if got[2]["type"] != "tool_use" {
		t.Fatalf("expected third block to be tool_use, got %v: %s", got[2]["type"], raw)
	}
}
