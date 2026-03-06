package tui

import (
	"context"

	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/auth/keyring"
	"github.com/usetero/cli/internal/infrastructure/auth/workos"
	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
	"github.com/usetero/cli/internal/infrastructure/logging"
	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/powersync/uploader"
	infraPreferences "github.com/usetero/cli/internal/infrastructure/preferences"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
	"github.com/usetero/cli/internal/interfaces/cli/config"
	"github.com/usetero/cli/internal/runtime/onboarding"
	"github.com/usetero/cli/internal/runtime/session"
)

type onboardingDependencies struct {
	onboarding *onboarding.Runtime
	session    *session.Runtime
}

func newOnboardingDependencies(cfg config.RuntimeConfig, scope logging.Scope) (*onboardingDependencies, error) {
	preferencesStore, err := infraPreferences.NewStore(string(cfg.Env))
	if err != nil {
		return nil, err
	}
	preferencesService := preferences.NewService(preferencesStore)

	keyringStore, err := keyring.NewStore(string(cfg.Env))
	if err != nil {
		return nil, err
	}
	workosClient := workos.NewClient(
		cfg.WorkOS.ClientID,
		[]string{cfg.API.URL, cfg.PowerSync.URL, cfg.Chat.URL},
	)

	identityService := identity.NewService(
		identity.NewWorkOSProvider(workosClient),
		identity.NewKeyringTokenStore(keyringStore),
		identity.NopLogger{},
	)

	apiClient := controlplane.NewClient(cfg.API.URL, identityService)

	organizationService := tenancy.NewRemoteOrganizationService(apiClient)
	workspaceService := tenancy.NewRemoteWorkspaceService(apiClient)
	datadogService := integrations.NewRemoteDatadogService(apiClient)
	accountsFactory := func(organizationID tenancy.OrganizationID) tenancy.AccountService {
		return tenancy.NewRemoteAccountService(apiClient, organizationID)
	}

	sessionStorage := newSessionStorage(string(cfg.Env))
	sessionRuntime := session.NewRuntime(
		sessionStorage,
		func() (session.Syncer, error) {
			return syncer.New(
				cfg.PowerSync.URL,
				syncerTokenSource{identity: identityService},
				scope.Child("root/session/syncer"),
			)
		},
		func(db *sqlite.DB, notifier interface {
			NotifyUploadCompleted(ctx context.Context) error
		}) (session.Uploader, error) {
			return uploader.New(
				psdb.NewStore(db),
				psclient.NewClient(cfg.PowerSync.URL),
				uploaderTokenSource{identity: identityService},
				scope.Child("root/session/uploader"),
				uploader.WithSyncNotifier(notifier),
			), nil
		},
		scope.Child("root/session"),
	)

	onboardingRuntime := onboarding.NewRuntime(
		preferencesService,
		organizationService,
		accountsFactory,
		workspaceService,
		datadogService,
		sessionRuntime,
	)

	return &onboardingDependencies{
		onboarding: onboardingRuntime,
		session:    sessionRuntime,
	}, nil
}
