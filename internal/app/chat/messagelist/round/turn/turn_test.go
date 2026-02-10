package turn

import (
	"errors"
	"testing"

	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func newTestTurn(t *testing.T) *Model {
	t.Helper()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	// DB and chat client are nil — these tests exercise the state machine,
	// not persistence or streaming. The returned tea.Cmds are never executed.
	return New(theme, "conv-1", "acct-1", "user-1",
		msgs.UserSubmittedInput{Text: "hi"}, 80, nil, nil, nil, scope)
}

func TestHandleStreamUpdate(t *testing.T) {
	t.Parallel()

	t.Run("end_turn completes", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming
		m.assistantMessage.SetID("asst-1") // skip SetContent path

		m.handleStreamUpdate(streamUpdate{
			message: &domain.Message{
				ID:         "asst-1",
				StopReason: "end_turn",
			},
			done: true,
		})

		if m.state != StateComplete {
			t.Errorf("expected StateComplete, got %d", m.state)
		}
	})

	t.Run("tool_use transitions to awaiting", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming
		m.assistantMessage.SetID("asst-1")

		m.handleStreamUpdate(streamUpdate{
			message: &domain.Message{
				ID:         "asst-1",
				StopReason: "tool_use",
				Content: []domain.Block{
					{Index: 1, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{ID: "tool-1", Name: "query"}},
				},
			},
			done: true,
		})

		if m.state != StateAwaitingToolResults {
			t.Errorf("expected StateAwaitingToolResults, got %d", m.state)
		}
		if m.pendingTools != 1 {
			t.Errorf("expected 1 pending tool, got %d", m.pendingTools)
		}
	})

	t.Run("tools already completed skips awaiting", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming
		m.assistantMessage.SetID("asst-1")
		// Tool completed during streaming, before stream finished
		m.toolResults = []tools.Result{{ToolUseID: "tool-1"}}

		m.handleStreamUpdate(streamUpdate{
			message: &domain.Message{
				ID:         "asst-1",
				StopReason: "tool_use",
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{ID: "tool-1", Name: "query"}},
				},
			},
			done: true,
		})

		if m.state != StateComplete {
			t.Errorf("expected StateComplete (tools already done), got %d", m.state)
		}
	})

	t.Run("stream error completes", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming

		m.handleStreamUpdate(streamUpdate{
			err:  errors.New("connection failed"),
			done: true,
		})

		if m.state != StateComplete {
			t.Errorf("expected StateComplete on error, got %d", m.state)
		}
	})

	t.Run("intermediate update stays streaming", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming
		m.assistantMessage.SetID("asst-1")

		m.handleStreamUpdate(streamUpdate{
			message: &domain.Message{
				ID: "asst-1",
			},
			done: false,
		})

		if m.state != StateStreaming {
			t.Errorf("expected StateStreaming during intermediate update, got %d", m.state)
		}
	})
}

func TestCancel(t *testing.T) {
	t.Parallel()

	t.Run("sets state to complete", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming

		m.Cancel()

		if m.state != StateComplete {
			t.Errorf("expected StateComplete after Cancel, got %d", m.state)
		}
	})

	t.Run("suppresses stream error after cancel", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming
		m.stream = &streamState{
			updates: make(chan streamUpdate),
			cancel:  func() {},
			done:    false,
		}

		m.Cancel()

		// Simulate the error that arrives after context cancellation
		cmd := m.handleStreamUpdate(streamUpdate{
			err:  errors.New("context canceled"),
			done: true,
		})

		// Should return nil (no error toast), not an error command
		if cmd != nil {
			t.Error("expected nil command after cancel, got non-nil (error was not suppressed)")
		}
	})

	t.Run("idempotent on idle turn", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)

		m.Cancel() // no stream, no panic

		if m.state != StateComplete {
			t.Errorf("expected StateComplete, got %d", m.state)
		}
	})
}

func TestHandleToolCompleted(t *testing.T) {
	t.Parallel()

	t.Run("collects results while streaming", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming

		m.handleToolCompleted("tool-1", tools.Result{ToolUseID: "tool-1"})

		if len(m.toolResults) != 1 {
			t.Errorf("expected 1 collected result, got %d", len(m.toolResults))
		}
		// Should not change state — still streaming
		if m.state != StateStreaming {
			t.Errorf("expected StateStreaming, got %d", m.state)
		}
	})

	t.Run("fires results when all collected and persisted", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateAwaitingToolResults
		m.persisted = true
		m.pendingTools = 2
		m.toolResults = []tools.Result{{ToolUseID: "tool-1"}}

		cmd := m.handleToolCompleted("tool-2", tools.Result{ToolUseID: "tool-2"})

		if m.state != StateComplete {
			t.Errorf("expected StateComplete when all tools done, got %d", m.state)
		}
		if len(m.toolResults) != 2 {
			t.Errorf("expected 2 results, got %d", len(m.toolResults))
		}
		if cmd == nil {
			t.Error("expected non-nil cmd (fireToolResults)")
		}
	})

	t.Run("does not fire until all tools complete", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateAwaitingToolResults
		m.pendingTools = 2

		m.handleToolCompleted("tool-1", tools.Result{ToolUseID: "tool-1"})

		if m.state != StateAwaitingToolResults {
			t.Errorf("expected StateAwaitingToolResults with 1 of 2 tools, got %d", m.state)
		}
	})

	t.Run("waits for persist before firing results", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateAwaitingToolResults
		m.persisted = false
		m.pendingTools = 1

		cmd := m.handleToolCompleted("tool-1", tools.Result{ToolUseID: "tool-1"})

		if m.state != StateComplete {
			t.Errorf("expected StateComplete, got %d", m.state)
		}
		if cmd != nil {
			t.Error("expected nil cmd — persist hasn't completed yet")
		}
	})
}

func TestPersistBeforeFireToolResults(t *testing.T) {
	t.Parallel()

	t.Run("tools pre-completed fires after persist", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateComplete
		m.pendingTools = 1
		m.toolResults = []tools.Result{{ToolUseID: "tool-1"}}
		m.persisted = false

		// Simulate assistantPersisted arriving
		cmd := m.Update(assistantPersisted{messageID: "asst-1"})

		if !m.persisted {
			t.Error("expected persisted = true")
		}
		if cmd == nil {
			t.Error("expected non-nil cmd (fireToolResults)")
		}
	})

	t.Run("tools complete after persist fires immediately", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateAwaitingToolResults
		m.persisted = true
		m.pendingTools = 1

		cmd := m.handleToolCompleted("tool-1", tools.Result{ToolUseID: "tool-1"})

		if cmd == nil {
			t.Error("expected non-nil cmd (fireToolResults) — already persisted")
		}
	})

	t.Run("no-op when no pending tools", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateComplete
		m.pendingTools = 0

		cmd := m.Update(assistantPersisted{messageID: "asst-1"})

		if !m.persisted {
			t.Error("expected persisted = true")
		}
		if cmd != nil {
			t.Error("expected nil cmd — no tools to fire")
		}
	})
}
