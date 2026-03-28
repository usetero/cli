package cli

import (
	"context"

	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/infrastructure/logging"
	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

type accountRuntimeFactory struct {
	identity        *identity.Service
	env             string
	powerSyncOrigin string
	scope           logging.Scope
}

func newAccountRuntimeFactory(exec *runner, identityService *identity.Service) *accountRuntimeFactory {
	if identityService == nil {
		panic("account runtime factory requires identity service")
	}
	return &accountRuntimeFactory{
		identity:        identityService,
		env:             string(exec.cfg.Env),
		powerSyncOrigin: exec.cfg.PowerSync.Origin,
		scope:           exec.scope.Child("tui"),
	}
}

func (f *accountRuntimeFactory) New(ctx context.Context, accountScope accountruntime.Scope) (*accountruntime.Runtime, error) {
	return accountruntime.New(ctx, accountScope, accountruntime.Config{
		Env:             f.env,
		PowerSyncOrigin: f.powerSyncOrigin,
		SyncerTokens:    syncerTokenSource{identity: f.identity},
		UploaderTokens:  uploaderTokenSource{identity: f.identity},
	}, f.scope.Child("app/account"))
}

type syncerTokenSource struct {
	identity *identity.Service
}

func (s syncerTokenSource) GetAccessToken(ctx context.Context) (pssyncer.AccessToken, error) {
	token, err := s.identity.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return pssyncer.AccessToken(token), nil
}

func (s syncerTokenSource) ForceRefreshAccessToken(ctx context.Context) (pssyncer.AccessToken, error) {
	return s.GetAccessToken(ctx)
}

type uploaderTokenSource struct {
	identity *identity.Service
}

func (s uploaderTokenSource) GetAccessToken(ctx context.Context) (psclient.AccessToken, error) {
	token, err := s.identity.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return psclient.AccessToken(token), nil
}
