package commandbar

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

func TestShortHelpShowsCommandsWhenInputNotCapturing(t *testing.T) {
	m := New(theme.Default())
	m.SetCommands([]core.Command{{ID: core.CommandQuit, Title: "Quit"}})
	m.SetState(&core.Input{Kind: core.InputText, Placeholder: "Type here"}, nil, nil, nil)

	bindings := m.ShortHelp()
	if len(bindings) != 1 || bindings[0].Help().Key != "/" {
		t.Fatalf("expected slash commands help, got %#v", bindings)
	}
}

func TestShortHelpHidesCommandsWhenInputCaptures(t *testing.T) {
	m := New(theme.Default())
	m.SetCommands([]core.Command{{ID: core.CommandQuit, Title: "Quit"}})
	m.SetState(&core.Input{Kind: core.InputText, Placeholder: "Type here"}, nil, nil, nil)

	if _, consumed := m.HandleKey(tea.KeyPressMsg{Text: "a", Code: 'a'}); !consumed {
		t.Fatal("expected text input to consume typed key")
	}

	bindings := m.ShortHelp()
	for _, binding := range bindings {
		if binding.Help().Key == "/" {
			t.Fatalf("expected slash help to be hidden while input captures, got %#v", bindings)
		}
	}
}

func TestHandleKeyDoesNotOpenCommandsWhileInputCaptures(t *testing.T) {
	m := New(theme.Default())
	m.SetCommands([]core.Command{{ID: core.CommandQuit, Title: "Quit"}})
	m.SetState(&core.Input{Kind: core.InputText, Placeholder: "Type here"}, nil, nil, nil)

	if _, consumed := m.HandleKey(tea.KeyPressMsg{Text: "a", Code: 'a'}); !consumed {
		t.Fatal("expected text input to consume typed key")
	}

	if _, consumed := m.HandleKey(tea.KeyPressMsg{Text: "/", Code: '/'}); !consumed {
		t.Fatal("expected slash to be consumed by focused text input")
	}
	if m.paletteOpen {
		t.Fatal("expected palette to remain closed while text input captures")
	}
}
