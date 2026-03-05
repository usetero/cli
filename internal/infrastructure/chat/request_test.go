package chat

import (
	"encoding/json"
	"testing"
)

func TestRequestValidate(t *testing.T) {
	stop := StopReasonEndTurn
	req := Request{
		ConversationID: "e7fdf7ec-fce5-4ca6-a572-bfd6bf8df3c8",
		Messages: []Message{
			{Role: RoleUser, Content: []Block{{Type: BlockTypeText, Text: &Text{Content: "hi"}}}},
			{Role: RoleAssistant, StopReason: &stop, Content: []Block{{Type: BlockTypeThinking, Thinking: &Thinking{Content: "..."}}}},
		},
		Tools: []Tool{{Name: "query", Description: "query logs", InputSchema: map[string]any{"type": "object"}}},
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestRequestPayload(t *testing.T) {
	req := Request{ConversationID: "e7fdf7ec-fce5-4ca6-a572-bfd6bf8df3c8", Messages: []Message{{Role: RoleUser, Content: []Block{{Type: BlockTypeText, Text: &Text{Content: "hi"}}}}}}
	p, err := req.payload()
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	if p.ChatProtocolVersion != protocolVersion {
		t.Fatalf("version mismatch: %q", p.ChatProtocolVersion)
	}
}

func TestToolResultValidation(t *testing.T) {
	ok := json.RawMessage(`{"ok":true}`)
	if err := (Block{Type: BlockTypeToolResult, ToolResult: &ToolResult{ToolUseID: "u1", Content: ok}}).Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := (Block{Type: BlockTypeToolResult, ToolResult: &ToolResult{ToolUseID: "u1", IsError: true}}).Validate(); err == nil {
		t.Fatal("expected error")
	}
}
