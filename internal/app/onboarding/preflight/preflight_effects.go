package preflight

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
)

func (m *Model) checkAuth() tea.Cmd {
	return func() tea.Msg {
		hasValidAuth := false
		var user *auth.User
		if m.auth.IsAuthenticated() {
			if _, err := m.auth.GetAccessToken(m.ctx); err == nil {
				hasValidAuth = true
				if userID, err := m.auth.GetUserID(m.ctx); err == nil && userID != "" {
					user = &auth.User{ID: userID}
				} else {
					// Avoid getting stuck in sync with a valid token but no user identity.
					_ = m.auth.ClearTokens()
					hasValidAuth = false
				}
			} else {
				_ = m.auth.ClearTokens()
			}
		}
		return preflightAuthCheckCompletedMsg{hasValidAuth: hasValidAuth, user: user}
	}
}

func (m *Model) loadOrganizations() tea.Cmd {
	return func() tea.Msg {
		orgs, err := m.services.WithAccountID("").Organizations.List(m.ctx)
		return preflightOrganizationsLoadedMsg{orgs: orgs, err: err}
	}
}

func (m *Model) loadAccounts(orgID domain.OrganizationID) tea.Cmd {
	return func() tea.Msg {
		accounts, err := m.services.WithAccountID("").Accounts.List(m.ctx, orgID)
		return preflightAccountsLoadedMsg{accounts: accounts, err: err}
	}
}

func (m *Model) emitResult() tea.Cmd {
	m.stage = stageFinalizing
	return func() tea.Msg {
		return preflightResolutionCompletedMsg{state: m.state}
	}
}

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

func preflightOutcomeForError(err error) (bootstrap.PreflightOutcome, string) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return bootstrap.PreflightOutcomeFailed, err.Error()
	}
	return bootstrap.PreflightOutcomeInconclusive, err.Error()
}
