package show

import (
	"encoding/json"
	"errors"
	"testing"

	chattools "github.com/usetero/cli/internal/api/chatclient/tools"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("executes on InputComplete and emits ToolCompleted", func(t *testing.T) {
		t.Parallel()

		m := New(styles.NewTheme(true), 0, "turn-1", "tool-1", 80, nil, logtest.NewScope(t))

		cmd := m.Update(msgs.StreamCompleted{
			Message: domain.Message{
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{
						ID:            "tool-1",
						Input:         json.RawMessage(`{"entity":"policy","id":"p-1"}`),
						InputComplete: true,
					}},
				},
			},
		})

		if m.state != 2 { // tools.StateComplete
			t.Fatalf("expected complete state, got %d", m.state)
		}

		msg := cmd()
		completed, ok := msg.(msgs.ToolCompleted)
		if !ok {
			t.Fatalf("expected ToolCompleted, got %T", msg)
		}
		if completed.TurnID != "turn-1" {
			t.Fatalf("TurnID = %q, want turn-1", completed.TurnID)
		}
		if completed.ToolUseID != "tool-1" {
			t.Fatalf("ToolUseID = %q, want tool-1", completed.ToolUseID)
		}
		if completed.Error == nil {
			t.Fatal("expected error (no executor)")
		}
	})

	t.Run("does not execute when index mismatches", func(t *testing.T) {
		t.Parallel()
		m := New(styles.NewTheme(true), 0, "turn-1", "tool-1", 80, nil, logtest.NewScope(t))
		cmd := m.Update(msgs.AssistantContentUpdated{
			Message: domain.Message{
				Content: []domain.Block{
					{Index: 1, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{
						ID:            "tool-1",
						Input:         json.RawMessage(`{"entity":"policy","id":"p-1"}`),
						InputComplete: true,
					}},
				},
			},
		})
		if cmd != nil {
			t.Fatal("expected nil cmd for mismatched index")
		}
	})
}

func TestResultAndName(t *testing.T) {
	t.Parallel()

	m := New(styles.NewTheme(true), 0, "turn-1", "tool-1", 80, nil, logtest.NewScope(t))

	m.err = errors.New("fail")
	if got := m.Result(); got != "Failed" {
		t.Fatalf("Result() = %q, want Failed", got)
	}

	m.err = nil
	m.result = domaintools.ShowResult{IDShort: "abcd"}
	m.policy = &domain.Policy{
		Category:            domain.CategoryPIILeakage,
		CategoryDisplayName: "PII",
		ServiceName:         "api",
		LogEventName:        "request",
	}
	if got := m.Name(); got == "" {
		t.Fatal("Name() should not be empty")
	}
	if got := m.Result(); got != "api / request" {
		t.Fatalf("Result() = %q, want %q", got, "api / request")
	}

	m.policy.LogEventName = ""
	if got := m.Result(); got != "api" {
		t.Fatalf("Result() = %q, want %q", got, "api")
	}
}

func TestAutoExpand(t *testing.T) {
	t.Parallel()
	m := New(styles.NewTheme(true), 0, "turn-1", "tool-1", 80, &chattools.ShowTool{}, logtest.NewScope(t))
	if !m.AutoExpand() {
		t.Fatal("AutoExpand() = false, want true")
	}
}

func TestUpdate_IgnoresExecutionCompletionFromDifferentToolInstance(t *testing.T) {
	t.Parallel()

	m := New(styles.NewTheme(true), 0, "turn-1", "tool-1", 80, nil, logtest.NewScope(t))
	cmd := m.Update(showExecutedMsg{
		toolID: "tool-2",
		result: domaintools.ShowResult{Entity: domaintools.EntityPolicy, ID: "p-2"},
	})
	if cmd != nil {
		t.Fatal("expected nil cmd for foreign completion")
	}
	if m.state != tools.StateAccumulating {
		t.Fatalf("state = %d, want StateAccumulating", m.state)
	}
	if m.result.ID != "" {
		t.Fatalf("unexpected result id %q", m.result.ID)
	}
}
