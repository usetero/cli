package preflight

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/core/bootstrap"
)

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case preflightAuthCheckCompletedMsg:
		return m.handleAuthChecked(msg)
	case preflightOrganizationsLoadedMsg:
		return m.handleOrganizationsLoaded(msg)
	case preflightAccountsLoadedMsg:
		return m.handleAccountsLoaded(msg)
	case preflightResolutionCompletedMsg:
		return m.handleResult(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}
	return nil
}

func (m *Model) handleAuthChecked(msg preflightAuthCheckCompletedMsg) tea.Cmd {
	m.state.HasValidAuth = msg.hasValidAuth
	if !m.state.HasValidAuth {
		return m.emitResult()
	}
	m.stage = stageOrganizations
	return m.loadOrganizations()
}

func (m *Model) handleOrganizationsLoaded(msg preflightOrganizationsLoadedMsg) tea.Cmd {
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
}

func (m *Model) handleAccountsLoaded(msg preflightAccountsLoadedMsg) tea.Cmd {
	if msg.err != nil {
		m.scope.Warn("preflight account lookup failed", "error", msg.err, "org_id", m.state.Org.ID)
		m.state.Outcome, m.state.Error = preflightOutcomeForError(msg.err)
		return m.emitResult()
	}
	m.state.Account = resolveAccount(msg.accounts, m.state.DefaultAccountID)
	return m.emitResult()
}

func (m *Model) handleResult(msg preflightResolutionCompletedMsg) tea.Cmd {
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
		return bootstrap.PreflightResolved{State: msg.state}
	}
}
