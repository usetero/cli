package datadogapikey

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_SubmitValidationAndMessage(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected nil command when api key is empty")
	}

	model.Update(tea.KeyPressMsg{Text: "k", Code: 'k'})
	model.Update(tea.KeyPressMsg{Text: "e", Code: 'e'})
	model.Update(tea.KeyPressMsg{Text: "y", Code: 'y'})

	_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected submit command")
	}
	msg := cmd()
	submitted, ok := msg.(SubmittedMsg)
	if !ok {
		t.Fatalf("expected SubmittedMsg, got %T", msg)
	}
	if submitted.APIKey != "key" {
		t.Fatalf("expected api key key, got %q", submitted.APIKey)
	}
}
