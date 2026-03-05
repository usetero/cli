package protocolv2

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	newValidRequest := func() Request {
		ok := false
		return Request{
			ChatProtocolVersion: Version,
			ConversationID:      "00000000-0000-0000-0000-000000000001",
			Messages: []Message{
				{
					Role: RoleUser,
					Content: []Block{
						{
							Type: BlockTypeText,
							Text: &Text{Content: "run query"},
						},
					},
				},
				{
					Role: RoleAssistant,
					Content: []Block{
						{
							Type: BlockTypeToolUse,
							ToolUse: &ToolUse{
								ID:    "toolu_1",
								Name:  "query",
								Input: json.RawMessage(`{"sql":"select 1"}`),
							},
						},
					},
				},
				{
					Role: RoleUser,
					Content: []Block{
						{
							Type: BlockTypeToolResult,
							ToolResult: &ToolResult{
								ToolUseID: "toolu_1",
								IsError:   &ok,
								Content:   json.RawMessage(`{"rows":[1]}`),
							},
						},
					},
				},
			},
			ContextEntities: []ContextEntity{
				{
					EntityType: ContextEntityTypeService,
					EntityID:   "00000000-0000-0000-0000-000000000002",
				},
			},
		}
	}

	t.Run("valid request passes", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		if err := Validate(req); err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
	})

	t.Run("missing protocol version fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.ChatProtocolVersion = ""
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `"chat_protocol_version" is required`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("unknown protocol version fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.ChatProtocolVersion = "v9"
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `"chat_protocol_version" must be "v2"`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("invalid conversation id fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.ConversationID = "conv-1"
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `"conversation_id" must be a valid UUID`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("empty messages fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.Messages = nil
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `"messages" is required`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("empty message content fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.Messages[0].Content = nil
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), "content is required") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("invalid role fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.Messages[0].Role = Role("invalid")
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), "unsupported role") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("invalid stop reason fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		bad := StopReason("invalid")
		req.Messages[0].StopReason = &bad
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), "unsupported stop_reason") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("tool_result requires is_error", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.Messages[2].Content[0].ToolResult.IsError = nil
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `"is_error" is required`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("tool_result error requires error text and empty content", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		yes := true
		req.Messages[2].Content[0].ToolResult.IsError = &yes
		req.Messages[2].Content[0].ToolResult.Content = json.RawMessage(`{"rows":[1]}`)
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `"error" is required when is_error is true`) {
			t.Fatalf("error = %v", err)
		}

		req.Messages[2].Content[0].ToolResult.Error = "boom"
		err = Validate(req)
		if err == nil || !strings.Contains(err.Error(), `"content" must be empty when is_error is true`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("tool_result success requires content and forbids embedded tool_use_id", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		no := false
		req.Messages[2].Content[0].ToolResult.IsError = &no
		req.Messages[2].Content[0].ToolResult.Content = nil
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `"content" is required when is_error is false`) {
			t.Fatalf("error = %v", err)
		}

		req.Messages[2].Content[0].ToolResult.Content = json.RawMessage(`{"tool_use_id":"toolu_1"}`)
		err = Validate(req)
		if err == nil || !strings.Contains(err.Error(), "must not contain tool_use_id") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("invalid context entity type fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.ContextEntities[0].EntityType = ContextEntityType("policy")
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), "invalid entity_type") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("invalid context entity id fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.ContextEntities[0].EntityID = "not-a-uuid"
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `"entity_id" must be a valid UUID`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("unknown tool_use_id fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.Messages[2].Content = append(req.Messages[2].Content, Block{
			Type: BlockTypeToolResult,
			ToolResult: &ToolResult{
				ToolUseID: "toolu_missing",
				IsError:   req.Messages[2].Content[0].ToolResult.IsError,
				Content:   json.RawMessage(`{"rows":[2]}`),
			},
		})
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `unknown tool_use_id "toolu_missing"`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("duplicate tool_use id fails", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.Messages = append(req.Messages, Message{
			Role: RoleAssistant,
			Content: []Block{{
				Type: BlockTypeToolUse,
				ToolUse: &ToolUse{
					ID:    "toolu_1",
					Name:  "query",
					Input: json.RawMessage(`{"sql":"select 2"}`),
				},
			}},
		}, Message{
			Role: RoleUser,
			Content: []Block{{
				Type: BlockTypeToolResult,
				ToolResult: &ToolResult{
					ToolUseID: "toolu_1",
					IsError:   req.Messages[2].Content[0].ToolResult.IsError,
					Content:   json.RawMessage(`{"rows":[9]}`),
				},
			}},
		})
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `duplicate tool_use id "toolu_1"`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("tool_use must be followed by immediate user tool_result message", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.Messages = []Message{
			req.Messages[0],
			req.Messages[1],
		}
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), "tool_use requires an immediate following user tool_result message") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("assistant turn between tool_use and tool_result fails ordering", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.Messages = []Message{
			req.Messages[0],
			req.Messages[1],
			{
				Role: RoleAssistant,
				Content: []Block{{
					Type: BlockTypeText,
					Text: &Text{Content: "intermediate"},
				}},
			},
			req.Messages[2],
		}
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `role "assistant" cannot repeat consecutively`) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("consecutive roles fail", func(t *testing.T) {
		t.Parallel()
		req := newValidRequest()
		req.Messages = []Message{
			req.Messages[0],
			{
				Role: RoleUser,
				Content: []Block{{
					Type: BlockTypeText,
					Text: &Text{Content: "again"},
				}},
			},
		}
		err := Validate(req)
		if err == nil || !strings.Contains(err.Error(), `role "user" cannot repeat consecutively`) {
			t.Fatalf("error = %v", err)
		}
	})
}
