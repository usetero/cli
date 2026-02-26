package chat

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/usetero/cli/internal/domain"
)

func TestToWireRequest_StripsInternalFields(t *testing.T) {
	t.Parallel()

	req := Request{
		ConversationID: "00000000-0000-0000-0000-000000000001",
		Messages: []domain.Message{{
			ID:             "msg-1",
			ConversationID: "00000000-0000-0000-0000-000000000001",
			Role:           domain.RoleUser,
			CreatedAt:      time.Now(),
			Content: []domain.Block{{
				Index: 7,
				Type:  domain.BlockTypeText,
				Text:  &domain.TextBlock{Content: "hello"},
			}},
		}},
		Tools: []Tool{{
			Name:        "query",
			Description: "Run SQL",
			InputSchema: NewObjectSchema(map[string]Property{"sql": {Type: "string"}}, []string{"sql"}),
		}},
	}

	wireReq, err := toWireRequest(req)
	if err != nil {
		t.Fatalf("toWireRequest() error = %v", err)
	}
	data, err := json.Marshal(wireReq)
	if err != nil {
		t.Fatalf("json.Marshal(wireReq) error = %v", err)
	}
	payload := string(data)

	forbidden := []string{"\"id\"", "\"created_at\"", "\"index\""}
	for _, key := range forbidden {
		if strings.Contains(payload, key) {
			t.Fatalf("payload contains forbidden field %s: %s", key, payload)
		}
	}
	if !strings.Contains(payload, `"chat_protocol_version":"v2"`) {
		t.Fatalf("payload missing chat_protocol_version=v2: %s", payload)
	}
	if !strings.Contains(payload, `"conversation_id":"00000000-0000-0000-0000-000000000001"`) {
		t.Fatalf("payload missing conversation_id: %s", payload)
	}
	if !strings.Contains(payload, `"messages"`) || !strings.Contains(payload, `"content"`) {
		t.Fatalf("payload missing expected fields: %s", payload)
	}
}

func TestToWireRequest_RejectsInvalidDomainBlock(t *testing.T) {
	t.Parallel()

	_, err := toWireRequest(Request{
		ConversationID: "00000000-0000-0000-0000-000000000001",
		Messages: []domain.Message{{
			Role: domain.RoleUser,
			Content: []domain.Block{{
				Type: domain.BlockTypeText,
				Text: nil,
			}},
		}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestToWireRequest_RejectsInvalidRole(t *testing.T) {
	t.Parallel()

	_, err := toWireRequest(Request{
		ConversationID: "00000000-0000-0000-0000-000000000001",
		Messages: []domain.Message{{
			Role: domain.Role("invalid"),
			Content: []domain.Block{{
				Type: domain.BlockTypeText,
				Text: &domain.TextBlock{Content: "hello"},
			}},
		}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestToWireRequest_RejectsInvalidStopReason(t *testing.T) {
	t.Parallel()

	_, err := toWireRequest(Request{
		ConversationID: "00000000-0000-0000-0000-000000000001",
		Messages: []domain.Message{{
			Role:       domain.RoleAssistant,
			StopReason: "invalid",
			Content: []domain.Block{{
				Type: domain.BlockTypeText,
				Text: &domain.TextBlock{Content: "hello"},
			}},
		}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestToWireRequest_SanitizesToolResultContent(t *testing.T) {
	t.Parallel()

	wireReq, err := toWireRequest(Request{
		ConversationID: "00000000-0000-0000-0000-000000000001",
		Messages: []domain.Message{
			{
				Role: domain.RoleAssistant,
				Content: []domain.Block{{
					Type: domain.BlockTypeToolUse,
					ToolUse: &domain.ToolUse{
						ID:    "tool-1",
						Name:  "query",
						Input: json.RawMessage(`{"sql":"select 1"}`),
					},
				}},
			},
			{
				Role: domain.RoleUser,
				Content: []domain.Block{{
					Type: domain.BlockTypeToolResult,
					ToolResult: &domain.ToolResult{
						ToolUseID: "tool-1",
						Content: map[string]any{
							"tool_use_id": "tool-1",
							"rows":        []map[string]any{{"id": "svc-1"}},
						},
					},
				}},
			},
		},
	})
	if err != nil {
		t.Fatalf("toWireRequest() error = %v", err)
	}
	if len(wireReq.Messages) != 2 || len(wireReq.Messages[1].Content) != 1 {
		t.Fatalf("unexpected wire request shape: %#v", wireReq)
	}
	content := wireReq.Messages[1].Content[0].ToolResult.Content
	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("tool_result.content should be JSON object, got error: %v", err)
	}
	if _, ok := parsed["tool_use_id"]; ok {
		t.Fatalf("tool_use_id leaked into tool_result.content: %#v", parsed)
	}
}

func TestToWireRequest_RejectsEmptyMessages(t *testing.T) {
	t.Parallel()

	_, err := toWireRequest(Request{
		ConversationID: "00000000-0000-0000-0000-000000000001",
	})
	if err == nil || !strings.Contains(err.Error(), `"messages" is required`) {
		t.Fatalf("error = %v", err)
	}
}

func TestToWireRequest_RejectsMessageWithEmptyContent(t *testing.T) {
	t.Parallel()

	_, err := toWireRequest(Request{
		ConversationID: "00000000-0000-0000-0000-000000000001",
		Messages: []domain.Message{{
			Role:    domain.RoleUser,
			Content: nil,
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "content is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestToWireRequest_RejectsInvalidContextEntity(t *testing.T) {
	t.Parallel()

	_, err := toWireRequest(Request{
		ConversationID: "00000000-0000-0000-0000-000000000001",
		Messages: []domain.Message{{
			Role: domain.RoleUser,
			Content: []domain.Block{{
				Type: domain.BlockTypeText,
				Text: &domain.TextBlock{Content: "hello"},
			}},
		}},
		ContextEntities: []domain.ContextEntity{{
			EntityType: domain.ContextEntityType("policy"),
			EntityID:   "not-a-uuid",
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "invalid entity_type") {
		t.Fatalf("error = %v", err)
	}
}

func TestToWireRequest_RejectsUnknownToolUseReference(t *testing.T) {
	t.Parallel()

	_, err := toWireRequest(Request{
		ConversationID: "00000000-0000-0000-0000-000000000001",
		Messages: []domain.Message{
			{
				Role: domain.RoleUser,
				Content: []domain.Block{{
					Type: domain.BlockTypeText,
					Text: &domain.TextBlock{Content: "hello"},
				}},
			},
			{
				Role: domain.RoleUser,
				Content: []domain.Block{{
					Type: domain.BlockTypeToolResult,
					ToolResult: &domain.ToolResult{
						ToolUseID: "toolu_missing",
						Content:   map[string]any{"rows": []any{}},
					},
				}},
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), `unknown tool_use_id "toolu_missing"`) {
		t.Fatalf("error = %v", err)
	}
}
