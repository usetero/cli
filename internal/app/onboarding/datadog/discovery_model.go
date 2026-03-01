package datadog

import (
	"context"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	api "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/progress"
)

// DiscoveryModel polls for Datadog discovery status.
type DiscoveryModel struct {
	ctx              context.Context
	theme            styles.Theme
	services         api.APIServices
	scope            log.Scope
	datadogAccountID domain.DatadogAccountID

	spinner  spinner.Model
	progress *progress.Model
	status   *api.DatadogAccountStatus
	err      error
	width    int
	height   int
}

// NewDiscovery creates a new discovery step.
func NewDiscovery(
	ctx context.Context,
	theme styles.Theme,
	datadogAccountID domain.DatadogAccountID,
	services api.APIServices,
	scope log.Scope,
) *DiscoveryModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if datadogAccountID == "" {
		panic("datadogAccountID is empty")
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Accent).Background(theme.Bg)

	return &DiscoveryModel{
		ctx:              ctx,
		theme:            theme,
		services:         services,
		scope:            scope,
		datadogAccountID: datadogAccountID,
		spinner:          sp,
		progress:         progress.New(theme, 50),
	}
}

// Init starts polling for status.
func (m *DiscoveryModel) Init() tea.Cmd {
	m.scope.Info("starting datadog discovery", "datadogAccountID", m.datadogAccountID)
	return tea.Batch(m.spinner.Tick, m.pollStatus())
}

// SetSize updates dimensions.
func (m *DiscoveryModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.progress.SetWidth(min(width, 50))
}

// ShortHelp returns the key bindings for the short help view.
func (m *DiscoveryModel) ShortHelp() []key.Binding {
	if m.err != nil {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}
	}
	return nil
}
