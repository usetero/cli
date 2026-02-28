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

type gateDefinition struct {
	newStep func(m *Model) Step
}

func (m *Model) newStepForGate(gate Gate) Step {
	switch gate {
	case GatePreflight:
		return preflight.New(m.ctx, m.theme, m.services, m.auth, m.userPrefs, m.orgPrefs, m.scope)
	case GateAuthenticate:
		return auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.scope)
	case GateRoleSelect:
		return role.New(m.theme, m.userPrefs, m.scope)
	case GateOrgSelect:
		return organizations.NewSelect(m.ctx, m.theme, m.services, m.userPrefs, m.auth, m.scope)
	case GateOrgCreate:
		return organizations.NewCreate(m.ctx, m.theme, m.services, m.userPrefs, m.scope)
	case GateAccountSelect:
		return accounts.NewSelect(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope)
	case GateAccountCreate:
		return accounts.NewCreate(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope)
	case GateRuntimeInit:
		return runtimeinit.New(m.theme, *m.state.Org, *m.state.Account, m.scope)
	case GateDatadogCheck:
		return datadog.NewCheck(m.ctx, m.theme, *m.state.Account, m.services, m.scope)
	case GateDatadogRegion:
		return datadog.NewRegion(m.theme, m.scope)
	case GateDatadogAPIKey:
		return datadog.NewAPIKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.services, m.scope)
	case GateDatadogAppKey:
		return datadog.NewAppKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.state.DDAPIKey, m.services, m.scope)
	case GateDatadogDiscovery:
		return datadog.NewDiscovery(m.ctx, m.theme, m.state.DDAccount, m.services, m.scope)
	case GateWorkspaceSelect:
		return workspaces.NewSelect(m.ctx, m.theme, *m.state.Account, m.services, m.orgPrefs, m.scope)
	case GateSync:
		return sync.New(m.theme, m.syncer, m.scope)
	default:
		panic("unsupported onboarding gate: " + gate.String())
	}
}

func defaultGateDefinitions() map[Gate]gateDefinition {
	return map[Gate]gateDefinition{
		GatePreflight: {newStep: func(m *Model) Step {
			return preflight.New(m.ctx, m.theme, m.services, m.auth, m.userPrefs, m.orgPrefs, m.scope)
		}},
		GateAuthenticate: {newStep: func(m *Model) Step {
			return auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.scope)
		}},
		GateRoleSelect: {newStep: func(m *Model) Step {
			return role.New(m.theme, m.userPrefs, m.scope)
		}},
		GateOrgSelect: {newStep: func(m *Model) Step {
			return organizations.NewSelect(m.ctx, m.theme, m.services, m.userPrefs, m.auth, m.scope)
		}},
		GateOrgCreate: {newStep: func(m *Model) Step {
			return organizations.NewCreate(m.ctx, m.theme, m.services, m.userPrefs, m.scope)
		}},
		GateAccountSelect: {newStep: func(m *Model) Step {
			return accounts.NewSelect(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope)
		}},
		GateAccountCreate: {newStep: func(m *Model) Step {
			return accounts.NewCreate(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope)
		}},
		GateRuntimeInit: {newStep: func(m *Model) Step {
			return runtimeinit.New(m.theme, *m.state.Org, *m.state.Account, m.scope)
		}},
		GateDatadogCheck: {newStep: func(m *Model) Step {
			return datadog.NewCheck(m.ctx, m.theme, *m.state.Account, m.services, m.scope)
		}},
		GateDatadogRegion: {newStep: func(m *Model) Step {
			return datadog.NewRegion(m.theme, m.scope)
		}},
		GateDatadogAPIKey: {newStep: func(m *Model) Step {
			return datadog.NewAPIKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.services, m.scope)
		}},
		GateDatadogAppKey: {newStep: func(m *Model) Step {
			return datadog.NewAppKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.state.DDAPIKey, m.services, m.scope)
		}},
		GateDatadogDiscovery: {newStep: func(m *Model) Step {
			return datadog.NewDiscovery(m.ctx, m.theme, m.state.DDAccount, m.services, m.scope)
		}},
		GateWorkspaceSelect: {newStep: func(m *Model) Step {
			return workspaces.NewSelect(m.ctx, m.theme, *m.state.Account, m.services, m.orgPrefs, m.scope)
		}},
		GateSync: {newStep: func(m *Model) Step {
			return sync.New(m.theme, m.syncer, m.scope)
		}},
	}
}
