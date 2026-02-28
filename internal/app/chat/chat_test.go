package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api/chatclient"
	"github.com/usetero/cli/internal/api/chatclient/chattest"
	"github.com/usetero/cli/internal/app/chat/msgs"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
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

func recordingCompletingClient(requests *[]chat.Request) *chattest.MockClient {
	var mu sync.Mutex
	call := 0

	return &chattest.MockClient{
		StreamSnapshotsFunc: func(_ context.Context, req chat.Request, onSnapshot func(chat.StreamSnapshot)) (*chat.StreamResult, error) {
			mu.Lock()
			*requests = append(*requests, req)
			call++
			asstID := domain.MessageID(fmt.Sprintf("asst-%d", call))
			mu.Unlock()

			msg := &domain.Message{
				ID:         asstID,
				Model:      "test-model",
				StopReason: "end_turn",
				Content:    []domain.Block{{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "hello"}}},
			}

			if onSnapshot != nil {
				onSnapshot(chat.StreamSnapshot{
					ConversationID: "conv-1",
					TurnID:         "turn-1",
					Seq:            1,
					Status:         chat.StreamStatusStreaming,
					Message:        msg,
				})
				onSnapshot(chat.StreamSnapshot{
					ConversationID: "conv-1",
					TurnID:         "turn-1",
					Seq:            2,
					Status:         chat.StreamStatusCompleted,
					Done:           true,
					Message:        msg,
				})
			}

			return &chat.StreamResult{Message: msg}, nil
		},
	}
}

// submitAndDrain sends a UserSubmittedInput and drains the cmd loop.
func submitAndDrain(m *Model, text string, maxSteps int) {
	cmd := m.Update(msgs.UserSubmittedInput{Text: text})
	teatest.DrainCmds(m.Update, cmd, maxSteps)
}

