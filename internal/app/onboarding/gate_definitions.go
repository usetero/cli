package onboarding

import (
	"fmt"

	"github.com/usetero/cli/internal/app/onboarding/accounts"
	"github.com/usetero/cli/internal/app/onboarding/auth"
	"github.com/usetero/cli/internal/app/onboarding/datadog"
	"github.com/usetero/cli/internal/app/onboarding/organizations"
	"github.com/usetero/cli/internal/app/onboarding/preflight"
	"github.com/usetero/cli/internal/app/onboarding/role"
	"github.com/usetero/cli/internal/app/onboarding/runtimeinit"
	"github.com/usetero/cli/internal/app/onboarding/workspaces"
	"github.com/usetero/cli/internal/core/bootstrap"
)

func (m *Model) newStepForGate(gate Gate) (Step, error) {
	if m.gateBuildHook != nil {
		if err := m.gateBuildHook(gate); err != nil {
			return nil, err
		}
	}

	if err := m.validateGateState(gate); err != nil {
		return nil, err
	}

	switch gate {
	case bootstrap.GatePreflight:
		return preflight.New(m.ctx, m.theme, m.services, m.auth, m.userPrefs, m.orgPrefs, m.scope), nil
	case bootstrap.GateAuthenticate:
		return auth.NewAuthenticate(m.ctx, m.theme, m.auth, m.scope), nil
	case bootstrap.GateRoleSelect:
		return role.New(m.theme, m.userPrefs, m.scope), nil
	case bootstrap.GateOrgSelect:
		return organizations.NewSelect(m.ctx, m.theme, m.services, m.userPrefs, m.auth, m.scope), nil
	case bootstrap.GateOrgCreate:
		return organizations.NewCreate(m.ctx, m.theme, m.services, m.userPrefs, m.scope), nil
	case bootstrap.GateAccountSelect:
		return accounts.NewSelect(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope), nil
	case bootstrap.GateAccountCreate:
		return accounts.NewCreate(m.ctx, m.theme, *m.state.Org, m.services, m.orgPrefs, m.scope), nil
	case bootstrap.GateRuntimeInit:
		return runtimeinit.New(m.theme, *m.state.Org, *m.state.Account, m.scope), nil
	case bootstrap.GateDatadogCheck:
		return datadog.NewCheck(m.ctx, m.theme, *m.state.Account, m.services, m.scope), nil
	case bootstrap.GateDatadogRegion:
		return datadog.NewRegion(m.theme, m.scope), nil
	case bootstrap.GateDatadogAPIKey:
		return datadog.NewAPIKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.services, m.scope), nil
	case bootstrap.GateDatadogAppKey:
		return datadog.NewAppKey(m.ctx, m.theme, *m.state.Account, m.state.DDSite, m.state.DDAPIKey, m.services, m.scope), nil
	case bootstrap.GateDatadogDiscovery:
		return datadog.NewDiscovery(m.ctx, m.theme, m.state.DDAccount, m.services, m.scope), nil
	case bootstrap.GateWorkspaceSelect:
		return workspaces.NewSelect(m.ctx, m.theme, *m.state.Account, m.services, m.orgPrefs, m.scope), nil
	default:
		return nil, fmt.Errorf("unsupported gate %q", gate)
	}
}

func (m *Model) validateGateState(gate Gate) error {
	req := bootstrap.RequirementForGate(gate)
	switch {
	case req.NeedsOrg && m.state.Org == nil:
		return fmt.Errorf("gate %q requires org", gate)
	case req.NeedsAccount && m.state.Account == nil:
		return fmt.Errorf("gate %q requires account", gate)
	case req.NeedsWorkspace && m.state.Workspace == nil:
		return fmt.Errorf("gate %q requires workspace", gate)
	case req.NeedsDDSite && m.state.DDSite == "":
		return fmt.Errorf("gate %q requires datadog site", gate)
	case req.NeedsDDAPIKey && m.state.DDAPIKey == "":
		return fmt.Errorf("gate %q requires datadog api key", gate)
	case req.NeedsDDAccount && m.state.DDAccount == "":
		return fmt.Errorf("gate %q requires datadog account", gate)
	default:
		return nil
	}
}
