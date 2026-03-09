package form

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_UpdatesActiveField(t *testing.T) {
	model := New(theme.New(false), FieldSpec{ID: "name", Label: "Name: ", Placeholder: "Name"})
	model.Update(tea.KeyPressMsg{Text: "a", Code: 'a'})
	model.Update(tea.KeyPressMsg{Text: "b", Code: 'b'})

	if got := model.Value("name"); got != "ab" {
		t.Fatalf("expected value ab, got %q", got)
	}
}

func TestModel_TabSwitchesFields(t *testing.T) {
	model := New(
		theme.New(false),
		FieldSpec{ID: "name", Label: "Name: ", Placeholder: "Name"},
		FieldSpec{ID: "key", Label: "Key: ", Placeholder: "Key"},
	)

	model.Update(tea.KeyPressMsg{Text: "a", Code: 'a'})
	model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	model.Update(tea.KeyPressMsg{Text: "x", Code: 'x'})

	if model.ActiveField() != "key" {
		t.Fatalf("expected active field key, got %q", model.ActiveField())
	}
	if got := model.Value("name"); got != "a" {
		t.Fatalf("expected name value a, got %q", got)
	}
	if got := model.Value("key"); got != "x" {
		t.Fatalf("expected key value x, got %q", got)
	}
}

func TestModel_ViewRendersLabelsAndValues(t *testing.T) {
	model := New(theme.New(false), FieldSpec{ID: "name", Label: "Name: ", Placeholder: "Name"})
	model.SetValue("name", "Acme")
	view := ansi.Strip(model.View().Content)

	if !strings.Contains(view, "Name: ") || !strings.Contains(view, "Acme") {
		t.Fatalf("expected rendered field, got %q", view)
	}
}
