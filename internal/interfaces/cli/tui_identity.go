package cli

import (
	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/infrastructure/auth/keyring"
	"github.com/usetero/cli/internal/infrastructure/auth/workos"
)

func newIdentityService(exec *runner) (*identity.Service, error) {
	keyringStore, err := keyring.NewStore(string(exec.cfg.Env))
	if err != nil {
		return nil, err
	}

	workosClient := workos.NewClient(
		exec.cfg.WorkOS.ClientID,
		[]string{exec.cfg.API.Origin, exec.cfg.PowerSync.Origin, exec.cfg.Chat.Origin},
	)

	return identity.NewService(
		workos.NewProvider(workosClient),
		keyring.NewTokenStore(keyringStore),
		identity.NopLogger{},
	), nil
}
