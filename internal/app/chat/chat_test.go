package chat

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/chattest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/teatest"
)

// newTestChat creates a chat model with a real DB and a mock streaming client.
func newTestChat(t *testing.T, client chat.Client) *Model {
	t.Helper()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)
	db := dbtest.OpenTestDB(t)

	m := New(nil, domain.Account{ID: "acct-1"}, domain.Workspace{ID: "ws-1"}, theme, db, client, nil, scope)
	m.SetSize(80, 40)
	return m
}

// blockingClient returns a mock client whose stream calls onMessage once then
// blocks until cancelled. Suitable for testing cancel mid-stream.
func blockingClient() *chattest.MockClient {
	return &chattest.MockClient{
		StreamFunc: func(ctx context.Context, _ chat.Request, onMessage func(*domain.Message)) (*chat.StreamResult, error) {
			onMessage(&domain.Message{
				ID:      "asst-1",
				Content: []domain.Block{{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "hello"}}},
			})
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}
}

// completingClient returns a mock client that immediately completes with a
// text response.
func completingClient() *chattest.MockClient {
	return &chattest.MockClient{
		StreamFunc: func(_ context.Context, _ chat.Request, onMessage func(*domain.Message)) (*chat.StreamResult, error) {
			onMessage(&domain.Message{
				ID:         "asst-1",
				Model:      "test-model",
				StopReason: "end_turn",
				Content:    []domain.Block{{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "hello"}}},
			})
			return &chat.StreamResult{}, nil
		},
	}
}

// failingClient returns a mock client that returns an error immediately.
func failingClient() *chattest.MockClient {
	return &chattest.MockClient{
		StreamFunc: func(_ context.Context, _ chat.Request, _ func(*domain.Message)) (*chat.StreamResult, error) {
			return nil, errors.New("connection failed")
		},
	}
}

func abortedClient(reason string) *chattest.MockClient {
	return &chattest.MockClient{
		StreamSnapshotsFunc: func(_ context.Context, _ chat.Request, onSnapshot func(chat.StreamSnapshot)) (*chat.StreamResult, error) {
			msg := &domain.Message{
				ID:      "asst-1",
				Model:   "test-model",
				Content: []domain.Block{{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "partial"}}},
			}
			onSnapshot(chat.StreamSnapshot{
				ConversationID: "conv-1",
				TurnID:         "turn-1",
				Seq:            1,
				Status:         chat.StreamStatusAborted,
				AbortReason:    reason,
				Done:           true,
				Message:        msg,
			})
			return &chat.StreamResult{Message: msg}, nil
		},
	}
}

// submitAndDrain sends a UserSubmittedInput and drains the cmd loop.
func submitAndDrain(m *Model, text string, maxSteps int) {
	cmd := m.Update(msgs.UserSubmittedInput{Text: text})
	teatest.DrainCmds(m.Update, cmd, maxSteps)
}

func listMessages(t *testing.T, m *Model) []domain.Message {
	t.Helper()
	messages, err := m.db.Messages().List(context.Background(), m.conversationID)
	if err != nil {
		t.Fatalf("failed to list messages: %v", err)
	}
	return messages
}

func TestCancelActiveRound(t *testing.T) {
	t.Parallel()

	t.Run("cleans up orphaned user message from DB", func(t *testing.T) {
		t.Parallel()
		m := newTestChat(t, blockingClient())

		// Submit triggers conversation creation + user message persistence.
		submitAndDrain(m, "hello", 20)

		// User message should be in the DB.
		messages := listMessages(t, m)
		if len(messages) == 0 {
			t.Fatal("expected user message in DB after submit")
		}

		// Cancel the active round (simulates ESC).
		m.CancelActiveRound()

		// The orphaned user message should be cleaned up.
		messages = listMessages(t, m)
		if len(messages) != 0 {
			t.Errorf("expected 0 messages after cancel, got %d (roles: %v)", len(messages), messageRoles(messages))
		}
	})

	t.Run("next submit after cancel produces valid history", func(t *testing.T) {
		t.Parallel()
		m := newTestChat(t, completingClient())

		// First submit + cancel.
		submitAndDrain(m, "first", 20)
		m.CancelActiveRound()

		// Second submit — should complete normally.
		submitAndDrain(m, "second", 50)

		messages := listMessages(t, m)
		if err := validateAlternation(messages); err != nil {
			t.Errorf("invalid message history after cancel + resubmit: %v (roles: %v)", err, messageRoles(messages))
		}
	})
}

func TestStreamFailed(t *testing.T) {
	t.Parallel()

	t.Run("cleans up orphaned user message from DB", func(t *testing.T) {
		t.Parallel()
		m := newTestChat(t, failingClient())

		submitAndDrain(m, "hello", 20)

		messages := listMessages(t, m)
		if len(messages) != 0 {
			t.Errorf("expected 0 messages after stream failure, got %d (roles: %v)", len(messages), messageRoles(messages))
		}
	})
}

func TestStreamAborted(t *testing.T) {
	t.Parallel()

	t.Run("non-user abort persists partial assistant message", func(t *testing.T) {
		t.Parallel()
		m := newTestChat(t, abortedClient("context_canceled"))

		submitAndDrain(m, "hello", 40)

		messages := listMessages(t, m)
		if len(messages) != 2 {
			t.Fatalf("expected 2 messages, got %d (roles: %v)", len(messages), messageRoles(messages))
		}
		if messages[0].Role != domain.RoleUser {
			t.Fatalf("message 0 role = %s, want user", messages[0].Role)
		}
		if messages[1].Role != domain.RoleAssistant {
			t.Fatalf("message 1 role = %s, want assistant", messages[1].Role)
		}
		if messages[1].StopReason != "aborted" {
			t.Fatalf("assistant stop_reason = %q, want %q", messages[1].StopReason, "aborted")
		}
	})

	t.Run("user_cancelled abort does not persist assistant message", func(t *testing.T) {
		t.Parallel()
		m := newTestChat(t, abortedClient("user_cancelled"))

		submitAndDrain(m, "hello", 40)

		messages := listMessages(t, m)
		if len(messages) != 1 {
			t.Fatalf("expected 1 message, got %d (roles: %v)", len(messages), messageRoles(messages))
		}
		if messages[0].Role != domain.RoleUser {
			t.Fatalf("message role = %s, want user", messages[0].Role)
		}
	})
}

// validateAlternation checks that messages strictly alternate roles,
// starting with user. Returns an error describing the violation if any.
func validateAlternation(messages []domain.Message) error {
	if len(messages) == 0 {
		return nil
	}
	if messages[0].Role != domain.RoleUser {
		return errors.New("first message must be user")
	}
	for i := 1; i < len(messages); i++ {
		if messages[i].Role == messages[i-1].Role {
			return errors.New("consecutive messages with same role: " + string(messages[i].Role))
		}
	}
	return nil
}

// messageRoles returns a slice of role strings for debugging.
func messageRoles(messages []domain.Message) []string {
	roles := make([]string, len(messages))
	for i, m := range messages {
		roles[i] = string(m.Role)
	}
	return roles
}
