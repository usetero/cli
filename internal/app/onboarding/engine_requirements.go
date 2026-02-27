package onboarding

type gateRequirement struct {
	needsOrg       bool
	needsAccount   bool
	needsDDSite    bool
	needsDDAPIKey  bool
	needsDDAccount bool
	needsWorkspace bool
}

var gateRequirements = map[Gate]gateRequirement{
	GateAccountSelect:    {needsOrg: true},
	GateAccountCreate:    {needsOrg: true},
	GateDatadogCheck:     {needsOrg: true, needsAccount: true},
	GateDatadogAPIKey:    {needsOrg: true, needsAccount: true, needsDDSite: true},
	GateDatadogAppKey:    {needsOrg: true, needsAccount: true, needsDDSite: true, needsDDAPIKey: true},
	GateDatadogDiscovery: {needsOrg: true, needsAccount: true, needsDDAccount: true},
	GateWorkspaceSelect:  {needsOrg: true, needsAccount: true},
	GateSync:             {needsOrg: true, needsAccount: true, needsWorkspace: true},
}

func rewindGateFor(target Gate, state engineState) Gate {
	req, ok := gateRequirements[target]
	if !ok {
		return target
	}

	if req.needsOrg && state.org == nil {
		return GateOrgSelect
	}
	if req.needsAccount && state.account == nil {
		return GateAccountSelect
	}
	if req.needsDDSite && state.ddSite == "" {
		return GateDatadogRegion
	}
	if req.needsDDAPIKey && state.ddAPIKey == "" {
		return GateDatadogAPIKey
	}
	if req.needsDDAccount && state.ddAccount == "" {
		return GateDatadogCheck
	}
	if req.needsWorkspace && state.workspace == nil {
		return GateWorkspaceSelect
	}
	return target
}
