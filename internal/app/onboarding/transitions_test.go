package onboarding

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	iauth "github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/auth/authtest"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/preferences/preferencestest"
	"github.com/usetero/cli/internal/styles"
)

func TestHandleTransitionPreflightRouting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    bootstrap.PreflightState
		wantGate Gate
	}{
		{
			name: "unauthenticated routes to authenticate",
			state: bootstrap.PreflightState{
				HasValidAuth: false,
			},
			wantGate: bootstrap.GateAuthenticate,
		},
		{
			name: "missing role routes to role select",
			state: bootstrap.PreflightState{
				HasValidAuth: true,
			},
			wantGate: bootstrap.GateRoleSelect,
		},
		{
			name: "missing org routes to org select",
			state: bootstrap.PreflightState{
				HasValidAuth: true,
				Role:         bootstrap.RolePlatform,
			},
			wantGate: bootstrap.GateOrgSelect,
		},
		{
			name: "resolved org only routes to account select",
			state: bootstrap.PreflightState{
				HasValidAuth: true,
				Role:         bootstrap.RolePlatform,
				Org:          ptrOrg("org-1"),
			},
			wantGate: bootstrap.GateAccountSelect,
		},
		{
			name: "resolved account routes to runtime init",
			state: bootstrap.PreflightState{
				HasValidAuth: true,
				Role:         bootstrap.RolePlatform,
				Org:          ptrOrg("org-1"),
				Account:      ptrAccount("acc-1"),
			},
			wantGate: bootstrap.GateRuntimeInit,
		},
		{
			name: "failed preflight routes to authenticate even with valid auth",
			state: bootstrap.PreflightState{
				Outcome:      bootstrap.PreflightOutcomeFailed,
				HasValidAuth: true,
				Role:         bootstrap.RolePlatform,
				Org:          ptrOrg("org-1"),
				Account:      ptrAccount("acc-1"),
			},
			wantGate: bootstrap.GateAuthenticate,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newTestModel(t)
			cmd := m.handleTransition(bootstrap.PreflightResolved{State: tc.state})
			if cmd == nil {
				t.Fatalf("expected command for %s", tc.name)
			}
			if m.gate != tc.wantGate {
				t.Fatalf("gate = %s, want %s", m.gate, tc.wantGate)
			}
			if tc.wantGate == bootstrap.GateRuntimeInit {
				msg := cmd()
				if _, ok := msg.(bootstrap.EnsureRuntime); !ok {
					t.Fatalf("runtime init should emit EnsureRuntime, got %T", msg)
				}
			}
		})
	}
}

func TestHandleTransitionDatadogBranchRouting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		msg      any
		wantGate Gate
	}{
		{name: "datadog ready goes to workspace select", msg: bootstrap.DatadogReady{}, wantGate: bootstrap.GateWorkspaceSelect},
		{name: "datadog needed goes to region", msg: bootstrap.DatadogNeeded{}, wantGate: bootstrap.GateDatadogRegion},
		{name: "discovery complete goes to workspace select", msg: bootstrap.DatadogDiscoveryComplete{}, wantGate: bootstrap.GateWorkspaceSelect},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newTestModel(t)
			m.state.Org = ptrOrg("org-1")
			m.state.Account = ptrAccount("acc-1")
			_ = m.handleTransition(tc.msg)
			if m.gate != tc.wantGate {
				t.Fatalf("gate = %s, want %s", m.gate, tc.wantGate)
			}
		})
	}
}

func TestHandleTransitionWorkspaceSelectedSetsState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.state.Org = ptrOrg("org-1")
	m.state.Account = ptrAccount("acc-1")
	workspace := domain.Workspace{ID: "ws-1", Name: "Workspace 1"}

	if cmd := m.handleTransition(bootstrap.WorkspaceSelected{Workspace: workspace}); cmd == nil {
		t.Fatal("expected transition command")
	}
	if m.gate != bootstrap.GateSync {
		t.Fatalf("gate = %s, want %s", m.gate, bootstrap.GateSync)
	}
	if m.state.Workspace == nil || m.state.Workspace.ID != workspace.ID {
		t.Fatalf("workspace state not set correctly: %+v", m.state.Workspace)
	}
}

func TestHandleTransitionDatadogState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.state.Org = ptrOrg("org-1")
	m.state.Account = ptrAccount("acc-1")

	if _ = m.handleTransition(bootstrap.DatadogRegionSelected{Site: "US1"}); m.gate != bootstrap.GateDatadogAPIKey {
		t.Fatalf("expected datadog region to route to api key gate")
	}
	if m.state.DDSite != "US1" {
		t.Fatalf("ddSite = %s, want US1", m.state.DDSite)
	}

	if _ = m.handleTransition(bootstrap.DatadogAPIKeyEntered{APIKey: "api-key"}); m.gate != bootstrap.GateDatadogAppKey {
		t.Fatalf("expected api key entered to route to app key gate")
	}
	if m.state.DDAPIKey != "api-key" {
		t.Fatalf("ddAPIKey = %q, want %q", m.state.DDAPIKey, "api-key")
	}

	if _ = m.handleTransition(bootstrap.DatadogAccountCreated{DatadogAccountID: "dd-1"}); m.gate != bootstrap.GateDatadogDiscovery {
		t.Fatalf("expected account created to route to discovery gate")
	}
	if m.state.DDAccount != "dd-1" {
		t.Fatalf("ddAccount = %q, want %q", m.state.DDAccount, "dd-1")
	}
}