func submitToolResultsAndDrain(m *Model, results []domaintools.Result, maxSteps int) {
	cmd := m.Update(msgs.UserSubmittedInput{ToolResults: results})
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

func TestRequestHistoryUsesInMemorySessionNotDBRead(t *testing.T) {
	t.Parallel()

	var requests []chat.Request
	m := newTestChat(t, recordingCompletingClient(&requests))

	submitAndDrain(m, "first", 50)

	// Simulate durability drift: assistant row missing in SQLite.
	stored := listMessages(t, m)
	for _, msg := range stored {
		if msg.Role == domain.RoleAssistant {
			if err := m.db.Messages().Delete(context.Background(), msg.ID); err != nil {
				t.Fatalf("delete assistant: %v", err)
			}
		}
	}

	submitAndDrain(m, "second", 70)

	if len(requests) < 2 {
		t.Fatalf("requests = %d, want at least 2", len(requests))
	}
	second := requests[1].Messages
	if len(second) < 3 {
		t.Fatalf("second request message count = %d, want >= 3", len(second))
	}
	if second[0].Role != domain.RoleUser {
		t.Fatalf("second[0].role = %q, want user", second[0].Role)
	}
	if second[1].Role != domain.RoleAssistant {
		t.Fatalf("second[1].role = %q, want assistant (session history)", second[1].Role)
	}
	if second[len(second)-1].Role != domain.RoleUser {
		t.Fatalf("last role = %q, want user", second[len(second)-1].Role)
	}
}

func TestToolResultFollowupRequestContainsAssistantAndToolResult(t *testing.T) {
	t.Parallel()

	var requests []chat.Request
	m := newTestChat(t, recordingCompletingClient(&requests))

	// Turn 1: user prompt -> assistant response.
	submitAndDrain(m, "run a query", 50)

	// Turn 2: user tool_result follow-up should include prior assistant + tool_result.
	submitToolResultsAndDrain(m, []domaintools.Result{{
		ToolUseID: "toolu_1",
		Content: map[string]any{
			"rows": []map[string]any{
				{"service_id": "svc-1", "weekly_volume": 12345},
			},
		},
	}}, 60)

	if len(requests) < 2 {
		t.Fatalf("requests = %d, want >= 2", len(requests))
	}
	second := requests[1].Messages
	if len(second) != 3 {
		t.Fatalf("second request message count = %d, want 3", len(second))
	}
	if second[0].Role != domain.RoleUser {
		t.Fatalf("second[0].role = %q, want user", second[0].Role)
	}
	if second[1].Role != domain.RoleAssistant {
		t.Fatalf("second[1].role = %q, want assistant", second[1].Role)
	}
	if second[2].Role != domain.RoleUser {
		t.Fatalf("second[2].role = %q, want user(tool_result)", second[2].Role)
	}
	if len(second[2].Content) != 1 || second[2].Content[0].Type != domain.BlockTypeToolResult {
		t.Fatalf("second[2] content = %#v, want single tool_result block", second[2].Content)
	}
	if got := second[2].Content[0].ToolResult.ToolUseID; got != "toolu_1" {
		t.Fatalf("tool_use_id = %q, want %q", got, "toolu_1")
	}
}

func TestToolResultFollowupKeepsAssistantWhenStreamMessageIDMissing(t *testing.T) {
	t.Parallel()

	var requests []chat.Request
	var mu sync.Mutex
	call := 0
	client := &chattest.MockClient{
		StreamSnapshotsFunc: func(_ context.Context, req chat.Request, onSnapshot func(chat.StreamSnapshot)) (*chat.StreamResult, error) {
			mu.Lock()
			requests = append(requests, req)
			call++
			n := call
			mu.Unlock()

			if n == 1 {
				msg := &domain.Message{
					// Intentionally empty ID: mirrors stream payloads that do not carry message IDs.
					ID:         "",
					Model:      "test-model",
					StopReason: "tool_use",
					Content: []domain.Block{{
						Index: 0,
						Type:  domain.BlockTypeToolUse,
						ToolUse: &domain.ToolUse{
							ID:            "toolu_1",
							Name:          "query",
							Input:         json.RawMessage(`{"sql":"select 1"}`),
							InputComplete: true,
						},
					}},
				}
				onSnapshot(chat.StreamSnapshot{
					ConversationID: req.ConversationID,
					TurnID:         "turn-1",
					Seq:            1,
					Status:         chat.StreamStatusToolUse,
					Done:           true,
					Message:        msg,
				})
				return &chat.StreamResult{Message: msg}, nil
			}

			msg := &domain.Message{
				ID:         "asst-2",
				Model:      "test-model",
				StopReason: "end_turn",
				Content:    []domain.Block{{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "done"}}},
			}
			onSnapshot(chat.StreamSnapshot{
				ConversationID: req.ConversationID,
				TurnID:         "turn-2",
				Seq:            1,
				Status:         chat.StreamStatusCompleted,
				Done:           true,
				Message:        msg,
			})
			return &chat.StreamResult{Message: msg}, nil
		},
	}

	m := newTestChat(t, client)
	submitAndDrain(m, "run a query", 60)
	submitToolResultsAndDrain(m, []domaintools.Result{{
		ToolUseID: "toolu_1",
		Content: map[string]any{
			"rows": []map[string]any{{"service_id": "svc-1"}},
		},
	}}, 60)

	mu.Lock()
	defer mu.Unlock()
	if len(requests) < 2 {
		t.Fatalf("requests = %d, want >= 2", len(requests))
	}
	second := requests[1].Messages
	if len(second) != 3 {
		t.Fatalf("second request message count = %d, want 3", len(second))
	}
	if second[1].Role != domain.RoleAssistant {
		t.Fatalf("second[1].role = %q, want assistant", second[1].Role)
	}
	if len(second[1].Content) != 1 || second[1].Content[0].Type != domain.BlockTypeToolUse {
		t.Fatalf("second[1].content = %#v, want single tool_use block", second[1].Content)
	}
	if got := second[1].Content[0].ToolUse.ID; got != "toolu_1" {
		t.Fatalf("assistant tool_use.id = %q, want toolu_1", got)
	}
}

