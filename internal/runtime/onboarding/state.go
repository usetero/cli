package onboarding

import (
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

// DatadogDraft stores in-progress integration data while onboarding is incomplete.
type DatadogDraft struct {
	Site      integrations.DatadogSite
	HasAPIKey bool
	apiKey    string
}

// State is the onboarding truth projection used to derive the next step.
type State struct {
	Role preferences.Role

	Organizations        Organizations
	SelectedOrganization *tenancy.Organization

	Accounts        Accounts
	SelectedAccount *tenancy.Account

	Workspaces        Workspaces
	SelectedWorkspace *tenancy.Workspace

	DatadogAccount *integrations.DatadogAccount
	DatadogStatus  *integrations.DatadogStatus
	DatadogDraft   DatadogDraft

	PowerSyncReady bool
	NextStep       Step
}

type Organizations []tenancy.Organization

type Accounts []tenancy.Account

type Workspaces []tenancy.Workspace

func (o Organizations) Select(preferred tenancy.OrganizationID) *tenancy.Organization {
	if preferred != "" {
		for i := range o {
			if o[i].ID == preferred {
				return &o[i]
			}
		}
	}
	if len(o) == 1 {
		return &o[0]
	}
	return nil
}

func (a Accounts) Select(preferred tenancy.AccountID) *tenancy.Account {
	if preferred != "" {
		for i := range a {
			if a[i].ID == preferred {
				return &a[i]
			}
		}
	}
	if len(a) == 1 {
		return &a[0]
	}
	return nil
}

func (w Workspaces) Select(preferred tenancy.WorkspaceID) *tenancy.Workspace {
	if preferred != "" {
		for i := range w {
			if w[i].ID == preferred {
				return &w[i]
			}
		}
	}
	if len(w) == 1 {
		return &w[0]
	}
	return nil
}
