package preflight

import (
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
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
