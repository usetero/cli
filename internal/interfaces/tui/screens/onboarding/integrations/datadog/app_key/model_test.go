package datadogappkey

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_SubmitRequiresBothFields(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.Update(tea.KeyPressMsg{Text: "n", Code: 'n'})
	model.Update(tea.KeyPressMsg{Text: "a", Code: 'a'})
	model.Update(tea.KeyPressMsg{Text: "m", Code: 'm'})
	model.Update(tea.KeyPressMsg{Text: "e", Code: 'e'})

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected nil command when app key missing")
	}
}

func TestModel_SubmitEmitsMessage(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))

	// Fill name.
	model.Update(tea.KeyPressMsg{Text: "a", Code: 'a'})
	model.Update(tea.KeyPressMsg{Text: "c", Code: 'c'})
	model.Update(tea.KeyPressMsg{Text: "c", Code: 'c'})
	model.Update(tea.KeyPressMsg{Text: "t", Code: 't'})

	// Move to app key and fill it.
	model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	model.Update(tea.KeyPressMsg{Text: "x", Code: 'x'})
	model.Update(tea.KeyPressMsg{Text: "y", Code: 'y'})
	model.Update(tea.KeyPressMsg{Text: "z", Code: 'z'})

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected submit command")
	}
	msg := cmd()
	submitted, ok := msg.(SubmittedMsg)
	if !ok {
		t.Fatalf("expected SubmittedMsg, got %T", msg)
	}
	if submitted.Name != "acct" || submitted.AppKey != "xyz" {
		t.Fatalf("unexpected submitted payload: %#v", submitted)
	}
}
