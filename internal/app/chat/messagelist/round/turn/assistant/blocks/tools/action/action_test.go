package action

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log/logtest"
)

func testConfig() Config {
	return Config{
		DisplayName: func(_ json.RawMessage) string { return "Test Tool" },
		Status:      func(_ json.RawMessage) string { return "Running test" },
		Result:      func(_ domaintools.Result) string { return "Done" },
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("accumulates input from content blocks", func(t *testing.T) {
		t.Parallel()

		m := New(0, "turn-1", "tool-1", 80, testConfig(), nil, logtest.NewScope(t))

		m.Update(msgs.AssistantContentUpdated{
			Message: domain.Message{
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{
						ID:            "tool-1",
						Input:         json.RawMessage(`{"key":"val`),
						InputComplete: false,
					}},
				},
			},
		})

		if m.state != tools.StateAccumulating {
			t.Errorf("expected StateAccumulating, got %d", m.state)
		}
	})

	t.Run("ignores blocks with wrong index", func(t *testing.T) {
		t.Parallel()

		m := New(0, "turn-1", "tool-1", 80, testConfig(), nil, logtest.NewScope(t))

		m.Update(msgs.AssistantContentUpdated{
			Message: domain.Message{
				Content: []domain.Block{
					{Index: 5, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{
						ID:            "tool-1",
						Input:         json.RawMessage(`{"key":"value"}`),
						InputComplete: true,
					}},
				},
			},
		})

		if m.state != tools.StateAccumulating {
			t.Errorf("expected StateAccumulating (wrong index ignored), got %d", m.state)
		}
	})

	t.Run("executes on InputComplete and fires ToolCompleted", func(t *testing.T) {
		t.Parallel()

		executor := func(input json.RawMessage) (domaintools.Result, error) {
			return domaintools.Result{Content: map[string]any{"ok": true}}, nil
		}

		m := New(0, "turn-1", "tool-1", 80, testConfig(), executor, logtest.NewScope(t))

		cmd := m.Update(msgs.StreamCompleted{
			Message: domain.Message{
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{
						ID:            "tool-1",
						Input:         json.RawMessage(`{}`),
						InputComplete: true,
					}},
				},
			},
		})

		if m.state != tools.StateComplete {
			t.Fatalf("expected StateComplete, got %d", m.state)
		}
		if m.err != nil {
			t.Fatalf("unexpected error: %v", m.err)
		}

		// Execute the command and check the message
		msg := cmd()
		completed, ok := msg.(msgs.ToolCompleted)
		if !ok {
			t.Fatalf("expected msgs.ToolCompleted, got %T", msg)
		}
		if completed.ToolUseID != "tool-1" {
			t.Errorf("ToolUseID = %q, want %q", completed.ToolUseID, "tool-1")
		}
		if completed.TurnID != "turn-1" {
			t.Errorf("TurnID = %q, want %q", completed.TurnID, "turn-1")
		}
		if completed.Error != nil {
			t.Errorf("unexpected error in completed: %v", completed.Error)
		}
	})

	t.Run("fires ToolCompleted with error on failure", func(t *testing.T) {
		t.Parallel()

		executor := func(input json.RawMessage) (domaintools.Result, error) {
			return domaintools.Result{}, errors.New("exec failed")
		}

		m := New(0, "turn-1", "tool-1", 80, testConfig(), executor, logtest.NewScope(t))

		cmd := m.Update(msgs.StreamCompleted{
			Message: domain.Message{
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{
						ID:            "tool-1",
						Input:         json.RawMessage(`{}`),
						InputComplete: true,
					}},
				},
			},
		})

		if m.state != tools.StateComplete {
			t.Fatalf("expected StateComplete, got %d", m.state)
		}
		if m.err == nil {
			t.Fatal("expected error, got nil")
		}

		msg := cmd()
		completed, ok := msg.(msgs.ToolCompleted)
		if !ok {
			t.Fatalf("expected msgs.ToolCompleted, got %T", msg)
		}
		if completed.Error == nil {
			t.Error("expected error in completed message")
		}
	})

	t.Run("does not re-execute after completion", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		executor := func(input json.RawMessage) (domaintools.Result, error) {
			callCount++
			return domaintools.Result{Content: map[string]any{"ok": true}}, nil
		}

		m := New(0, "turn-1", "tool-1", 80, testConfig(), executor, logtest.NewScope(t))

		content := []domain.Block{
			{Index: 0, Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{
				ID:            "tool-1",
				Input:         json.RawMessage(`{}`),
				InputComplete: true,
			}},
		}

		m.Update(msgs.StreamCompleted{Message: domain.Message{Content: content}})
		m.Update(msgs.StreamCompleted{Message: domain.Message{Content: content}})

		if callCount != 1 {
			t.Errorf("executor called %d times, want 1", callCount)
		}
	})
}

func TestConfigDelegation(t *testing.T) {
	t.Parallel()

	config := Config{
		DisplayName: func(input json.RawMessage) string {
			var m map[string]any
			if err := json.Unmarshal(input, &m); err != nil {
				return "Enable"
			}
			if m["action"] == "disable" {
				return "Disable"
			}
			return "Enable"
		},
		Status: func(_ json.RawMessage) string { return "Working" },
		Result: func(r domaintools.Result) string {
			name, _ := r.Content["name"].(string)
			return name + " done"
		},
	}

	m := New(0, "turn-1", "tool-1", 80, config, nil, logtest.NewScope(t))
	m.input = json.RawMessage(`{"action":"disable"}`)
	m.result = domaintools.Result{Content: map[string]any{"name": "svc"}}

	if got := m.Name(); got != "Disable" {
		t.Errorf("Name() = %q, want %q", got, "Disable")
	}
	if got := m.Status(); got != "Working" {
		t.Errorf("Status() = %q, want %q", got, "Working")
	}
	if got := m.Result(); got != "svc done" {
		t.Errorf("Result() = %q, want %q", got, "svc done")
	}
}
