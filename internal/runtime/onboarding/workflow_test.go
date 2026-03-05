package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

type behaviorPrefs struct {
	snapshot preferences.Snapshot
}

func (f *behaviorPrefs) Snapshot(context.Context) (preferences.Snapshot, error) {
	return f.snapshot, nil
}
func (f *behaviorPrefs) SetRole(_ context.Context, role preferences.Role) error {
	f.snapshot.Role = role
	return nil
}
func (f *behaviorPrefs) SetOrganization(_ context.Context, orgID tenancy.OrganizationID) error {
	f.snapshot.Organization = orgID
	f.snapshot.Account = ""
	f.snapshot.Workspace = ""
	return nil
}
func (f *behaviorPrefs) SetAccount(_ context.Context, accountID tenancy.AccountID) error {
	f.snapshot.Account = accountID
	f.snapshot.Workspace = ""
	return nil
}
func (f *behaviorPrefs) SetWorkspace(_ context.Context, workspaceID tenancy.WorkspaceID) error {
	f.snapshot.Workspace = workspaceID
	return nil
}
func (f *behaviorPrefs) ClearScope(_ context.Context) error {
	f.snapshot.Organization = ""
	f.snapshot.Account = ""
	f.snapshot.Workspace = ""
	return nil
}

type behaviorOrgs struct {
	list      []tenancy.Organization
	bootstrap tenancy.OrganizationBootstrap
}

func (f *behaviorOrgs) List(context.Context) ([]tenancy.Organization, error) { return f.list, nil }
func (f *behaviorOrgs) Create(context.Context, string) (tenancy.OrganizationBootstrap, error) {
	return f.bootstrap, nil
}

type behaviorAccounts struct {
	list []tenancy.Account
	next tenancy.AccountID
}

func (f *behaviorAccounts) Create(context.Context, string) (tenancy.AccountID, error) {
	if f.next == "" {
		f.next = "acct_new"
	}
	return f.next, nil
}
func (f *behaviorAccounts) Delete(context.Context, tenancy.AccountID) error { return nil }
func (f *behaviorAccounts) List(context.Context) ([]tenancy.Account, error) { return f.list, nil }

type behaviorWorkspaces struct {
	list map[tenancy.AccountID][]tenancy.Workspace
}

func (f *behaviorWorkspaces) Create(context.Context, tenancy.AccountID, string) (tenancy.WorkspaceID, error) {
	return "", nil
}
func (f *behaviorWorkspaces) Delete(context.Context, tenancy.WorkspaceID) error { return nil }
func (f *behaviorWorkspaces) ListByAccount(_ context.Context, accountID tenancy.AccountID) ([]tenancy.Workspace, error) {
	return f.list[accountID], nil
}

type behaviorDatadog struct {
	byAccount map[tenancy.AccountID]*integrations.DatadogAccount
	status    map[integrations.DatadogAccountID]*integrations.DatadogStatus

	validateOK  bool
	validateMsg string
	validateErr error
}

func (f *behaviorDatadog) GetByAccount(_ context.Context, accountID tenancy.AccountID) (*integrations.DatadogAccount, error) {
	return f.byAccount[accountID], nil
}
func (f *behaviorDatadog) ValidateAPIKey(context.Context, integrations.DatadogSite, string) (bool, string, error) {
	return f.validateOK, f.validateMsg, f.validateErr
}
func (f *behaviorDatadog) Create(context.Context, integrations.CreateDatadogAccountInput) (integrations.DatadogAccountID, error) {
	return "dd_1", nil
}
func (f *behaviorDatadog) Status(_ context.Context, datadogAccountID integrations.DatadogAccountID) (*integrations.DatadogStatus, error) {
	return f.status[datadogAccountID], nil
}

type behaviorReady struct{ ready bool }

func (r behaviorReady) Ready(context.Context) (bool, error) { return r.ready, nil }