func TestHandleTransitionRuntimeReady(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	org := ptrOrg("org-1")
	account := ptrAccount("acc-1")

	if _ = m.handleTransition(bootstrap.RuntimeReady{Org: *org, Account: *account}); m.gate != bootstrap.GateDatadogCheck {
		t.Fatalf("expected runtime ready to route to datadog check")
	}
	if m.state.Org == nil || m.state.Org.ID != org.ID {
		t.Fatalf("org state not set from runtime ready")
	}
	if m.state.Account == nil || m.state.Account.ID != account.ID {
		t.Fatalf("account state not set from runtime ready")
	}
}

func TestHandleTransitionRuntimeReadySyncsServiceAccountScope(t *testing.T) {
	t.Parallel()

	m, client := newTestModelWithClient(t)
	var scopedAccountID domain.AccountID
	client.SetAccountIDFunc = func(accountID domain.AccountID) {
		scopedAccountID = accountID
	}

	org := ptrOrg("org-1")
	account := ptrAccount("acc-1")

	_ = m.handleTransition(bootstrap.RuntimeReady{Org: *org, Account: *account})
	if scopedAccountID != "acc-1" {
		t.Fatalf("scoped account id = %q, want %q", scopedAccountID, "acc-1")
	}
}

func TestHandleTransitionOrgSelectedClearsServiceAccountScope(t *testing.T) {
	t.Parallel()

	m, client := newTestModelWithClient(t)
	var scopedAccountID domain.AccountID
	client.SetAccountIDFunc = func(accountID domain.AccountID) {
		scopedAccountID = accountID
	}

	m.state.Account = ptrAccount("acc-1")
	_ = m.handleTransition(bootstrap.OrgSelected{Org: *ptrOrg("org-2")})
	if scopedAccountID != "" {
		t.Fatalf("scoped account id = %q, want empty", scopedAccountID)
	}
}

func TestHandleTransitionSyncComplete(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.state.User = ptrUser("user-1")
	m.state.Org = ptrOrg("org-1")
	m.state.Account = ptrAccount("acc-1")
	m.state.Workspace = ptrWorkspace("ws-1")

	cmd := m.handleTransition(bootstrap.SyncComplete{})
	if cmd == nil {
		t.Fatal("expected completion command")
	}
	msg := cmd()
	complete, ok := msg.(bootstrap.OnboardingComplete)
	if !ok {
		t.Fatalf("message type = %T, want bootstrap.OnboardingComplete", msg)
	}
	if complete.Org.ID != "org-1" || complete.Account.ID != "acc-1" || complete.Workspace.ID != "ws-1" || complete.User.ID != "user-1" {
		t.Fatalf("unexpected completion payload: %+v", complete)
	}
}

func TestHandleTransitionSyncCompleteMissingStateNoops(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.state.User = ptrUser("user-1")
	m.state.Org = ptrOrg("org-1")
	// Missing account/workspace should not panic or emit completion payload.

	cmd := m.handleTransition(bootstrap.SyncComplete{})
	if cmd != nil {
		t.Fatal("expected nil command when completion state is incomplete")
	}
}

func newTestModel(t *testing.T) *Model {
	t.Helper()
	m, _ := newTestModelWithClient(t)
	return m
}

func newTestModelWithClient(t *testing.T) (*Model, *apitest.MockClient) {
	t.Helper()

	scope := logtest.NewScope(t)
	client := apitest.NewMockClient()
	services := api.NewAPIServices(client, scope)
	userPrefs := preferencestest.NewMockUserPreferences()
	orgPrefs := preferencestest.NewMockOrgPreferences()
	authSvc := &authtest.MockAuth{}
	syncer := powersynctest.NewMockSyncer()

	m := New(context.Background(), styles.NewTheme(true), services, userPrefs, orgPrefs, authSvc, syncer, scope)
	m.SetSize(120, 40)
	return m, client
}

func ptrOrg(id string) *domain.Organization {
	return &domain.Organization{ID: domain.OrganizationID(id), Name: id}
}

func ptrAccount(id string) *domain.Account {
	return &domain.Account{ID: domain.AccountID(id), Name: id}
}

func ptrWorkspace(id string) *domain.Workspace {
	return &domain.Workspace{ID: domain.WorkspaceID(id), Name: id}
}

func ptrUser(id string) *iauth.User {
	return &iauth.User{ID: id}
}
