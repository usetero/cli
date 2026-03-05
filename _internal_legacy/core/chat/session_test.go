package chat

import (
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestSession_RecordAssistantMessageUpsert(t *testing.T) {
	t.Parallel()

	s := NewSession("conv-1", []domain.Message{{
		ID:   "asst-1",
		Role: domain.RoleAssistant,
		Content: []domain.Block{{
			Type: domain.BlockTypeText,
			Text: &domain.TextBlock{Content: "old"},
		}},
	}})

	s.RecordAssistantMessage(domain.Message{
		ID:   "asst-1",
		Role: domain.RoleAssistant,
		Content: []domain.Block{{
			Type: domain.BlockTypeText,
			Text: &domain.TextBlock{Content: "new"},
		}},
	})

	msgs := s.Messages()
	if len(msgs) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(msgs))
	}
	if got := msgs[0].Content[0].Text.Content; got != "new" {
		t.Fatalf("assistant content = %q, want %q", got, "new")
	}
}

func TestSession_AppendUserToolResultsMessage(t *testing.T) {
	t.Parallel()

	s := NewSession("conv-1", nil)
	s.AppendUserToolResultsMessage("tool-result-1", []domain.ToolResult{
		{
			ToolUseID: "tool-a",
			Content:   map[string]any{"rows": []any{1, 2}},
		},
		{
			ToolUseID: "tool-b",
			IsError:   true,
			Error:     "failed",
		},
	})

	msgs := s.Messages()
	if len(msgs) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(msgs))
	}
	msg := msgs[0]
	if msg.Role != domain.RoleUser {
		t.Fatalf("role = %q, want %q", msg.Role, domain.RoleUser)
	}
	if len(msg.Content) != 2 {
		t.Fatalf("len(content) = %d, want 2", len(msg.Content))
	}
	if msg.Content[0].Type != domain.BlockTypeToolResult || msg.Content[0].ToolResult.ToolUseID != "tool-a" {
		t.Fatalf("unexpected first block: %#v", msg.Content[0])
	}
	if msg.Content[1].Type != domain.BlockTypeToolResult || msg.Content[1].ToolResult.ToolUseID != "tool-b" {
		t.Fatalf("unexpected second block: %#v", msg.Content[1])
	}
}

func TestSession_AppendUserTextMessage(t *testing.T) {
	t.Parallel()

	s := NewSession("conv-1", nil)
	s.AppendUserTextMessage("msg-1", "hello")

	msgs := s.Messages()
	if len(msgs) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(msgs))
	}
	msg := msgs[0]
	if msg.Role != domain.RoleUser {
		t.Fatalf("role = %q, want %q", msg.Role, domain.RoleUser)
	}
	if len(msg.Content) != 1 || msg.Content[0].Type != domain.BlockTypeText || msg.Content[0].Text == nil {
		t.Fatalf("unexpected content: %#v", msg.Content)
	}
	if msg.Content[0].Text.Content != "hello" {
		t.Fatalf("text = %q, want %q", msg.Content[0].Text.Content, "hello")
	}
}

func TestSession_MessagesReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	s := NewSession("conv-1", []domain.Message{{
		ID:   "msg-1",
		Role: domain.RoleUser,
		Content: []domain.Block{{
			Type: domain.BlockTypeText,
			Text: &domain.TextBlock{Content: "hello"},
		}},
	}})

	got := s.Messages()
	got[0].Content[0].Text.Content = "mutated"

	again := s.Messages()
	if again[0].Content[0].Text.Content != "hello" {
		t.Fatalf("session history was mutated through returned slice")
	}
}

func TestSession_RemoveMessagesByID(t *testing.T) {
	t.Parallel()

	s := NewSession("conv-1", []domain.Message{
		{ID: "a", Role: domain.RoleUser},
		{ID: "b", Role: domain.RoleAssistant},
		{ID: "c", Role: domain.RoleUser},
	})

	s.RemoveMessagesByID([]domain.MessageID{"b", "c"})

	msgs := s.Messages()
	if len(msgs) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(msgs))
	}
	if msgs[0].ID != "a" {
		t.Fatalf("remaining id = %q, want %q", msgs[0].ID, "a")
	}
}
