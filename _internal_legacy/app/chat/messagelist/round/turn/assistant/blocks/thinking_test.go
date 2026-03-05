package blocks

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/styles"
)

func newTestThinking(t *testing.T, text string) *ThinkingBlock {
	t.Helper()
	theme := styles.NewTheme(true)
	return NewThinkingBlock(theme, 0, text, 80)
}

func TestThinkingCollapsedByDefault(t *testing.T) {
	t.Parallel()
	m := newTestThinking(t, "some internal reasoning")
	view := m.View()
	plain := ansi.Strip(view)

	if strings.Contains(plain, "reasoning") {
		t.Error("expected collapsed view to hide content")
	}
	if !strings.Contains(plain, "▶") {
		t.Error("expected collapsed chevron")
	}
	if !strings.Contains(plain, "Thinking") {
		t.Error("expected Thinking label")
	}
}

func TestThinkingExpanded(t *testing.T) {
	t.Parallel()
	m := newTestThinking(t, "some internal reasoning")
	m.Toggle(0)
	view := m.View()
	plain := ansi.Strip(view)

	if !strings.Contains(plain, "reasoning") {
		t.Errorf("expected expanded view to show content, got:\n%s", plain)
	}
	if !strings.Contains(plain, "▼") {
		t.Error("expected expanded chevron")
	}
	if !strings.Contains(plain, "Thinking") {
		t.Error("expected Thinking label")
	}
}

func TestThinkingToggle(t *testing.T) {
	t.Parallel()
	m := newTestThinking(t, "reasoning text here")

	collapsedView := m.View()
	m.Toggle(0)
	expandedView := m.View()

	collapsedLines := lipgloss.Height(collapsedView)
	expandedLines := lipgloss.Height(expandedView)
	if collapsedLines >= expandedLines {
		t.Errorf("collapsed (%d lines) should be shorter than expanded (%d lines)", collapsedLines, expandedLines)
	}

	// Toggle back to collapsed
	m.Toggle(0)
	plain := ansi.Strip(m.View())
	if strings.Contains(plain, "reasoning") {
		t.Error("expected content hidden after toggling back")
	}
}

func TestThinkingEmptyText(t *testing.T) {
	t.Parallel()
	m := newTestThinking(t, "")
	m.Toggle(0)
	plain := ansi.Strip(m.View())

	// Should still render header even with empty text
	if !strings.Contains(plain, "Thinking") {
		t.Error("expected Thinking label even with empty text")
	}
}
