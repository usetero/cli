// Package sync provides the sync step for onboarding.
package sync

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/progress"
)

// Model waits for sync to complete.
type Model struct {
	theme    styles.Theme
	syncer   powersync.Syncer
	scope    log.Scope
	spinner  spinner.Model
	progress *progress.Model
	width    int
	height   int
}

// New creates a new sync step.
func New(theme styles.Theme, syncer powersync.Syncer, scope log.Scope) *Model {
	if syncer == nil {
		panic("syncer is nil")
	}

	scope.Debug("initialized")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Accent).Background(theme.Bg)

	return &Model{
		theme:    theme,
		syncer:   syncer,
		scope:    scope,
		spinner:  sp,
		progress: progress.New(theme, 50),
	}
}

// Init starts the spinner and checks if already ready.
func (m *Model) Init() tea.Cmd {
	if m.syncer.IsReady() {
		m.scope.Info("sync already complete")
		return func() tea.Msg { return bootstrap.SyncComplete{} }
	}
	m.scope.Debug("waiting for sync")
	return m.spinner.Tick
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.progress.SetWidth(min(width, 50))
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	// Sync is automatic, no user action needed.
	return nil
}
