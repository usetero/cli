package onboarding

import (
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/tenancy"
)

// State is the onboarding truth projection used to derive the next step.
type State struct {
	Organizations        []tenancy.Organization
	SelectedOrganization *tenancy.Organization

	Accounts        []tenancy.Account
	SelectedAccount *tenancy.Account

	Workspaces        []tenancy.Workspace
	SelectedWorkspace *tenancy.Workspace

	DatadogAccount *integrations.DatadogAccount
	DatadogStatus  *integrations.DatadogStatus
	DatadogDraft   DatadogDraft

	PowerSyncReady bool
	NextStep       Step
}

func selectOrganization(values []tenancy.Organization, preferred tenancy.OrganizationID) *tenancy.Organization {
	if preferred != "" {
		for i := range values {
			if values[i].ID == preferred {
				return &values[i]
			}
		}
	}
	if len(values) == 1 {
		return &values[0]
	}
	return nil
}

func selectAccount(values []tenancy.Account, preferred tenancy.AccountID) *tenancy.Account {
	if preferred != "" {
		for i := range values {
			if values[i].ID == preferred {
				return &values[i]
			}
		}
	}
	if len(values) == 1 {
		return &values[0]
	}
	return nil
}

func selectWorkspace(values []tenancy.Workspace, preferred tenancy.WorkspaceID) *tenancy.Workspace {
	if preferred != "" {
		for i := range values {
			if values[i].ID == preferred {
				return &values[i]
			}
		}
	}
	if len(values) == 1 {
		return &values[0]
	}
	return nil
}
