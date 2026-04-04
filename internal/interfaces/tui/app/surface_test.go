package app

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	chromedivider "github.com/usetero/cli/internal/interfaces/tui/components/chrome/divider"
	"github.com/usetero/cli/internal/interfaces/tui/components/commandbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/helpbar"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type switchScreenMsg struct{}

type stubScreen struct {
	page     core.Page
	help     []key.Binding
	commands []core.Command
	next     core.Screen
}

func (s *stubScreen) Init() tea.Cmd                        { return nil }
func (s *stubScreen) View() tea.View                       { return tea.NewView("") }
func (s *stubScreen) SetSize(_, _ int)                     {}
func (s *stubScreen) Page() core.Page                      { return s.page }
func (s *stubScreen) Commands() []core.Command             { return s.commands }
func (s *stubScreen) ShortHelp() []key.Binding             { return s.help }
func (s *stubScreen) Input() *core.Input                   { return nil }
func (s *stubScreen) Busy() *core.Busy                     { return nil }
func (s *stubScreen) Error() *core.Error                   { return nil }
func (s *stubScreen) Notice() *core.Notice                 { return nil }
func (s *stubScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(switchScreenMsg); ok && s.next != nil {
		return s.next, nil
	}
	return s, nil
}

func TestSurfaceUpdateRefreshesShellStateFromActiveScreen(t *testing.T) {
	appTheme := theme.Default()
	second := &stubScreen{
		page: core.Page{Title: "Second"},
		help: []key.Binding{
			key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "beta")),
		},
	}
	first := &stubScreen{
		page: core.Page{Title: "First"},
		help: []key.Binding{
			key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "alpha")),
		},
		next: second,
	}

	s := newSurface(
		first,
		statusbar.New("dev", appTheme),
		chromedivider.New(appTheme),
		commandbar.New(appTheme),
		helpbar.New(appTheme),
	)
	s.shell.helpbar.SetSize(80, 1)

	initial := s.shell.helpbar.View().Content
	if !strings.Contains(initial, "alpha") {
		t.Fatalf("expected initial help to include alpha, got %q", initial)
	}

	_ = s.Update(switchScreenMsg{})
	s.shell.helpbar.SetSize(80, 1)

	updated := s.shell.helpbar.View().Content
	if strings.Contains(updated, "alpha") {
		t.Fatalf("expected updated help to drop alpha, got %q", updated)
	}
	if !strings.Contains(updated, "beta") {
		t.Fatalf("expected updated help to include beta, got %q", updated)
	}
}
