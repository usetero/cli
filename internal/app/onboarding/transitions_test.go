package onboarding

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	iauth "github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/auth/authtest"
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
		state    msgs.PreflightState
		wantGate Gate
	}{
		{
			name: "unauthenticated routes to authenticate",
			state: msgs.PreflightState{
				HasValidAuth: false,
			},
			wantGate: GateAuthenticate,
		},
		{
			name: "missing role routes to role select",
			state: msgs.PreflightState{
				HasValidAuth: true,
			},
			wantGate: GateRoleSelect,
		},
		{
			name: "missing org routes to org select",
			state: msgs.PreflightState{
				HasValidAuth: true,
				Role:         msgs.RolePlatform,
			},
			wantGate: GateOrgSelect,
		},
		{
			name: "resolved org only routes to account select",
			state: msgs.PreflightState{
				HasValidAuth: true,
				Role:         msgs.RolePlatform,
				Org:          ptrOrg("org-1"),
			},
			wantGate: GateAccountSelect,
		},
		{
			name: "resolved account routes to runtime init",
			state: msgs.PreflightState{
				HasValidAuth: true,
				Role:         msgs.RolePlatform,
				Org:          ptrOrg("org-1"),
				Account:      ptrAccount("acc-1"),
			},
			wantGate: GateRuntimeInit,
		},
		{
			name: "failed preflight routes to authenticate even with valid auth",
			state: msgs.PreflightState{
				Outcome:      msgs.PreflightOutcomeFailed,
				HasValidAuth: true,
				Role:         msgs.RolePlatform,
				Org:          ptrOrg("org-1"),
				Account:      ptrAccount("acc-1"),
			},
			wantGate: GateAuthenticate,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newTestModel(t)
			cmd := m.handleTransition(msgs.PreflightResolved{State: tc.state})
			if cmd == nil {
				t.Fatalf("expected command for %s", tc.name)
			}
			if m.gate != tc.wantGate {
				t.Fatalf("gate = %s, want %s", m.gate, tc.wantGate)
			}
			if tc.wantGate == GateRuntimeInit {
				msg := cmd()
				if _, ok := msg.(msgs.EnsureRuntime); !ok {
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
		{name: "datadog ready goes to workspace select", msg: msgs.DatadogReady{}, wantGate: GateWorkspaceSelect},
		{name: "datadog needed goes to region", msg: msgs.DatadogNeeded{}, wantGate: GateDatadogRegion},
		{name: "discovery complete goes to workspace select", msg: msgs.DatadogDiscoveryComplete{}, wantGate: GateWorkspaceSelect},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newTestModel(t)
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
	workspace := domain.Workspace{ID: "ws-1", Name: "Workspace 1"}

	if cmd := m.handleTransition(msgs.WorkspaceSelected{Workspace: workspace}); cmd == nil {
		t.Fatal("expected transition command")
	}
	if m.gate != GateSync {
		t.Fatalf("gate = %s, want %s", m.gate, GateSync)
	}
	if m.state.workspace == nil || m.state.workspace.ID != workspace.ID {
		t.Fatalf("workspace state not set correctly: %+v", m.state.workspace)
	}
}

func TestHandleTransitionDatadogState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.state.account = ptrAccount("acc-1")

	if _ = m.handleTransition(msgs.DatadogRegionSelected{Site: "US1"}); m.gate != GateDatadogAPIKey {
		t.Fatalf("expected datadog region to route to api key gate")
	}
	if m.state.ddSite != "US1" {
		t.Fatalf("ddSite = %s, want US1", m.state.ddSite)
	}

	if _ = m.handleTransition(msgs.DatadogAPIKeyEntered{APIKey: "api-key"}); m.gate != GateDatadogAppKey {
		t.Fatalf("expected api key entered to route to app key gate")
	}
	if m.state.ddAPIKey != "api-key" {
		t.Fatalf("ddAPIKey = %q, want %q", m.state.ddAPIKey, "api-key")
	}

	if _ = m.handleTransition(msgs.DatadogAccountCreated{DatadogAccountID: "dd-1"}); m.gate != GateDatadogDiscovery {
		t.Fatalf("expected account created to route to discovery gate")
	}
	if m.state.ddAccount != "dd-1" {
		t.Fatalf("ddAccount = %q, want %q", m.state.ddAccount, "dd-1")
	}
}

func TestHandleTransitionRuntimeReady(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	org := ptrOrg("org-1")
	account := ptrAccount("acc-1")

	if _ = m.handleTransition(msgs.RuntimeReady{Org: *org, Account: *account}); m.gate != GateDatadogCheck {
		t.Fatalf("expected runtime ready to route to datadog check")
	}
	if m.state.org == nil || m.state.org.ID != org.ID {
		t.Fatalf("org state not set from runtime ready")
	}
	if m.state.account == nil || m.state.account.ID != account.ID {
		t.Fatalf("account state not set from runtime ready")
	}
}

func TestHandleTransitionSyncComplete(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.state.user = ptrUser("user-1")
	m.state.org = ptrOrg("org-1")
	m.state.account = ptrAccount("acc-1")
	m.state.workspace = ptrWorkspace("ws-1")

	cmd := m.handleTransition(msgs.SyncComplete{})
	if cmd == nil {
		t.Fatal("expected completion command")
	}
	msg := cmd()
	complete, ok := msg.(msgs.OnboardingComplete)
	if !ok {
		t.Fatalf("message type = %T, want msgs.OnboardingComplete", msg)
	}
	if complete.Org.ID != "org-1" || complete.Account.ID != "acc-1" || complete.Workspace.ID != "ws-1" || complete.User.ID != "user-1" {
		t.Fatalf("unexpected completion payload: %+v", complete)
	}
}

func newTestModel(t *testing.T) *Model {
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
	return m
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
