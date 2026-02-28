package onboarding

type gateRequirement struct {
	needsOrg       bool
	needsAccount   bool
	needsDDSite    bool
	needsDDAPIKey  bool
	needsDDAccount bool
	needsWorkspace bool
}

func (m *Model) rewindGateFor(target Gate) Gate {
	def, ok := m.definitions[target]
	if !ok {
		return target
	}
	req := def.requirement

	if req.needsOrg && m.state.org == nil {
		return GateOrgSelect
	}
	if req.needsAccount && m.state.account == nil {
		return GateAccountSelect
	}
	if req.needsDDSite && m.state.ddSite == "" {
		return GateDatadogRegion
	}
	if req.needsDDAPIKey && m.state.ddAPIKey == "" {
		return GateDatadogAPIKey
	}
	if req.needsDDAccount && m.state.ddAccount == "" {
		return GateDatadogCheck
	}
	if req.needsWorkspace && m.state.workspace == nil {
		return GateWorkspaceSelect
	}
	return target
}
