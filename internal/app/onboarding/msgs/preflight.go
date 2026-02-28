package msgs

import "github.com/usetero/cli/internal/domain"

// PreflightOutcome indicates confidence in preflight-derived routing.
type PreflightOutcome string

const (
	PreflightOutcomeResolved     PreflightOutcome = "resolved"
	PreflightOutcomeInconclusive PreflightOutcome = "inconclusive"
	PreflightOutcomeFailed       PreflightOutcome = "failed"
)

// PreflightState captures startup readiness signals used to choose the first gate.
type PreflightState struct {
	Outcome            PreflightOutcome
	HasValidAuth       bool
	Role               string
	ActiveOrgID        domain.OrganizationID
	DefaultAccountID   domain.AccountID
	DefaultWorkspaceID domain.WorkspaceID
	Org                *domain.Organization
	Account            *domain.Account
	Workspace          *domain.Workspace
	HasDatadog         bool
	Error              string
}

// PreflightResolved is emitted after startup preflight completes.
type PreflightResolved struct {
	State PreflightState
}
