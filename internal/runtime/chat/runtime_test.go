package chat

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	domainchat "github.com/usetero/cli/internal/domains/chat"
	domchattest "github.com/usetero/cli/internal/domains/chat/chattest"
	chattools "github.com/usetero/cli/internal/domains/chat/tools"
	infrachat "github.com/usetero/cli/internal/infrastructure/chat"
	"github.com/usetero/cli/internal/infrastructure/chat/chattest"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func newConversationService() *domchattest.ConversationService {
	var id domainchat.ConversationID
	return &domchattest.ConversationService{
		CreateFn: func(context.Context, *string) (domainchat.ConversationID, error) {
			if id == "" {
				id = "conv_1"
			}
			return id, nil
		},
	}
}

func newMessageService() *domchattest.MessageService {
	return &domchattest.MessageService{
		CreateUserMessageFn: func(_ context.Context, _ domainchat.ConversationID, _ string) (domainchat.MessageID, error) {
			return "msg_user_1", nil
		},
	}
}

func TestRuntime_SendUserText(t *testing.T) {
	rt, err := New(newConversationService(), newMessageService(), chattest.Client{StreamFn: func(_ context.Context, req infrachat.Request, onEvent func(infrachat.Event)) (infrachat.StreamResult, error) {
		onEvent(infrachat.Event{Type: infrachat.EventTypeTextDelta, TextContent: "hello"})
		onEvent(infrachat.Event{Done: true})
		return infrachat.StreamResult{ConversationID: req.ConversationID, TurnID: "turn_1", LastSeq: 1}, nil
	}})
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer rt.Close()

	if err := rt.SendUserText(context.Background(), "hi"); err != nil {
		t.Fatalf("send: %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		st := rt.State()
		if !st.Streaming {
			if len(st.Messages) < 2 {
				t.Fatalf("expected 2 messages, got %+v", st.Messages)
			}
			if st.Messages[1].Content != "hello" {
				t.Fatalf("assistant text mismatch: %q", st.Messages[1].Content)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for stream completion")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestRuntime_Cancel(t *testing.T) {
	rt, err := New(newConversationService(), newMessageService(), chattest.Client{StreamFn: func(ctx context.Context, req infrachat.Request, onEvent func(infrachat.Event)) (infrachat.StreamResult, error) {
		<-ctx.Done()
		return infrachat.StreamResult{}, ctx.Err()
	}})
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer rt.Close()

	if err := rt.SendUserText(context.Background(), "hi"); err != nil {
		t.Fatalf("send: %v", err)
	}
	if err := rt.Cancel(); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		st := rt.State()
		if !st.Streaming {
			if st.Error == "" {
				t.Fatal("expected cancel error")
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for cancel")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestRuntime_RejectConcurrentSend(t *testing.T) {
	block := make(chan struct{})
	rt, err := New(newConversationService(), newMessageService(), chattest.Client{StreamFn: func(ctx context.Context, req infrachat.Request, onEvent func(infrachat.Event)) (infrachat.StreamResult, error) {
		<-block
		return infrachat.StreamResult{}, errors.New("done")
	}})
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer rt.Close()

	if err := rt.SendUserText(context.Background(), "hi"); err != nil {
		t.Fatalf("send: %v", err)
	}
	if err := rt.SendUserText(context.Background(), "second"); err == nil {
		t.Fatal("expected error")
	}
	close(block)
}

func TestRuntime_ToolUseRunsSecondTurn(t *testing.T) {
	sqlite.SetExtensionPath("")
	db, err := sqlite.OpenBare(context.Background(), filepath.Join(t.TempDir(), "runtime-chat-tools.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	call := 0
	client := chattest.Client{StreamFn: func(_ context.Context, req infrachat.Request, onEvent func(infrachat.Event)) (infrachat.StreamResult, error) {
		call++
		switch call {
		case 1:
			onEvent(infrachat.Event{
				Type: infrachat.EventTypeToolUse,
				ToolUse: &infrachat.ToolUse{
					ID:    "tool_1",
					Name:  "query",
					Input: json.RawMessage(`{"sql":"select 1"}`),
				},
			})
			onEvent(infrachat.Event{Done: true})
			return infrachat.StreamResult{ConversationID: req.ConversationID, TurnID: "turn_1", LastSeq: 1}, nil
		default:
			onEvent(infrachat.Event{Type: infrachat.EventTypeTextDelta, TextContent: "after tool"})
			onEvent(infrachat.Event{Done: true})
			return infrachat.StreamResult{ConversationID: req.ConversationID, TurnID: "turn_2", LastSeq: 2}, nil
		}
	}}

	rt, err := NewWithTools(
		newConversationService(),
		newMessageService(),
		client,
		chattools.Toolset{Query: chattools.NewQueryTool(db)},
	)
	if err != nil {
		t.Fatalf("new runtime with tools: %v", err)
	}
	defer rt.Close()

	if err := rt.SendUserText(context.Background(), "run query"); err != nil {
		t.Fatalf("send: %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		st := rt.State()
		if !st.Streaming {
			if call < 2 {
				t.Fatalf("expected second stream turn after tool use, calls=%d", call)
			}
			if len(st.Messages) == 0 || st.Messages[len(st.Messages)-1].Content != "after tool" {
				t.Fatalf("expected final assistant message after tool, got %+v", st.Messages)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for tool loop completion")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestRuntime_ToolUseUnknownAndErrorStillContinue(t *testing.T) {
	call := 0
	client := chattest.Client{StreamFn: func(_ context.Context, req infrachat.Request, onEvent func(infrachat.Event)) (infrachat.StreamResult, error) {
		call++
		switch call {
		case 1:
			onEvent(infrachat.Event{
				Type: infrachat.EventTypeToolUse,
				ToolUse: &infrachat.ToolUse{
					ID:    "tool_unknown_1",
					Name:  "unknown_tool",
					Input: json.RawMessage(`{}`),
				},
			})
			onEvent(infrachat.Event{
				Type: infrachat.EventTypeToolUse,
				ToolUse: &infrachat.ToolUse{
					ID:    "tool_approve_1",
					Name:  "approve_policy",
					Input: json.RawMessage(`{"policy_id":"pol_1"}`),
				},
			})
			onEvent(infrachat.Event{Done: true})
			return infrachat.StreamResult{ConversationID: req.ConversationID, TurnID: "turn_1", LastSeq: 1}, nil
		default:
			onEvent(infrachat.Event{Type: infrachat.EventTypeTextDelta, TextContent: "continued"})
			onEvent(infrachat.Event{Done: true})
			return infrachat.StreamResult{ConversationID: req.ConversationID, TurnID: "turn_2", LastSeq: 2}, nil
		}
	}}

	rt, err := NewWithTools(
		newConversationService(),
		newMessageService(),
		client,
		chattools.Toolset{
			ApprovePolicy: chattools.NewApprovePolicyTool(func(context.Context, chattools.PolicyID) error {
				return errors.New("approve failed")
			}),
		},
	)
	if err != nil {
		t.Fatalf("new runtime with tools: %v", err)
	}
	defer rt.Close()

	if err := rt.SendUserText(context.Background(), "run tools"); err != nil {
		t.Fatalf("send: %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		st := rt.State()
		if !st.Streaming {
			if call < 2 {
				t.Fatalf("expected second stream turn after tool errors, calls=%d", call)
			}
			if st.Error != "" {
				t.Fatalf("expected runtime error to stay empty, got %q", st.Error)
			}

			sawSummary := false
			sawFinalAssistant := false
			for i := range st.Messages {
				msg := st.Messages[i]
				if msg.Role == domainchat.RoleUser &&
					msg.Content == "tool run: unknown_tool: error, approve_policy: error" {
					sawSummary = true
				}
				if msg.Role == domainchat.RoleAssistant && msg.Content == "continued" {
					sawFinalAssistant = true
				}
			}
			if !sawSummary {
				t.Fatalf("expected tool summary message, got %+v", st.Messages)
			}
			if !sawFinalAssistant {
				t.Fatalf("expected final assistant message after tool errors, got %+v", st.Messages)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for tool error loop completion")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestRuntime_NewValidation(t *testing.T) {
	t.Parallel()

	if _, err := New(nil, newMessageService(), chattest.Client{}); err == nil {
		t.Fatalf("expected conversations validation error")
	}
	if _, err := New(newConversationService(), nil, chattest.Client{}); err == nil {
		t.Fatalf("expected messages validation error")
	}
	if _, err := New(newConversationService(), newMessageService(), nil); err == nil {
		t.Fatalf("expected client validation error")
	}
}
