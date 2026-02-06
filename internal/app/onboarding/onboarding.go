// Package onboarding provides the onboarding flow for new users.
package onboarding

import (
	"context"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app/onboarding/accounts"
	"github.com/usetero/cli/internal/app/onboarding/auth"
	"github.com/usetero/cli/internal/app/onboarding/datadog"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/app/onboarding/organizations"
	"github.com/usetero/cli/internal/app/onboarding/role"
	"github.com/usetero/cli/internal/app/onboarding/sync"
	"github.com/usetero/cli/internal/app/onboarding/workspaces"
	iauth "github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

// Model is the onboarding orchestrator.
type Model struct {
	// Dependencies available from start
	ctx      context.Context
	theme    *styles.Theme
	services api.APIServices
	prefs    preferences.Preferences
	auth     iauth.Auth
	syncer   powersync.Syncer
	scope    log.Scope

	// Accumulated state from step completions
	org       *domain.Organization
	account   *domain.Account
	workspace *domain.Workspace
	ddSite    domain.DatadogSite
	ddAPIKey  string

	// Current step
	step   Step
	width  int
	height int
}

// New creates a new onboarding model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	services api.APIServices,
	prefs preferences.Preferences,
	authService iauth.Auth,
	syncer powersync.Syncer,
	scope log.Scope,
) *Model {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}
	if syncer == nil {
		panic("syncer is nil")
	}

	scope = scope.Child("onboarding")

	return &Model{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		auth:     authService,
		syncer:   syncer,
		scope:    scope,
	}
}

// Init starts the onboarding flow with auth check.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("onboarding started")
	return m.setStep(auth.NewCheck(m.ctx, m.theme, m.auth, m.scope))
}

// Update handles messages and orchestrates step transitions.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.step != nil {
			m.step.SetSize(m.width, m.height)
		}
		return nil

	// Auth messages
	case msgs.AuthChecked:
		m.scope.Debug("auth check complete", slog.Bool("needs_auth", msg.NeedsAuth))
		if msg.NeedsAuth {
			return m.setStep(auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.scope))
		}
		return m.setStep(role.New(m.theme, m.prefs, m.scope))

	case msgs.Authenticated:
		m.scope.Info("user authenticated")
		return m.setStep(role.New(m.theme, m.prefs, m.scope))

	// Role messages
	case msgs.RoleSelected:
		m.scope.Info("role selected", slog.String("role", msg.Role))
		return m.setStep(organizations.NewSelect(m.ctx, m.theme, m.services, m.prefs, m.auth, m.scope))

	// Organization messages
	case msgs.OrgSelected:
		m.scope.Info("organization selected", slog.String("org_id", msg.Org.ID.String()))
		m.org = &msg.Org
		return m.setStep(accounts.NewSelect(m.ctx, m.theme, msg.Org, m.services, m.prefs, m.scope))

	case msgs.NoOrgs:
		m.scope.Debug("no organizations found")
		return m.setStep(organizations.NewCreate(m.ctx, m.theme, m.services, m.prefs, m.scope))

	case msgs.OrgCreated:
		m.scope.Info("organization created", slog.String("org_id", msg.Org.ID.String()))
		m.org = &msg.Org
		return m.setStep(accounts.NewSelect(m.ctx, m.theme, msg.Org, m.services, m.prefs, m.scope))

	// Account messages
	case msgs.AccountSelected:
		m.scope.Info("account selected", slog.String("account_id", msg.Account.ID.String()))
		m.org = &msg.Org
		m.account = &msg.Account
		return m.setStep(datadog.NewCheck(m.ctx, m.theme, msg.Account, m.services, m.scope))

	case msgs.NoAccounts:
		m.scope.Debug("no accounts found")
		return m.setStep(accounts.NewCreate(m.ctx, m.theme, msg.Org, m.services, m.prefs, m.scope))

	case msgs.AccountCreated:
		m.scope.Info("account created", slog.String("account_id", msg.Account.ID.String()))
		m.org = &msg.Org
		m.account = &msg.Account
		return m.setStep(datadog.NewCheck(m.ctx, m.theme, msg.Account, m.services, m.scope))

	// Datadog messages
	case msgs.DatadogReady:
		m.scope.Debug("datadog ready")
		return m.setStep(workspaces.NewSelect(m.ctx, m.theme, *m.account, m.services, m.prefs, m.scope))

	case msgs.DatadogNeeded:
		m.scope.Debug("datadog setup needed")
		return m.setStep(datadog.NewRegion(m.theme, m.scope))

	case msgs.DatadogRegionSelected:
		m.scope.Info("datadog region selected", slog.String("site", string(msg.Site)))
		m.ddSite = msg.Site
		return m.setStep(datadog.NewAPIKey(m.ctx, m.theme, *m.account, m.ddSite, m.services, m.scope))

	case msgs.DatadogAPIKeyEntered:
		m.scope.Debug("datadog api key validated")
		m.ddAPIKey = msg.APIKey
		return m.setStep(datadog.NewAppKey(m.ctx, m.theme, *m.account, m.ddSite, m.ddAPIKey, m.services, m.scope))

	case msgs.DatadogAccountCreated:
		m.scope.Info("datadog account created", slog.String("datadog_account_id", msg.DatadogAccountID.String()))
		return m.setStep(datadog.NewDiscovery(m.ctx, m.theme, msg.DatadogAccountID, m.services, m.scope))

	case msgs.DatadogDiscoveryComplete:
		m.scope.Info("datadog discovery complete")
		return m.setStep(workspaces.NewSelect(m.ctx, m.theme, *m.account, m.services, m.prefs, m.scope))

	// Workspace messages
	case msgs.WorkspaceSelected:
		m.scope.Info("workspace selected", slog.String("workspace_id", string(msg.Workspace.ID)))
		m.workspace = &msg.Workspace
		return m.setStep(sync.New(m.theme, m.syncer, m.scope))

	// Sync messages
	case msgs.SyncComplete:
		m.scope.Info("onboarding complete",
			slog.String("org_id", m.org.ID.String()),
			slog.String("account_id", m.account.ID.String()),
			slog.String("workspace_id", string(m.workspace.ID)),
		)
		return func() tea.Msg {
			return msgs.OnboardingComplete{
				Org:       *m.org,
				Account:   *m.account,
				Workspace: *m.workspace,
			}
		}
	}

	// Delegate to current step
	if m.step != nil {
		return m.step.Update(msg)
	}
	return nil
}

// setStep sets the current step and initializes it.
func (m *Model) setStep(step Step) tea.Cmd {
	m.step = step
	m.step.SetSize(m.width, m.height)
	return m.step.Init()
}

// View renders the current step.
func (m *Model) View() string {
	if m.step == nil {
		return ""
	}

	// Bottom-align step content in available space
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignVertical(lipgloss.Bottom).
		Render(m.step.View())
}

// SetSize updates the model's dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.step != nil {
		m.step.SetSize(width, height)
	}
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	if m.step != nil {
		return m.step.ShortHelp()
	}
	return nil
}
