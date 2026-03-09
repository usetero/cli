package thinking

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_ViewIncludesLabel(t *testing.T) {
	model := New(theme.New(false), Settings{Size: 4, Label: "Thinking"})
	model.startTime = time.Now().Add(-birthDuration)
	view := ansi.Strip(model.View().Content)
	if !strings.Contains(view, "Thinking") {
		t.Fatalf("expected label in view, got %q", view)
	}
}

func TestModel_SetLabel(t *testing.T) {
	model := New(theme.New(false), Settings{Size: 4, Label: "One"})
	model.SetLabel("Two")
	model.startTime = time.Now().Add(-birthDuration)
	view := ansi.Strip(model.View().Content)
	if !strings.Contains(view, "Two") {
		t.Fatalf("expected updated label in view, got %q", view)
	}
}
