package bootstrap

// PreflightOutcome indicates confidence in preflight-derived routing.
type PreflightOutcome string

const (
	PreflightOutcomeResolved     PreflightOutcome = "resolved"
	PreflightOutcomeInconclusive PreflightOutcome = "inconclusive"
	PreflightOutcomeFailed       PreflightOutcome = "failed"
)

// Role constants.
const (
	RolePlatform = "platform"
	RoleEngineer = "engineer"
)

// PreflightInput captures readiness signals needed for initial gate selection.
type PreflightInput struct {
	Outcome      PreflightOutcome
	HasValidAuth bool
	Role         string
	HasOrg       bool
	HasAccount   bool
}

// DecideNextGate maps preflight input to the next deterministic onboarding gate.
func DecideNextGate(in PreflightInput) Gate {
	if in.Outcome == PreflightOutcomeFailed || !in.HasValidAuth {
		return GateAuthenticate
	}
	if in.Role != RolePlatform && in.Role != RoleEngineer {
		return GateRoleSelect
	}
	if !in.HasOrg {
		return GateOrgSelect
	}
	if !in.HasAccount {
		return GateAccountSelect
	}
	return GateRuntimeInit
}
