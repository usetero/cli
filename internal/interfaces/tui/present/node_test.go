package present

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestNotice(t *testing.T) {
	out := ansi.Strip(Render(theme.New(false), Notice("Welcome", "Body copy")))
	if !strings.Contains(out, "Welcome") || !strings.Contains(out, "Body copy") {
		t.Fatalf("expected notice content, got %q", out)
	}
}

func TestErrorCard(t *testing.T) {
	out := Render(theme.New(false), ErrorCard(BlockGap(1, Error("Failed"), Body("boom"))))
	if !strings.Contains(out, "Failed") || !strings.Contains(out, "boom") {
		t.Fatalf("expected error card content, got %q", out)
	}
	if !strings.Contains(out, "╭") {
		t.Fatalf("expected bordered card, got %q", out)
	}
	if strings.Contains(out, "\x1b[m     \x1b[m") {
		t.Fatalf("expected padded tail to keep background styling, got %q", out)
	}
}

func TestField(t *testing.T) {
	out := ansi.Strip(Render(theme.New(false), Field("> ", "Name: ", Body("Acme"))))
	if !strings.Contains(out, "Name: ") || !strings.Contains(out, "Acme") {
		t.Fatalf("expected field content, got %q", out)
	}
}

func TestStatusBlock(t *testing.T) {
	out := ansi.Strip(Render(theme.New(false), StatusBlock("Syncing", Body("Connecting..."), Muted("3 / 10 rows"))))
	if !strings.Contains(out, "Syncing") || !strings.Contains(out, "Connecting...") || !strings.Contains(out, "3 / 10 rows") {
		t.Fatalf("expected status block content, got %q", out)
	}
	if !strings.Contains(out, "╭") {
		t.Fatalf("expected bordered card, got %q", out)
	}
}
