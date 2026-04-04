package onboardingtest

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/integrations/integrationstest"
	domainprefs "github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/domains/tenancy/tenancytest"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncertest"
	prefstoretest "github.com/usetero/cli/internal/infrastructure/preferences/preferencestest"
	runtimeonboarding "github.com/usetero/cli/internal/runtime/onboarding"
)

type Config struct {
	Snapshot         domainprefs.Snapshot
	Organizations    []tenancy.Organization
	Accounts         map[tenancy.OrganizationID][]tenancy.Account
	Workspaces       map[tenancy.AccountID][]tenancy.Workspace
	DatadogByAccount map[tenancy.AccountID]*integrations.DatadogAccount
	DatadogStatus    map[integrations.DatadogAccountID]*integrations.DatadogStatus
	Ready            bool
	Bootstrap        tenancy.OrganizationBootstrap
}

type Harness struct {
	Workflow        *runtimeonboarding.Workflow
	PreferenceStore *prefstoretest.Store
	Datadog         *integrationstest.MockDatadogService
}

func NewHarness(t testing.TB, cfg Config) *Harness {
	t.Helper()

	if cfg.Accounts == nil {
		cfg.Accounts = map[tenancy.OrganizationID][]tenancy.Account{}
	}
	if cfg.Workspaces == nil {
		cfg.Workspaces = map[tenancy.AccountID][]tenancy.Workspace{}
	}
	if cfg.DatadogByAccount == nil {
		cfg.DatadogByAccount = map[tenancy.AccountID]*integrations.DatadogAccount{}
	}
	if cfg.DatadogStatus == nil {
		cfg.DatadogStatus = map[integrations.DatadogAccountID]*integrations.DatadogStatus{}
	}
	if cfg.Bootstrap.Organization.ID == "" {
		cfg.Bootstrap = tenancy.OrganizationBootstrap{
			Organization: tenancy.Organization{ID: "org_1", Name: "Org 1"},
			Account:      tenancy.Account{ID: "acct_1", Name: "Account 1"},
			Workspace:    tenancy.Workspace{ID: "ws_1", AccountID: "acct_1", Name: "Workspace 1"},
		}
	}

	store := prefstoretest.NewStore()
	store.Snapshot = cfg.Snapshot
	prefs := domainprefs.NewService(store)

	datadog := &integrationstest.MockDatadogService{
		GetFn: func(context.Context) (*integrations.DatadogAccount, error) {
			return nil, nil
		},
		ValidateAPIKeyFn: func(context.Context, integrations.DatadogAPIKeyValidation) (bool, string, error) {
			return true, "", nil
		},
		StatusFn: func(_ context.Context, accountID integrations.DatadogAccountID) (*integrations.DatadogStatus, error) {
			return cfg.DatadogStatus[accountID], nil
		},
	}

	workflow := runtimeonboarding.NewWorkflow(
		prefs,
		&tenancytest.MockOrganizationService{
			ListFn: func(context.Context) ([]tenancy.Organization, error) { return cfg.Organizations, nil },
			CreateFn: func(context.Context, tenancy.OrganizationCreate) (tenancy.OrganizationBootstrap, error) {
				return cfg.Bootstrap, nil
			},
		},
		func(orgID tenancy.OrganizationID) tenancy.AccountService {
			return &tenancytest.MockAccountService{
				ListFn: func(context.Context) ([]tenancy.Account, error) { return cfg.Accounts[orgID], nil },
				CreateFn: func(context.Context, tenancy.AccountCreate) (tenancy.AccountID, error) {
					return "acct_new", nil
				},
			}
		},
		func(accountID tenancy.AccountID) tenancy.WorkspaceService {
			return &tenancytest.MockWorkspaceService{
				ListFn: func(context.Context) ([]tenancy.Workspace, error) {
					return cfg.Workspaces[accountID], nil
				},
			}
		},
		func(accountID tenancy.AccountID) integrations.DatadogService {
			datadog.GetFn = func(context.Context) (*integrations.DatadogAccount, error) {
				return cfg.DatadogByAccount[accountID], nil
			}
			datadog.CreateFn = func(_ context.Context, input integrations.DatadogAccountCreate) (integrations.DatadogAccountID, error) {
				id := integrations.DatadogAccountID("dd_1")
				cfg.DatadogByAccount[accountID] = &integrations.DatadogAccount{
					ID:   id,
					Name: input.Name.String(),
					Site: input.Site,
				}
				if cfg.DatadogStatus[id] == nil {
					cfg.DatadogStatus[id] = &integrations.DatadogStatus{ReadyForUse: false}
				}
				return id, nil
			}
			return datadog
		},
		syncertest.MockReadinessService{
			ReadyFn: func(context.Context) (bool, error) { return cfg.Ready, nil },
		},
	)

	return &Harness{
		Workflow:        workflow,
		PreferenceStore: store,
		Datadog:         datadog,
	}
}
