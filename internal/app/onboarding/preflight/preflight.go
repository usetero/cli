// Package preflight resolves initial onboarding context before rendering step UIs.
package preflight

import (
	"context"
	"errors"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

type resultMsg struct {
	state msgs.PreflightState
}

type authCheckedMsg struct {
	hasValidAuth bool
}

type orgsLoadedMsg struct {
	orgs []domain.Organization
	err  error
}

type accountsLoadedMsg struct {
	accounts []domain.Account
	err      error
}

type stage int

const (
	stageStarting stage = iota
	stageAuth
	stageOrganizations
	stageAccounts
	stageFinalizing
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
	state    msgs.PreflightState
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

	return &Model{
		ctx:      ctx,
		theme:    theme,
		services: services,
		auth:     authService,
		userPref: userPref,
		orgPref:  orgPref,
		scope:    scope,
		stage:    stageStarting,
		spinner: func() spinner.Model {
			sp := spinner.New()
			sp.Spinner = spinner.Dot
			sp.Style = lipgloss.NewStyle().Foreground(theme.Accent).Background(theme.Bg)
			return sp
		}(),
	}
}

func (m *Model) Init() tea.Cmd {
	m.started = time.Now()
	m.state = msgs.PreflightState{
		Outcome:            msgs.PreflightOutcomeResolved,
		Role:               m.userPref.GetRole(),
		ActiveOrgID:        m.userPref.GetActiveOrgID(),
		DefaultAccountID:   m.orgPref.GetDefaultAccountID(),
		DefaultWorkspaceID: m.orgPref.GetDefaultWorkspaceID(),
	}
	m.stage = stageAuth
	return tea.Batch(m.spinner.Tick, m.checkAuth())
}

func (m *Model) checkAuth() tea.Cmd {
	return func() tea.Msg {
		hasValidAuth := false
		if m.auth.IsAuthenticated() {
			if _, err := m.auth.GetAccessToken(m.ctx); err == nil {
				hasValidAuth = true
			} else {
				_ = m.auth.ClearTokens()
			}
		}
		return authCheckedMsg{hasValidAuth: hasValidAuth}
	}
}

func (m *Model) loadOrganizations() tea.Cmd {
	return func() tea.Msg {
		orgs, err := m.services.WithAccountID("").Organizations.List(m.ctx)
		return orgsLoadedMsg{orgs: orgs, err: err}
	}
}

func (m *Model) loadAccounts(orgID domain.OrganizationID) tea.Cmd {
	return func() tea.Msg {
		accounts, err := m.services.WithAccountID("").Accounts.List(m.ctx, orgID)
		return accountsLoadedMsg{accounts: accounts, err: err}
	}
}

func (m *Model) emitResult() tea.Cmd {
	m.stage = stageFinalizing
	return func() tea.Msg {
		return resultMsg{state: m.state}
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case authCheckedMsg:
		m.state.HasValidAuth = msg.hasValidAuth
		if !m.state.HasValidAuth {
			return m.emitResult()
		}
		m.stage = stageOrganizations
		return m.loadOrganizations()

	case orgsLoadedMsg:
		if msg.err != nil {
			m.scope.Warn("preflight org lookup failed", "error", msg.err)
			m.state.Outcome, m.state.Error = preflightOutcomeForError(msg.err)
			return m.emitResult()
		}
		m.state.Org = resolveOrg(msg.orgs, m.state.ActiveOrgID)
		if m.state.Org == nil {
			return m.emitResult()
		}
		m.stage = stageAccounts
		return m.loadAccounts(m.state.Org.ID)

	case accountsLoadedMsg:
		if msg.err != nil {
			m.scope.Warn("preflight account lookup failed", "error", msg.err, "org_id", m.state.Org.ID)
			m.state.Outcome, m.state.Error = preflightOutcomeForError(msg.err)
			return m.emitResult()
		}
		m.state.Account = resolveAccount(msg.accounts, m.state.DefaultAccountID)
		return m.emitResult()

	case resultMsg:
		elapsed := time.Since(m.started)
		m.scope.Debug("preflight resolved",
			"has_valid_auth", msg.state.HasValidAuth,
			"role", msg.state.Role,
			"active_org_id", msg.state.ActiveOrgID,
			"default_account_id", msg.state.DefaultAccountID,
			"default_workspace_id", msg.state.DefaultWorkspaceID,
			"outcome", msg.state.Outcome,
			"org", msg.state.Org != nil,
			"account", msg.state.Account != nil,
			"error", msg.state.Error,
			"elapsed_ms", elapsed.Milliseconds())
		return func() tea.Msg {
			return msgs.PreflightResolved{State: msg.state}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}
	return nil
}

func (m *Model) View() string {
	s := m.theme.Styles
	title := s.Title.Render("Getting ready")
	statusLine := m.spinner.View() + " " + s.Body.Render(m.stageText())
	return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)
}

func (m *Model) SetSize(width, height int) {}

func (m *Model) ShortHelp() []key.Binding { return nil }

func resolveOrg(orgs []domain.Organization, activeOrgID domain.OrganizationID) *domain.Organization {
	if activeOrgID != "" {
		for _, org := range orgs {
			if org.ID == activeOrgID {
				resolved := org
				return &resolved
			}
		}
	}
	if len(orgs) == 1 {
		resolved := orgs[0]
		return &resolved
	}
	return nil
}

func resolveAccount(accounts []domain.Account, defaultAccountID domain.AccountID) *domain.Account {
	if defaultAccountID != "" {
		for _, account := range accounts {
			if account.ID == defaultAccountID {
				resolved := account
				return &resolved
			}
		}
	}
	if len(accounts) == 1 {
		resolved := accounts[0]
		return &resolved
	}
	return nil
}

func preflightOutcomeForError(err error) (msgs.PreflightOutcome, string) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return msgs.PreflightOutcomeFailed, err.Error()
	}
	return msgs.PreflightOutcomeInconclusive, err.Error()
}

func (m *Model) stageText() string {
	switch m.stage {
	case stageAuth:
		return "Checking authentication"
	case stageOrganizations:
		return "Loading organizations"
	case stageAccounts:
		return "Loading accounts"
	case stageFinalizing:
		return "Finalizing setup"
	default:
		return "Preparing onboarding..."
	}
}
