package onboarding

import (
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
)

// Runtime is the onboarding runtime lifecycle type.
type Runtime = Service

// NewRuntime constructs the onboarding runtime.
func NewRuntime(
	preferences preferences.PreferenceService,
	orgs tenancy.OrganizationService,
	accounts func(organizationID tenancy.OrganizationID) tenancy.AccountService,
	workspaces tenancy.WorkspaceService,
	datadog integrations.DatadogService,
	readiness pssyncer.ReadinessService,
) *Runtime {
	return NewService(preferences, orgs, accounts, workspaces, datadog, readiness)
}
