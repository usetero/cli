package chat

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const protocolVersion = "v2"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

func (r Role) Valid() bool {
	return r == RoleUser || r == RoleAssistant
}

type StopReason string

const (
	StopReasonEndTurn StopReason = "end_turn"
	StopReasonToolUse StopReason = "tool_use"
)

func (s StopReason) Valid() bool {
	return s == StopReasonEndTurn || s == StopReasonToolUse
}

type BlockType string

const (
	BlockTypeText       BlockType = "text"
	BlockTypeThinking   BlockType = "thinking"
	BlockTypeToolUse    BlockType = "tool_use"
	BlockTypeToolResult BlockType = "tool_result"
)

func (b BlockType) Valid() bool {
	switch b {
	case BlockTypeText, BlockTypeThinking, BlockTypeToolUse, BlockTypeToolResult:
		return true
	default:
		return false
	}
}

type ContextEntityType string

const (
	ContextEntityTypeService  ContextEntityType = "service"
	ContextEntityTypeLogEvent ContextEntityType = "log_event"
)

func (c ContextEntityType) Valid() bool {
	return c == ContextEntityTypeService || c == ContextEntityTypeLogEvent
}

type Request struct {
	ChatProtocolVersion string          `json:"chat_protocol_version,omitempty"`
	ConversationID      string          `json:"conversation_id"`
	Messages            []Message       `json:"messages"`
	ContextEntities     []ContextEntity `json:"context_entities,omitempty"`
	Tools               []Tool          `json:"tools,omitempty"`
}

type Message struct {
	Role       Role        `json:"role"`
	Content    []Block     `json:"content"`
	Model      string      `json:"model,omitempty"`
	StopReason *StopReason `json:"stop_reason,omitempty"`
}

type Block struct {
	Type       BlockType   `json:"type"`
	Text       *Text       `json:"text,omitempty"`
	Thinking   *Thinking   `json:"thinking,omitempty"`
	ToolUse    *ToolUse    `json:"tool_use,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

type Text struct {
	Content string `json:"content"`
}

type Thinking struct {
	Content string `json:"content"`
}

type ToolUse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type ToolResult struct {
	ToolUseID string          `json:"tool_use_id"`
	IsError   bool            `json:"-"`
	Error     string          `json:"error,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

type ContextEntity struct {
	EntityType ContextEntityType `json:"entity_type"`
	EntityID   string            `json:"entity_id"`
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

func (r Request) Validate() error {
	if r.ConversationID == "" {
		return fmt.Errorf("conversation id is required")
	}
	if _, err := uuid.Parse(r.ConversationID); err != nil {
		return fmt.Errorf("conversation id must be a uuid")
	}
	if len(r.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}
	for i, msg := range r.Messages {
		if !msg.Role.Valid() {
			return fmt.Errorf("messages[%d]: invalid role %q", i, msg.Role)
		}
		if len(msg.Content) == 0 {
			return fmt.Errorf("messages[%d]: content is required", i)
		}
		if msg.StopReason != nil && !msg.StopReason.Valid() {
			return fmt.Errorf("messages[%d]: invalid stop_reason %q", i, *msg.StopReason)
		}
		for j, b := range msg.Content {
			if err := b.Validate(); err != nil {
				return fmt.Errorf("messages[%d].content[%d]: %w", i, j, err)
			}
		}
	}
	if r.Messages[0].Role != RoleUser {
		return fmt.Errorf("first message must be user")
	}
	for i := 1; i < len(r.Messages); i++ {
		if r.Messages[i].Role == r.Messages[i-1].Role {
			return fmt.Errorf("messages[%d]: adjacent roles must alternate", i)
		}
	}
	for i, t := range r.Tools {
		if t.Name == "" {
			return fmt.Errorf("tools[%d]: name is required", i)
		}
		if t.Description == "" {
			return fmt.Errorf("tools[%d]: description is required", i)
		}
		if t.InputSchema == nil {
			return fmt.Errorf("tools[%d]: input schema is required", i)
		}
	}
	for i, e := range r.ContextEntities {
		if !e.EntityType.Valid() {
			return fmt.Errorf("context_entities[%d]: invalid entity type %q", i, e.EntityType)
		}
		if e.EntityID == "" {
			return fmt.Errorf("context_entities[%d]: entity id is required", i)
		}
		if _, err := uuid.Parse(e.EntityID); err != nil {
			return fmt.Errorf("context_entities[%d]: entity id must be a uuid", i)
		}
	}
	return nil
}

func (b Block) Validate() error {
	switch b.Type {
	case BlockTypeText:
		if b.Text == nil {
			return fmt.Errorf("text block missing payload")
		}
	case BlockTypeThinking:
		if b.Thinking == nil {
			return fmt.Errorf("thinking block missing payload")
		}
	case BlockTypeToolUse:
		if b.ToolUse == nil || b.ToolUse.ID == "" || b.ToolUse.Name == "" {
			return fmt.Errorf("tool_use block missing id/name")
		}
	case BlockTypeToolResult:
		if b.ToolResult == nil {
			return fmt.Errorf("tool_result block missing payload")
		}
		if b.ToolResult.ToolUseID == "" {
			return fmt.Errorf("tool_result missing tool_use_id")
		}
		if b.ToolResult.IsError {
			if b.ToolResult.Error == "" {
				return fmt.Errorf("tool_result error is required when is_error is true")
			}
			if len(b.ToolResult.Content) > 0 {
				return fmt.Errorf("tool_result content must be empty when is_error is true")
			}
		} else {
			if len(b.ToolResult.Content) == 0 {
				return fmt.Errorf("tool_result content is required when is_error is false")
			}
			if !isJSONObject(b.ToolResult.Content) {
				return fmt.Errorf("tool_result content must be a json object")
			}
		}
	default:
		return fmt.Errorf("unsupported block type %q", b.Type)
	}
	return nil
}

func isJSONObject(raw json.RawMessage) bool {
	var m map[string]any
	return json.Unmarshal(raw, &m) == nil
}

func (r Request) payload() (Request, error) {
	if err := r.Validate(); err != nil {
		return Request{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}
	r.ChatProtocolVersion = protocolVersion
	return r, nil
}

func (r ToolResult) MarshalJSON() ([]byte, error) {
	type alias struct {
		ToolUseID string          `json:"tool_use_id"`
		IsError   *bool           `json:"is_error"`
		Error     string          `json:"error,omitempty"`
		Content   json.RawMessage `json:"content,omitempty"`
	}
	v := r.IsError
	return json.Marshal(alias{
		ToolUseID: r.ToolUseID,
		IsError:   &v,
		Error:     r.Error,
		Content:   r.Content,
	})
}
