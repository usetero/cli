package turn

import (
	"errors"
	"testing"

	msgs "github.com/usetero/cli/internal/app/chat/events"
	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func newTestTurn(t *testing.T) *Model {
	t.Helper()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	// Stream runner and persister are nil — these tests exercise the state machine,
	// not persistence or streaming. The returned tea.Cmds are never executed.
	return New(theme, "conv-1", "acct-1", "user-1",
		msgs.UserSubmittedInput{Text: "hi"}, 80, nil, nil, nil, nil, scope)
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
		if m.toolTracker.pendingTools != 1 {
			t.Errorf("expected 1 pending tool, got %d", m.toolTracker.pendingTools)
		}
	})

	t.Run("tools already completed skips awaiting", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming
		m.assistantMessage.SetID("asst-1")
		// Tool completed during streaming, before stream finished
		m.toolTracker.results = []tools.Result{{ToolUseID: "tool-1"}}

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
			cancel:  func(error) {},
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

	t.Run("aborted user_cancelled does not emit completion cmd", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming

		cmd := m.handleStreamUpdate(streamUpdate{
			message: &domain.Message{ID: "asst-1"},
			status:  corechat.StreamStatusAborted,
			abort:   "user_cancelled",
			done:    true,
		})

		if cmd != nil {
			t.Error("expected nil cmd for user_cancelled abort")
		}
	})

	t.Run("aborted non-user emits completion cmd", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateStreaming

		cmd := m.handleStreamUpdate(streamUpdate{
			message: &domain.Message{ID: "asst-1"},
			status:  corechat.StreamStatusAborted,
			abort:   "context_canceled",
			done:    true,
		})

		if cmd == nil {
			t.Error("expected non-nil cmd for non-user abort")
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

		if len(m.toolTracker.results) != 1 {
			t.Errorf("expected 1 collected result, got %d", len(m.toolTracker.results))
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
		m.toolTracker.persisted = true
		m.toolTracker.pendingTools = 2
		m.toolTracker.results = []tools.Result{{ToolUseID: "tool-1"}}

		cmd := m.handleToolCompleted("tool-2", tools.Result{ToolUseID: "tool-2"})

		if m.state != StateComplete {
			t.Errorf("expected StateComplete when all tools done, got %d", m.state)
		}
		if len(m.toolTracker.results) != 2 {
			t.Errorf("expected 2 results, got %d", len(m.toolTracker.results))
		}
		if cmd == nil {
			t.Error("expected non-nil cmd (fireToolResults)")
		}
	})

	t.Run("does not fire until all tools complete", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateAwaitingToolResults
		m.toolTracker.pendingTools = 2

		m.handleToolCompleted("tool-1", tools.Result{ToolUseID: "tool-1"})

		if m.state != StateAwaitingToolResults {
			t.Errorf("expected StateAwaitingToolResults with 1 of 2 tools, got %d", m.state)
		}
	})

	t.Run("waits for persist before firing results", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateAwaitingToolResults
		m.toolTracker.persisted = false
		m.toolTracker.pendingTools = 1

		cmd := m.handleToolCompleted("tool-1", tools.Result{ToolUseID: "tool-1"})

		if m.state != StateComplete {
			t.Errorf("expected StateComplete, got %d", m.state)
		}
		if cmd != nil {
			t.Error("expected nil cmd — persist hasn't completed yet")
		}
	})

	t.Run("unknown tool completion increments protocol violation counter", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateAwaitingToolResults
		m.toolTracker.pendingTools = 1
		m.toolTracker.pendingToolIDs = map[string]bool{"tool-1": true}

		m.handleToolCompleted("tool-x", tools.Result{ToolUseID: "tool-x"})

		if got := m.protocolViolationCount; got != 1 {
			t.Fatalf("protocolViolationCount = %d, want 1", got)
		}
		if len(m.toolTracker.results) != 0 {
			t.Fatalf("toolResults len = %d, want 0", len(m.toolTracker.results))
		}
	})
}

