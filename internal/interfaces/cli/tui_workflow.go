package cli

import (
	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/domains/integrations"
	domainprefs "github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
	prefstore "github.com/usetero/cli/internal/infrastructure/preferences"
	runtimeonboarding "github.com/usetero/cli/internal/runtime/onboarding"
)

func newOnboardingWorkflow(exec *runner, identityService *identity.Service) (*runtimeonboarding.Workflow, error) {
	if identityService == nil {
		panic("onboarding workflow requires identity service")
	}

	store, err := prefstore.NewStore(string(exec.cfg.Env))
	if err != nil {
		return nil, err
	}

	prefs := domainprefs.NewService(store)
	client := controlplane.NewClient(exec.cfg.API.Origin, identityService)
	orgs := tenancy.NewRemoteOrganizationService(client)
	workspaces := tenancy.NewRemoteWorkspaceService(client)
	datadog := integrations.NewRemoteDatadogService(client)

	return runtimeonboarding.NewWorkflow(
		prefs,
		orgs,
		func(organizationID tenancy.OrganizationID) tenancy.AccountService {
			return tenancy.NewRemoteAccountService(client, organizationID)
		},
		workspaces,
		datadog,
		nil,
	), nil
}
