package accountcreate

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_InputAndSubmit(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))

	model.Update(tea.KeyPressMsg{Text: "A", Code: 'a'})
	model.Update(tea.KeyPressMsg{Text: "c", Code: 'c'})
	model.Update(tea.KeyPressMsg{Text: "m", Code: 'm'})
	model.Update(tea.KeyPressMsg{Text: "e", Code: 'e'})

	if got := model.Name(); got != "Acme" {
		t.Fatalf("expected name Acme, got %q", got)
	}

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected create command")
	}
	msg := cmd()
	created, ok := msg.(CreatedMsg)
	if !ok {
		t.Fatalf("expected CreatedMsg, got %T", msg)
	}
	if created.Create.Name != "Acme" {
		t.Fatalf("expected create name Acme, got %q", created.Create.Name)
	}
}
