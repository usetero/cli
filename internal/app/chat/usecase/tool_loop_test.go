package usecase

import (
	"context"
	"testing"

	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
)

func TestMemoryToolLoop_PrepareNextTurn_AppendsToolResultToSession(t *testing.T) {
	t.Parallel()

	loop := NewMemoryToolLoop()

	session := corechat.NewSession("conv-1", []domain.Message{
		{
			ID:   "user-1",
			Role: domain.RoleUser,
			Content: []domain.Block{
				{Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "hi"}},
			},
		},
	})

	out, err := loop.PrepareNextTurn(context.Background(), PrepareNextTurnInput{
		AccountID:      "acct-1",
		ConversationID: "conv-1",
		Results: []domaintools.Result{
			{
				ToolUseID: "tool-1",
				Content:   map[string]any{"ok": true},
			},
		},
		Session: session,
	})
	if err != nil {
		t.Fatalf("PrepareNextTurn() error = %v", err)
	}
	if out.MessageID == "" {
		t.Fatal("MessageID is empty")
	}
	if out.ToolResultMessage.ID != out.MessageID {
		t.Fatalf("tool result message id = %q, want %q", out.ToolResultMessage.ID, out.MessageID)
	}
	if len(out.Messages) != 2 {
		t.Fatalf("messages len = %d, want 2", len(out.Messages))
	}
	last := out.Messages[len(out.Messages)-1]
	if last.Role != domain.RoleUser {
		t.Fatalf("last role = %q, want %q", last.Role, domain.RoleUser)
	}
	if len(last.Content) != 1 || last.Content[0].Type != domain.BlockTypeToolResult {
		t.Fatalf("last content mismatch: %+v", last.Content)
	}
	if last.Content[0].ToolResult == nil || last.Content[0].ToolResult.ToolUseID != "tool-1" {
		t.Fatalf("tool_result mismatch: %+v", last.Content[0].ToolResult)
	}
}

func TestMemoryToolLoop_PrepareNextTurn_StartsSessionWhenNil(t *testing.T) {
	t.Parallel()

	loop := NewMemoryToolLoop()

	out, err := loop.PrepareNextTurn(context.Background(), PrepareNextTurnInput{
		AccountID:      "acct-1",
		ConversationID: "conv-1",
		Results: []domaintools.Result{
			{ToolUseID: "tool-1", Content: map[string]any{"k": "v"}},
		},
		Session: nil,
	})
	if err != nil {
		t.Fatalf("PrepareNextTurn() error = %v", err)
	}
	if out.MessageID == "" {
		t.Fatal("expected a generated MessageID")
	}
	if len(out.Messages) != 1 {
		t.Fatalf("messages len = %d, want 1", len(out.Messages))
	}
	if out.ToolResultMessage.ID == "" || out.ToolResultMessage.ID != out.MessageID {
		t.Fatalf("toolResultMessage ID mismatch: %+v", out.ToolResultMessage)
	}
}
