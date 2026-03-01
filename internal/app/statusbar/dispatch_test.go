package statusbar

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

func TestToggleDrawerRequiresData(t *testing.T) {
	t.Parallel()

	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")
	m.tabs = []drawerTab{
		stubDrawerTab{label: "A", hasData: false},
		stubDrawerTab{label: "B", hasData: true},
	}

	m.ToggleDrawer()
	if !m.drawerOpen {
		t.Fatalf("drawer should open when at least one tab has data")
	}

	m.ToggleDrawer()
	if m.drawerOpen {
		t.Fatalf("drawer should close when toggled while open")
	}
}

func TestHandleEscDelegatesToActiveTab(t *testing.T) {
	t.Parallel()

	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")
	closed := false
	m.tabs = []drawerTab{
		stubDrawerTab{
			label:       "A",
			detail:      true,
			onClose:     func() { closed = true },
			hasData:     true,
			interactive: true,
		},
	}
	m.activeTab = 0

	if !m.HandleEsc() {
		t.Fatalf("expected esc to be consumed by active detail tab")
	}
	if !closed {
		t.Fatalf("expected esc to close active tab detail")
	}
}

func TestHandleKeyPressUsesInteractiveTabsOnly(t *testing.T) {
	t.Parallel()

	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")
	called := false
	m.tabs = []drawerTab{
		stubDrawerTab{
			label:       "A",
			interactive: false,
			onHandle: func(msg tea.KeyPressMsg) tea.Cmd {
				called = true
				return nil
			},
		},
		stubDrawerTab{
			label:       "B",
			interactive: true,
			onHandle: func(msg tea.KeyPressMsg) tea.Cmd {
				called = true
				return nil
			},
		},
	}

	m.activeTab = 0
	_ = m.HandleKeyPress(tea.KeyPressMsg{})
	if called {
		t.Fatalf("non-interactive tab should not receive key presses")
	}

	m.activeTab = 1
	_ = m.HandleKeyPress(tea.KeyPressMsg{})
	if !called {
		t.Fatalf("interactive tab should receive key presses")
	}
}

type stubDrawerTab struct {
	label       string
	hasData     bool
	interactive bool
	detail      bool
	onClose     func()
	onHandle    func(msg tea.KeyPressMsg) tea.Cmd
}

func (s stubDrawerTab) Label() string                { return s.label }
func (s stubDrawerTab) SetDB(_ sqlite.DB) tea.Cmd    { return nil }
func (s stubDrawerTab) Init() tea.Cmd                { return nil }
func (s stubDrawerTab) Update(_ tea.Msg) tea.Cmd     { return nil }
func (s stubDrawerTab) HasData() bool                { return s.hasData }
func (s stubDrawerTab) CompactView() string          { return "" }
func (s stubDrawerTab) ExpandedView(_, _ int) string { return "" }
func (s stubDrawerTab) Interactive() bool            { return s.interactive }
func (s stubDrawerTab) InDetail() bool               { return s.detail }
func (s stubDrawerTab) CloseDetail() {
	if s.onClose != nil {
		s.onClose()
	}
}
func (s stubDrawerTab) HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if s.onHandle == nil {
		return nil
	}
	return s.onHandle(msg)
}
