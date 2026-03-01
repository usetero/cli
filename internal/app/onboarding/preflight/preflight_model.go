// Package preflight resolves initial onboarding context before rendering step UIs.
package preflight

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/auth"
	api "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

// Model computes onboarding preconditions up-front.
type Model struct {
	ctx      context.Context
	theme    styles.Theme
	services api.APIServices
	auth     auth.Auth
	userPref preferences.UserPreferences
	orgPref  preferences.OrgPreferences
	scope    log.Scope
	state    bootstrap.PreflightState
	stage    stage
	spinner  spinner.Model
	started  time.Time
}

func New(
	ctx context.Context,
	theme styles.Theme,
	services api.APIServices,
	authService auth.Auth,
	userPref preferences.UserPreferences,
	orgPref preferences.OrgPreferences,
	scope log.Scope,
) *Model {
	if ctx == nil {
		panic("ctx is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}
	if userPref == nil {
		panic("userPref is nil")
	}
	if orgPref == nil {
		panic("orgPref is nil")
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Accent).Background(theme.Bg)

	return &Model{
		ctx:      ctx,
		theme:    theme,
		services: services,
		auth:     authService,
		userPref: userPref,
		orgPref:  orgPref,
		scope:    scope,
		stage:    stageStarting,
		spinner:  sp,
	}
}

func (m *Model) Init() tea.Cmd {
	m.started = time.Now()
	m.state = bootstrap.PreflightState{
		Outcome:            bootstrap.PreflightOutcomeResolved,
		Role:               m.userPref.GetRole(),
		ActiveOrgID:        m.userPref.GetActiveOrgID(),
		DefaultAccountID:   m.orgPref.GetDefaultAccountID(),
		DefaultWorkspaceID: m.orgPref.GetDefaultWorkspaceID(),
	}
	m.stage = stageAuth
	return tea.Batch(m.spinner.Tick, m.checkAuth())
}

func (m *Model) SetSize(width, height int) {}

func (m *Model) ShortHelp() []key.Binding { return nil }