func TestInterleavedToolUseFlow(t *testing.T) {
	t.Parallel()

	m := newTestTurn(t)
	m.state = StateStreaming
	m.assistantMessage.SetID("asst-1")

	// Stream completes with two tool_use blocks from interleaved tool input deltas.
	m.handleStreamUpdate(streamUpdate{
		message: &domain.Message{
			ID:         "asst-1",
			StopReason: "tool_use",
			Content: []domain.Block{
				{Index: 0, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{ID: "tool-a", Name: "query"}},
				{Index: 1, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{ID: "tool-b", Name: "query"}},
			},
		},
		done: true,
	})

	if m.state != StateAwaitingToolResults {
		t.Fatalf("expected StateAwaitingToolResults, got %d", m.state)
	}
	if m.toolTracker.pendingTools != 2 {
		t.Fatalf("expected 2 pending tools, got %d", m.toolTracker.pendingTools)
	}

	// Unknown tool result is ignored once pending IDs are fixed.
	m.handleToolCompleted("tool-c", tools.Result{ToolUseID: "tool-c"})
	if len(m.toolTracker.results) != 0 {
		t.Fatalf("expected 0 collected results after unknown tool, got %d", len(m.toolTracker.results))
	}

	// Interleaved completions should only complete once all known tools are done.
	m.handleToolCompleted("tool-b", tools.Result{ToolUseID: "tool-b"})
	if m.state != StateAwaitingToolResults {
		t.Fatalf("expected awaiting state after first completion, got %d", m.state)
	}
	m.handleToolCompleted("tool-a", tools.Result{ToolUseID: "tool-a"})
	if m.state != StateComplete {
		t.Fatalf("expected complete state after all tools, got %d", m.state)
	}
}

func TestPersistBeforeFireToolResults(t *testing.T) {
	t.Parallel()

	t.Run("tools pre-completed fires after persist", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateComplete
		m.toolTracker.pendingTools = 1
		m.toolTracker.results = []tools.Result{{ToolUseID: "tool-1"}}
		m.toolTracker.persisted = false

		// Simulate assistantPersisted arriving
		cmd := m.Update(assistantPersisted{turnID: m.userMessage.ID(), messageID: "asst-1"})

		if !m.toolTracker.persisted {
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
		m.toolTracker.persisted = true
		m.toolTracker.pendingTools = 1

		cmd := m.handleToolCompleted("tool-1", tools.Result{ToolUseID: "tool-1"})

		if cmd == nil {
			t.Error("expected non-nil cmd (fireToolResults) — already persisted")
		}
	})

	t.Run("no-op when no pending tools", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateComplete
		m.toolTracker.pendingTools = 0

		cmd := m.Update(assistantPersisted{turnID: m.userMessage.ID(), messageID: "asst-1"})

		if !m.toolTracker.persisted {
			t.Error("expected persisted = true")
		}
		if cmd != nil {
			t.Error("expected nil cmd — no tools to fire")
		}
	})
}

func TestFireToolResultsOnlyOnce(t *testing.T) {
	t.Parallel()

	t.Run("second call returns nil", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)
		m.state = StateAwaitingToolResults
		m.toolTracker.persisted = true
		m.toolTracker.pendingTools = 1

		cmd1 := m.handleToolCompleted("tool-1", tools.Result{ToolUseID: "tool-1"})
		if cmd1 == nil {
			t.Fatal("expected non-nil cmd on first fire")
		}

		cmd2 := m.fireToolResults()
		if cmd2 != nil {
			t.Error("expected nil cmd on second fireToolResults call")
		}
	})
}

func TestUpdate_TurnScopedRouting(t *testing.T) {
	t.Parallel()

	t.Run("tool completed turn mismatch is ignored", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)

		m.Update(msgs.ToolCompleted{
			TurnID:    "user-other",
			ToolUseID: "tool-1",
		})

		if got := m.protocolViolationCount; got != 0 {
			t.Fatalf("protocolViolationCount = %d, want 0", got)
		}
	})

	t.Run("assistant persisted turn mismatch is ignored", func(t *testing.T) {
		t.Parallel()
		m := newTestTurn(t)

		m.Update(assistantPersisted{
			turnID:    "user-other",
			messageID: "asst-1",
		})

		if got := m.protocolViolationCount; got != 0 {
			t.Fatalf("protocolViolationCount = %d, want 0", got)
		}
		if m.toolTracker.persisted {
			t.Fatal("persisted should remain false for mismatched turn")
		}
	})
}
