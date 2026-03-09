package datadogappkey

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/browser"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_SubmitRequiresBothFields(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.SetSite("US1")
	model.Update(tea.KeyPressMsg{Text: "n", Code: 'n'})
	model.Update(tea.KeyPressMsg{Text: "a", Code: 'a'})
	model.Update(tea.KeyPressMsg{Text: "m", Code: 'm'})
	model.Update(tea.KeyPressMsg{Text: "e", Code: 'e'})

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected open command when app key is empty")
	}
	if _, ok := cmd().(browser.OpenRequestedMsg); !ok {
		t.Fatalf("expected OpenRequestedMsg when app key is empty, got %T", cmd())
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

func TestModel_OpenEmitsBrowserIntent(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.SetSite("US1")

	_, cmd := model.Update(tea.KeyPressMsg{Text: "o", Code: 'o'})
	if cmd == nil {
		t.Fatal("expected open command")
	}
	msg := cmd()
	open, ok := msg.(browser.OpenRequestedMsg)
	if !ok {
		t.Fatalf("expected OpenRequestedMsg, got %T", msg)
	}
	if !strings.Contains(open.URL, "/organization-settings/service-accounts") {
		t.Fatalf("expected service accounts url, got %q", open.URL)
	}
}

func TestModel_EnterOpensBrowserWhenAppKeyEmpty(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.SetSite("US1")

	model.Update(tea.KeyPressMsg{Text: "a", Code: 'a'})
	model.Update(tea.KeyPressMsg{Text: "c", Code: 'c'})
	model.Update(tea.KeyPressMsg{Text: "c", Code: 'c'})
	model.Update(tea.KeyPressMsg{Text: "t", Code: 't'})

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected open command")
	}
	msg := cmd()
	open, ok := msg.(browser.OpenRequestedMsg)
	if !ok {
		t.Fatalf("expected OpenRequestedMsg, got %T", msg)
	}
	if !strings.Contains(open.URL, "/organization-settings/service-accounts") {
		t.Fatalf("expected service accounts url, got %q", open.URL)
	}
}
