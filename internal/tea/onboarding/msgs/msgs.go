// Package msgs defines the messages emitted by onboarding steps.
// These are the contracts between steps and the onboarding orchestrator.
package msgs

import "github.com/usetero/cli/internal/domain"

// AuthChecked is emitted when the auth check completes.
type AuthChecked struct {
	NeedsAuth bool
}

// Authenticated is emitted when authentication succeeds.
type Authenticated struct{}

// RoleSelected is emitted when user selects their role.
type RoleSelected struct {
	Role string
}

// Role constants.
const (
	RolePlatform = "platform"
	RoleEngineer = "engineer"
)

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

// DatadogReady is emitted when datadog is already configured.
type DatadogReady struct{}

// DatadogNeeded is emitted when datadog setup is required.
type DatadogNeeded struct{}

// DatadogConfigured is emitted when datadog setup completes.
type DatadogConfigured struct{}

// WorkspaceSelected is emitted when user selects a workspace.
type WorkspaceSelected struct {
	Workspace domain.Workspace
}

// SyncComplete is emitted when initial sync finishes.
type SyncComplete struct{}

// OnboardingComplete is emitted to the root model when onboarding finishes.
type OnboardingComplete struct {
	Org       domain.Organization
	Account   domain.Account
	Workspace domain.Workspace
}
