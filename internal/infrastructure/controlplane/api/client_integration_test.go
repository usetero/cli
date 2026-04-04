//go:build integration_live

package api

import (
	"context"
	"os"
	"testing"

	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/infrastructure/auth/keyring"
	"github.com/usetero/cli/internal/infrastructure/auth/workos"
	"github.com/usetero/cli/internal/interfaces/cli/config"
)

func TestIntegrationLive_AccountClient_GetDatadogAccount_DemoAccountPresent(t *testing.T) {
	cfg, err := config.Resolve(config.RuntimeConfig{Env: config.ProfileDev})
	if err != nil {
		t.Fatalf("resolve dev config: %v", err)
	}

	store, err := openSystemStore("dev")
	if err != nil {
		t.Fatalf("open dev auth store: %v", err)
	}

	identityService := identity.NewService(
		workos.NewProvider(workos.NewClient(
			cfg.WorkOS.ClientID,
			[]string{cfg.API.Origin, cfg.PowerSync.Origin, cfg.Chat.Origin},
		)),
		keyring.NewTokenStore(store),
		identity.NopLogger{},
	)
	client := NewBootstrapClient(cfg.API.Origin, identityService)

	orgs, err := client.ListOrganizations(context.Background())
	if err != nil {
		t.Fatalf("list organizations: %v", err)
	}

	var demoOrg *Organization
	for i := range orgs {
		if orgs[i].Name == "Tero Demo" {
			demoOrg = &orgs[i]
			break
		}
	}
	if demoOrg == nil {
		t.Fatal("expected demo organization \"Tero Demo\" to exist")
	}

	accounts, err := client.ListAccounts(context.Background(), demoOrg.ID)
	if err != nil {
		t.Fatalf("list demo accounts: %v", err)
	}

	var demoAccount *Account
	for i := range accounts {
		if accounts[i].Name == "Tero Demo" {
			demoAccount = &accounts[i]
			break
		}
	}
	if demoAccount == nil {
		t.Fatal("expected demo account \"Tero Demo\" to exist")
	}

	accountClient, err := client.ForAccount(demoAccount.ID)
	if err != nil {
		t.Fatalf("bind demo account: %v", err)
	}

	datadogAccount, err := accountClient.GetDatadogAccount(context.Background())
	if err != nil {
		t.Fatalf("get demo datadog account: %v", err)
	}
	if datadogAccount == nil {
		t.Fatal("expected demo account to have a datadog account")
	}
}

func openSystemStore(env string) (*keyring.Store, error) {
	prevBackend, hadBackend := os.LookupEnv(keyring.EnvBackend)
	prevPath, hadPath := os.LookupEnv(keyring.EnvPath)
	_ = os.Unsetenv(keyring.EnvBackend)
	_ = os.Unsetenv(keyring.EnvPath)
	defer func() {
		if hadBackend {
			_ = os.Setenv(keyring.EnvBackend, prevBackend)
		} else {
			_ = os.Unsetenv(keyring.EnvBackend)
		}
		if hadPath {
			_ = os.Setenv(keyring.EnvPath, prevPath)
		} else {
			_ = os.Unsetenv(keyring.EnvPath)
		}
	}()
	return keyring.NewStore(env)
}
