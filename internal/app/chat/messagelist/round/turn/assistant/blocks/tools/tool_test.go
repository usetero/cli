package tools

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/styles"
)

// stubChild implements Child with fixed return values.
type stubChild struct {
	name   string
	status string
	result string
	state  State
	toolID string
	err    error
	width  int
	view   string
}

func (s *stubChild) Update(tea.Msg) tea.Cmd { return nil }
func (s *stubChild) View() string           { return s.view }
func (s *stubChild) SetWidth(w int)         { s.width = w }
func (s *stubChild) Name() string           { return s.name }
func (s *stubChild) Status() string         { return s.status }
func (s *stubChild) Result() string         { return s.result }
func (s *stubChild) State() State           { return s.state }
func (s *stubChild) ToolID() string         { return s.toolID }
func (s *stubChild) Err() error             { return s.err }

func newTestTool(t *testing.T, child *stubChild) *Model {
	t.Helper()
	theme := styles.NewTheme(true)
	return New(theme, 0, "tool-1", 80, child)
}

func TestStatusRendering(t *testing.T) {
	t.Parallel()

	t.Run("pending shows icon and name", func(t *testing.T) {
		t.Parallel()
		child := &stubChild{name: "Query", state: StateAccumulating}
		m := newTestTool(t, child)
		m.updateStatus()
		view := m.View()

		if !strings.Contains(view, IconPending) {
			t.Error("expected pending icon")
		}
		if !strings.Contains(view, "Query") {
			t.Error("expected tool name")
		}
	})

	t.Run("running shows status message", func(t *testing.T) {
		t.Parallel()
		child := &stubChild{name: "Query", state: StateExecuting, status: "Checking services"}
		m := newTestTool(t, child)
		m.updateStatus()
		view := m.View()

		if !strings.Contains(view, IconPending) {
			t.Error("expected pending icon for running state")
		}
		if !strings.Contains(view, "Checking services") {
			t.Error("expected status message in view")
		}
	})

	t.Run("success shows result with chevron", func(t *testing.T) {
		t.Parallel()
		child := &stubChild{name: "Query", state: StateComplete, result: "Found 14 services"}
		m := newTestTool(t, child)
		m.updateStatus()
		view := m.View()

		if !strings.Contains(view, IconSuccess) {
			t.Error("expected success icon")
		}
		if !strings.Contains(view, "Found 14 services") {
			t.Error("expected result message in view")
		}
		if !strings.Contains(view, "▶") {
			t.Error("expected collapsed chevron when default collapsed")
		}
	})

	t.Run("error shows error tag and message", func(t *testing.T) {
		t.Parallel()
		child := &stubChild{
			name:  "Query",
			state: StateComplete,
			err:   errors.New("connection timeout"),
		}
		m := newTestTool(t, child)
		m.updateStatus()
		view := m.View()

		if !strings.Contains(view, IconError) {
			t.Error("expected error icon")
		}
		if !strings.Contains(view, "ERROR") {
			t.Error("expected ERROR tag in view")
		}
		if !strings.Contains(view, "connection timeout") {
			t.Error("expected error message in view")
		}
	})
}

func TestToggle(t *testing.T) {
	t.Parallel()

	child := &stubChild{
		name:   "Query",
		state:  StateComplete,
		result: "Found items",
		view:   "detailed output here",
	}
	m := newTestTool(t, child)
	m.updateStatus()

	// Default is collapsed
	collapsedView := m.View()
	if strings.Contains(collapsedView, "detailed output here") {
		t.Error("expected no body content when collapsed by default")
	}
	if !strings.Contains(collapsedView, "▶") {
		t.Error("expected collapsed chevron")
	}

	// Toggle to expand
	m.Toggle()
	expandedView := m.View()
	if !strings.Contains(expandedView, "detailed output here") {
		t.Error("expected body content when expanded")
	}
	if !strings.Contains(expandedView, "▼") {
		t.Error("expected expanded chevron")
	}

	// Expanded should be taller
	expandedLines := strings.Count(expandedView, "\n")
	collapsedLines := strings.Count(collapsedView, "\n")
	if collapsedLines >= expandedLines {
		t.Errorf("collapsed (%d lines) should be shorter than expanded (%d lines)", collapsedLines, expandedLines)
	}
}

func TestKind(t *testing.T) {
	t.Parallel()
	child := &stubChild{name: "Query", state: StateAccumulating}
	m := newTestTool(t, child)
	if m.Kind() != block.KindTool {
		t.Errorf("expected KindTool, got %d", m.Kind())
	}
}
