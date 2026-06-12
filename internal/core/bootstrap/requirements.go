package bootstrap

// GateRequirement encodes prerequisites for entering a gate.
type GateRequirement struct {
	NeedsOrg       bool
	NeedsAccount   bool
	NeedsDDSite    bool
	NeedsDDAPIKey  bool
	NeedsDDAccount bool
}

// RewindGate returns the earliest gate that satisfies the missing requirement.
// If all requirements are met, it returns target.
func RewindGate(target Gate, req GateRequirement, state State) Gate {
	if req.NeedsOrg && state.Org == nil {
		return GateOrgSelect
	}
	if req.NeedsAccount && state.Account == nil {
		return GateAccountSelect
	}
	if req.NeedsDDSite && state.DDSite == "" {
		return GateDatadogRegion
	}
	if req.NeedsDDAPIKey && state.DDAPIKey == "" {
		return GateDatadogAPIKey
	}
	if req.NeedsDDAccount && state.DDAccount == "" {
		return GateDatadogCheck
	}
	return target
}