func TestCreateOrganization_AppliesBootstrapScope(t *testing.T) {
	t.Parallel()

	prefs := &behaviorPrefs{snapshot: preferences.Snapshot{Role: preferences.RoleEngineer}}
	orgs := &behaviorOrgs{
		list: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}},
		bootstrap: tenancy.OrganizationBootstrap{
			Organization: tenancy.Organization{ID: "org_1", Name: "Org 1"},
			Account:      tenancy.Account{ID: "acct_1", Name: "Account 1"},
			Workspace:    tenancy.Workspace{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
		},
	}

	svc, err := NewService(
		prefs,
		orgs,
		func(orgID tenancy.OrganizationID) tenancy.AccountService {
			if orgID == "org_1" {
				return &behaviorAccounts{list: []tenancy.Account{{ID: "acct_1", Name: "Account 1"}}}
			}
			return &behaviorAccounts{}
		},
		&behaviorWorkspaces{
			list: map[tenancy.AccountID][]tenancy.Workspace{
				"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
			},
		},
		&behaviorDatadog{byAccount: map[tenancy.AccountID]*integrations.DatadogAccount{}, status: map[integrations.DatadogAccountID]*integrations.DatadogStatus{}, validateOK: true},
		behaviorReady{ready: false},
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	state, err := svc.CreateOrganization(context.Background(), "Org 1")
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	if prefs.snapshot.Organization != "org_1" || prefs.snapshot.Account != "acct_1" || prefs.snapshot.Workspace != "ws_1" {
		t.Fatalf("bootstrap scope not persisted: %+v", prefs.snapshot)
	}
	if state.SelectedOrganization == nil || state.SelectedOrganization.ID != "org_1" {
		t.Fatalf("expected selected organization org_1, got %+v", state.SelectedOrganization)
	}
	if state.SelectedAccount == nil || state.SelectedAccount.ID != "acct_1" {
		t.Fatalf("expected selected account acct_1, got %+v", state.SelectedAccount)
	}
	if state.SelectedWorkspace == nil || state.SelectedWorkspace.ID != "ws_1" {
		t.Fatalf("expected selected workspace ws_1, got %+v", state.SelectedWorkspace)
	}
	if state.NextStep != StepDatadogRegion {
		t.Fatalf("expected datadog region next, got %q", state.NextStep)
	}
}

func TestCreateAccount_RequiresSelectedOrganization(t *testing.T) {
	t.Parallel()

	svc, err := NewService(
		&behaviorPrefs{snapshot: preferences.Snapshot{Role: preferences.RolePlatform}},
		&behaviorOrgs{list: []tenancy.Organization{}},
		func(tenancy.OrganizationID) tenancy.AccountService { return &behaviorAccounts{} },
		&behaviorWorkspaces{list: map[tenancy.AccountID][]tenancy.Workspace{}},
		&behaviorDatadog{byAccount: map[tenancy.AccountID]*integrations.DatadogAccount{}, status: map[integrations.DatadogAccountID]*integrations.DatadogStatus{}, validateOK: true},
		behaviorReady{ready: true},
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = svc.CreateAccount(context.Background(), "A")
	if err == nil || !strings.Contains(err.Error(), "organization must be selected") {
		t.Fatalf("expected selected-organization guard error, got %v", err)
	}
}

func TestSubmitDatadogAPIKey_UsesValidationMessage(t *testing.T) {
	t.Parallel()

	svc, err := NewService(
		&behaviorPrefs{snapshot: preferences.Snapshot{
			Role:         preferences.RolePlatform,
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		}},
		&behaviorOrgs{list: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}}},
		func(tenancy.OrganizationID) tenancy.AccountService {
			return &behaviorAccounts{list: []tenancy.Account{{ID: "acct_1", Name: "Account 1"}}}
		},
		&behaviorWorkspaces{list: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		}},
		&behaviorDatadog{
			byAccount:   map[tenancy.AccountID]*integrations.DatadogAccount{},
			status:      map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
			validateOK:  false,
			validateMsg: "datadog rejected key",
		},
		behaviorReady{ready: false},
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if _, err := svc.SetDatadogSite(context.Background(), integrations.DatadogSiteUS1); err != nil {
		t.Fatalf("set site: %v", err)
	}
	_, err = svc.SubmitDatadogAPIKey(context.Background(), "bad")
	if err == nil || !strings.Contains(err.Error(), "datadog rejected key") {
		t.Fatalf("expected validation message error, got %v", err)
	}
}

