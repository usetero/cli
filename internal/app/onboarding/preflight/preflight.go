// Package preflight resolves initial onboarding context before rendering step UIs.
package preflight

import (
	"context"
	"errors"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

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

// Model computes onboarding preconditions up-front.
type Model struct {
	ctx      context.Context
	theme    styles.Theme
	services api.APIServices
	auth     auth.Auth
	userPref preferences.UserPreferences
	orgPref  preferences.OrgPreferences
	scope    log.Scope
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
	}
}

func (m *Model) Init() tea.Cmd {
	return m.resolve()
}

func (m *Model) resolve() tea.Cmd {
	return func() tea.Msg {
		hasValidAuth := false
		if m.auth.IsAuthenticated() {
			if _, err := m.auth.GetAccessToken(m.ctx); err == nil {
				hasValidAuth = true
			} else {
				_ = m.auth.ClearTokens()
			}
		}

		state := msgs.PreflightState{
			Outcome:            msgs.PreflightOutcomeResolved,
			HasValidAuth:       hasValidAuth,
			Role:               m.userPref.GetRole(),
			ActiveOrgID:        m.userPref.GetActiveOrgID(),
			DefaultAccountID:   m.orgPref.GetDefaultAccountID(),
			DefaultWorkspaceID: m.orgPref.GetDefaultWorkspaceID(),
		}
		if !state.HasValidAuth {
			return resultMsg{state: state}
		}

		unscoped := m.services.WithAccountID("")

		orgs, err := unscoped.Organizations.List(m.ctx)
		if err != nil {
			m.scope.Warn("preflight org lookup failed", "error", err)
			state.Outcome, state.Error = preflightOutcomeForError(err)
			return resultMsg{state: state}
		}
		state.Org = resolveOrg(orgs, state.ActiveOrgID)
		if state.Org == nil {
			return resultMsg{state: state}
		}

		accounts, err := unscoped.Accounts.List(m.ctx, state.Org.ID)
		if err != nil {
			m.scope.Warn("preflight account lookup failed", "error", err, "org_id", state.Org.ID)
			state.Outcome, state.Error = preflightOutcomeForError(err)
			return resultMsg{state: state}
		}
		state.Account = resolveAccount(accounts, state.DefaultAccountID)
		if state.Account == nil {
			return resultMsg{state: state}
		}

		accountScoped := m.services.WithAccountID(state.Account.ID)
		hasDatadog, err := accountScoped.DatadogAccounts.HasAccount(m.ctx, state.Account.ID)
		if err != nil {
			m.scope.Warn("preflight datadog lookup failed", "error", err, "account_id", state.Account.ID)
			state.Outcome, state.Error = preflightOutcomeForError(err)
			return resultMsg{state: state}
		}
		state.HasDatadog = hasDatadog
		if !state.HasDatadog {
			return resultMsg{state: state}
		}

		workspaces, err := accountScoped.Workspaces.List(m.ctx, state.Account.ID)
		if err != nil {
			m.scope.Warn("preflight workspace lookup failed", "error", err, "account_id", state.Account.ID)
			state.Outcome, state.Error = preflightOutcomeForError(err)
			return resultMsg{state: state}
		}
		state.Workspace = resolveWorkspace(workspaces, state.DefaultWorkspaceID)

		return resultMsg{state: state}
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case resultMsg:
		m.scope.Debug("preflight resolved",
			"has_valid_auth", msg.state.HasValidAuth,
			"role", msg.state.Role,
			"active_org_id", msg.state.ActiveOrgID,
			"default_account_id", msg.state.DefaultAccountID,
			"default_workspace_id", msg.state.DefaultWorkspaceID,
			"outcome", msg.state.Outcome,
			"org", msg.state.Org != nil,
			"account", msg.state.Account != nil,
			"workspace", msg.state.Workspace != nil,
			"has_datadog", msg.state.HasDatadog,
			"error", msg.state.Error)
		return func() tea.Msg {
			return msgs.PreflightResolved{State: msg.state}
		}
	}
	return nil
}

func (m *Model) View() string {
	return m.theme.Styles.Title.Render("Preparing onboarding...")
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

func resolveWorkspace(workspaces []domain.Workspace, defaultWorkspaceID domain.WorkspaceID) *domain.Workspace {
	if defaultWorkspaceID != "" {
		for _, workspace := range workspaces {
			if workspace.ID == defaultWorkspaceID {
				resolved := workspace
				return &resolved
			}
		}
	}
	if len(workspaces) == 1 {
		resolved := workspaces[0]
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
