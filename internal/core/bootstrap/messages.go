package bootstrap

import (
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

// Message is a typed onboarding/bootstrap message contract.
type Message interface {
	bootstrapMessage()
}

type Authenticated struct {
	User auth.User
}

type RoleSelected struct {
	Role string
}

type OrgSelected struct {
	Org domain.Organization
}

type NoOrgs struct{}

type OrgCreated struct {
	Org domain.Organization
}

type AccountSelected struct {
	Org     domain.Organization
	Account domain.Account
}

type NoAccounts struct {
	Org domain.Organization
}

type AccountCreated struct {
	Org     domain.Organization
	Account domain.Account
}

type EnsureRuntime struct {
	Org     domain.Organization
	Account domain.Account
}

type RuntimeReady struct {
	Org     domain.Organization
	Account domain.Account
}

type DatadogReady struct{}

type DatadogNeeded struct{}

type DatadogRegionSelected struct {
	Site domain.DatadogSite
}

type DatadogAPIKeyEntered struct {
	APIKey string
}

type DatadogAccountCreated struct {
	DatadogAccountID domain.DatadogAccountID
}

type DatadogDiscoveryComplete struct{}

type WorkspaceSelected struct {
	Workspace domain.Workspace
}

type SyncComplete struct{}

type OnboardingComplete struct {
	User      *auth.User
	Org       domain.Organization
	Account   domain.Account
	Workspace domain.Workspace
}

type PreflightState struct {
	Outcome            PreflightOutcome
	HasValidAuth       bool
	User               *auth.User
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

type PreflightResolved struct {
	State PreflightState
}

func (Authenticated) bootstrapMessage()         {}
func (RoleSelected) bootstrapMessage()          {}
func (OrgSelected) bootstrapMessage()           {}
func (NoOrgs) bootstrapMessage()                {}
func (OrgCreated) bootstrapMessage()            {}
func (AccountSelected) bootstrapMessage()       {}
func (NoAccounts) bootstrapMessage()            {}
func (AccountCreated) bootstrapMessage()        {}
func (RuntimeReady) bootstrapMessage()          {}
func (DatadogReady) bootstrapMessage()          {}
func (DatadogNeeded) bootstrapMessage()         {}
func (DatadogRegionSelected) bootstrapMessage() {}
func (DatadogAPIKeyEntered) bootstrapMessage()  {}
func (DatadogAccountCreated) bootstrapMessage() {}
func (DatadogDiscoveryComplete) bootstrapMessage() {
}
func (WorkspaceSelected) bootstrapMessage() {}
func (SyncComplete) bootstrapMessage()      {}
func (PreflightResolved) bootstrapMessage() {}
