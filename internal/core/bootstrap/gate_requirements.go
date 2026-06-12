package bootstrap

// RequirementForGate returns rewind requirements for a gate.
func RequirementForGate(gate Gate) GateRequirement {
	switch gate {
	case GateAccountSelect, GateAccountCreate:
		return GateRequirement{NeedsOrg: true}
	case GateRuntimeInit, GateDatadogCheck:
		return GateRequirement{NeedsOrg: true, NeedsAccount: true}
	case GateDatadogAPIKey:
		return GateRequirement{NeedsOrg: true, NeedsAccount: true, NeedsDDSite: true}
	case GateDatadogAppKey:
		return GateRequirement{NeedsOrg: true, NeedsAccount: true, NeedsDDSite: true, NeedsDDAPIKey: true}
	case GateDatadogDiscovery:
		return GateRequirement{NeedsOrg: true, NeedsAccount: true, NeedsDDAccount: true}
	default:
		return GateRequirement{}
	}
}
