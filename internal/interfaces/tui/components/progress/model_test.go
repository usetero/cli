package progress

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_WidthClamp(t *testing.T) {
	model := New(theme.New(false), 1)
	view := ansi.Strip(model.ViewAs(50))
	if !strings.Contains(view, "%") {
		t.Fatalf("expected percentage in progress bar, got %q", view)
	}
	if strings.Count(view, "█")+strings.Count(view, "░") < 10 {
		t.Fatalf("expected minimum width 10, got %q", view)
	}
}

func TestModel_PercentClamp(t *testing.T) {
	model := New(theme.New(false), 10)

	low := ansi.Strip(model.ViewAs(-10))
	high := ansi.Strip(model.ViewAs(200))

	if strings.Count(low, "█") != 0 {
		t.Fatalf("expected 0 filled at low clamp, got %q", low)
	}
	if strings.Count(high, "░") != 0 {
		t.Fatalf("expected full bar at high clamp, got %q", high)
	}
}
