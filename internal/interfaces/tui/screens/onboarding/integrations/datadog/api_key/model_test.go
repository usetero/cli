package datadogapikey

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/browser"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_SubmitValidationAndMessage(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.SetSite("US1")

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected open command when api key is empty")
	}
	if _, ok := cmd().(browser.OpenRequestedMsg); !ok {
		t.Fatalf("expected OpenRequestedMsg when api key is empty, got %T", cmd())
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
	if !strings.Contains(open.URL, "/organization-settings/api-keys") {
		t.Fatalf("expected api key url, got %q", open.URL)
	}
}

func TestModel_EnterOpensBrowserWhenEmpty(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.SetSite("US1")

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected open command")
	}
	msg := cmd()
	open, ok := msg.(browser.OpenRequestedMsg)
	if !ok {
		t.Fatalf("expected OpenRequestedMsg, got %T", msg)
	}
	if !strings.Contains(open.URL, "/organization-settings/api-keys") {
		t.Fatalf("expected api key url, got %q", open.URL)
	}
}