func TestInternalToolLoopKeepsTopLevelSessionAligned(t *testing.T) {
	t.Parallel()

	var requests []chat.Request
	var mu sync.Mutex
	call := 0
	client := &chattest.MockClient{
		StreamSnapshotsFunc: func(_ context.Context, req chat.Request, onSnapshot func(chat.StreamSnapshot)) (*chat.StreamResult, error) {
			mu.Lock()
			requests = append(requests, req)
			call++
			n := call
			mu.Unlock()

			if n == 1 {
				msg := &domain.Message{
					ID:         "asst-1",
					Model:      "test-model",
					StopReason: "tool_use",
					Content: []domain.Block{{
						Index: 0,
						Type:  domain.BlockTypeToolUse,
						ToolUse: &domain.ToolUse{
							ID:            "toolu_1",
							Name:          "query",
							Input:         json.RawMessage(`{"sql":"select 1"}`),
							InputComplete: true,
						},
					}},
				}
				onSnapshot(chat.StreamSnapshot{
					ConversationID: req.ConversationID,
					TurnID:         "turn-1",
					Seq:            1,
					Status:         chat.StreamStatusToolUse,
					Done:           true,
					Message:        msg,
				})
				return &chat.StreamResult{Message: msg}, nil
			}

			msg := &domain.Message{
				ID:         domain.MessageID(fmt.Sprintf("asst-%d", n)),
				Model:      "test-model",
				StopReason: "end_turn",
				Content:    []domain.Block{{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "done"}}},
			}
			onSnapshot(chat.StreamSnapshot{
				ConversationID: req.ConversationID,
				TurnID:         fmt.Sprintf("turn-%d", n),
				Seq:            1,
				Status:         chat.StreamStatusCompleted,
				Done:           true,
				Message:        msg,
			})
			return &chat.StreamResult{Message: msg}, nil
		},
	}

	m := newTestChat(t, client)
	submitAndDrain(m, "run a query", 80)

	stored := listMessages(t, m)
	if len(stored) == 0 {
		t.Fatal("expected first user message")
	}
	firstTurnID := stored[0].ID

	cmd := m.Update(msgs.ToolCompleted{
		TurnID:    firstTurnID,
		ToolUseID: "toolu_1",
		Result: domaintools.Result{
			ToolUseID: "toolu_1",
			Content: map[string]any{
				"rows": []map[string]any{{"service_id": "svc-1"}},
			},
		},
	})
	teatest.DrainCmds(m.Update, cmd, 120)

	submitAndDrain(m, "what are my top disabled services?", 120)

	mu.Lock()
	defer mu.Unlock()
	if len(requests) < 3 {
		t.Fatalf("requests = %d, want >= 3", len(requests))
	}
	third := requests[2].Messages
	if len(third) < 5 {
		t.Fatalf("third request message count = %d, want >= 5", len(third))
	}
	if third[1].Role != domain.RoleAssistant || third[1].StopReason != "tool_use" {
		t.Fatalf("third[1] = role=%q stop_reason=%q, want assistant tool_use", third[1].Role, third[1].StopReason)
	}
	if third[2].Role != domain.RoleUser || len(third[2].Content) == 0 || third[2].Content[0].Type != domain.BlockTypeToolResult {
		t.Fatalf("third[2] = %#v, want user tool_result message", third[2])
	}
	if got := third[2].Content[0].ToolResult.ToolUseID; got != "toolu_1" {
		t.Fatalf("third[2] tool_use_id = %q, want toolu_1", got)
	}
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

	t.Run("maps protocol errors to user-friendly toast", func(t *testing.T) {
		t.Parallel()
		m := newTestChat(t, completingClient())
		cmd := m.Update(msgs.StreamFailed{TurnID: "turn-1", Err: errors.New("protocol error: unknown event type")})
		errMsg, ok := extractErrorToast(cmd)
		if !ok {
			t.Fatal("expected error toast command")
		}
		if errMsg.Message != "The chat service returned an unexpected stream format. Please retry." {
			t.Fatalf("toast message = %q", errMsg.Message)
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

func extractErrorToast(cmd tea.Cmd) (appmsg.Error, bool) {
	if cmd == nil {
		return appmsg.Error{}, false
	}
	msg := cmd()
	if msg == nil {
		return appmsg.Error{}, false
	}
	if e, ok := msg.(appmsg.Error); ok {
		return e, true
	}
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return appmsg.Error{}, false
	}
	for _, sub := range batch {
		if sub == nil {
			continue
		}
		subMsg := sub()
		if e, ok := subMsg.(appmsg.Error); ok {
			return e, true
		}
	}
	return appmsg.Error{}, false
}
