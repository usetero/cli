package onboarding

import (
	"testing"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

func TestNextStep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state State
		want  Step
	}{
		{
			name:  "role first",
			state: State{},
			want:  StepRoleSelect,
		},
		{
			name: "organization create when none exist",
			state: State{
				Role: preferences.RoleEngineer,
			},
			want: StepOrganizationCreate,
		},
		{
			name: "organization select when choices exist",
			state: State{
				Role:          preferences.RoleEngineer,
				Organizations: []tenancy.Organization{{ID: "org_1"}},
			},
			want: StepOrganizationSelect,
		},
		{
			name: "account create when org selected and no accounts",
			state: State{
				Role:                 preferences.RoleEngineer,
				SelectedOrganization: &tenancy.Organization{ID: "org_1"},
			},
			want: StepAccountCreate,
		},
		{
			name: "account select when choices exist",
			state: State{
				Role:                 preferences.RoleEngineer,
				SelectedOrganization: &tenancy.Organization{ID: "org_1"},
				Accounts:             []tenancy.Account{{ID: "acct_1"}},
			},
			want: StepAccountSelect,
		},
		{
			name: "workspace select after account",
			state: State{
				Role:                 preferences.RoleEngineer,
				SelectedOrganization: &tenancy.Organization{ID: "org_1"},
				SelectedAccount:      &tenancy.Account{ID: "acct_1"},
			},
			want: StepWorkspaceSelect,
		},
		{
			name: "datadog region before site chosen",
			state: State{
				Role:                 preferences.RoleEngineer,
				SelectedOrganization: &tenancy.Organization{ID: "org_1"},
				SelectedAccount:      &tenancy.Account{ID: "acct_1"},
				SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1"},
			},
			want: StepDatadogRegion,
		},
		{
			name: "datadog api key after site",
			state: State{
				Role:                 preferences.RoleEngineer,
				SelectedOrganization: &tenancy.Organization{ID: "org_1"},
				SelectedAccount:      &tenancy.Account{ID: "acct_1"},
				SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1"},
				DatadogDraft:         DatadogDraft{Site: integrations.DatadogSiteUS1},
			},
			want: StepDatadogAPIKey,
		},
		{
			name: "datadog app key after validated api key",
			state: State{
				Role:                 preferences.RoleEngineer,
				SelectedOrganization: &tenancy.Organization{ID: "org_1"},
				SelectedAccount:      &tenancy.Account{ID: "acct_1"},
				SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1"},
				DatadogDraft:         DatadogDraft{Site: integrations.DatadogSiteUS1, HasAPIKey: true},
			},
			want: StepDatadogAppKey,
		},
		{
			name: "datadog discovery until ready",
			state: State{
				Role:                 preferences.RoleEngineer,
				SelectedOrganization: &tenancy.Organization{ID: "org_1"},
				SelectedAccount:      &tenancy.Account{ID: "acct_1"},
				SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1"},
				DatadogAccount:       &integrations.DatadogAccount{ID: "dd_1"},
				DatadogStatus:        &integrations.DatadogStatus{ReadyForUse: false},
			},
			want: StepDatadogDiscovery,
		},
		{
			name: "powersync ready after datadog ready",
			state: State{
				Role:                 preferences.RoleEngineer,
				SelectedOrganization: &tenancy.Organization{ID: "org_1"},
				SelectedAccount:      &tenancy.Account{ID: "acct_1"},
				SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1"},
				DatadogAccount:       &integrations.DatadogAccount{ID: "dd_1"},
				DatadogStatus:        &integrations.DatadogStatus{ReadyForUse: true},
			},
			want: StepPowerSyncReady,
		},
		{
			name: "done when everything is ready",
			state: State{
				Role:                 preferences.RoleEngineer,
				SelectedOrganization: &tenancy.Organization{ID: "org_1"},
				SelectedAccount:      &tenancy.Account{ID: "acct_1"},
				SelectedWorkspace:    &tenancy.Workspace{ID: "ws_1"},
				DatadogAccount:       &integrations.DatadogAccount{ID: "dd_1"},
				DatadogStatus:        &integrations.DatadogStatus{ReadyForUse: true},
				PowerSyncReady:       true,
			},
			want: StepDone,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := nextStep(tc.state); got != tc.want {
				t.Fatalf("nextStep() = %q, want %q", got, tc.want)
			}
		})
	}
}
