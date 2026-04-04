package statusbar

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar/alerts"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar/estate"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar/pressure"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar/scope"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar/sync"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type children struct {
	scope    *scope.Model
	estate   *estate.Model
	alerts   *alerts.Model
	pressure *pressure.Model
	sync     *sync.Model
}

// Model renders the global status bar header.
type Model struct {
	children core.Children
	env      string
	theme    theme.Theme
	width    int
	parts    children
}

var _ core.Model = (*Model)(nil)

// New constructs a status bar model.
func New(env string, appTheme theme.Theme) *Model {
	parts := children{
		scope:    scope.New(appTheme),
		estate:   estate.New(appTheme),
		alerts:   alerts.New(appTheme),
		pressure: pressure.New(appTheme),
		sync:     sync.New(appTheme),
	}

	return &Model{
		env:   env,
		theme: appTheme,
		parts: parts,
		children: core.Children{
			parts.scope,
			parts.estate,
			parts.alerts,
			parts.pressure,
			parts.sync,
		},
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd {
	return m.children.Init()
}

// Update satisfies tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.children.Update(msg)
}

func (m *Model) renderBrand() string {
	label := strings.ToUpper(theme.AppName)
	switch strings.ToLower(strings.TrimSpace(m.env)) {
	case "", "prod", "production":
		return m.theme.Shell.HeaderBrand.Bold(true).Render(label)
	default:
		return m.theme.Shell.HeaderBrand.Bold(true).Render(label + " " + strings.ToUpper(strings.TrimSpace(m.env)))
	}
}
