package loading

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestSpinnerModel_ViewRendersLabelAndDetail(t *testing.T) {
	model := NewSpinner(theme.New(false), "Connecting")
	model.SetDetail("Waiting for service")

	view := ansi.Strip(model.View().Content)
	if !strings.Contains(view, "Connecting") {
		t.Fatalf("expected label in view, got %q", view)
	}
	if !strings.Contains(view, "Waiting for service") {
		t.Fatalf("expected detail in view, got %q", view)
	}
}

func TestThinkingModel_SetLabelUpdatesView(t *testing.T) {
	model := NewThinking(theme.New(false), "Loading")
	model.SetLabel("Discovering")
	model.SetDetail("This can take a moment")

	if got := model.label; got != "Discovering" {
		t.Fatalf("expected label to update, got %q", got)
	}

	cmd := model.Init()
	if cmd == nil {
		t.Fatal("expected initial animation command")
	}
	model.Update(cmd())
	view := ansi.Strip(model.View().Content)
	if !strings.Contains(view, "This can take a moment") {
		t.Fatalf("expected detail in view, got %q", view)
	}
}

func TestThinkingModel_InitReturnsTick(t *testing.T) {
	model := NewThinking(theme.New(false), "Loading")
	if cmd := model.Init(); cmd == nil {
		t.Fatal("expected initial animation command")
	}
}

func TestSpinnerModel_UpdateIgnoresNonSpinnerMessages(t *testing.T) {
	model := NewSpinner(theme.New(false), "Connecting")
	next, cmd := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if next != model {
		t.Fatal("expected spinner loading to keep same model instance")
	}
	if cmd != nil {
		t.Fatal("expected no command for unrelated message")
	}
}
