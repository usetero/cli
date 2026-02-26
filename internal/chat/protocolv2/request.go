package protocolv2

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const Version = "v2"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type StopReason string

const (
	StopReasonEndTurn StopReason = "end_turn"
	StopReasonToolUse StopReason = "tool_use"
)

type BlockType string

const (
	BlockTypeText       BlockType = "text"
	BlockTypeThinking   BlockType = "thinking"
	BlockTypeToolUse    BlockType = "tool_use"
	BlockTypeToolResult BlockType = "tool_result"
)

type ContextEntityType string

const (
	ContextEntityTypeService  ContextEntityType = "service"
	ContextEntityTypeLogEvent ContextEntityType = "log_event"
)

type Request struct {
	ChatProtocolVersion string          `json:"chat_protocol_version"`
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
	IsError   *bool           `json:"is_error"`
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

func Validate(req Request) error {
	if req.ChatProtocolVersion == "" {
		return fmt.Errorf(`"chat_protocol_version" is required`)
	}
	if req.ChatProtocolVersion != Version {
		return fmt.Errorf(`"chat_protocol_version" must be %q`, Version)
	}
	if req.ConversationID == "" {
		return fmt.Errorf(`"conversation_id" is required`)
	}
	if _, err := uuid.Parse(req.ConversationID); err != nil {
		return fmt.Errorf(`"conversation_id" must be a valid UUID`)
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf(`"messages" is required`)
	}

	for i, msg := range req.Messages {
		if err := validateMessage(msg); err != nil {
			return fmt.Errorf("messages[%d]: %w", i, err)
		}
	}
	for i, entity := range req.ContextEntities {
		if err := validateContextEntity(entity); err != nil {
			return fmt.Errorf("context_entities[%d]: %w", i, err)
		}
	}
	return nil
}

func validateMessage(msg Message) error {
	switch msg.Role {
	case RoleUser, RoleAssistant:
	default:
		return fmt.Errorf("unsupported role %q", msg.Role)
	}
	if len(msg.Content) == 0 {
		return fmt.Errorf("content is required")
	}
	if msg.StopReason != nil {
		switch *msg.StopReason {
		case StopReasonEndTurn, StopReasonToolUse:
		default:
			return fmt.Errorf("unsupported stop_reason %q", *msg.StopReason)
		}
	}

	for i, block := range msg.Content {
		if err := validateBlock(block); err != nil {
			return fmt.Errorf("content[%d]: %w", i, err)
		}
	}
	return nil
}

func validateBlock(block Block) error {
	switch block.Type {
	case BlockTypeText:
		if block.Text == nil {
			return fmt.Errorf("text block missing payload")
		}
	case BlockTypeThinking:
		if block.Thinking == nil {
			return fmt.Errorf("thinking block missing payload")
		}
	case BlockTypeToolUse:
		if block.ToolUse == nil {
			return fmt.Errorf("tool_use block missing payload")
		}
		if block.ToolUse.ID == "" || block.ToolUse.Name == "" {
			return fmt.Errorf("tool_use missing id/name")
		}
	case BlockTypeToolResult:
		if block.ToolResult == nil {
			return fmt.Errorf("tool_result block missing payload")
		}
		if err := validateToolResult(*block.ToolResult); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported block type %q", block.Type)
	}
	return nil
}

func validateToolResult(result ToolResult) error {
	if result.ToolUseID == "" {
		return fmt.Errorf("tool_result missing tool_use_id")
	}
	if result.IsError == nil {
		return fmt.Errorf(`tool_result "is_error" is required`)
	}
	if *result.IsError {
		if result.Error == "" {
			return fmt.Errorf(`tool_result "error" is required when is_error is true`)
		}
		if len(result.Content) > 0 {
			return fmt.Errorf(`tool_result "content" must be empty when is_error is true`)
		}
		return nil
	}

	if len(result.Content) == 0 {
		return fmt.Errorf(`tool_result "content" is required when is_error is false`)
	}
	var parsed map[string]any
	if err := json.Unmarshal(result.Content, &parsed); err != nil {
		return fmt.Errorf("tool_result content must be valid JSON object")
	}
	if _, ok := parsed["tool_use_id"]; ok {
		return fmt.Errorf("tool_result.content must not contain tool_use_id")
	}
	return nil
}

func validateContextEntity(entity ContextEntity) error {
	if entity.EntityType == "" {
		return fmt.Errorf(`"entity_type" is required`)
	}
	if entity.EntityType != ContextEntityTypeService && entity.EntityType != ContextEntityTypeLogEvent {
		return fmt.Errorf(`invalid entity_type %q`, entity.EntityType)
	}
	if entity.EntityID == "" {
		return fmt.Errorf(`"entity_id" is required`)
	}
	if _, err := uuid.Parse(entity.EntityID); err != nil {
		return fmt.Errorf(`"entity_id" must be a valid UUID`)
	}
	return nil
}
