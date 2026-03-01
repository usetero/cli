package preflight

import (
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
)

type preflightResolvedMsg struct {
	state bootstrap.PreflightState
}

type preflightAuthCheckedMsg struct {
	hasValidAuth bool
}

type preflightOrganizationsLoadedMsg struct {
	orgs []domain.Organization
	err  error
}

type preflightAccountsLoadedMsg struct {
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
