// Package msgs defines the messages emitted by onboarding steps.
// These are the contracts between steps and the onboarding orchestrator.
package msgs

import (
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

// Role constants.
const (
	RolePlatform = "platform"
	RoleEngineer = "engineer"
)

// Auth messages.

// Authenticated is emitted when authentication succeeds.
type Authenticated struct {
	User auth.User
}

// Role messages.

// RoleSelected is emitted when user selects their role.
type RoleSelected struct {
	Role string
}

// Organization messages.

// OrgSelected is emitted when user selects an organization.
type OrgSelected struct {
	Org domain.Organization
}

// NoOrgs is emitted when no organizations exist.
type NoOrgs struct{}

// OrgCreated is emitted when a new organization is created.
type OrgCreated struct {
	Org domain.Organization
}

// Account messages.

// AccountSelected is emitted when user selects an account.
type AccountSelected struct {
	Org     domain.Organization
	Account domain.Account
}

// NoAccounts is emitted when no accounts exist for the org.
type NoAccounts struct {
	Org domain.Organization
}

// AccountCreated is emitted when a new account is created.
type AccountCreated struct {
	Org     domain.Organization
	Account domain.Account
}

// Datadog messages.

// DatadogReady is emitted when datadog is already configured.
type DatadogReady struct{}

// DatadogNeeded is emitted when datadog setup is required.
type DatadogNeeded struct{}

// DatadogRegionSelected is emitted when user selects a datadog region.
type DatadogRegionSelected struct {
	Site domain.DatadogSite
}

// DatadogAPIKeyEntered is emitted when user enters API key.
type DatadogAPIKeyEntered struct {
	APIKey string
}

// DatadogAccountCreated is emitted when datadog account is created.
type DatadogAccountCreated struct {
	DatadogAccountID domain.DatadogAccountID
}

// DatadogDiscoveryComplete is emitted when discovery finishes.
type DatadogDiscoveryComplete struct{}

// Workspace messages.

// WorkspaceSelected is emitted when user selects a workspace.
type WorkspaceSelected struct {
	Workspace domain.Workspace
}

// Sync messages.

// SyncComplete is emitted when initial sync finishes.
type SyncComplete struct{}

// OnboardingComplete is emitted to the root model when onboarding finishes.
type OnboardingComplete struct {
	User      *auth.User
	Org       domain.Organization
	Account   domain.Account
	Workspace domain.Workspace
}

// Preflight messages.

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
