package usecase

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
)

func TestSQLiteAssistantPersister_PersistAssistant_Success(t *testing.T) {
	t.Parallel()

	db := dbtest.OpenTestDB(t)
	p := NewSQLiteAssistantPersister(db)

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

	got, err := db.Messages().Get(context.Background(), msgID)
	if err != nil {
		t.Fatalf("db.Messages().Get() error = %v", err)
	}
	if got.Role != domain.RoleAssistant {
		t.Fatalf("role = %q, want %q", got.Role, domain.RoleAssistant)
	}
	if got.Model != "claude-3" {
		t.Fatalf("model = %q, want claude-3", got.Model)
	}
	if got.StopReason != "end_turn" {
		t.Fatalf("stop_reason = %q, want end_turn", got.StopReason)
	}
	if len(got.Content) != 1 || got.Content[0].Text == nil || got.Content[0].Text.Content != "hello" {
		t.Fatalf("content mismatch: %+v", got.Content)
	}
}

func TestSQLiteAssistantPersister_PersistAssistant_ErrorOnMissingSchema(t *testing.T) {
	t.Parallel()

	db := sqlitetest.OpenBareDB(t) // no schema tables applied
	p := NewSQLiteAssistantPersister(db)

	_, err := p.PersistAssistant(context.Background(), PersistAssistantInput{
		AccountID:      "acct-1",
		ConversationID: "conv-1",
		Message:        domain.Message{Model: "claude-3"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