func TestSubmitDatadogAppKey_RequiresValidatedAPIKey(t *testing.T) {
	t.Parallel()

	svc, err := NewService(
		&behaviorPrefs{snapshot: preferences.Snapshot{
			Role:         preferences.RoleEngineer,
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		}},
		&behaviorOrgs{list: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}}},
		func(tenancy.OrganizationID) tenancy.AccountService {
			return &behaviorAccounts{list: []tenancy.Account{{ID: "acct_1", Name: "Account 1"}}}
		},
		&behaviorWorkspaces{list: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"}},
		}},
		&behaviorDatadog{byAccount: map[tenancy.AccountID]*integrations.DatadogAccount{}, status: map[integrations.DatadogAccountID]*integrations.DatadogStatus{}, validateOK: true},
		behaviorReady{ready: false},
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if _, err := svc.SetDatadogSite(context.Background(), integrations.DatadogSiteUS1); err != nil {
		t.Fatalf("set site: %v", err)
	}
	_, err = svc.SubmitDatadogAppKey(context.Background(), "DD", "app-key")
	if err == nil || !strings.Contains(err.Error(), "api key must be validated first") {
		t.Fatalf("expected api-key guard error, got %v", err)
	}
}

func TestDatadogDraft_ResetsAcrossAccountSwitch(t *testing.T) {
	t.Parallel()

	prefs := &behaviorPrefs{snapshot: preferences.Snapshot{
		Role:         preferences.RolePlatform,
		Organization: "org_1",
		Account:      "acct_1",
		Workspace:    "ws_1",
	}}
	dd := &behaviorDatadog{
		byAccount:   map[tenancy.AccountID]*integrations.DatadogAccount{},
		status:      map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
		validateOK:  false,
		validateMsg: "bad",
	}
	svc, err := NewService(
		prefs,
		&behaviorOrgs{list: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}}},
		func(tenancy.OrganizationID) tenancy.AccountService {
			return &behaviorAccounts{list: []tenancy.Account{
				{ID: "acct_1", Name: "A1"},
				{ID: "acct_2", Name: "A2"},
			}}
		},
		&behaviorWorkspaces{list: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "W1"}},
			"acct_2": {{ID: "ws_2", AccountID: "acct_2", Name: "W2"}},
		}},
		dd,
		behaviorReady{ready: false},
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if _, err := svc.SetDatadogSite(context.Background(), integrations.DatadogSiteUS1); err != nil {
		t.Fatalf("set site: %v", err)
	}
	// Force draft HasAPIKey=true so we can validate reset behavior.
	if _, err := svc.SubmitDatadogAPIKey(context.Background(), "k"); err == nil {
		// expected validation error, draft should remain unvalidated
	}
	svc.setDraft(func(d *DatadogDraft) {
		d.HasAPIKey = true
	})

	if _, err := svc.SelectAccount(context.Background(), "acct_2"); err != nil {
		t.Fatalf("select account: %v", err)
	}
	state, err := svc.State(context.Background())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.DatadogDraft.Site.Valid() || state.DatadogDraft.HasAPIKey {
		t.Fatalf("expected draft reset after account switch, got %+v", state.DatadogDraft)
	}
}

func TestSubmitDatadogAPIKey_PropagatesValidationErrors(t *testing.T) {
	t.Parallel()

	svc, err := NewService(
		&behaviorPrefs{snapshot: preferences.Snapshot{
			Role:         preferences.RolePlatform,
			Organization: "org_1",
			Account:      "acct_1",
			Workspace:    "ws_1",
		}},
		&behaviorOrgs{list: []tenancy.Organization{{ID: "org_1", Name: "Org 1"}}},
		func(tenancy.OrganizationID) tenancy.AccountService {
			return &behaviorAccounts{list: []tenancy.Account{{ID: "acct_1", Name: "A1"}}}
		},
		&behaviorWorkspaces{list: map[tenancy.AccountID][]tenancy.Workspace{
			"acct_1": {{ID: "ws_1", AccountID: "acct_1", Name: "W1"}},
		}},
		&behaviorDatadog{
			byAccount:   map[tenancy.AccountID]*integrations.DatadogAccount{},
			status:      map[integrations.DatadogAccountID]*integrations.DatadogStatus{},
			validateOK:  false,
			validateErr: errors.New("datadog unavailable"),
		},
		behaviorReady{ready: false},
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if _, err := svc.SetDatadogSite(context.Background(), integrations.DatadogSiteUS1); err != nil {
		t.Fatalf("set site: %v", err)
	}
	_, err = svc.SubmitDatadogAPIKey(context.Background(), "k")
	if err == nil || !strings.Contains(err.Error(), "datadog unavailable") {
		t.Fatalf("expected validation transport error, got %v", err)
	}
}
