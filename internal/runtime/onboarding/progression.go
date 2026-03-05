package onboarding

func nextStep(state State) Step {
	if state.SelectedOrganization == nil {
		return organizationStep(state.Organizations)
	}
	if state.SelectedAccount == nil {
		return accountStep(state.Accounts)
	}
	if state.SelectedWorkspace == nil {
		return StepWorkspaceSelect
	}
	if state.DatadogAccount == nil {
		if !state.DatadogDraft.Site.Valid() {
			return StepDatadogRegion
		}
		if !state.DatadogDraft.HasAPIKey {
			return StepDatadogAPIKey
		}
		return StepDatadogAppKey
	}
	if state.DatadogStatus == nil || !state.DatadogStatus.ReadyForUse {
		return StepDatadogDiscovery
	}
	if !state.PowerSyncReady {
		return StepPowerSyncReady
	}
	return StepDone
}

func organizationStep(orgs Organizations) Step {
	if len(orgs) == 0 {
		return StepOrganizationCreate
	}
	return StepOrganizationSelect
}

func accountStep(accounts Accounts) Step {
	if len(accounts) == 0 {
		return StepAccountCreate
	}
	return StepAccountSelect
}
