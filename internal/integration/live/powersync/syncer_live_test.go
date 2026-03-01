//go:build integration_live

package powersynclive_test

import (
	"context"
	"testing"
	"time"

	"github.com/usetero/cli/internal/auth"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/workos"
)

// Live integration tests run against non-production services.
//
// Prerequisites:
//  1. task auth:login
//  2. task run (complete onboarding to set default account)
//  3. task test:integration:live
func TestIntegrationLive_Syncer(t *testing.T) {
	cliConfig := config.LoadCLIConfig()
	logger := logtest.NewScope(t)
	env := cliConfig.Environment()

	t.Logf("Environment: %s", env)
	t.Logf("API Endpoint: %s", cliConfig.APIEndpoint)
	t.Logf("PowerSync Endpoint: %s", cliConfig.PowerSyncEndpoint)

	storage := keyring.New(env)
	oauthProvider := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.ChatEndpoint, cliConfig.PowerSyncEndpoint)
	authSvc := auth.NewService(oauthProvider, storage, logger)

	if _, err := authSvc.GetAccessToken(context.Background()); err != nil {
		t.Fatalf("failed to get access token: %v (run: task auth:login)", err)
	}

	orgCfg, err := config.LoadOrgPreferences(env, config.ActiveOrgID(env))
	if err != nil {
		t.Fatalf("org preferences not found: %v (run: task run)", err)
	}
	orgPrefs := preferences.NewOrgService(orgCfg, logger)
	accountID := orgPrefs.GetDefaultAccountID()
	if accountID == "" {
		t.Fatalf("no default account (run: task run)")
	}

	services := graphql.NewServiceSet(cliConfig.APIEndpoint+"/graphql", authSvc, logger).WithAccountID(accountID)

	account, err := services.Accounts.Get(context.Background(), accountID)
	if err != nil {
		t.Fatalf("failed to fetch account: %v", err)
	}
	if account == nil {
		t.Fatalf("account %s not found", accountID)
	}

	t.Logf("Account: %s (%s)", account.Name, account.ID)

	t.Run("connects and syncs data", func(t *testing.T) {
		database := dbtest.OpenTestDB(t)
		syncer := powersync.NewSyncer(cliConfig.PowerSyncEndpoint, authSvc, logtest.NewScope(t))

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		firstSyncDone := make(chan struct{})
		err := syncer.Start(ctx, database, accountID.String(), func() {
			close(firstSyncDone)
		})
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer syncer.Stop()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		lastState := ""
		for {
			select {
			case <-firstSyncDone:
				t.Log("first sync completed")
				goto done
			case <-ticker.C:
				state := syncer.State()
				summary := summarizeState(state)
				if summary != lastState {
					t.Logf("State: %s", summary)
					if errState, ok := state.(*powersync.Error); ok {
						t.Fatalf("sync error: %v", errState.Err)
					}
					lastState = summary
				}
			case <-ctx.Done():
				t.Fatalf("timeout waiting for first sync, last state: %s", summarizeState(syncer.State()))
			}
		}
	done:

		buckets, err := countBuckets(database)
		if err != nil {
			t.Fatalf("countBuckets() error = %v", err)
		}
		t.Logf("Synced %d buckets", buckets)

		if buckets == 0 {
			t.Fatal("no buckets synced")
		}
		if !syncer.IsReady() {
			t.Error("IsReady() should be true after first sync")
		}
	})
}

func countBuckets(db sqlite.DB) (int64, error) {
	var count int64
	err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM ps_buckets").Scan(&count)
	return count, err
}

func summarizeState(state powersync.State) string {
	switch s := state.(type) {
	case *powersync.Disconnected:
		return "disconnected"
	case *powersync.Connecting:
		return "connecting"
	case *powersync.Syncing:
		if s.Progress != nil {
			return "syncing " + s.Progress.String()
		}
		return "syncing"
	case *powersync.Ready:
		return "ready"
	case *powersync.Reconnecting:
		if s.Degraded {
			return "reconnecting (degraded)"
		}
		return "reconnecting"
	case *powersync.Error:
		return "error: " + s.Err.Error()
	default:
		return "unknown"
	}
}
