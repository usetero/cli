package onboarding

import (
	"github.com/usetero/cli/internal/app/onboarding/accounts"
	"github.com/usetero/cli/internal/app/onboarding/auth"
	"github.com/usetero/cli/internal/app/onboarding/datadog"
	"github.com/usetero/cli/internal/app/onboarding/organizations"
	"github.com/usetero/cli/internal/app/onboarding/preflight"
	"github.com/usetero/cli/internal/app/onboarding/role"
	"github.com/usetero/cli/internal/app/onboarding/runtimeinit"
	"github.com/usetero/cli/internal/app/onboarding/sync"
	"github.com/usetero/cli/internal/app/onboarding/workspaces"
)

func (m *Model) newStepForGate(gate Gate) (StepCore, bool) {
	switch gate {
	case GatePreflight:
		return preflight.New(m.ctx, m.theme, m.services, m.auth, m.userPrefs, m.orgPrefs, m.scope), true
	case GateAuthenticate:
		return auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.scope), true
	case GateRoleSelect:
		return role.New(m.theme, m.userPrefs, m.scope), true
	case GateOrgSelect:
		return organizations.NewSelect(m.ctx, m.theme, m.services, m.userPrefs, m.auth, m.scope), true
	case GateOrgCreate:
		return organizations.NewCreate(m.ctx, m.theme, m.services, m.userPrefs, m.scope), true
	case GateAccountSelect:
		return accounts.NewSelect(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope), true
	case GateAccountCreate:
		return accounts.NewCreate(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope), true
	case GateRuntimeInit:
		return runtimeinit.New(m.theme, *m.state.Org, *m.state.Account, m.scope), true
	case GateDatadogCheck:
		return datadog.NewCheck(m.ctx, m.theme, *m.state.Account, m.services, m.scope), true
	case GateDatadogRegion:
		return datadog.NewRegion(m.theme, m.scope), true
	case GateDatadogAPIKey:
		return datadog.NewAPIKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.services, m.scope), true
	case GateDatadogAppKey:
		return datadog.NewAppKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.state.DDAPIKey, m.services, m.scope), true
	case GateDatadogDiscovery:
		return datadog.NewDiscovery(m.ctx, m.theme, m.state.DDAccount, m.services, m.scope), true
	case GateWorkspaceSelect:
		return workspaces.NewSelect(m.ctx, m.theme, *m.state.Account, m.services, m.orgPrefs, m.scope), true
	case GateSync:
		return sync.New(m.theme, m.syncer, m.scope), true
	default:
		return nil, false
	}
}
