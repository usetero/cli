package usecase

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestMemoryAssistantPersister_PersistAssistant_MintsID(t *testing.T) {
	t.Parallel()

	p := NewMemoryAssistantPersister()

	msgID, err := p.PersistAssistant(context.Background(), PersistAssistantInput{
		AccountID:      "acct-1",
		ConversationID: "conv-1",
		Message: domain.Message{
			Model:      "claude-3",
			StopReason: "end_turn",
			Content: []domain.Block{
				{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "hello"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("PersistAssistant() error = %v", err)
	}
	if msgID == "" {
		t.Fatal("PersistAssistant() returned empty message ID")
	}
}

func TestMemoryAssistantPersister_PersistAssistant_UniqueIDs(t *testing.T) {
	t.Parallel()

	p := NewMemoryAssistantPersister()
	input := PersistAssistantInput{AccountID: "acct-1", ConversationID: "conv-1"}

	first, err := p.PersistAssistant(context.Background(), input)
	if err != nil {
		t.Fatalf("PersistAssistant() error = %v", err)
	}
	second, err := p.PersistAssistant(context.Background(), input)
	if err != nil {
		t.Fatalf("PersistAssistant() error = %v", err)
	}
	if first == second {
		t.Fatalf("expected unique message IDs, got %q twice", first)
	}
}
