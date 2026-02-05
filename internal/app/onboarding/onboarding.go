// Package onboarding provides the onboarding flow for new users.
package onboarding

import (
	"context"
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app/layouts/header"
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
	logger   log.Logger
	layout   *header.Model

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
	logger log.Logger,
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
	if logger == nil {
		panic("logger is nil")
	}
	return &Model{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		auth:     authService,
		syncer:   syncer,
		logger:   logger,
		layout:   header.New(theme),
	}
}

// Init starts the onboarding flow with auth check.
func (m *Model) Init() tea.Cmd {
	m.logger.Info("onboarding started")
	return m.setStep(auth.NewCheck(m.ctx, m.theme, m.auth, m.logger))
}

// Update handles messages and orchestrates step transitions.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout.SetSize(msg.Width, msg.Height)
		if m.step != nil {
			contentWidth, contentHeight := m.layout.ContentSize()
			m.step.SetSize(contentWidth, contentHeight)
		}
		return nil

	// Auth messages
	case msgs.AuthChecked:
		m.logger.Debug("auth check complete", slog.Bool("needs_auth", msg.NeedsAuth))
		if msg.NeedsAuth {
			return m.setStep(auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.logger))
		}
		return m.setStep(role.New(m.theme, m.prefs, m.logger))

	case msgs.Authenticated:
		m.logger.Info("user authenticated")
		return m.setStep(role.New(m.theme, m.prefs, m.logger))

	// Role messages
	case msgs.RoleSelected:
		m.logger.Info("role selected", slog.String("role", string(msg.Role)))
		return m.setStep(organizations.NewSelect(m.ctx, m.theme, m.services, m.prefs, m.auth, m.logger))

	// Organization messages
	case msgs.OrgSelected:
		m.logger.Info("organization selected", slog.String("org_id", msg.Org.ID.String()))
		m.org = &msg.Org
		return m.setStep(accounts.NewSelect(m.ctx, m.theme, msg.Org, m.services, m.prefs, m.logger))

	case msgs.NoOrgs:
		m.logger.Debug("no organizations found")
		return m.setStep(organizations.NewCreate(m.ctx, m.theme, m.services, m.prefs, m.logger))

	case msgs.OrgCreated:
		m.logger.Info("organization created", slog.String("org_id", msg.Org.ID.String()))
		m.org = &msg.Org
		return m.setStep(accounts.NewSelect(m.ctx, m.theme, msg.Org, m.services, m.prefs, m.logger))

	// Account messages
	case msgs.AccountSelected:
		m.logger.Info("account selected", slog.String("account_id", msg.Account.ID.String()))
		m.org = &msg.Org
		m.account = &msg.Account
		return m.setStep(datadog.NewCheck(m.ctx, m.theme, msg.Account, m.services, m.logger))

	case msgs.NoAccounts:
		m.logger.Debug("no accounts found")
		return m.setStep(accounts.NewCreate(m.ctx, m.theme, msg.Org, m.services, m.prefs, m.logger))

	case msgs.AccountCreated:
		m.logger.Info("account created", slog.String("account_id", msg.Account.ID.String()))
		m.org = &msg.Org
		m.account = &msg.Account
		return m.setStep(datadog.NewCheck(m.ctx, m.theme, msg.Account, m.services, m.logger))

	// Datadog messages
	case msgs.DatadogReady:
		m.logger.Debug("datadog ready")
		return m.setStep(workspaces.NewSelect(m.ctx, m.theme, *m.account, m.services, m.prefs, m.logger))

	case msgs.DatadogNeeded:
		m.logger.Debug("datadog setup needed")
		return m.setStep(datadog.NewRegion(m.theme, m.logger))

	case msgs.DatadogRegionSelected:
		m.logger.Info("datadog region selected", slog.String("site", string(msg.Site)))
		m.ddSite = msg.Site
		return m.setStep(datadog.NewAPIKey(m.ctx, m.theme, *m.account, m.ddSite, m.services, m.logger))

	case msgs.DatadogAPIKeyEntered:
		m.logger.Debug("datadog api key validated")
		m.ddAPIKey = msg.APIKey
		return m.setStep(datadog.NewAppKey(m.ctx, m.theme, *m.account, m.ddSite, m.ddAPIKey, m.services, m.logger))

	case msgs.DatadogAccountCreated:
		m.logger.Info("datadog account created", slog.String("datadog_account_id", msg.DatadogAccountID.String()))
		return m.setStep(datadog.NewDiscovery(m.ctx, m.theme, msg.DatadogAccountID, m.services, m.logger))

	case msgs.DatadogDiscoveryComplete:
		m.logger.Info("datadog discovery complete")
		return m.setStep(workspaces.NewSelect(m.ctx, m.theme, *m.account, m.services, m.prefs, m.logger))

	// Workspace messages
	case msgs.WorkspaceSelected:
		m.logger.Info("workspace selected", slog.String("workspace_id", string(msg.Workspace.ID)))
		m.workspace = &msg.Workspace
		return m.setStep(sync.New(m.theme, m.syncer, m.logger))

	// Sync messages
	case msgs.SyncComplete:
		m.logger.Info("onboarding complete",
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
	contentWidth, contentHeight := m.layout.ContentSize()
	m.step.SetSize(contentWidth, contentHeight)
	return m.step.Init()
}

// View renders the current step.
func (m *Model) View() string {
	if m.step == nil {
		return ""
	}

	// Get content dimensions from layout
	contentWidth, contentHeight := m.layout.ContentSize()

	// Get step content
	stepContent := m.step.View()

	// Bottom-align step content in available space
	content := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		AlignVertical(lipgloss.Bottom).
		Render(stepContent)

	// Wrap in layout
	return m.layout.Render(content)
}

// SetSize updates the model's dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.layout.SetSize(width, height)
	if m.step != nil {
		contentWidth, contentHeight := m.layout.ContentSize()
		m.step.SetSize(contentWidth, contentHeight)
	}
}
